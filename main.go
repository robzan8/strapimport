package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Article struct {
	Title, Slug, PublishDate, Excerpt, Content string

	FeatureImage FeatureImage
	Tags         []string
}

type Tag struct {
	Tag string `json:"tag"`
}

type FeatureImage struct {
	Id               int       `json:"id"`
	Name             string    `json:"name"`
	Hash             string    `json:"hash"`
	Sha256           string    `json:"sha256"`
	Ext              string    `json:"ext"`
	Mime             string    `json:"mime"`
	Size             float64   `json:"size"`
	Url              string    `json:"url"`
	Provider         string    `json:"provider"`
	ProviderMetadata *struct{} `json:"provider_metadata"` // always nil
	CreatedAt        string    `json:"created_at"`
	UpdatedAt        string    `json:"updated_at"`
}

func slugOf(title string) string {
	punct, err := regexp.Compile(`[^a-zA-Z0-9 ]+`)
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

const baseUrl = "https://webdata.gnucoop.io"

func main() {
	flag.StringVar(&token, "token", "", "auth token")
	flag.Parse()
	log.SetFlags(0)

	f, err := os.Open("blog_with_tags.json")
	if err != nil {
		log.Fatal(err)
	}
	dec := json.NewDecoder(f)
	var articles []Article
	err = dec.Decode(&articles)
	if err != nil {
		log.Fatal(err)
	}
	tags := make(map[string]bool)
	for _, article := range articles {
		for _, tag := range article.Tags {
			tags[tag] = true
		}
	}
	for tag := range tags {
		postTag(Tag{tag})
	}
}

func postTag(tag Tag) {
	body, err := json.Marshal(tag)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", baseUrl+"/tags", bytes.NewReader(body))
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
	log.Printf("%s\n", body)
}

func downloadImages() {
	imgSrcExp, err := regexp.Compile(`!\[\]\(([^\)]+)\)`)
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
		fmt.Fprintf(f, "\t\tTitle:        %q,\n", article.Title)
		fmt.Fprintf(f, "\t\tSlug:         %q,\n", article.Slug)
		fmt.Fprintf(f, "\t\tPublishDate:  %q,\n", article.PublishDate)
		fmt.Fprintf(f, "\t\tFeatureImage: FeatureImage{Name: %q},\n", article.FeatureImage.Name)
		fmt.Fprintf(f, "\t\tExcerpt:      %q,\n", article.Excerpt)
		f.WriteString("\t\tContent: `")
		f.WriteString(article.Content)
		f.WriteString("`,\n")
		f.WriteString("\t},\n")
	}
	f.WriteString("}\n")
}

func readFeatureImages() {
	f, err := os.Open("feature_images.json")
	if err != nil {
		log.Fatal(err)
	}
	dec := json.NewDecoder(f)
	var images []FeatureImage
	err = dec.Decode(&images)
	if err != nil {
		log.Fatal(err)
	}
	for i := range blog {
		article := &blog[i]
		article.FeatureImage = findFeatureImage(images, article.FeatureImage.Name)
	}
}

func findFeatureImage(images []FeatureImage, name string) FeatureImage {
	for _, img := range images {
		if img.Name == name {
			return img
		}
	}
	panic(name + " not found")
}

func postArticle(article *Article) {
	body, err := json.Marshal(article)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", baseUrl+"/articles", bytes.NewReader(body))
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

func postFeatureImage(fileName string) {
	img, err := os.Open("feature_images/" + fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer img.Close()

	var buf bytes.Buffer
	multiw := multipart.NewWriter(&buf)
	filew, err := multiw.CreateFormFile("files", fileName)
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(filew, img)
	err = multiw.Close()
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("POST", baseUrl+"/upload", &buf)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", multiw.FormDataContentType())
	if token != "" {
		req.Header.Add("Authorization", "Token "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Fatalf("Image %s: unexpected %d response:\n%s", fileName, resp.StatusCode, body)
	}
	log.Printf("%s\n", body)
}
