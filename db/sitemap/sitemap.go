package sitemap

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/jeffail/tunny"
	"github.com/jinzhu/gorm"
	"github.com/suicidegang/spider-srv/db/url"
	"gopkg.in/redis.v5"

	"log"
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
	var scraper func(string, uint64) func()

	scraper = func(urlStr string, depth uint64) func() {
		return func() {
			ourl, err := url.Prepare(db, r, urlStr, sitemap.ID)
			if err != nil {
				log.Printf("[err] %v/%v: %v", urlStr, depth, err)
				return
			}

			hash := ourl.Hash()

			if _, exists := urls[hash]; exists {
				log.Printf("[skip] %v:%v", hash, ourl.FullURL())
				return
			}

			urls[hash] = 1
			log.Printf("[url] %+v", ourl)

			if depth+1 > sitemap.Depth {
				return
			}

			doc, err := ourl.Document(r)
			if err != nil {
				log.Printf("[err] %v/%v: %v", urlStr, depth, err)
				return
			}

			doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
				link, exists := s.Attr("href")

				if exists && len(link) > 0 {
					pool.SendWorkAsync(scraper(link, depth+1), func(data interface{}, err error) {})
				}
			})
		}
	}

	pool.SendWorkAsync(scraper(sitemap.EntryUrl, 0), func(data interface{}, err error) {})

	return sitemap, nil
}
