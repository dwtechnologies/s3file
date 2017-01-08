package s3file

import (
	"bytes"
	"io"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// FileExists takes a *FileRequest c and checks if the requested file exists or not.
// It returns true if the file exists.
func FileExists(c *FileRequest) bool {
	checkAwsRegion()

	s3bucket := c.S3Bucket
	prefix := c.S3Prefix
	file := c.S3Filename

	pf := filepath.Join(prefix, file)
	svc := s3.New(session.New(), &aws.Config{Region: aws.String(awsRegion)})

	params := &s3.GetObjectInput{
		Bucket: aws.String(s3bucket),
		Key:    aws.String(pf),
	}

	_, err := svc.GetObject(params)

	if err != nil {
		return false
	}

	return true
}

// GetFile takes a *FileRequest c and tries to download the requested file from s3.
// It returns the content of the requested file as a string and any download / read errors.
func GetFile(c *FileRequest) (string, error) {
	checkAwsRegion()

	s3bucket := c.S3Bucket
	prefix := c.S3Prefix
	file := c.S3Filename

	pf := filepath.Join(prefix, file)
	svc := s3.New(session.New(), &aws.Config{Region: aws.String(awsRegion)})

	params := &s3.GetObjectInput{
		Bucket: aws.String(s3bucket),
		Key:    aws.String(pf),
	}

	resp, err := svc.GetObject(params)
	// If we get an error, return an empty string
	if err != nil {
		return "", err
	}

	// Create a new buffer to store the Body of the response in
	buf := bytes.NewBuffer(nil)
	io.Copy(buf, resp.Body)

	resp.Body.Close()
	return string(buf.Bytes()), nil
}

// RemoveFile takes a *FileRequest c and tries to remove the requested file from s3.
// It returns any errors that might occure.
func RemoveFile(c *FileRequest) error {
	checkAwsRegion()

	s3bucket := c.S3Bucket
	prefix := c.S3Prefix
	file := c.S3Filename

	pf := filepath.Join(prefix, file)
	svc := s3.New(session.New(), &aws.Config{Region: aws.String(awsRegion)})

	params := &s3.DeleteObjectInput{
		Bucket: aws.String(s3bucket),
		Key:    aws.String(pf),
	}

	_, err := svc.DeleteObject(params)

	if err != nil {
		return err
	}
	return nil
}

// Returns the correct chunksize
func getChunkSizes(fileSize int64, chunkMaxSize int64) int64 {
	if fileSize < chunkMaxSize {
		return fileSize
	}
	return chunkMaxSize
}

// Returns the number of parts that are needed for an multipart upload
func getPartsNum(fileSize int64, chunkMaxSize int64) int64 {
	div := float64(float64(fileSize) / float64(chunkMaxSize))
	divint64 := int64(div)
	ret := int64(1)

	if div > float64(divint64) {
		return ret + divint64
	}

	if div == float64(divint64) && div >= float64(1) {
		return divint64
	}
	return ret + divint64
}

func checkAwsRegion() {
	// If the AWS Region env var is not set, default to eu-west-1
	if awsRegion == "" {
		awsRegion = "eu-west-1"
	}
}
