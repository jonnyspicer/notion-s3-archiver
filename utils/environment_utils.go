package utils

import (
	"os"
)

func GetNotionApiKey() string {
	return os.Getenv(NotionApiKey)
}

func GetTopLevelPageId() string {
	return os.Getenv(TopLevelPageId)
}

func GetOutputDir() string {
	return os.Getenv(OutputDir)
}

func GetBucketName() string {
	return os.Getenv(BucketName)
}

func GetSnsTopicArn() string {
	return os.Getenv(SnsTopicArn)
}

func GetAwsRegion() string {
	return os.Getenv(AwsRegion)
}
