# Reddit Post Exporter
Fetch and save the JSON data of Reddit posts and their comments, from specified subreddit and specific category (top, controversial, hot or rising) up to a certain limit, which is specified by the user, too.

### Example
Check [example](https://github.com/sncelta/reddit-post-exporter/tree/example) branch to see results.

## Prerequisites
- Go
- Git (optional)
- Stable internet connection

## Install

### Build
1. Clone the repository using Git:
```
git clone https://github.com/sncelta/reddit-post-exporter.git
```
...or download the source from GitHub.

2. Navigate to `src` folder and run `go build` in terminal. You may skip building the program if you have Go installed and want to run it from source.

That will create an executable called `reddit-post-exporter`.

If you don't feel like building it from source, go to Releases page.

## Usage
Run the following command:

```
reddit-post-exporter -subreddit <subreddit> -limit <limit> -categoryID <categoryID>
```
Where:

    - <subreddit>: The subreddit you want to fetch posts from. (default: "programming")
    - <limit>: The maximum number of posts to fetch. (default: 5)
    - <categoryID>: The category of posts to fetch (0 - top, 1 - controversial, 2 - hot, 3 - rising). (default: 0)

### Example

```
reddit-post-exporter -subreddit golang -limit 10 -categoryID 0
```
This will fetch 10 posts from "Top" sort from the "golang" subreddit and save their JSON data, along with their comments, in the following directory structure:
```
└── <subreddit>
    └── Day-Month-Year
        └── Hour-Minutes-Seconds
            └── Category (e.g rising)
                └── posts
                    └── post-<postID>
                        ├── post-<index>.json
                        └── comments
                            └── comment-<index>.json
```

## How it works
The program first checks if the given subreddit exists by sending a GET request to `https://www.reddit.com/r/<subreddit>/about.json`. If the request is successful and the response contains a "data" field, then the subreddit exists.

It then constructs a URL to fetch the posts from, based on the input subreddit, maximum amount of posts (specified in `limit` flag), and category ID. The URL is sent as a GET request to the Reddit API, and the response JSON data is parsed to extract the posts.

For each post, the program creates a directory structure to save its JSON data and comments. The path for a post is structured as follows: `<subreddit>/<date>/<time>/<category>/<postID>`. The post's JSON data is saved in a file named post-<postID>.json.

If the post has comments, it creates a comments directory within the post's directory and saves each comment's JSON data in a file named comment-<commentID>.json. The comments are fetched by sending a GET request to the post's permalink with the .json extension.

## Disclaimer
This program is not affiliated with or endorsed by Reddit. Use at your own risk.

## Contributing
If you find a bug or have a suggestion for this program, please feel free to open an issue or submit a pull request on GitHub.

## License
This project is licensed under the MIT License. See the LICENSE file for details.

## Notes
If you want to remove the delay while exporting comments, go to line 183 and remove `time.Sleep(1 * time.Second)`.
You may as well adjust the delay to your own preference.
