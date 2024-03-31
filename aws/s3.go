package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/jonnyspicer/notion-s3-archiver/utils"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func UploadDirectoryToS3Glacier(bucketName, directoryPath string, sess *session.Session) error {
	start := time.Now()
	log.Println("Beginning S3 upload")
	uploader := s3manager.NewUploader(sess)

	err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				log.Printf("failed to open file %q: %v", path, err)
				return err
			}
			defer file.Close()

			_, err = uploader.Upload(&s3manager.UploadInput{
				Bucket:       aws.String(bucketName),
				Key:          aws.String(prepareFileKey(path)), // Adjust this if you need a different key structure
				Body:         file,
				StorageClass: aws.String("DEEP_ARCHIVE"), // Use DEEP_ARCHIVE for deeper savings
			})
			if err != nil {
				log.Printf("failed to upload file %q: %v", path, err)
				return err
			}
			log.Printf("Successfully uploaded %q to %q with Glacier storage class\n", path, bucketName)
		}
		return nil
	})

	log.Printf("Upload time was: %s", time.Since(start))

	return err
}

func prepareFileKey(path string) string {
	path = strings.ReplaceAll(path, utils.GetOutputDir(), "")
	return strings.ReplaceAll(path, "|", "/")
}
