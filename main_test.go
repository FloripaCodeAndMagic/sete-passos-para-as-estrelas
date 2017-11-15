package main

import (
	"fmt"
	"net/url"
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
