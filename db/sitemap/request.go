package sitemap

import (
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/gorm"
	"github.com/suicidegang/spider-srv/db/url"
	"gopkg.in/redis.v5"

	"log"
	urls "net/url"
	"regexp"
	"strings"
)

// Request represents an URL that must
type Request struct {
	URL          string
	Entry        string
	Depth        int
	Patterns     map[string]*regexp.Regexp
	SitemapID    uint
	FinalDepth   uint64
	UniqueParams bool
	Strict       bool
	DB           *gorm.DB
	R            *redis.Client
	Done         *sync.Map
	Retries      *sync.Map
}

type enqueue func(job Request)

// Stuff to be done when worker get to this job.
func (req Request) Work(next enqueue) {

	// Iterate over patterns to see if any of them matches the url & process it.
	for group, pattern := range req.Patterns {
		if pattern.MatchString(req.URL) {
			params := pattern.SubexpNames()
			meta := map[string]string{}

			if len(params) > 1 {
				values := pattern.FindStringSubmatch(req.URL)
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
	if req.Strict == false && strings.HasPrefix(req.URL, req.Entry) {
		req.ProcessPageURL("site", map[string]string{}, next)
	}
}

// ProcessPageURL from current request.
func (req Request) ProcessPageURL(group string, meta map[string]string, next enqueue) {
	ourl, err := url.Prepare(req.DB, req.R, req.URL, group, meta, uint(req.SitemapID))
	if err != nil {
		req.ThrowError(err)
		return
	}

	// Keeps a map of hashes for further O(1) checks
	req.Done.Store(req.Hash(), 1)

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
			if link[0:1] == "/" || (len(link) > 4 && link[0:4] != "http") {
				link = url.FixRelative(link, req.URL)
			}

			lurl, err := urls.Parse(link)
			if err != nil {
				log.Printf("Invalid link %s", link)
				return
			}

			if !req.UniqueParams {
				lurl.RawQuery = ""
			}

			if _, exists := req.Done.Load(url.HashMD5(lurl.String())); exists {
				return
			}

			// Keep it kind of protected from other routines...
			req.Done.Store(url.HashMD5(lurl.String()), 1)

			w := Request{
				URL:        lurl.String(),
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

func (req Request) Hash() string {
	return url.HashMD5(req.URL)
}

func (req Request) ThrowError(err error) {
	n := 0
	h := req.Hash()
	if ns, exists := req.Retries.Load(h); exists {
		n = ns.(int)
	}

	if n >= 0 {
		log.Printf("[err] %v/%v: %v", req.URL, req.Depth, err)

		// Use sitemap table to avoid new retries
		req.Done.Store(h, 1)
		return
	}

	req.Retries.Store(h, n+1)
}
