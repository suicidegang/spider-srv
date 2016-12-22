package sitemap

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/redis.v5"

	"encoding/json"
	"log"
	"regexp"
)

type Sitemap struct {
	gorm.Model
	Name     string `gorm:"unique_index"`
	EntryUrl string
	Depth    uint64
	Patterns string `sql:"type:JSONB NOT NULL DEFAULT '{}'::JSONB"`
	Updating bool
}

func (sitemap Sitemap) Create(db *gorm.DB, r *redis.Client) (Sitemap, error) {

	log.Printf("[sg.micro.srv.spider] Sitemap::create")
	sitemap.Updating = true

	if err := db.Create(&sitemap).Error; err != nil {
		return sitemap, err
	}

	var groups map[string]string
	patterns := map[string]*regexp.Regexp{}

	if err := json.Unmarshal([]byte(sitemap.Patterns), &groups); err != nil {
		return sitemap, err
	}

	// Compile regexp patterns received as strings
	for group, pattern := range groups {
		r, err := regexp.Compile(pattern)

		if err != nil {
			return sitemap, err
		}

		patterns[group] = r
	}

	w := SitemapRequest{Url: sitemap.EntryUrl, Depth: 0, Patterns: patterns, SitemapID: sitemap.ID, FinalDepth: sitemap.Depth, DB: db, R: r}

	// Send the request to the queue
	Queue <- w

	return sitemap, nil
}
