package handler

import (
	"github.com/micro/go-micro/errors"
	"github.com/suicidegang/spider-srv/db"
	"github.com/suicidegang/spider-srv/db/dataset"
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
