package crawler

import (
	"context"
	"errors"
	"fmt"
	"github.com/jonnyspicer/go-notion"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type Page struct {
	id             string
	parentTreePath string
}

type Crawler struct {
	visited      map[string]string
	toVisit      []Page
	client       *notion.Client
	visitedMutex sync.Mutex
	newPagesChan chan Page
}

func NewCrawler() *Crawler {
	// A hashmap of page IDs to "filepaths"
	visited := make(map[string]string)

	// A queue of pages to visit
	toVisit := make([]Page, 0)

	// TODO: don't commit this!
	os.Setenv(NotionApiKey, "secret_ay7xFS49gT4LrrejvjNARMiA06OLs423rabnYRDXgOi")
	os.Setenv(TopLevelPageId, "6c67147fdb094ee6b47a1da836a7bf66")

	key := getNotionApiKey()

	client := notion.NewClient(key)

	return &Crawler{
		visited: visited,
		toVisit: toVisit,
		client:  client,
	}
}

func (c *Crawler) CrawlPage(parentPage Page) error {
	ctx := context.Background()
	cursor := ""
	more := true

	fmt.Printf("Crawling page: %s", parentPage.parentTreePath)

	for more {
		pq := &notion.PaginationQuery{
			StartCursor: cursor, // maybe this field needs to not be here at all?
			PageSize:    MaxPageSize,
		}

		resp, err := c.client.FindBlockChildrenByID(ctx, parentPage.id, pq)
		if err != nil {
			log.Fatalf("Failed to find block children: %v", err)
		}

		for _, result := range resp.Results {
			// ignore other block children
			if result.BlockType() == "child_page" {
				// we're only interested in blocks we haven't already visited
				if _, ok := c.visited[result.ID()]; !ok {
					// this is pretty sloppy and I don't like it
					childPage, err := c.pageFromBlock(parentPage.parentTreePath, result)
					if err != nil {
						return err
					}
					fmt.Printf("Sending child page to channel: %s", childPage.parentTreePath)
					c.newPagesChan <- *childPage
				}
			}
		}

		more = resp.HasMore
		if resp.NextCursor != nil {
			cursor = *resp.NextCursor
		}
	}

	c.visited[parentPage.id] = parentPage.parentTreePath

	return nil
}

func (c *Crawler) FullCrawl() {
	start := time.Now()

	toVisitChan := make(chan Page, 10)
	c.newPagesChan = make(chan Page, 100)

	go func() {
		for newPage := range c.newPagesChan {
			c.toVisit = append(c.toVisit, newPage)
			fmt.Printf("to visit: %+v", c.toVisit)
		}
	}()

	var wg sync.WaitGroup

	numWorkers := 5

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			for page := range toVisitChan {
				err := c.CrawlPage(page)
				if err != nil {
					fmt.Printf("Worker %d: %v\n", workerId, err)
				}
			}
		}(i)
	}

	id := getTopLevelPageId()
	rootPath := "/"
	c.toVisit = append(c.toVisit, Page{
		id:             id,
		parentTreePath: rootPath,
	})

	for len(c.toVisit) > 0 {
		page := c.toVisit[0]
		c.toVisit = c.toVisit[1:]

		toVisitChan <- page
	}

	wg.Wait()
	close(toVisitChan)
	close(c.newPagesChan)

	//fmt.Printf("%+v", c.listAllPageIds())
	fmt.Println(len(c.visited))
	log.Printf("Crawl time was: %s", time.Since(start))
}

func (c *Crawler) listAllPageIds() []string {
	ids := make([]string, len(c.visited))

	i := 0
	for j := range c.visited {
		ids[i] = j
		i++
	}

	return ids
}

func (c *Crawler) pageFromBlock(parentTreePath string, block notion.Block) (*Page, error) {
	id := block.ID()
	if title, ok := block.Extras().(map[string]string)["title"]; ok {
		return &Page{
			id:             id,
			parentTreePath: c.normalisePath(parentTreePath + title + "/"),
		}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("error retrieving title from extras field on block with id %s", block.ID()))
	}
}

func (c *Crawler) normalisePath(path string) string {
	path = strings.ToLower(path)
	path = strings.Replace(path, " ", "-", -1)
	return strings.TrimSpace(path)
}
