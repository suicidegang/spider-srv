package dataset

import (
	"github.com/jinzhu/gorm"
	"github.com/suicidegang/spider-srv/db/selector"
	"github.com/suicidegang/spider-srv/db/url"
	"gopkg.in/redis.v5"

	"encoding/json"
)

const HEALTHY_STATUS = "healthy"
const DIRTY_STATUS = "dirty"

type Dataset struct {
	gorm.Model

	Selector selector.Selector
	Url      url.Url

	SelectorID uint
	UrlID      uint
	Hash       string
	Status     string
	Data       string `sql:"type:JSONB NOT NULL DEFAULT '{}'::JSONB"`
	Revision   uint
}

func (dataset Dataset) Document() (map[string]string, error) {
	var d map[string]string

	err := json.Unmarshal([]byte(dataset.Data), &d)
	if err != nil {
		return d, err
	}

	return d, nil
}

func Prepare(db *gorm.DB, r *redis.Client, selectorID, urlID uint) (Dataset, error) {
	var ds Dataset

	u, err := url.One(db, uint64(urlID))
	if err != nil {
		return ds, err
	}

	s, err := selector.One(db, uint64(selectorID))
	if err != nil {
		return ds, err
	}

	if db.Where("selector_id = ? AND url_id = ?", selectorID, urlID).Order("revision desc").First(&ds).RecordNotFound() {
		doc, err := u.Document(r)
		if err != nil {
			return ds, err
		}

		dataset, err := s.Query(doc)
		if err != nil {
			return ds, err
		}

		data, err := json.Marshal(dataset)
		if err != nil {
			return ds, err
		}

		ds = Dataset{
			SelectorID: selectorID,
			UrlID:      urlID,
			Hash:       HashMapMD5(dataset),
			Data:       string(data),
			Status:     HEALTHY_STATUS,
			Revision:   1,
		}

		db.Create(&ds)
		return ds, nil
	}

	return ds, nil
}
