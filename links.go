package grawler

import (
	"io"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func tokenIsAnchor(t *html.Token) (string, bool) {
	return "href", t.DataAtom == atom.A
}

func tokenIsScript(t *html.Token) (string, bool) {
	return "src", t.DataAtom == atom.Script
}

func tokenIsLink(t *html.Token) (string, bool) {
	return "href", t.DataAtom == atom.Link
}

func tokenIsImg(t *html.Token) (string, bool) {
	return "src", t.DataAtom == atom.Img
}

func isLink(t *html.Token) (string, bool) {
	return tokenIsAnchor(t)
}

func isAsset(t *html.Token) (string, bool) {
	if tag, ok := tokenIsScript(t); ok {
		return tag, ok
	} else if tag, ok := tokenIsLink(t); ok {
		return tag, ok
	} else if tag, ok := tokenIsImg(t); ok {
		return tag, ok
	}
	return "", false
}

func findAttr(name string, token *html.Token) *html.Attribute {
	for _, attr := range token.Attr {
		if attr.Key == name {
			return &attr
		}
	}
	return nil
}

func findLinksAndAssets(rdr io.Reader) ([]string, []string) {
	links := []string{}
	assets := []string{}

	doc := html.NewTokenizer(rdr)

	for tokenType := doc.Next(); tokenType != html.ErrorToken; tokenType = doc.Next() {
		switch tokenType {
		case html.StartTagToken:
			fallthrough
		case html.SelfClosingTagToken:
			token := doc.Token()
			// TODO XXX
			// Which HTML elements can have href=? and/or src=?
			// https://www.w3.org/TR/REC-html40/index/attributes.html
			if tag, ok := isLink(&token); ok {
				if attr := findAttr(tag, &token); attr != nil {
					links = append(links, attr.Val)
				}
			}
			if tag, ok := isAsset(&token); ok {
				if attr := findAttr(tag, &token); attr != nil {
					assets = append(assets, attr.Val)
				}
			}
		}
	}

	return links, assets
}
