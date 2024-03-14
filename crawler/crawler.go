package crawler

import (
	"context"
	"fmt"
	"github.com/jonnyspicer/go-notion"
	"log"
	"os"
)

func Crawl() {
	// A hashmap of page IDs to "filepaths"
	visited := make(map[string]string)

	// A queue of page IDs to visit
	to_visit := make([]string, 0)

	// TODO: don't commit this!
	os.Setenv(NotionApiKey, "good job Jonny, you remembered not to commit your API key!")
	os.Setenv(TopLevelPageId, "it probably would've been fine to commit this ID but better safe than sorry")

	key := getNotionApiKey()
	id := getTopLevelPageId()
	ctx := context.Background()

	client := notion.NewClient(key)

	cursor := ""
	more := true
	for more {
		pq := &notion.PaginationQuery{
			StartCursor: cursor, // maybe this field needs to not be here at all?
			PageSize:    MaxPageSize,
		}

		resp, err := client.FindBlockChildrenByID(ctx, id, pq)
		if err != nil {
			log.Fatalf("Failed to find block children: %v", err)
		}

		for _, result := range resp.Results {
			if result.BlockType() == "child_page" {
				if _, ok := visited[result.ID()]; !ok {
					to_visit = append(to_visit, result.ID())
				}
			}
		}

		more = resp.HasMore
		if resp.NextCursor != nil {
			cursor = *resp.NextCursor
		}
	}

	visited[id] = "/"

	fmt.Printf("%+v", to_visit)
	fmt.Printf("\n%+v", visited)
}
