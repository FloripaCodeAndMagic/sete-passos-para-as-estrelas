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

	articles, err := fetchPage(ts.URL, "Joseph Opala")
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

	articles, err := fetchPage(ts.URL, "Joseph Opala")
	if err != nil {
		t.Errorf("Unexpected fetch error")
	}

	expected := []string{}
	if !reflect.DeepEqual(articles, expected) {
		t.Errorf("Expected %v, got %v", expected, articles)
	}
}
