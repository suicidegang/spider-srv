package url

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/gorm"
	"gopkg.in/redis.v5"

	"database/sql"
	"errors"
	"log"
	"net/url"
	"time"
)

type Url struct {
	gorm.Model

	SitemapID sql.NullInt64

	Url         string
	QueryParams string
	Title       string
	Description string `gorm:"type:text"`
	Meta        string `sql:"type:JSONB NOT NULL DEFAULT '{}'::JSONB"`

	LastModified *time.Time `gorm:"null"`
	Expires      *time.Time `gorm:"null"`
}

func (u Url) FullURL() string {
	url := u.Url

	if len(u.QueryParams) > 0 {
		url = url + "?" + u.QueryParams
	}

	return url
}

func (u Url) Document(r *redis.Client) (doc *goquery.Document, err error) {
	doc, err = Document(r, u.FullURL())
	return
}

func Prepare(db *gorm.DB, r *redis.Client, urlStr string, sitemapID uint) (Url, error) {
	if len(urlStr) < 8 {
		log.Printf("[err] Url::prepare could not parse url: %v", urlStr)
		return Url{}, errors.New("Invalid url")
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("[err] Url::prepare could not parse url: %v", err.Error())
		return Url{}, err
	}

	if !u.IsAbs() {
		log.Printf("[err] Url::prepare did not get absolute url: %v", urlStr)
		return Url{}, errors.New("Not an absolute URL")
	}

	var ourl Url

	simpleUrl := u.Scheme + "://" + u.Host + u.EscapedPath()
	queryParams := u.RawQuery

	if db.Where("url = ? AND query_params = ?", simpleUrl, queryParams).First(&ourl).RecordNotFound() {
		doc, err := Document(r, urlStr)
		if err != nil {
			return Url{}, err
		}

		vsitemap := sitemapID != 0
		title := doc.Find("title").Text()
		description := doc.Find("meta[name='description']").AttrOr("content", "")

		ourl = Url{
			SitemapID:   sql.NullInt64{int64(sitemapID), vsitemap},
			Url:         simpleUrl,
			QueryParams: queryParams,
			Title:       title,
			Description: description,
			Meta:        "{}",
		}

		db.Create(&ourl)

		return ourl, nil
	}

	return ourl, nil
}
