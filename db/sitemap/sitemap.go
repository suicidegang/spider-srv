package sitemap

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/redis.v5"

	"encoding/json"
	"errors"
	"log"
	"regexp"
	"strings"
)

type Sitemap struct {
	gorm.Model
	Name     string
	EntryUrl string
	Depth    uint64
	Patterns string `sql:"type:JSONB NOT NULL DEFAULT '{}'::JSONB"`
	Updating bool
	Strict   bool
}

type Pattern struct {
	Name    string `json:"name"`
	Matches string `json:"matches"`
}

var InvalidPatternErr = errors.New("Could not compile URL pattern into regexp.")

var slugBind = regexp.MustCompile(`{([[:alnum:]|_]+):slug}`)
var numBind = regexp.MustCompile(`{([[:alnum:]]+):num}`)

func bslug(input string) string {
	return slugBind.ReplaceAllString(input, `(?P<$1>[[:alnum:]-_]+)`)
}

func bnum(input string) string {
	return numBind.ReplaceAllString(input, `(?P<$1>[[:digit:]]+)`)
}

func (sitemap Sitemap) Create(db *gorm.DB, r *redis.Client) (Sitemap, error) {

	log.Printf("[sg.micro.srv.spider] Sitemap::create")
	sitemap.Updating = true

	if err := db.Create(&sitemap).Error; err != nil {
		return sitemap, err
	}

	var groups []Pattern
	patterns := map[string]*regexp.Regexp{}

	if err := json.Unmarshal([]byte(sitemap.Patterns), &groups); err != nil {
		return sitemap, err
	}

	regexer := strings.NewReplacer(".", "\\.", "/", "\\/", "?", "\\?", "^", "^"+sitemap.EntryUrl)

	// Create regex patterns from patterns with bindings
	for _, pattern := range groups {
		reg := regexer.Replace(pattern.Matches)
		reg = bslug(reg)
		reg = bnum(reg)
		log.Printf("%s using %s", pattern.Name, reg)
		r, err := regexp.Compile(reg)

		if err != nil {
			return sitemap, InvalidPatternErr
		}

		patterns[pattern.Name] = r
	}

	w := SitemapRequest{Url: sitemap.EntryUrl, Strict: sitemap.Strict, Entry: sitemap.EntryUrl, UniqueParams: false, Depth: 0, Patterns: patterns, SitemapID: sitemap.ID, FinalDepth: sitemap.Depth, DB: db, R: r}

	// Send the request to the queue
	Queue <- w

	return sitemap, nil
}
