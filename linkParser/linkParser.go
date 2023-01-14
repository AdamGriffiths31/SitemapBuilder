package linkParser

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

type Link struct {
	Href string
	Text string
}

func Parse(r io.Reader) ([]Link, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	nodes := linkNodes(doc)
	var links []Link

	for _, node := range nodes {
		links = append(links, buildLink(node))
	}
	return links, nil
}

func buildLink(n *html.Node) Link {
	var res Link
	for _, atrr := range n.Attr {
		if atrr.Key == "href" {
			res.Href = atrr.Val
			break
		}
	}

	res.Text = text(n)
	return res
}

func text(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	if n.Type != html.ElementNode {
		return ""
	}

	var res string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res += text(c) //TODO Byte buffer
	}

	return strings.Join(strings.Fields(res), " ")
}

func linkNodes(n *html.Node) []*html.Node {
	if n.Type == html.ElementNode && n.Data == "a" {
		return []*html.Node{n}
	}

	var res []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res = append(res, linkNodes(c)...)
	}

	return res
}
