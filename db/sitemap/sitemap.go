package sitemap

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/jeffail/tunny"
	"github.com/jinzhu/gorm"
	"github.com/suicidegang/spider-srv/db/url"
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

func (sitemap Sitemap) Create(db *gorm.DB, r *redis.Client, pool *tunny.WorkPool) (Sitemap, error) {

	log.Printf("[sg.micro.srv.spider] Sitemap::create")
	sitemap.Updating = true

	if err := db.Create(&sitemap).Error; err != nil {
		return sitemap, err
	}

	urls := map[string]int{}
	patterns := map[string]*regexp.Regexp{}

	var groups map[string]string
	var scraper func(string, uint64) func()

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

	scraper = func(urlStr string, depth uint64) func() {
		return func() {
			// Iterate over patterns to see if any of them matches the url
			for group, pattern := range patterns {
				if pattern.MatchString(urlStr) {
					ourl, err := url.Prepare(db, r, urlStr, group, sitemap.ID)
					if err != nil {
						log.Printf("[err] %v/%v: %v", urlStr, depth, err)
						return
					}

					hash := ourl.Hash()
					if _, exists := urls[hash]; exists {
						log.Printf("[skip] %v:%v", hash, ourl.FullURL())
						return
					}

					// Keep map of hashes for further O(1) checks
					urls[hash] = 1
					log.Printf("[url] %+v", ourl)
					if depth+1 > sitemap.Depth {
						return
					}

					// Retrieve URL's queriable document
					doc, err := ourl.Document(r)
					if err != nil {
						log.Printf("[err] %v/%v: %v", urlStr, depth, err)
						return
					}

					doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
						link, exists := s.Attr("href")

						if exists && len(link) > 1 {

							// Relative path should be prefixed with base url
							if link[0:1] == "/" {
								link = url.FixRelative(link, urlStr)
							}

							pool.SendWorkAsync(scraper(link, depth+1), func(data interface{}, err error) {})
						}
					})

					break
				}
			}
		}
	}

	pool.SendWorkAsync(scraper(sitemap.EntryUrl, 0), func(data interface{}, err error) {})

	return sitemap, nil
}
