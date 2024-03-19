package crawler

import (
	"context"
	"errors"
	"fmt"
	"github.com/jonnyspicer/go-notion"
	"log"
	"strings"
	"time"
)

type Page struct {
	id             string
	parentTreePath string
}

type Crawler struct {
	visited map[string]string
	toVisit []Page
	client  *notion.Client
}

func NewCrawler() *Crawler {
	// A hashmap of page IDs to "filepaths"
	visited := make(map[string]string)

	// A queue of pages to visit
	toVisit := make([]Page, 0)

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

	for more {
		pq := &notion.PaginationQuery{
			StartCursor: cursor,
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
					childPage, err := c.pageFromBlock(parentPage.parentTreePath, result)
					if err != nil {
						return err
					}
					c.toVisit = append(c.toVisit, *childPage)
				}
			} else if result.BlockType() == "child_database" {
				pages, err := c.getDatabaseChildPages(ctx, result.ID())
				if err != nil {
					fmt.Println(err)
				}
				for _, page := range pages {
					if _, ok := c.visited[page.id]; !ok {
						title, _ := c.getDatabaseChildPageTitle(ctx, page.id)
						page.parentTreePath = c.normalisePath(parentPage.parentTreePath + title + "/")
						c.toVisit = append(c.toVisit, *page)
					}
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

	id := getTopLevelPageId()
	rootPath := "/"
	c.toVisit = append(c.toVisit, Page{
		id:             id,
		parentTreePath: rootPath,
	})

	for len(c.toVisit) > 0 {
		page := c.toVisit[0]
		c.toVisit = c.toVisit[1:]

		// It turns out that doing this with a worker pool runs into the rate limit
		// So for now we'll just do it synchronously
		err := c.CrawlPage(page)
		if err != nil {
			fmt.Println(err)
		}
	}

	log.Printf("Found %v pages", len(c.visited))
	log.Printf("Crawl time was: %s", time.Since(start))
}

func (c *Crawler) ListAllPageIds() []string {
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
	// TODO: the docs have a list of characters to avoid; decide whether to avoid them or not
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-keys.html
	//path = strings.Replace(path, " ", "-", -1)
	return strings.TrimSpace(path)
}

func (c *Crawler) getDatabaseChildPages(ctx context.Context, databaseId string) ([]*Page, error) {
	more := true
	cursor := ""
	pages := make([]*Page, 0)

	for more {
		dq := &notion.DatabaseQuery{
			Filter:      nil,
			Sorts:       nil,
			StartCursor: cursor,
			PageSize:    MaxPageSize,
		}
		resp, err := c.client.QueryDatabase(ctx, databaseId, dq)
		if err != nil {
			log.Fatalf("Failed to query database: %s", err)
		}

		for _, result := range resp.Results {
			pages = append(pages, &Page{
				id:             result.ID,
				parentTreePath: "",
			})
		}

		more = resp.HasMore
		if resp.NextCursor != nil {
			cursor = *resp.NextCursor
		}
	}

	return pages, nil
}

func (c *Crawler) getDatabaseChildPageTitle(ctx context.Context, pageId string) (string, error) {
	resp, err := c.client.FindPageByID(ctx, pageId)
	if err != nil {
		log.Fatalf("Failed to get page title: %s", err)
	}

	title := resp.ID

	if dbn, ok := resp.Properties.(notion.DatabasePageProperties)["Name"]; ok {
		title = dbn.Title[0].PlainText
	}

	return title, err
}
