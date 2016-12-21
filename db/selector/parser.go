package selector

import (
	"github.com/PuerkitoBio/goquery"

	"encoding/json"
	"errors"
)

type Parser interface {
	Query(*goquery.Document) (interface{}, error)
}

type ParsersMap map[string]Parser

func (pm *ParsersMap) UnmarshalJSON(data []byte) error {
	fields := make(map[string]json.RawMessage)
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return err
	}

	pfields := make(ParsersMap)

	// Iter over fields to determine parsers
	for k, v := range fields {
		parser := map[string]interface{}{}
		err := json.Unmarshal(v, &parser)
		if err != nil {
			return err
		}

		if name, exists := parser["use"].(string); exists {
			switch name {
			case TEXT_PARSER:
				t := TextParser{}
				err := json.Unmarshal(v, &t)
				if err != nil {
					return err
				}
				pfields[k] = t
			default:
				return errors.New("Unrecognized parser")
			}
		}
	}
	*pm = pfields
	return nil
}
