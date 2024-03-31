# notion-s3-archiver
A utility to back up any [Notion](https://www.notion.so) page, and all its subpages, to AWS S3 Glacier Deep Archive.

## Basic Usage

Assumes you several environment variables set; `NOTION_API_KEY`,  `NOTION_TOP_LEVEL_PAGE_ID`, `NOTION_OUTPUT_DIR`, 
`NOTION_BUCKET_NAME`, `NOTION_SNS_TOPIC_ARN` and `NOTION_AWS_REGION`. You get an API
key when you create an [integration](https://www.notion.so/integrations), and the page ID is the hexadecimal suffix in the URL 
of whichever page you are interested in having the crawler start at, eg for
`https://spill.notion.site/Careers-at-Spill-93c02882604b4b3ebfb7be0222692847` the id is
`93c02882604b4b3ebfb7be0222692847`. The output directory is where the downloaded json data will be stored - 
NB this directory will be created by the script, so you only need to provide the path. The bucket and
SNS names are the bucket to be uploaded to, and the topic which log messages will get published to - both of these must
be in the specified AWS rwegion.

There are many ways to set environment variables - for testing, the easiest is to use `export`:
```shell
export NOTION_API_KEY="123apiKEY"
export NOTION_TOP_LEVEL_PAGE_ID="93c02882604b4b3ebfb7be0222692847"
export NOTION_OUTPUT_DIR="/users/linus/repositories/notion-s3-archiver/out"
export NOTION_BUCKET_NAME="notion-s3-archive-bucket"
export NOTION_SNS_TOPIC_ARN="arn:aws:sns:eu-north-1:000000000000:my-sns-topic"
export NOTION_AWS_REGION="us-east-1"
```

Then download the repo, build the module and run the code.
```shell
git clone https://github.com/jonnyspicer/notion-s3-archiver.git
cd notion-s3-archiver
go build
./notion-s3-archiver
```

Don't be alarmed if this seems to take a while - the Notion API is rate-limited and so all operations in the script
are executed synchronously. As an example, my workspace contains roughly 4500 pages and takes around 3 hours for the
crawl alone.

## Future Improvements
1. Make the tool more customisable (eg SNS topic no longer required)
2. Add alternate places to store archived data, including on disk
3. Add options for eg bucket versioning, other S3 storage classes
4. Extract the crawler and the archiver out, document them etc