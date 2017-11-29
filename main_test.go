package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestBuildQuery(t *testing.T) {
	query := buildQuery("Joseph Opala")

	v := url.Values{}
	v.Set("action", "query")
	v.Add("titles", "Joseph Opala")
	v.Add("prop", "links")
	v.Add("pllimit", "max")
	v.Add("format", "json")

	expected := fmt.Sprintf("/w/api.php?%s", v.Encode())

	if query != expected {
		t.Errorf("Expected %s got %s", expected, query)
	}
}

func newTestWikiResponse(titles []string) WikiResponse {
	links := []Link{}
	for _, l := range titles {
		links = append(links, Link{l})
	}

	wr := WikiResponse{
		Query: WikiQuery{
			Pages: map[string]Page{
				"123": Page{
					Title: "Joseph Opala",
					Links: links,
				},
			},
		},
	}

	return wr
}

func TestFetchPage(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := newTestWikiResponse([]string{"whatever"})

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
	articles, err := fetcher.FetchPage(ts.URL, "Joseph Opala")
	if err != nil {
		t.Errorf("Unexpected fetch error")
	}

	expected := []string{"whatever"}
	if !reflect.DeepEqual(articles, expected) {
		t.Errorf("Expected %v, got %v", expected, articles)
	}
}

func TestFetchPageNoTitles(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := newTestWikiResponse([]string{})

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
	articles, err := fetcher.FetchPage(ts.URL, "Joseph Opala")
	if err != nil {
		t.Errorf("Unexpected fetch error")
	}

	expected := []string{}
	if !reflect.DeepEqual(articles, expected) {
		t.Errorf("Expected %v, got %v", expected, articles)
	}
}

type MockFetcher struct {
	returnIndex int
	returnMap   map[string][]string
}

func (m *MockFetcher) FetchPage(_ string, pageName string) ([]string, error) {
	return m.returnMap[pageName], nil
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
