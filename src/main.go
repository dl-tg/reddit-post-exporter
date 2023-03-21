package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var categoryID = [4]string{"top", "controversial", "hot", "rising"}

// I don't want to set f*ckton of headers for every request...

func rpe_request(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, path, body)

	if err != nil {
		return nil, err
	}

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
func checkError(err error) {
	// Error handler function
	if err != nil {
		log.Fatal(err)
	}
}

func subredditValid(subreddit string) bool {
	// Validate the subreddit to avoid unexpected errors by checking if it has "data" key
	url := fmt.Sprintf("https://www.reddit.com/r/%s/about.json", subreddit)
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return false
	}

	if _, ok := data["data"]; !ok {
		return false
	}

	return true
}

func fetchPosts(subreddit string, id int, limit int) {
	// Construct the URL to get posts from, based on input subreddit, maximum amount of posts and category id
	var url string = fmt.Sprintf("https://www.reddit.com/r/%s/%s.json?limit=%d", subreddit, categoryID[id], limit)

	client := &http.Client{}
	req, err := rpe_request("GET", url, nil)
	checkError(err)

	res, err := client.Do(req)
	checkError(err)

	defer res.Body.Close()

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

	// Create an empty slice of maps to hold each post's data
	posts := []map[string]interface{}{}

	// Iterate over each child in the data struct's Children slice, extract its data field and append it to the posts slice
	for _, child := range data.Data.Children {
		post := child.Data
		posts = append(posts, post)
	}

	exportPosts(posts, subreddit, id)
}

func exportPosts(posts []map[string]interface{}, subreddit string, id int) {
	now := time.Now()

	day, month, year := now.Day(), now.Month(), now.Year()
	hour, minute, second := now.Hour(), now.Minute(), now.Second()

	for i, post := range posts {
		// Iterate over each post in the posts slice
		jsonData, err := json.MarshalIndent(post, "", "  ")
		checkError(err)

		var dateNow string = fmt.Sprintf("%d-%v-%d", day, month, year)
		var timeNow string = fmt.Sprintf("%d-%d-%d", hour, minute, second)
		var filename string = fmt.Sprintf("post-%d.json", i)

		path := filepath.Join(".", subreddit, dateNow, timeNow, categoryID[id], fmt.Sprintf("post-%s", post["id"].(string)))
		filePath := filepath.Join(path, filename)

		checkError(os.MkdirAll(path, 0755))

		file, err := os.Create(filePath)
		checkError(err)

		defer file.Close()

		n, err := io.WriteString(file, string(jsonData))
		checkError(err)

		fmt.Printf("Saved post %d to path %s\nBytes: %d\n", i, path, n)

		fetchComments(post, path, i)
	}
}

func fetchComments(post map[string]interface{}, path string, postIndex int) {
	// If the post has comments, fetch them and save them to a directory within the post's directory
	if post["num_comments"].(float64) != 0 {
		permalink := post["permalink"].(string) + ".json"

		// Construct the URL for the post's comments using post's permalink key
		commentsURL := "https://www.reddit.com" + permalink
		commentsReq, err := rpe_request("GET", commentsURL, nil)
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
	os.MkdirAll(commentPath, 0755)

	if len(commentsData) >= 2 {
		data, ok := commentsData[1].(map[string]interface{})["data"].(map[string]interface{})
		if ok && data != nil {
			if children, ok := data["children"].([]interface{}); ok {
				for j, child := range children {
					comment := child.(map[string]interface{})

					commentJSONData, err := json.MarshalIndent(comment, "", "  ")
					checkError(err)

					commentFilename := fmt.Sprintf("comment-%d.json", j)
					commentFile, err := os.Create(filepath.Join(commentPath, commentFilename))
					checkError(err)

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
	var limit int
	var id int

	flag.StringVar(&subreddit, "subreddit", "programming", "Subreddit to fetch posts from")
	flag.IntVar(&limit, "limit", 5, "Amount of posts to fetch")
	flag.IntVar(&id, "categoryID", 0, "Category of posts to fetch\n0 - top\n1 - controversial\n2 - hot\n3 - rising")

	flag.Parse()

	if !subredditValid(subreddit) {
		log.Fatal("Specified subreddit is invalid. Are you sure it exists or isn't banned/private?")
	}

	id = int(math.Min(float64(id), 3))

	fetchPosts(subreddit, id, limit)

}
