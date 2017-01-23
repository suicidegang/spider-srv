package selector

import (
	"github.com/PuerkitoBio/goquery"

	"strings"
)

const TEXT_PARSER string = "text"
const PIPE_PARSER string = "pipe"
const INNER_PARSER string = "innerHTML"
const ATTR_PARSER string = "attr"

type TextParser struct {
	Target  string   `json:"value"`
	Index   int      `json:"i"`
	Filters []string `json:"filters"`
}

func (text TextParser) Query(doc *goquery.Document) (interface{}, error) {
	n := doc.Find(text.Target).Eq(text.Index).Text()

	for _, filter := range text.Filters {
		switch filter {
		case "trim-space":
			n = strings.TrimSpace(n)
		}
	}

	return n, nil
}

type AttrParser struct {
	Target  string   `json:"value"`
	Attr    string   `json:"attr"`
	Index   int      `json:"i"`
	Filters []string `json:"filters"`
}

func (attr AttrParser) Query(doc *goquery.Document) (interface{}, error) {
	n := doc.Find(attr.Target).Eq(attr.Index).AttrOr(attr.Attr, "")

	for _, filter := range attr.Filters {
		switch filter {
		case "trim-space":
			n = strings.TrimSpace(n)
		}
	}

	return n, nil
}

type PipeParser struct {
	Target string `json:"value"`
	Index  int    `json:"i"`
	Nested int    `json:"then"`
}

func (pipe PipeParser) Query(doc *goquery.Document) (interface{}, error) {
	return nil, nil
}
