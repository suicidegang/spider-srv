package dataset

import (
	"github.com/jinzhu/gorm"
	"github.com/suicidegang/spider-srv/db/selector"
	"github.com/suicidegang/spider-srv/db/url"
	"gopkg.in/redis.v5"

	"encoding/json"
	"errors"
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
	Hits       uint
}

func (dataset Dataset) Document() (map[string]string, error) {
	var d map[string]string

	err := json.Unmarshal([]byte(dataset.Data), &d)
	if err != nil {
		return d, err
	}

	return d, nil
}

func (dataset Dataset) Invalidate(db *gorm.DB, r *redis.Client) (Dataset, error) {
	var (
		u url.Url
		s selector.Selector
	)

	if db.Model(&dataset).Related(&u).RecordNotFound() {
		return dataset, errors.New("Dataset url not found.")
	}

	if db.Model(&dataset).Related(&s).RecordNotFound() {
		return dataset, errors.New("Selector not found.")
	}

	doc, err := u.Document(r)
	if err != nil {
		return dataset, err
	}

	dset, err := s.Query(doc)
	if err != nil {
		return dataset, err
	}

	data, err := json.Marshal(dset)
	if err != nil {
		return dataset, err
	}

	ds := Dataset{
		SelectorID: dataset.SelectorID,
		UrlID:      dataset.UrlID,
		Hash:       HashMapMD5(dset),
		Data:       string(data),
		Status:     HEALTHY_STATUS,
		Revision:   dataset.Revision + 1,
		Hits:       0,
	}

	db.Create(&ds)
	return ds, nil
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
			Hits:       1,
		}

		db.Create(&ds)
		return ds, nil
	}

	if s.UpdatedAt.After(ds.UpdatedAt) || u.UpdatedAt.After(ds.UpdatedAt) {
		ds, err = ds.Invalidate(db, r)

		if err != nil {
			return ds, err
		}
	}

	db.Model(&ds).Update("hits", gorm.Expr("hits + 1"))

	return ds, nil
}
