package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Blog struct {
	Articles []Article `xml:"channel>item"`
}

type Article struct {
	Title   string `xml:"title"`
	Slug    string `xml:"-"`
	Excerpt string `xml:"excerpt"`
	Content string `xml:"content"`
}

func slugOf(title string) string {
	punct, err := regexp.Compile("[^a-zA-Z0-9 ]+")
	if err != nil {
		log.Fatal(err)
	}
	title = punct.ReplaceAllString(title, "")
	title = strings.ToLower(title)
	title = strings.ReplaceAll(title, " ", "-")
	return title
}

var (
	token  string
	client http.Client
)

const (
	blogFile = "gnucoop_blog.xml"
	postUrl  = "https://webdata.gnucoop.io/articles"
)

func main() {
	flag.StringVar(&token, "token", "", "auth token")
	flag.Parse()
	log.SetFlags(0)

	var blog Blog
	f, err := os.Open(blogFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	dec := xml.NewDecoder(f)
	err = dec.Decode(&blog)
	if err != nil {
		log.Fatal(err)
	}
	for i := range blog.Articles {
		article := &blog.Articles[i]
		article.Slug = slugOf(article.Title)
		postArticle(article)
	}
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
	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Unexpected response with code %d:\n%s", resp.StatusCode, body)
	}
}
