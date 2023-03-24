package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var categoryID = [5]string{"new", "top", "controversial", "hot", "rising"}

// I don't want to set f*ckton of headers for every request...
// http.NewRequest wrapper with additional headers

func rpeRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, path, body)

	if err != nil {
		return nil, err
	}

	// Set headers to prevent Reddit from blocking and rate-limiting
	req.Header.Set("authority", "www.reddit.com")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="89", "Chromium";v="89", ";Not A Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("upgrade-insecure-requests", "1")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36")
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("dnt", "1")
	req.Header.Set("sec-fetch-site", "none")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("accept-language", "en-GB,en;q=0.9")

	return req, nil
}

// Error handler function
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

/* Wrapper function for creating directories with permissions and instantly
saving file to the directory with error handling */

func saveToDir(path, filename string, perm fs.FileMode) *os.File {
	checkError(os.MkdirAll(path, perm))

	file, err := os.Create(filepath.Join(".", path, filename))
	checkError(err)

	return file
}
func subredditValid(subreddit string) bool {
	// Construct the URL of the subreddit's about.json page
	url := fmt.Sprintf("https://www.reddit.com/r/%s/about.json", subreddit)

	client := &http.Client{}

	req, err := rpeRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return false
	}
	defer resp.Body.Close()

	// Parse the JSON response
	var data struct {
		Data struct {
			SubredditType     string `json:"subreddit_type"`
			Subscribers       int    `json:"subscribers"`
		} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return false
	}

	/* Check if the subreddit exists and is public.
	If it doesn't exist, it won't have the Subscribers field, thus it will return false */

	return data.Data.Subscribers >= 0 && data.Data.SubredditType != "private"

}

func fetchPosts(subreddit string, id, limit int, export_comments bool) {
	// Construct the URL to get posts from, based on input subreddit, maximum amount of posts and category id
	var url string = fmt.Sprintf("https://www.reddit.com/r/%s/%s.json?limit=%d", subreddit, categoryID[id], limit)

	client := &http.Client{}
	req, err := rpeRequest("GET", url, nil)
	checkError(err)

	/* Sends an HTTP request and returns an HTTP response, following policy
	   (such as redirects, cookies, auth) as configured on the client. */

	res, err := client.Do(req)
	checkError(err)

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatalf("HTTP request failed with status code %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	checkError(err)

	// Data struct
	var data struct {
		Data struct {
			Children []struct {
				Data map[string]interface{} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}

	// Unmarshal JSON from the body variable into the data struct
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatal(err)
	}

	// Extract data field for the posts
	posts := make([]map[string]interface{}, len(data.Data.Children))

	for i := range data.Data.Children {
		posts[i] = data.Data.Children[i].Data
	}

	exportPosts(posts, subreddit, id, export_comments)
}

func exportPosts(posts []map[string]interface{}, subreddit string, id int, export_comments bool) {
	now := time.Now()

	for i, post := range posts {
		// Iterate over each post in the posts slice
		jsonData, err := json.MarshalIndent(post, "", "  ")
		checkError(err)

		var dateNow string = now.Format("01-Jan-2006")
		var timeNow string = now.Format("15-04-05")
		var filename string = fmt.Sprintf("post-%d.json", i)

		var path string = filepath.Join(".", subreddit, dateNow, timeNow, categoryID[id], fmt.Sprintf("post-%s", post["id"].(string)))

		file := saveToDir(path, filename, 0700)

		defer file.Close()

		n, err := io.WriteString(file, string(jsonData))
		checkError(err)

		fmt.Printf("Saved post %d to path %s\nBytes: %d\n", i, path, n)

		if export_comments {
			fetchComments(post, path, i)
		}
	}
}

func fetchComments(post map[string]interface{}, path string, postIndex int) {
	// If the post has comments, fetch them and save them to a directory within the post's directory
	if post["num_comments"].(float64) != 0 {
		permalink := post["permalink"].(string) + ".json"

		// Construct the URL for the post's comments using post's permalink key
		commentsURL := fmt.Sprintf("https://www.reddit.com%s", permalink)
		commentsReq, err := rpeRequest("GET", commentsURL, nil)
		checkError(err)

		cfClient := &http.Client{}
		commentsRes, err := cfClient.Do(commentsReq)
		checkError(err)

		defer commentsRes.Body.Close()

		commentsBody, err := io.ReadAll(commentsRes.Body)
		checkError(err)

		var commentsData []interface{}
		checkError(json.Unmarshal(commentsBody, &commentsData))

		exportComments(commentsData, path, postIndex)
	}
}
func exportComments(commentsData []interface{}, path string, postIndex int) {
	commentPath := filepath.Join(path, "comments")

	if len(commentsData) >= 2 {
		data, ok := commentsData[1].(map[string]interface{})["data"].(map[string]interface{})
		if ok && data != nil {
			if children, ok := data["children"].([]interface{}); ok {
				for j, child := range children {
					comment := child.(map[string]interface{})

					commentJSONData, err := json.MarshalIndent(comment, "", "  ")
					checkError(err)

					commentFilename := fmt.Sprintf("comment-%d.json", j)
					commentFile := saveToDir(commentPath, commentFilename, 0700)

					defer commentFile.Close()

					n, err := io.WriteString(commentFile, string(commentJSONData))
					checkError(err)

					fmt.Printf("Saved comment %d for post %d to path %s\nBytes: %d\n", j, postIndex, commentPath, n)
				}
			}
		}
	}
}

func main() {
	// Define flags
	var subreddit string
	var limit, id int
	var export_comments bool

	flag.StringVar(&subreddit, "subreddit", "programming", "Subreddit to fetch posts from")
	flag.IntVar(&limit, "limit", 5, "Amount of posts to fetch")
	flag.IntVar(&id, "categoryID", 0, "Category of posts to fetch\n0 - new\n1 - top\n2 - controversial\n3 - hot\nr4 - rising")
	flag.BoolVar(&export_comments, "exportComments", true, "Toggle comment exporting")

	flag.Parse()

	if !subredditValid(subreddit) {
		log.Fatal("Specified subreddit is invalid. Are you sure it exists or isn't banned/private?")
	}

	id = int(math.Min(float64(id), 3))

	fetchPosts(subreddit, id, limit, export_comments)
}
