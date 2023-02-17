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

	"github.com/SitemapBuilder/linkParser"
)

type loc struct {
	Value string `xml:"loc"`
}

type urlSet struct {
	Urls []loc `xml:"url"`
}

func main() {
	urlFlag := flag.String("url", "https://www.bbc.co.uk/", "URL for building sitemap")
	maxDepth := flag.Int("depth", 1, "max depth of sitemap")
	flag.Parse()

	pages := bfs(*urlFlag, *maxDepth)

	var toXml urlSet
	for _, page := range pages {
		toXml.Urls = append(toXml.Urls, loc{page})
	}

	fmt.Printf(xml.Header)
	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("", " ")
	if err := enc.Encode(toXml); err != nil {
		fmt.Printf("Problem with NewEncoder %v", err)
	}

}

func bfs(urlString string, maxDepth int) []string {
	seen := make(map[string]struct{})
	var q map[string]struct{}
	nq := map[string]struct{}{
		urlString: {},
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
			for _, l := range get(url) {
				if _, ok := seen[l]; !ok {
					nq[l] = struct{}{}
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

func get(urlString string) []string {
	resp, err := http.Get(urlString)
	if err != nil {
		fmt.Printf("Problem with GET %v - %v", urlString, err)
	}

	defer resp.Body.Close()

	requestUrl := resp.Request.URL
	baseUrl := &url.URL{
		Scheme: requestUrl.Scheme,
		Host:   requestUrl.Host,
	}
	base := baseUrl.String()

	return filter(base, getHrefs(resp.Body, base))
}

func getHrefs(r io.Reader, base string) []string {
	links, err := linkParser.Parse(r)
	if err != nil {
		fmt.Printf("Problem with Parse")
	}

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

func filter(base string, links []string) []string {
	var ret []string
	for _, l := range links {
		if strings.HasPrefix(l, base) {
			ret = append(ret, l)
		}
	}

	return ret
}
