package main

import (
	"bytes"
	"fmt"
	"github.com/jonnyspicer/notion-s3-archiver/archiver"
	"github.com/jonnyspicer/notion-s3-archiver/aws"
	"github.com/jonnyspicer/notion-s3-archiver/crawler"
	"github.com/jonnyspicer/notion-s3-archiver/utils"
	"log"
	"time"
)

func main() {
	start := time.Now()
	sess, err := aws.NewSession()
	var buf bytes.Buffer
	log.SetOutput(&buf)

	mess := fmt.Sprintf("Began running script at %v", start.String())
	if err = aws.PublishToTopic(mess, sess, false); err != nil {
		log.Printf("Failed to publish message to SNS: %v\n", err)
	}

	c := crawler.NewCrawler()
	c.FullCrawl()
	pages := c.ListAlLPages()

	a := archiver.NewArchiver()
	err = a.DownloadAllPages(pages)
	if err != nil {
		log.Println(err)
		err = aws.PublishToTopic(buf.String(), sess, true)
	}

	if err = aws.UploadDirectoryToS3Glacier(utils.GetBucketName(), utils.GetOutputDir(), sess); err != nil {
		log.Printf("Failed to upload directory to S3 Glacier: %v\n", err)
		err = aws.PublishToTopic(buf.String(), sess, true)
	}

	log.Printf("Total time was: %s", time.Since(start))
	err = aws.PublishToTopic(buf.String(), sess, true)
}
