package s3file

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// PutFile takes an *PutFileRequest c and tries to send the requested file to s3.
// It returns any file upload errors.
func PutFile(c *PutFileRequest) error {
	const minSize = 5242880    // 5mb
	const defSize = 1073741824 // 1gb
	checkAwsRegion()

	var wg sync.WaitGroup

	s3bucket := c.S3Bucket
	prefix := c.S3Prefix
	filename := c.S3Filename
	path := c.LocalFile

	// Set partsize
	size := defSize
	if int(c.PartSize) > minSize {
		size = int(c.PartSize)
	}

	fs, err := os.Stat(path)
	if err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}

	// Set content type, but default to text/plain; charset=utf-8
	contentType := "text/plain; charset=utf-8"
	if c.ContentType != "" {
		contentType = c.ContentType
	}

	fileSize := fs.Size()
	chunkMaxSize := int64(size)
	numParts := getPartsNum(fileSize, chunkMaxSize)
	chunkSize := getChunkSizes(fileSize, chunkMaxSize)
	chunk := make([]byte, chunkSize)
	partCounter := int64(1)
	completedParts := make([]*s3.CompletedPart, 0, 0)
	finishMultiPart := true

	// Create the parts Channel
	partsChannel := make(chan *completedPart, numParts)

	svc := s3.New(session.New(), &aws.Config{Region: aws.String(awsRegion)})
	pf := filepath.Join(prefix, filename)

	params := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(s3bucket),
		Key:    aws.String(pf),
		Metadata: map[string]*string{
			"Key": aws.String(pf),
		},
		ContentType: aws.String(contentType),
	}

	resp, err := svc.CreateMultipartUpload(params)
	if err != nil {
		return err
	}
	multiPartId := *resp.UploadId

	// Read the first chunk of the file
	_, err = f.Read(chunk)
	for err != io.EOF {
		if err != nil {
			finishMultiPart = false
			break
		}

		params := &s3.UploadPartInput{
			Bucket:     aws.String(s3bucket),
			Key:        aws.String(pf),
			PartNumber: aws.Int64(partCounter),
			UploadId:   aws.String(multiPartId),
			Body:       bytes.NewReader(chunk),
		}

		// Run each of the chunks of the multipart upload as a separate go-routine
		wg.Add(1)
		go func(params *s3.UploadPartInput, partCounter int64, partsChannel chan<- *completedPart) {
			completedPart := new(completedPart)

			etag := ""
			resp, err := svc.UploadPart(params)
			if err != nil {
				completedPart.err = err
				finishMultiPart = false
				etag = "failed"
			} else {
				etag = *resp.ETag
			}

			part := &s3.CompletedPart{
				ETag:       aws.String(etag),
				PartNumber: aws.Int64(partCounter),
			}
			completedPart.part = part
			partsChannel <- completedPart
			wg.Done()
		}(params, partCounter, partsChannel)

		fileSize = fileSize - chunkSize
		chunkSize = getChunkSizes(fileSize, chunkMaxSize)
		// If chunkSize is 0, indicate that we have reached EOF
		if chunkSize == 0 {
			err = io.EOF
		}

		partCounter++
		// Read the next chunk of the file
		_, err = f.Read(chunk)
	}
	wg.Wait()
	close(partsChannel)

	// Add all the completed parts a slice of completedParts
	for part := range partsChannel {
		if part.err != nil {
			return part.err
		}

		completedParts = append(completedParts, part.part)
	}

	// If we got an indication that the multi part upload was finished, finalize it
	// Otherwise tell s3 to discard the multipart upload completely
	if finishMultiPart {
		params := &s3.CompleteMultipartUploadInput{
			Bucket:   aws.String(s3bucket),
			Key:      aws.String(pf),
			UploadId: aws.String(multiPartId),
			MultipartUpload: &s3.CompletedMultipartUpload{
				Parts: completedParts,
			},
		}

		_, err := svc.CompleteMultipartUpload(params)
		if err == nil {
			return nil
		}
	}

	// Tell that we couldn't upload the file
	paramsAbort := &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(s3bucket),
		Key:      aws.String(pf),
		UploadId: aws.String(multiPartId),
	}
	_, err = svc.AbortMultipartUpload(paramsAbort)
	if err != nil {
		err = fmt.Errorf("Couldn't abort the failed MultipartUpload. Do it manually. ID: %v", multiPartId)
		return err
	}

	err = fmt.Errorf("MutlipartUploadRequest removed")
	return err
}
