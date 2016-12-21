package selector

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/gorm"

	"encoding/json"
	"errors"
)

type Selector struct {
	gorm.Model
	Name     string
	Patterns string `sql:"type:JSONB NOT NULL DEFAULT '{}'::JSONB"`
	Active   bool
}

func (selector Selector) List() (map[string]Parser, error) {
	var fields ParsersMap

	if err := json.Unmarshal([]byte(selector.Patterns), &fields); err != nil {
		return fields, err
	}

	return fields, nil
}

func (selector Selector) Query(doc *goquery.Document) (map[string]string, error) {
	data := map[string]string{}

	fields, err := selector.List()
	if err != nil {
		return data, err
	}

	for name, parser := range fields {
		v, err := parser.Query(doc)
		if err != nil {
			return data, err
		}

		data[name] = v.(string)
	}

	return data, nil
}

func One(db *gorm.DB, id uint64) (Selector, error) {
	var selector Selector

	if db.First(&selector, id).RecordNotFound() {
		return selector, errors.New("Selector not found.")
	}

	return selector, nil
}
