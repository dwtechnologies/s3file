# s3file

The s3file package is a heler package to the official aws-go-sdk.
It will make it easier to

- Upload files to S3 (using multi-part upload)
- Download files from S3
- Checking if a file exists on S3
- Deleting a file on S3

## Function

### GetFile

```text
GetFile takes a *FileRequest c and tries to download the requested file from s3.
It returns the content of the requested file as a string and any download / read errors.
```

### RemoveFile

```text
RemoveFile takes a *FileRequest c and tries to remove the requested file from s3.
It returns any errors that might occure.
```

### FileExists

```text
FileExists takes a *FileRequest c and checks if the requested file exists or not.
It returns true if the file exists.
```

### PutFile

```text
PutFile takes an *PutFileRequest c and tries to send the requested file to s3.
It returns any file upload errors.
```

----------

## Types

### PutFileRequest

```go
type PutFileRequest struct {
    S3Bucket    string // Bucket to upload file to
    S3Prefix    string // Prefix to use for upload
    S3Filename  string // Filename to use on S3
    LocalFile   string // Local file on filesystem (whole path)
    PartSize    int64  // Size of parts (multi-part upload) in bytes, will default to 1gb
    ContentType string // Set the content type, will default to text/plain; charset=utf-8
}
```

### FileRequest

```go
type FileRequest struct {
    S3Bucket   string // Bucket to upload file to
    S3Prefix   string // Prefix to use for upload
    S3Filename string // Filename to use on S3
}
```