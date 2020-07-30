package s3

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func UploadDir(bucket string, keyPrefix string, dir string) error {
	return filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if isDirectory(path) {
			return nil
		}

		key := fmt.Sprintf("%s/%s", keyPrefix, path)
		file, err := os.Open(path) // TODO handle error
		if err != nil {
			return err
		}

		defer file.Close()
		return UploadFile(bucket, key, file)
	})
}

func UploadFile(bucket string, key string, file io.ReadSeeker) error {
	client := s3.New(session.New())

	input := &s3.PutObjectInput{
		Body:   file,
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := client.PutObject(input)
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
