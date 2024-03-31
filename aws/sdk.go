package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jonnyspicer/notion-s3-archiver/utils"
)

func NewSession() (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region: aws.String(utils.GetAwsRegion()),
	})
}
