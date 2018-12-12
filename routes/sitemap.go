package routes

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/suicidegang/spider-srv/db"
	"github.com/suicidegang/spider-srv/db/sitemap"
	"github.com/suicidegang/spider-srv/db/url"
	"gopkg.in/ikeikeikeike/go-sitemap-generator.v1/stm"
)

// TrackSitemap : POST /sitemap
func TrackSitemap(c echo.Context) error {
	var site sitemap.Sitemap
	if err := c.Bind(&site); err != nil {
		return err
	}

	site.Patterns = `[{"name": "page", "matches": "^"}]`
	prf, err := site.Create(db.Db, db.Redis)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, prf)
}

// Sitemap info.
func Sitemap(c echo.Context) error {
	sitemap, err := sitemap.FindByRef(db.Db, c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, sitemap)
}

func SitemapURLs(c echo.Context) error {
	urls, err := url.FindBySitemap(db.Db, c.Param("id"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, urls)
}

func UpdateURL(c echo.Context) error {
	var (
		err   error
		model url.Url
	)
	if err = c.Bind(&model); err != nil {
		return err
	}
	url, err := url.FindByRef(db.Db, model.Ref)
	if err != nil {
		return err
	}

	err = db.Db.Model(&url).Update(map[string]interface{}{
		"Group":           model.Group,
		"Meta":            model.Meta,
		"Priority":        model.Priority,
		"ChangeFrequency": model.ChangeFrequency,
		"Enabled":         model.Enabled,
	}).Error
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, url)
}

// SitemapFile generator.
func SitemapFile(c echo.Context) error {
	sitemap, err := sitemap.FindByRef(db.Db, c.Param("id"))
	if err != nil {
		return err
	}

	urls, err := url.FindBySitemap(db.Db, sitemap.Ref)
	if err != nil {
		return err
	}

	entryLen := len(sitemap.EntryURL)
	sm := stm.NewSitemap()
	sm.Create()
	sm.SetDefaultHost(sitemap.EntryURL)
	for _, url := range urls {
		if url.Enabled == false {
			continue
		}
		loc := url.FullURL()
		loc = loc[entryLen:]
		sm.Add(stm.URL{"loc": loc, "changefreq": url.ChangeFrequency, "priority": url.Priority})
	}
	return c.Blob(http.StatusOK, "application/xml; charset=UTF-8", sm.XMLContent())
}
