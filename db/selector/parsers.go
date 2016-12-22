package selector

import (
	"github.com/PuerkitoBio/goquery"

	"strings"
)

const TEXT_PARSER string = "text"
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
