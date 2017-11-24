package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestBuildQueriesSinglePage(t *testing.T) {
	query := buildQueries([]string{"Joseph Opala"})

	v := url.Values{}
	v.Set("action", "query")
	v.Add("titles", "Joseph Opala")
	v.Add("prop", "links")
	v.Add("pllimit", "max")
	v.Add("format", "json")

	expected := fmt.Sprintf("/w/api.php?%s", v.Encode())

	if query[0] != expected {
		t.Errorf("Expected %s got %s", expected, query)
	}
}

func TestBuildQueriesMultiplePages(t *testing.T) {
	query := buildQueries([]string{"Joseph Opala", "Ferrari"})

	v := url.Values{}
	v.Set("action", "query")
	v.Add("titles", "Joseph Opala|Ferrari")
	v.Add("prop", "links")
	v.Add("pllimit", "max")
	v.Add("format", "json")

	expected := fmt.Sprintf("/w/api.php?%s", v.Encode())

	if query[0] != expected {
		t.Errorf("Expected %s got %s", expected, query)
	}
}

func TestBuildQueriesFiftyPages(t *testing.T) {
	fiftyTwoStrings := []string{}
	fiftyStrings := []string{}
	remainingStrings := []string{}

	for i := 0; i < 52; i++ {
		if i < 50 {
			fiftyStrings = append(fiftyStrings, strconv.Itoa(i))
		} else {
			remainingStrings = append(remainingStrings, strconv.Itoa(i))
		}

		fiftyTwoStrings = append(fiftyTwoStrings, strconv.Itoa(i))
	}

	queries := buildQueries(fiftyTwoStrings)

	firstQuery := url.Values{}
	firstQuery.Set("action", "query")
	firstQuery.Add("titles", strings.Join(fiftyStrings, "|"))
	firstQuery.Add("prop", "links")
	firstQuery.Add("pllimit", "max")
	firstQuery.Add("format", "json")

	expectedFirstQuery := fmt.Sprintf("/w/api.php?%s", firstQuery.Encode())

	secondQuery := url.Values{}
	secondQuery.Set("action", "query")
	secondQuery.Add("titles", strings.Join(remainingStrings, "|"))
	secondQuery.Add("prop", "links")
	secondQuery.Add("pllimit", "max")
	secondQuery.Add("format", "json")

	expectedSecondQuery := fmt.Sprintf("/w/api.php?%s", secondQuery.Encode())

	if len(queries) != 2 {
		t.Errorf("Expected two queries to be returned.")
	}

	if queries[0] != expectedFirstQuery {
		t.Errorf("Expected %s got %s", expectedFirstQuery, queries[0])
	}

	if queries[1] != expectedSecondQuery {
		t.Errorf("Expected %s got %s", expectedSecondQuery, queries[1])
	}
}

func newTestWikiResponse(responseMap map[string][]string) WikiResponse {
	pageMap := map[string]Page{}

	i := 0

	for title, links := range responseMap {
		linkInstances := []Link{}

		for _, l := range links {
			linkInstances = append(linkInstances, Link{l})
		}

		page := Page{
			Title: title,
			Links: linkInstances,
		}

		pageMap[strconv.Itoa(i)] = page
		i++
	}

	wr := WikiResponse{
		Query: WikiQuery{
			Pages: pageMap,
		},
	}

	return wr
}

func TestFetchPage(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		payload := newTestWikiResponse(map[string][]string{
			"Joseph Opala": []string{"whatever"},
			"Ferrari":      []string{"lamborghini"},
		})

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			t.Errorf("Failed to encode mocked data")
		}
		w.Write(jsonPayload)

		wikiResponse := WikiResponse{}
		err = json.Unmarshal(jsonPayload, &wikiResponse)

		if err != nil {
			t.Errorf("Failed to encode mocked data")
		}

	}
	ts := httptest.NewServer(http.HandlerFunc(testHandler))
	defer ts.Close()

	fetcher := WikiFetcher{}
	articles, err := fetcher.FetchPage(ts.URL, []string{"Joseph Opala", "Ferrari"})
	if err != nil {
		t.Errorf("Unexpected fetch error")
	}

	expected := map[string][]string{
		"Joseph Opala": []string{"whatever"},
		"Ferrari":      []string{"lamborghini"},
	}

	if !reflect.DeepEqual(articles, expected) {
		t.Errorf("Expected %v, got %v", expected, articles)
	}
}

