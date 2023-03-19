# Reddit Post Exporter
Fetch and save the JSON data of Reddit posts and their corresponding comments, for a given subreddit and specific category (top, controversial, hot or rising) up to a certain limit.

## Usage
Run the following command:

```go run RPCe.go -subreddit=<subreddit> -limit=<limit> -categoryID=<categoryID>
```
Where:

    - <subreddit>: The subreddit you want to fetch posts from. (default: "programming")
    - <limit>: The maximum number of posts to fetch. (default: 5)
    - <categoryID>: The category of posts to fetch (0 - top, 1 - controversial, 2 - hot, 3 - rising). (default: 0)

### Example

```
go run RPCe.go -subreddit=golang -limit=10 -categoryID=0
```
This will fetch 10 posts from "Top" sort from the "golang" subreddit and save their JSON data, along with their comments, in the following directory structure:

└── Day-Month-Year
    └── Hour-Minutes-Seconds
        └── top
            ├── post-<postID>.json
            └── comments
                ├── comment-<index>.json
                ├── comment-<index>.json
                └── ...

## How it works
The program first checks if the given subreddit exists by sending a GET request to https://www.reddit.com/r/<subreddit>/about.json. If the request is successful and the response contains a "data" field, then the subreddit exists.

It then constructs a URL to fetch the posts from, based on the input subreddit, maximum amount of posts, and category ID. The URL is sent as a GET request to the Reddit API, and the response JSON data is parsed to extract the posts.

For each post, the program creates a directory structure to save its JSON data and comments. The path for a post is structured as follows: <subreddit>/<date>/<time>/<category>/<postID>. The post's JSON data is saved in a file named post-<postID>.json.

If the post has comments, it creates a comments directory within the post's directory and saves each comment's JSON data in a file named comment-<commentID>.json. The comments are fetched by sending a GET request to the post's permalink with the .json extension.

## Disclaimer
This program is not affiliated with or endorsed by Reddit. Use at your own risk.

## License
