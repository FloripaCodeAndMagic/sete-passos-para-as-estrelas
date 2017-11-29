package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
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
	Title  string `json:"title"`
	Links  []Link `json:"links"`
	Parent *Page
}

func (p *Page) articles() []string {
	articles := []string{}

	for _, l := range p.Links {
		articles = append(articles, l.Title)
	}

	return articles
}

func (p *Page) GetHierarchy() []string {
	path := []string{}
	currentPage := p

	for currentPage != nil {
		path = append(path, currentPage.Title)
		currentPage = currentPage.Parent
	}

	return path
}

type WikiQuery struct {
	Pages map[string]Page `json:"pages"`
}

type WikiResponse struct {
	Query WikiQuery `json:"query"`
}

type Fetcher interface {
	FetchPage(string, string) ([]string, error)
}

type WikiFetcher struct{}

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

func (fetcher *WikiFetcher) FetchPage(wikiURL string, pageName string) ([]string, error) {
	query := buildQuery(pageName)
	url := fmt.Sprintf("%s%s", wikiURL, query)
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

func getPathToPage(fetcher Fetcher, wikiURL string, startingPage string, endingPage string) ([]string, error) {
	rootPage := Page{startingPage, nil, nil}

	// Space, the final frontier
	frontier := []Page{rootPage}

	for len(frontier) > 0 {
		children := []Page{}

		for i, page := range frontier {
			childrenTitles, _ := fetcher.FetchPage(wikiURL, page.Title)

			for _, childTitle := range childrenTitles {
				childPage := Page{Title: childTitle, Parent: &frontier[i]}

				if strings.ToLower(childTitle) == strings.ToLower(endingPage) {
					return childPage.GetHierarchy(), nil
				}

				children = append(children, childPage)
			}
		}

		frontier = children
	}

	return nil, errors.New("No more references could be found.")
}
