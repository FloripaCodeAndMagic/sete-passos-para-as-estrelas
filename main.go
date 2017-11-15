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
	v.Add("prop", "links")
	v.Add("pllimit", "max")
	v.Add("format", "json")

	return fmt.Sprintf("/w/api.php?%s", v.Encode())
}
