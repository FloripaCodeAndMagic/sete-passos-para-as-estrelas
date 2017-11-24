package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	fetcher := WikiFetcher{}
	//page, _ := fetcher.FetchPage("https://pt.wikipedia.org", []string{"Ronaldo"})

	result, _ := getPathToPage(&fetcher, "https://pt.wikipedia.org", "Ronaldo", "Fausto Silva")
	fmt.Println(result)
}

func buildQueries(pageNames []string) []string {
	queries := []string{}

	for i := 0; i < int(math.Ceil(float64(len(pageNames))/50)); i++ {
		sliceLimit := 50 * (i + 1)
		if len(pageNames) < sliceLimit {
			sliceLimit = len(pageNames)
		}

		v := url.Values{}
		v.Set("action", "query")
		v.Add("titles", strings.Join(pageNames[50*i:sliceLimit], "|"))
		v.Add("prop", "links")
		v.Add("pllimit", "max")
		v.Add("format", "json")

		queries = append(queries, fmt.Sprintf("/w/api.php?%s", v.Encode()))
	}

	return queries
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
	FetchPage(string, []string) (map[string][]string, error)
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

func (fetcher *WikiFetcher) FetchPage(wikiURL string, pageNames []string) (map[string][]string, error) {
	queries := buildQueries(pageNames)

	articles := map[string][]string{}

	for _, query := range queries {
		fmt.Println(query)
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

		for _, page := range wikiResponse.Query.Pages {
			links := []string{}
			for _, link := range page.Links {
				links = append(links, link.Title)
			}

			articles[page.Title] = links
		}
	}

	return articles, nil
}

func getPathToPage(fetcher Fetcher, wikiURL string, startingPage string, endingPage string) ([]string, error) {
	pageMap := map[string]*Page{
		startingPage: &Page{startingPage, nil, nil},
	}

	// Space, the final frontier
	frontier := []string{startingPage}

	for len(frontier) > 0 {
		nextFrontier := []string{}

		linkMap, _ := fetcher.FetchPage(wikiURL, frontier)

		for parentTitle, childLinks := range linkMap {
			fmt.Println(parentTitle, len(childLinks))
			for _, childLink := range childLinks {
				childPage := Page{Title: childLink, Parent: pageMap[parentTitle]}
				pageMap[childLink] = &childPage

				if strings.ToLower(childLink) == strings.ToLower(endingPage) {
					return childPage.GetHierarchy(), nil
				}

				nextFrontier = append(nextFrontier, childLink)
			}
		}

		frontier = nextFrontier
	}

	return nil, errors.New("No more references could be found.")
}
