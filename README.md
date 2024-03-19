# notion-utils
A collection of small utilities for managing [Notion](https://www.notion.s) workspaces.

## Crawler

Recursively searches for any child pages from a given parent page, and returns a list of all links and paths relative to the parent.
Handy for backups - consider moving everything in your workspace under a single top-level file if you'd like to back your whole workspace up.
Runs synchronously due to Notion's API rate limiting.

### Basic Usage

Assumes you have environment variables for `NOTION_API_KEY` and `TOP_LEVEL_PAGE_ID` set. You get an API
key when you create an [integration](https://www.notion.so/integrations), and the page ID is the hexadecimal suffix in the URL 
of whichever page you are interested in having the crawler start at, eg for
`https://spill.notion.site/Careers-at-Spill-93c02882604b4b3ebfb7be0222692847` the id is
`93c02882604b4b3ebfb7be0222692847`.

There are many ways to set environment variables - for testing, the easiest is to use `export`:
```shell
export NOTION_API_KEY="123apiKEY"
export TOP_LEVEL_PAGE_ID="93c02882604b4b3ebfb7be0222692847"
```

Then install the `notion-utils` module. You'll need to have at least go 1.21 installed, or it'll complain.
```shell
go get github.com/jonnyspicer/notion-utils
go install github.com/jonnyspicer/notion-utils
```

Now the module is available for import using: `import "github.com/jonnyspicer/notion-utils"`. The "Hello World"
version of the crawler looks something like this:

```go
package main

import (
	"github.com/jonnyspicer/notion-utils/crawler"
)

func main() {
	c := crawler.NewCrawler()
	c.FullCrawl() 
	c.ListAllPageIds()
}
```
