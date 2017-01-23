package handler

import (
	"github.com/micro/go-micro/errors"
	"github.com/suicidegang/spider-srv/db"
	"github.com/suicidegang/spider-srv/db/sitemap"
	"github.com/suicidegang/spider-srv/db/url"
	proto "github.com/suicidegang/spider-srv/proto/spider"
	"golang.org/x/net/context"

	"encoding/json"
	"log"
)

func (srv *Spider) TrackSitemap(ctx context.Context, req *proto.TrackSitemapRequest, res *proto.TrackSitemapResponse) error {

	log.Printf("Spider::trackSitemap %+v", req)

	patterns, err := json.Marshal(req.Patterns)
	if err != nil {
		return errors.InternalServerError("sg.micro.srv.spider.TrackSitemap", err.Error())
	}

	smap := sitemap.Sitemap{
		Name:     req.Name,
		EntryUrl: req.From,
		Depth:    req.Depth,
		Patterns: string(patterns),
	}

	smap, err = smap.Create(db.Db, db.Redis)
	if err != nil {
		return errors.InternalServerError("sg.micro.srv.spider.TrackSitemap", err.Error())
	}

	res.Id = uint64(smap.ID)
	return nil
}

func (src *Spider) TrackURL(ctx context.Context, req *proto.TrackURLRequest, res *proto.TrackURLResponse) error {
	log.Printf("Spider::trackURL %+v", req)

	ourl, err := url.Prepare(db.Db, db.Redis, req.Url, req.Group, uint(0))
	if err != nil {
		return errors.InternalServerError("sg.micro.srv.spider.TrackURL", err.Error())
	}

	res.Id = uint64(ourl.ID)
	return nil
}

func (srv *Spider) FetchPages(ctx context.Context, req *proto.FetchPagesRequest, res *proto.FetchPagesResponse) error {

	log.Printf("Spider::fetchPages %+v", req)

	urls := url.FindFullText(db.Db, req.Search)
	results := make([]*proto.Url, len(urls))

	for i, u := range urls {
		results[i] = &proto.Url{
			Id:        uint64(u.ID),
			Url:       u.FullURL(),
			Title:     u.Title,
			Group:     u.Group,
			SitemapId: uint64(u.SitemapID.Int64),
		}
	}

	res.Results = results

	return nil
}
