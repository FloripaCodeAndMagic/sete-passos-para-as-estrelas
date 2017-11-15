package main

import (
	"fmt"
	"net/url"
)

func main() {

}

func buildQuery(pageName string) string {
	v := url.Values{}
	v.Set("action", "query")
	v.Add("titles", pageName)
	v.Add("prop", "revisions")
	v.Add("rvprop", "content")
	v.Add("format", "json")

	return fmt.Sprintf("/w/api.php?%s", v.Encode())
}
