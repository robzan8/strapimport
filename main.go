package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Article struct {
	Title, Slug, Excerpt, Content string
}

func slugOf(title string) string {
	punct, err := regexp.Compile("[^a-zA-Z0-9 ]+")
	if err != nil {
		log.Fatal(err)
	}
	title = punct.ReplaceAllString(title, "")
	title = strings.ToLower(title)
	title = strings.ReplaceAll(title, " ", "-")
	title = strings.ReplaceAll(title, "--", "-")
	return title
}

var (
	token  string
	client http.Client
)

const postUrl = "https://webdata.gnucoop.io/articles"

func main() {
	flag.StringVar(&token, "token", "", "auth token")
	flag.Parse()
	log.SetFlags(0)

	/* read blog from gnucoop_blog.json
	var blog []Article
	f, err := os.Open(blogFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	err = dec.Decode(&blog)
	if err != nil {
		log.Fatal(err)
	}
	for i := range blog {
		article := &blog[i]
		article.Slug = slugOf(article.Title)
		//postArticle(article)
	}*/

	f, err := os.Create("gnucoop_blog.go")
	if err != nil {
		log.Fatal(err)
	}
	f.WriteString("package main\n\n")
	f.WriteString("var blog = []Article{\n")
	for _, article := range blog {
		f.WriteString("{\n")
		fmt.Fprintf(f, "Title: %q,\n", article.Title)
		fmt.Fprintf(f, "Slug: %q,\n", article.Slug)
		fmt.Fprintf(f, "Excerpt: %q,\n", article.Excerpt)
		f.WriteString("Content: `")
		f.WriteString(article.Content)
		f.WriteString("`,\n")
		f.WriteString("},\n")
	}
	f.WriteString("}\n")
}

func postArticle(article *Article) {
	body, err := json.Marshal(article)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", postUrl, bytes.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	if token != "" {
		req.Header.Add("Authorization", "Token "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Fatalf("Unexpected response with code %d:\n%s", resp.StatusCode, body)
	}
}
