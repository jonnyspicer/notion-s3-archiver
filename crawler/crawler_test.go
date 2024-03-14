package crawler_test

import (
	"github.com/jonnyspicer/notion-utils/crawler"
	"testing"
)

func TestCrawl(t *testing.T) {
	// This test is only really used for debugging purposes
	crawler := crawler.NewCrawler()
	crawler.FullCrawl()
}
