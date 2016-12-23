package handler

import (
	"github.com/micro/go-micro/errors"
	"github.com/suicidegang/spider-srv/db"
	"github.com/suicidegang/spider-srv/db/dataset"
	"github.com/suicidegang/spider-srv/db/url"
	proto "github.com/suicidegang/spider-srv/proto/spider"
	"golang.org/x/net/context"

	"log"
)

func (srv *Spider) FetchDataset(ctx context.Context, req *proto.FetchDatasetRequest, res *proto.FetchDatasetResponse) error {
	log.Printf("Spider::fetchDataset %+v", req)

	ds, err := dataset.Prepare(db.Db, db.Redis, uint(req.Id), uint(req.UrlId))
	if err != nil {
		return errors.InternalServerError("sg.micro.srv.spiderFetchDataset", err.Error())
	}

	res.Data, err = ds.Document()
	if err != nil {
		return errors.InternalServerError("sg.micro.srv.spiderFetchDataset", err.Error())
	}

	return nil
}

func (srv *Spider) PrepareDatasets(ct context.Context, req *proto.PrepareDatasetsRequest, res *proto.PrepareDatasetsResponse) error {
	log.Printf("Spider::prepareDatasets %+v", req)

	urls := url.All(db.Db.Where("\"group\" = ?", req.Group))
	urls.Each(func(u url.Url) {
		job := dataset.PrepareJob{UrlID: u.ID, SelectorID: uint(req.SelectorId), DB: db.Db, R: db.Redis}

		dataset.Prepares <- job
	})

	res.Count = uint64(len(urls))
	return nil
}
