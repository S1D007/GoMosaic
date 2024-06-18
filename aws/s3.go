package aws

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var s3Client *s3.S3
var awsSession *session.Session

func S3() {
	creds := credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "")

	cfg := aws.NewConfig().
		WithCredentials(creds).
		WithRegion("ap-south-1")

	awsSession = session.Must(session.NewSession(cfg))

	s3Client = s3.New(awsSession)
}

func UploadImageFromBuffer(buffer *bytes.Buffer, folder, fileName string) (string, error) {
	bucket := "gkh-images"
	key := fmt.Sprintf("mosaic/%s/%s.png", folder, fileName)
	reader := bytes.NewReader(buffer.Bytes())

	_, err := s3Client.PutObjectWithContext(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String("image/png"),
		ACL:         aws.String("public-read"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, key), nil
}
