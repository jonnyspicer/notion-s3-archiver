package archiver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jonnyspicer/go-notion"
	"github.com/jonnyspicer/notion-s3-archiver/utils"
	"log"
	"os"
	"time"
)

type Archiver struct {
	client  *notion.Client
	ctx     context.Context
	rootDir string
}

func NewArchiver() *Archiver {
	key := utils.GetNotionApiKey()

	client := notion.NewClient(key)
	ctx := context.Background()
	rootDir := utils.GetOutputDir()

	return &Archiver{
		client:  client,
		ctx:     ctx,
		rootDir: rootDir,
	}
}

func (a *Archiver) DownloadAllPages(pages map[string]string) error {
	start := time.Now()
	archived := 0

	exists, err := dirExists(a.rootDir)
	if err != nil {
		return err
	}

	if exists {
		err = os.RemoveAll(a.rootDir)
	}

	err = os.Mkdir(a.rootDir, 0755)

	for id, path := range pages {
		log.Printf("Archiving page: %v\n", path)
		err := a.DownloadPage(id, path)
		if err != nil {
			log.Printf("Error downloading page: %v", err)
		} else {
			archived += 1
		}
	}
	log.Printf("Archived %v pages\n", archived)
	log.Printf("Archive time was: %s\n", time.Since(start))
	return nil
}

// for a given page, marshall its content into JSON and write the resulting object to a file
// NB will not get child pages, including those that are database objects
func (a *Archiver) DownloadPage(pageId, filePath string) error {
	page, _ := a.client.FindPageByID(a.ctx, pageId)
	// actually needs to start by retrieving the page information, and then have the blocks under that
	blocks, _ := a.GetBlockChildren(pageId)
	page.Blocks = blocks
	blockson, _ := json.Marshal(page)
	err := a.WriteJsonToFile(blockson, filePath)
	return err
}

// writes a page object to file
func (a *Archiver) WriteJsonToFile(object []byte, fileName string) error {
	// Create the file, or open it if it already exists, with write-only permissions
	// trim the final "/" from the filename
	outputFile := fmt.Sprintf("%s%s.json", a.rootDir, fileName[:len(fileName)-1])
	log.Println(outputFile)
	file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("error opening file: %w\n", err)
	}
	defer file.Close() // Ensure the file is closed after we're done

	// Write the JSON bytes to the file
	_, err = file.Write(object)
	if err != nil {
		return fmt.Errorf("error writing JSON to file: %w\n", err)
	}

	return nil // If we get here, the operation was a success
}

// Gets one level down from a given block
func (a *Archiver) GetBlockChildren(blockId string) ([]notion.Block, error) {
	ctx := context.Background()
	cursor := ""
	more := true
	blocks := make([]notion.Block, 0)

	for more {
		pq := &notion.PaginationQuery{
			StartCursor: cursor,
			PageSize:    utils.MaxPageSize,
		}

		resp, err := a.client.FindBlockChildrenByID(ctx, blockId, pq)
		if err != nil {
			log.Printf("Failed to find block children: %v\n", err)
			return nil, err
		}

		// What does this do about child pages and databases?
		// needs to exclude child pages
		// needs to send a get database request if it's a database
		blocks = append(blocks, resp.Results...)
		for _, block := range blocks {
			if block.HasChildren() {
				if block.BlockType() == notion.BlockTypeChildDatabase {
					db, err := a.GetDatabase(block.ID())
					if err != nil {
						log.Printf("Failed to get database details: %v\n", err)
					} else {
						block.SetExtras(db)
					}
				} else if block.BlockType() != notion.BlockTypeChildPage {
					children, _ := a.GetBlockChildren(block.ID())
					block.SetChildBlocks(children)
				}
			}
		}

		more = resp.HasMore
		if resp.NextCursor != nil {
			cursor = *resp.NextCursor
		}
	}

	return blocks, nil
}

// queries a database to get information about its entries
func (a *Archiver) GetDatabase(databaseId string) (notion.Database, error) {
	return a.client.FindDatabaseByID(a.ctx, databaseId)
}

func dirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
