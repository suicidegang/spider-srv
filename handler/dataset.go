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

func (srv *Spider) FetchDatasetBy(ctx context.Context, req *proto.FetchDatasetByRequest, res *proto.FetchDatasetByResponse) error {
	log.Printf("Spider::fetchDatasetBy %+v", req)

	u, err := url.FindBy(db.Db, req.Conditions)
	if err != nil {
		return errors.InternalServerError("sg.micro.srv.spider.FetchDatasetBy", err.Error())
	}

	ds, err := dataset.Prepare(db.Db, db.Redis, uint(req.SelectorId), uint(u.ID))
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

	query := db.Db.Where("\"group\" = ?", req.Group)

	// Apply query conditions if any
	if len(req.Conditions) > 0 {
		statement, binds := req.Conditions[0], req.Conditions[1:]
		bind := make([]interface{}, len(binds))
		for i, v := range binds {
			bind[i] = v
		}

		query = query.Where(statement, bind...)
	}

	urls := url.All(query)
	urls.Each(func(u url.Url) {
		job := dataset.PrepareJob{UrlID: u.ID, SelectorID: uint(req.SelectorId), DB: db.Db, R: db.Redis}

		dataset.Prepares <- job
	})

	res.Count = uint64(len(urls))
	return nil
}
