package selector

import (
	"github.com/PuerkitoBio/goquery"
)

const TEXT_PARSER string = "text"
const INNER_PARSER string = "innerHTML"
const ATTR_PARSER string = "attr"

type TextParser struct {
	Target string `json:"value"`
	Index  int    `json:"i"`
}

func (text TextParser) Query(doc *goquery.Document) (interface{}, error) {

	n := doc.Find(text.Target).Eq(text.Index).Text()

	return n, nil
}
