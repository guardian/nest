package s3

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// UploadDir walks a directory and uploads all files to S3
func UploadDir(bucket string, keyPrefix string, dir string) error {
	return filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if isDirectory(path) {
			return nil
		}

		key := fmt.Sprintf("%s%s", keyPrefix, strings.TrimPrefix(path, dir))
		file, err := os.Open(path)
		if err != nil {
			return err
		}

		defer file.Close()
		return UploadFile(bucket, key, file)
	})
}

// UploadFile uploads files to S3
func UploadFile(bucket string, key string, file io.ReadSeeker) error {
	session, err := session.NewSession(aws.NewConfig().WithRegion("eu-west-1"))
	if err != nil {
		return err
	}

	client := s3.New(session)

	input := &s3.PutObjectInput{
		Body:   file,
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	fmt.Printf("%v", input)

	_, err = client.PutObject(input)
	return err
}

func isDirectory(path string) bool {
	fd, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}
	mode := fd.Mode()
	return mode.IsDir()
}
