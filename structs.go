package s3file

import "os"

var awsRegion = os.Getenv("AWS_REGION")

// PutFileRequest is used by the PutFile function. You specify the necessary paramters for the file upload.
type PutFileRequest struct {
	S3Bucket   string // Bucket to upload file to
	S3Prefix   string // Prefix to use for upload
	S3Filename string // Filename to use on S3
	LocalFile  string // Local file on filesystem (whole path)
	PartSize   int64  // Size of parts (multi-part upload) in bytes
}

// FileRequest is used by the FileExists, GetFile and RemoveFile functions. You specify the necessary paramters for the file upload.
type FileRequest struct {
	S3Bucket   string // Bucket to upload file to
	S3Prefix   string // Prefix to use for upload
	S3Filename string // Filename to use on S3
}
