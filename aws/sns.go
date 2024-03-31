package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/jonnyspicer/notion-s3-archiver/utils"
	"os"
)

func PublishToTopic(message string, sess *session.Session, critical bool) error {
	snsClient := sns.New(sess)
	topic := utils.GetSnsTopicArn()
	_, err := snsClient.Publish(&sns.PublishInput{
		Message:  aws.String(message),
		TopicArn: aws.String(topic),
	})

	if err != nil && critical {
		fmt.Printf("Error publishing logs to SNS: %v", err)
		os.Exit(1)
	}
	return err
}
