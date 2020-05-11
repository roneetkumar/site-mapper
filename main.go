package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	link "github.com/roneetkumar/html-link-parser"
)

const xmlns = "https://www.sitemaps.org/schemas/sitemap/0.9"

type loc struct {
	Value string `xml:"loc"`
}

type urlset struct {
	URLs  []loc  `xml:"url"`
	XMLns string `xml:"xmlns,attr"`
}

func main() {

	urlFlag := flag.String("url", "https://roneetkumar.github.io", "the url that you want to buld site map for")
	maxDepth := flag.Int("depth", 3, "the maximum number of links deep to traverse")

	flag.Parse()

	pages := bfs(*urlFlag, *maxDepth)

	toXML := urlset{
		XMLns: xmlns,
	}

	for _, page := range pages {
		toXML.URLs = append(toXML.URLs, loc{page})
	}

	fmt.Print(xml.Header)
	encoder := xml.NewEncoder(os.Stdout)
	encoder.Indent("", "	")
	if err := encoder.Encode(toXML); err != nil {
		panic(err)
	}
	fmt.Println()
}

func bfs(urlStr string, maxDepth int) []string {

	seen := make(map[string]struct{})

	var q map[string]struct{}

	nq := map[string]struct{}{
		urlStr: struct{}{},
	}

	for i := 0; i <= maxDepth; i++ {
		q, nq = nq, make(map[string]struct{})

		if len(q) == 0 {
			break
		}
		for url := range q {
			if _, ok := seen[url]; ok {
				continue
			}
			seen[url] = struct{}{}

			for _, link := range get(url) {
				if _, ok := seen[link]; !ok {
					nq[link] = struct{}{}
				}
			}
		}
	}

	ret := make([]string, 0, len(seen))

	for url := range seen {
		ret = append(ret, url)
	}

	return ret
}

func filter(links []string, keepFn func(string) bool) []string {

	var ret []string

	for _, link := range links {
		if keepFn(link) {
			ret = append(ret, link)
		}
	}

	return ret
}

func withPrefix(pfx string) func(string) bool {
	return func(link string) bool {
		return strings.HasPrefix(link, pfx)
	}
}

func get(urlStr string) []string {

	res, err := http.Get(urlStr)
	if err != nil {
		return []string{}
	}
	defer res.Body.Close()

	reqURL := res.Request.URL

	baseURL := &url.URL{
		Scheme: reqURL.Scheme,
		Host:   reqURL.Host,
	}

	base := baseURL.String()

	return filter(makeHREFs(res.Body, base), withPrefix(base))
}

func makeHREFs(r io.Reader, base string) []string {
	links, _ := link.Parse(r)
	var hrefs []string

	for _, l := range links {
		switch {
		case strings.HasPrefix(l.Href, "/"):
			hrefs = append(hrefs, base+l.Href)
		case strings.HasPrefix(l.Href, "http"):
			hrefs = append(hrefs, l.Href)
		}
	}
	return hrefs
}
