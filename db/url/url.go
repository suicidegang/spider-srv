package url

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/gorm"
	"gopkg.in/redis.v5"

	"database/sql"
	"encoding/json"
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
	Group       string

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

func (u Url) Hash() string {
	return HashMD5(u.FullURL())
}

func (u Url) Document(r *redis.Client) (doc *goquery.Document, err error) {
	doc, err = Document(r, u.FullURL())
	return
}

type Urls []Url

func (u Urls) Each(fn func(u Url)) {
	if len(u) > 0 {
		for _, one := range u {
			fn(one)
		}
	}
}

func All(db *gorm.DB) Urls {
	var list Urls

	db.Find(&list)

	return list
}

func FindBy(db *gorm.DB, conditions []string) (Url, error) {
	var ourl Url

	if len(conditions) < 1 {
		return ourl, errors.New("Not enough conditions")
	}

	statement, binds := conditions[0], conditions[1:]
	bind := make([]interface{}, len(binds))
	for i, v := range binds {
		bind[i] = v
	}

	if db.Where(statement, bind...).First(&ourl).RecordNotFound() {
		return ourl, errors.New("Couldnt find any by those conditions.")
	}

	return ourl, nil
}

func FindFullText(db *gorm.DB, search string) Urls {
	query := db.Table("spider_urls_index uix")
	query = query.Select("u.*, ts_rank(document, plainto_tsquery('es', ?)) AS score", search)
	query = query.Joins("inner join spider_urls u ON u.id = uix.id")
	query = query.Where("document @@ plainto_tsquery('es', ?)", search)
	query = query.Order("score DESC")

	var urls Urls
	query.Scan(&urls)

	return urls
}

func Prepare(db *gorm.DB, r *redis.Client, urlStr, group string, meta map[string]string, sitemapID uint) (Url, error) {
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

		data, err := json.Marshal(meta)
		if err != nil {
			return Url{}, err
		}

		ourl = Url{
			SitemapID:   sql.NullInt64{int64(sitemapID), vsitemap},
			Url:         simpleUrl,
			QueryParams: queryParams,
			Title:       title,
			Description: description,
			Meta:        string(data),
			Group:       group,
		}

		db.Create(&ourl)

		return ourl, nil
	}

	return ourl, nil
}

func One(db *gorm.DB, id uint64) (Url, error) {
	ourl := Url{}

	if db.First(&ourl, id).RecordNotFound() {
		return ourl, errors.New("URL not found.")
	}

	return ourl, nil
}

func FixRelative(relative, absolute string) string {
	u, err := url.Parse(absolute)
	if err != nil {
		log.Printf("[err] Url::prepare could not parse url: %v", err.Error())
		return relative
	}

	return u.Scheme + "://" + u.Host + relative
}
