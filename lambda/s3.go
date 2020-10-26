package lambda

import (
	"bytes"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3API - bits from the S3 interface we need
type S3API interface {
	PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error)
}

// NewS3 returns a new S3 handler
func NewS3(c S3API) *S3Handler {
	return &S3Handler{s3: c}
}

// S3Handler points back to the interface
type S3Handler struct {
	s3 S3API
}

// performPut wrappers the S3 PutObject request
func (s *S3Handler) performPut(region string, bucket string, key string, buffer []byte, size int64) error {

	// Build the put params
	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		ACL:           aws.String("private"),
		Body:          bytes.NewReader(buffer),
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(http.DetectContentType(buffer)),
	}

	// Perform S3 put
	_, err := s.s3.PutObject(params)

	// Return
	return err
}
