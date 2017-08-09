package sitemap

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/gorm"
	"github.com/suicidegang/spider-srv/db/url"
	"gopkg.in/redis.v5"

	"log"
	urls "net/url"
	"regexp"
	"strings"
)

// Sitemap request represents an URL that must
type SitemapRequest struct {
	Url          string
	Entry        string
	Depth        int
	Patterns     map[string]*regexp.Regexp
	SitemapID    uint
	FinalDepth   uint64
	UniqueParams bool
	Strict       bool
	DB           *gorm.DB
	R            *redis.Client
}

type enqueue func(job SitemapRequest)

// Stuff to be done when worker get to this job.
func (req SitemapRequest) Work(next enqueue) {

	// Iterate over patterns to see if any of them matches the url & process it.
	for group, pattern := range req.Patterns {
		if pattern.MatchString(req.Url) {
			params := pattern.SubexpNames()
			meta := map[string]string{}

			if len(params) > 1 {
				values := pattern.FindStringSubmatch(req.Url)
				for i, name := range params {
					if i > 0 {
						meta[name] = values[i]
					}
				}
			}

			req.ProcessPageURL(group, meta, next)
			return
		}
	}

	// Uncategorized urls that match entry point must be processed as site pages.
	if req.Strict == false && strings.HasPrefix(req.Url, req.Entry) {
		req.ProcessPageURL("site", map[string]string{}, next)
	}
}

// Process page
func (req SitemapRequest) ProcessPageURL(group string, meta map[string]string, next enqueue) {

	ourl, err := url.Prepare(req.DB, req.R, req.Url, group, meta, uint(req.SitemapID))
	if err != nil {
		req.ThrowError(err)
		return
	}

	// Keeps a map of hashes for further O(1) checks
	SitemapTable.Set(req.Hash(), 1)

	// Retrieve URL's queriable document
	doc, err := ourl.Document(req.R)
	if err != nil {
		req.ThrowError(err)
		return
	}

	if uint64(req.Depth+1) > req.FinalDepth {
		return
	}

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		if link, exists := s.Attr("href"); exists && len(link) > 1 {

			// Relative path should be prefixed with base url
			if link[0:1] == "/" {
				link = url.FixRelative(link, req.Url)
			}

			lurl, err := urls.Parse(link)
			if err != nil {
				log.Printf("Invalid link %s", link)
				return
			}

			if !req.UniqueParams {
				lurl.RawQuery = ""
			}

			if SitemapTable.Has(url.HashMD5(lurl.String())) {
				return
			}

			// Keep it kind of protected from other routines...
			SitemapTable.Set(url.HashMD5(lurl.String()), 1)

			w := SitemapRequest{
				Url:        lurl.String(),
				Entry:      req.Entry,
				Depth:      req.Depth + 1,
				Patterns:   req.Patterns,
				SitemapID:  req.SitemapID,
				FinalDepth: req.FinalDepth,
				Strict:     req.Strict,
				DB:         req.DB, R: req.R,
			}

			for _, pattern := range req.Patterns {
				if pattern.MatchString(lurl.String()) {
					// Send the request to the queue
					next(w)
					return
				}
			}

			// Uncategorized urls that match entry point must be crawled
			if !req.Strict && strings.HasPrefix(lurl.String(), req.Entry) {

				// Send the request to the queue
				next(w)
			}
		}
	})
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
