package main

import (
	"github.com/jonnyspicer/notion-utils/crawler"
)

func main() {
	crawler := crawler.NewCrawler()
	crawler.FullCrawl()
}
