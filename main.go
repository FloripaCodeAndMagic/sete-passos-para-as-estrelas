package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func main() {
}

func buildQuery(pageName string) string {
	v := url.Values{}
	v.Set("action", "query")
	v.Add("titles", pageName)
	v.Add("prop", "links")
	v.Add("pllimit", "max")
	v.Add("format", "json")

	return fmt.Sprintf("/w/api.php?%s", v.Encode())
}

type Link struct {
	Title string `json:"title"`
}

type Page struct {
	Title string `json:"title"`
	Links []Link `json:"links"`
}

func (p *Page) articles() []string {
	articles := []string{}

	for _, l := range p.Links {
		articles = append(articles, l.Title)
	}

	return articles
}

type WikiQuery struct {
	Pages map[string]Page `json:"pages"`
}

type WikiResponse struct {
	Query WikiQuery `json:"query"`
}

func (wq *WikiQuery) articles() []string {
	articles := []string{}
	for _, page := range wq.Pages {
		articles = append(articles, page.articles()...)
	}

	return articles
}

func (wr *WikiResponse) articles() []string {
	return wr.Query.articles()
}

func fetchPage(wikiUrl string, pageName string) ([]string, error) {
	query := buildQuery(pageName)
	url := fmt.Sprintf("%s%s", wikiUrl, query)
	res, err := http.Get(url)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	wikiResponse := WikiResponse{}
	err = json.Unmarshal(body, &wikiResponse)
	if err != nil {
		return nil, err
	}

	articles := wikiResponse.articles()

	return articles, nil
}
