package handler

import (
	"github.com/micro/go-micro/errors"
	"github.com/suicidegang/spider-srv/db"
	"github.com/suicidegang/spider-srv/db/selector"
	"github.com/suicidegang/spider-srv/db/url"
	proto "github.com/suicidegang/spider-srv/proto/spider"
	"golang.org/x/net/context"

	"log"
)

func (srv *Spider) FetchDataset(ctx context.Context, req *proto.FetchDatasetRequest, res *proto.FetchDatasetResponse) error {
	log.Printf("Spider::fetchDataset %+v", req)

	u, err := url.One(db.Db, req.UrlId)
	if err != nil {
		return errors.InternalServerError("sg.micro.srv.spider.FetchDataset", err.Error())
	}

	s, err := selector.One(db.Db, req.Id)
	if err != nil {
		return errors.InternalServerError("sg.micro.srv.spider.FetchDataset", err.Error())
	}

	doc, err := u.Document(db.Redis)
	if err != nil {
		return errors.InternalServerError("sg.micro.srv.spider.FetchDataset", err.Error())
	}

	dataset, err := s.Query(doc)
	if err != nil {
		return errors.InternalServerError("sg.micro.srv.spider.FetchDataset", err.Error())
	}

	res.Data = dataset
	return nil
}
