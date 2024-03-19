package crawler

import (
	"os"
)

func getNotionApiKey() string {
	return os.Getenv(NotionApiKey)
}

func getTopLevelPageId() string {
	return os.Getenv(TopLevelPageId)
}
