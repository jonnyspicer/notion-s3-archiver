package crawler

import (
	"context"
	"errors"
	"fmt"
	"github.com/jonnyspicer/go-notion"
	"github.com/jonnyspicer/notion-s3-archiver/utils"
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
	ctx     context.Context
}

func NewCrawler() *Crawler {
	// A hashmap of page IDs to "filepaths"
	visited := make(map[string]string)

	// A queue of pages to visit
	toVisit := make([]Page, 0)

	key := utils.GetNotionApiKey()

	client := notion.NewClient(key)

	ctx := context.Background()

	return &Crawler{
		visited: visited,
		toVisit: toVisit,
		client:  client,
		ctx:     ctx,
	}
}

func (c *Crawler) CrawlPage(parentPage Page) error {
	cursor := ""
	more := true

	for more {
		pq := &notion.PaginationQuery{
			StartCursor: cursor,
			PageSize:    utils.MaxPageSize,
		}

		resp, err := c.client.FindBlockChildrenByID(c.ctx, parentPage.id, pq)
		if err != nil {
			log.Printf("Failed to find block children: %v\n", err)
			return err
		}

		for _, result := range resp.Results {
			// ignore other block children
			if result.BlockType() == "child_page" {
				// we're only interested in blocks we haven't already visited
				if _, ok := c.visited[result.ID()]; !ok {
					childPage, err := c.pageFromBlock(parentPage.parentTreePath, result)
					if err != nil {
						log.Printf("unable to get page from block, err: %v\n", err)
					}
					c.toVisit = append(c.toVisit, *childPage)
				}
			} else if result.BlockType() == "child_database" {
				pages, err := c.getDatabaseChildPages(result.ID())
				if err != nil {
					log.Printf("Failed to query database: %s\n", err)
				}
				for _, page := range pages {
					if _, ok := c.visited[page.id]; !ok {
						title, err := c.getDatabaseChildPageTitle(page.id)
						if err != nil {
							log.Println(err)
						} else {
							page.parentTreePath = c.normalisePath(parentPage.parentTreePath + title + "|")
						}
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

	id := utils.GetTopLevelPageId()
	rootPath := "Everything Everywhere All At Once|"
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
			log.Println(err)
		}
	}

	log.Printf("Found %v pages\n", len(c.visited))
	log.Printf("Crawl time was: %s\n", time.Since(start))
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

func (c *Crawler) ListAlLPages() map[string]string {
	return c.visited
}

func (c *Crawler) pageFromBlock(parentTreePath string, block notion.Block) (*Page, error) {
	id := block.ID()
	if title, ok := block.Extras().(map[string]string)["title"]; ok {
		return &Page{
			id:             id,
			parentTreePath: c.normalisePath(parentTreePath + title + "|"),
		}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("error retrieving title from extras field on block with id %s", block.ID()))
	}
}

func (c *Crawler) normalisePath(path string) string {
	// TODO: the docs have a list of characters to avoid; decide whether to avoid them or not
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-keys.html
	path = strings.Replace(path, " ", "%20", -1)
	return strings.TrimSpace(path)
}

func (c *Crawler) getDatabaseChildPages(databaseId string) ([]*Page, error) {
	more := true
	cursor := ""
	pages := make([]*Page, 0)

	for more {
		dq := &notion.DatabaseQuery{
			Filter:      nil,
			Sorts:       nil,
			StartCursor: cursor,
			PageSize:    utils.MaxPageSize,
		}
		resp, err := c.client.QueryDatabase(c.ctx, databaseId, dq)
		if err != nil {
			return nil, err
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

func (c *Crawler) getDatabaseChildPageTitle(pageId string) (string, error) {
	resp, err := c.client.FindPageByID(c.ctx, pageId)
	if err != nil {
		log.Printf("Failed to get page title: %s\n", err)
		return "", nil
	}

	title := resp.ID

	dbn, ok := resp.Properties.(notion.DatabasePageProperties)["Name"]

	if ok && dbn.Title != nil && len(dbn.Title) > 0 {
		title = dbn.Title[0].PlainText
	}

	return title, err
}
