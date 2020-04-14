package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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

func excerptOf(content string) string {
	// todo: remove images
	n := strings.Index(content, "\n")
	return content[0 : n+1]
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

	postImages()
}

func downloadImages() {
	imgSrcExp, err := regexp.Compile(`<img [^>]*src="([^"]+)"`)
	if err != nil {
		log.Fatal(err)
	}
	var imgSources []string
	for _, article := range blog {
		matches := imgSrcExp.FindAllStringSubmatch(article.Content, -1)
		for _, m := range matches {
			imgSources = append(imgSources, m[1]) // image src is submatch 1
		}
	}
	for _, src := range imgSources {
		resp, err := http.Get(src)
		if err != nil {
			log.Println(err)
			continue
		}
		defer resp.Body.Close()
		name := imageName(src)
		f, err := os.Create("images/" + name)
		if err != nil {
			log.Fatal(err)
		}
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Downloaded: %s\n", name)
	}
}

func imageName(src string) string {
	a := strings.LastIndex(src, "/")
	b := strings.LastIndex(src, "?")
	if b == -1 || b < a {
		b = len(src)
	}
	return src[a+1 : b]
}

func dumpBlog() {
	f, err := os.Create("gnucoop_blog.go")
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.WriteString("package main\n\n")
	if err != nil {
		log.Fatal(err)
	}
	f.WriteString("var blog = []Article{\n")
	for _, article := range blog {
		f.WriteString("\t{\n")
		fmt.Fprintf(f, "\t\tTitle:   %q,\n", article.Title)
		fmt.Fprintf(f, "\t\tSlug:    %q,\n", article.Slug)
		fmt.Fprintf(f, "\t\tExcerpt: %q,\n", article.Excerpt)
		f.WriteString("\t\tContent: `")
		f.WriteString(article.Content)
		f.WriteString("`,\n")
		f.WriteString("\t},\n")
	}
	f.WriteString("}\n")
}

func readBlogJson() {
	f, err := os.Open("gnucoop_blog.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	err = dec.Decode(&blog)
	if err != nil {
		log.Fatal(err)
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
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Fatalf("Unexpected response with code %d:\n%s", resp.StatusCode, body)
	}
}

func postImages() {

}