func TestFetchMoreThanFiftyPages(t *testing.T) {
	pageList := []string{}
	for i := 0; i < 51; i++ {
		pageList = append(pageList, strconv.Itoa(i))
	}

	requestCount := 0

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		payload := newTestWikiResponse(map[string][]string{})

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			t.Errorf("Failed to encode mocked data")
		}
		w.Write(jsonPayload)

		wikiResponse := WikiResponse{}
		err = json.Unmarshal(jsonPayload, &wikiResponse)

		if err != nil {
			t.Errorf("Failed to encode mocked data")
		}
	}

	ts := httptest.NewServer(http.HandlerFunc(testHandler))
	defer ts.Close()

	fetcher := WikiFetcher{}
	_, err := fetcher.FetchPage(ts.URL, pageList)
	if err != nil {
		t.Errorf("Unexpected fetch error")
	}

	if requestCount != 2 {
		t.Errorf("Expected 2 requests to be made, got %v", requestCount)
	}
}

func TestFetchPageNoTitles(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := newTestWikiResponse(map[string][]string{})

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			t.Errorf("Failed to encode mocked data")
		}
		w.Write(jsonPayload)

		wikiResponse := WikiResponse{}
		err = json.Unmarshal(jsonPayload, &wikiResponse)

		if err != nil {
			t.Errorf("Failed to encode mocked data")
		}

	}
	ts := httptest.NewServer(http.HandlerFunc(testHandler))
	defer ts.Close()

	fetcher := WikiFetcher{}
	articles, err := fetcher.FetchPage(ts.URL, []string{"Joseph Opala"})
	if err != nil {
		t.Errorf("Unexpected fetch error")
	}

	expected := map[string]string{}
	if len(articles) != 0 {
		t.Errorf("Expected %v, got %v", expected, articles)
	}
}

type MockFetcher struct {
	returnIndex int
	returnMap   map[string][]string
}

func (m *MockFetcher) FetchPage(_ string, pageNames []string) (map[string][]string, error) {
	return m.returnMap, nil
}

func (m *MockFetcher) SetReturnMap(returnMap map[string][]string) {
	m.returnMap = returnMap
}

func TestPageHierarchy(t *testing.T) {
	god := Page{Title: "God", Parent: nil}
	grandfather := Page{Title: "Grandfather", Parent: &god}
	father := Page{Title: "Father", Parent: &grandfather}
	son := Page{Title: "Son", Parent: &father}

	hierarchy := son.GetHierarchy()
	expected := []string{"Son", "Father", "Grandfather", "God"}

	if !reflect.DeepEqual(hierarchy, expected) {
		t.Errorf("Expected %v, got %v", expected, hierarchy)
	}
}

func TestGetPathToPage(t *testing.T) {
	mockFetcher := new(MockFetcher)
	mockFetcher.SetReturnMap(map[string][]string{
		"Joseph Opala": []string{"Batman", "Robin"},
		"Batman":       []string{"Belt", "Ferrari"},
		"Robin":        []string{"Joker"},
		"Belt":         []string{"Metal"},
		"Ferrari":      []string{"Lamborghini", "Bitcoin"},
	})

	path, err := getPathToPage(mockFetcher, "BatPedia.org", "Joseph Opala", "Bitcoin")

	if err != nil {
		t.Errorf("Unexpected fetch error")
	}

	expected := []string{"Bitcoin", "Ferrari", "Batman", "Joseph Opala"}

	if !reflect.DeepEqual(path, expected) {
		t.Errorf("Expected %v, got %v", expected, path)
	}
}
