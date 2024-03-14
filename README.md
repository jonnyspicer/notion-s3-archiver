# notion-utils

## TODO
1. ~~Non-programmatically move all the pages in my workspace under a single, top-level parent page. This not only makes it easier for the crawler, it also means I can easily share my entire workspace with an [integration](https://www.notion.so/integrations).~~
2. ~~Create a hashmap of `visited` pages, and a queue of pages marked as `to_visit`.~~
3. ~~Load in API key~~
4. ~~Send a `GET` request to the API for all the child-blocks of that page (even though it’s a page, it can be treated like a block, and requests can be made to the `blocks` API with its `id`).~~
5. ~~Handle pagination, as there is a max `page_size` of 100~~
6. ~~Add the top-level page to the `visited` hashmap.~~
7. ~~Parse the results for any links to other Notion pages. Check if they are in `visited`, and if not, add them to the `to_visit` queue.~~
8. Use a worker pool to take pages from the `to_visit` queue, and repeat steps 3-6.
9. Keep going until all threads are idle and the queue is empty.

## Notes 
- Need to fix voyager keybindings for hash, slashes, tildes and pipes
- Need to have API key and Page ID set as env vars
- ~~Should remove everything copied from go-notion, once import is sorted~~
- Remove setenv statements from crawler