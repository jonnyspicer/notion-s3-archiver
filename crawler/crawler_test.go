package crawler_test

import (
	"github.com/jonnyspicer/notion-utils/crawler"
	"testing"
)

func TestCrawl(t *testing.T) {
	crawler := crawler.NewCrawler()
	crawler.FullCrawl()
}
