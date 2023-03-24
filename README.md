# Reddit Post Exporter
![Preview](https://s10.gifyu.com/images/Desktop-3-21-2023-11-33-59-PM.gif)

### Unfortunately, 100 posts is the maximum because of how Reddit works. https://www.reddit.com/dev/api#GET_top see limit property

Fetch and export the JSON data of Reddit posts and their comments if needed, from specified subreddit and specific category (top, controversial, hot or rising) up to a certain limit, which is specified by the user, too, without using additional API wrappers, OAuth2, and many other third-party libraries/packages/etc. 

Note that it will export only those posts that fit the category. For example, I managed to export 100 posts and ~4,565 comments respectively from Hot category in AskReddit.

### Example
Check [example](https://github.com/sncelta/reddit-post-exporter/tree/example) branch to see results.

### Notes
If you want to know what each key in comment/post JSON means, read [here](https://www.reddit.com/dev/api/)

### Benchmarks
I tried exporting AskReddit again and it took me 1:01 minutes to export 100 posts and ~4126 comments with them.

```
$ time go run main.go -subreddit=AskReddit -limit=100 -categoryID=3
```

```
real    1m1.401s
user    0m2.384s
sys     0m0.697s
```

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

2. Navigate to `rpe` folder and run `go build` in terminal. You may skip building the program if you have Go installed and want to run it from source.

That will create an executable called `reddit-post-exporter`

If you don't feel like building it from source, go to Releases page.

## Usage
Run the following command:

```
reddit-post-exporter -subreddit=<string> -limit=<int> -categoryID=<int> -exportComments=<bool>
```

### Arguments
- `subreddit` : Subreddit to get data from
- `limit` : Amount of posts to fetch and export from the subreddit (default: 5)
- `categoryID` : Specifies from which sort/category posts will be retrieved (default: 0)

    0 - Top

    1 - Controversial
    
    2 - Hot
    
    3 - Rising

- `exportComments` : Whether it should retrieve comments or not. (default: true)

### Example

```
reddit-post-exporter -subreddit=golang -limit=10 -categoryID=0 -exportComments=true
```
This will fetch and export 10 posts from Top category from golang subreddit and save their JSON data, along with their comments (because `exportComments` is true), in the following directory structure:
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
The program first checks if the given subreddit is valid by sending a GET request to `https://www.reddit.com/r/<subreddit>/about.json`. If the request is successful and the response contains a "data" field, then the subreddit is valid.

It then constructs a URL to fetch the posts from, based on the input subreddit, maximum amount of posts (specified in `limit` flag), and category ID. The URL is sent as a GET request to the Reddit API, and the response JSON data is parsed to extract the posts.

For each post, it creates a directory structure to save its JSON data and comments. The path for a post is structured like: `<subreddit>/<date>/<time>/<category>/<postID>`. The post's JSON data is saved in `post-<postID>.json`.

If the post has comments, it creates a `comments` directory within the post's directory and saves each comment's JSON data in `comment-<index>.json`. The comments are fetched by sending a GET request to the post's permalink with the .json extension.

## Disclaimer
This program is not affiliated with or endorsed by Reddit. Use at your own risk.

## Contributing
If you find a bug or have a suggestion for this program, please feel free to open an issue or submit a pull request on GitHub.

## License
This project is licensed under the MIT License. See the LICENSE file for details.
