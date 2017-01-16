package sitemap

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/gorm"
	"github.com/suicidegang/spider-srv/db/url"
	"gopkg.in/redis.v5"

	"log"
	urls "net/url"
	"regexp"
)

type SitemapRequest struct {
	Url        string
	Depth      int
	Patterns   map[string]*regexp.Regexp
	SitemapID  uint
	FinalDepth uint64
	DB         *gorm.DB
	R          *redis.Client
}

func (req SitemapRequest) Work() {

	// First check to avoid dup ops within tree
	if SitemapTable.Has(url.HashMD5(req.Url)) {
		return
	}

	// Iterate over patterns to see if any of them matches the url
	for group, pattern := range req.Patterns {
		if pattern.MatchString(req.Url) {
			ourl, err := url.Prepare(req.DB, req.R, req.Url, group, uint(req.SitemapID))
			if err != nil {
				req.ThrowError(err)
				return
			}

			hash := ourl.Hash()
			if SitemapTable.Has(hash) {
				log.Printf("[skip] %v", hash)
				return
			}

			// Retrieve URL's queriable document
			doc, err := ourl.Document(req.R)
			if err != nil {
				req.ThrowError(err)
				return
			}

			// Keep map of hashes for further O(1) checks
			SitemapTable.Set(hash, 1)
			log.Printf("[url] %+v", ourl)
			if uint64(req.Depth+1) > req.FinalDepth {
				return
			}

			doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
				link, exists := s.Attr("href")

				if exists && len(link) > 1 {
					if _, err := urls.Parse(link); err != nil {
						return
					}

					// Relative path should be prefixed with base url
					if link[0:1] == "/" {
						link = url.FixRelative(link, req.Url)
					}

					if SitemapTable.Has(url.HashMD5(link)) {
						return
					}

					for _, pattern := range req.Patterns {
						if pattern.MatchString(link) {

							w := SitemapRequest{Url: link, Depth: req.Depth + 1, Patterns: req.Patterns, SitemapID: req.SitemapID, FinalDepth: req.FinalDepth, DB: req.DB, R: req.R}

							// Send the request to the queue
							Queue <- w
						}
					}

				}
			})

			break
		}
	}
}

func (req SitemapRequest) Hash() string {
	return url.HashMD5(req.Url)
}

func (req SitemapRequest) ThrowError(err error) {
	n := 0
	h := req.Hash()

	if SitemapRetries.Has(h) {
		ns, _ := SitemapRetries.Get(h)
		n = ns.(int)
	}

	if n >= 0 {
		log.Printf("[err] %v/%v: %v", req.Url, req.Depth, err)

		// Use sitemap table to avoid new retries
		SitemapTable.Set(h, 1)
		return
	}

	SitemapRetries.Set(h, n+1)
}
