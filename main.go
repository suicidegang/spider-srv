package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/micro/cli"
	"github.com/micro/go-micro"
	"github.com/suicidegang/spider-srv/db"
	"github.com/suicidegang/spider-srv/db/dataset"
	"github.com/suicidegang/spider-srv/db/sitemap"
	"github.com/suicidegang/spider-srv/handler"
	proto "github.com/suicidegang/spider-srv/proto/spider"
)

var (
	AsyncWorkers int = 4
)

func main() {
	service := micro.NewService(
		micro.Name("sg.micro.srv.spider"),
		micro.Version("0.1"),
		micro.RegisterTTL(time.Second*30),
		micro.RegisterInterval(time.Second*10),
		micro.Flags(
			cli.StringFlag{
				Name:   "redis_url",
				EnvVar: "REDIS_URL",
				Usage:  "Redis auth URL",
			},
			cli.StringFlag{
				Name:   "pgsql_url",
				EnvVar: "PGSQL",
				Usage:  "Postgresql auth URL",
			},
			cli.StringFlag{
				Name:   "async_workers",
				EnvVar: "ASYNC_WORKERS",
				Usage:  "How many workers per pool of async tasks.",
			},
		),
		micro.Action(func(c *cli.Context) {
			if len(c.String("redis_url")) > 0 {
				db.RedisUrl = c.String("redis_url")
			}

			if len(c.String("pgsql_url")) > 0 {
				db.DbUrl = c.String("pgsql_url")
			}

			if len(c.String("async_workers")) > 0 {
				AsyncWorkers = c.Int("async_workers")
			}
		}),
		micro.BeforeStart(func() error {
			log.Printf("[sg.micro.srv.spider] Starting service...")

			// Start the work queue dispatcher
			sitemap.Dispatcher()
			dataset.PQueueDispatcher(AsyncWorkers)

			go func() {
				log.Println(http.ListenAndServe("localhost:6060", nil))
			}()

			return nil
		}),
	)

	service.Init()
	db.Init()

	proto.RegisterSpiderHandler(service.Server(), new(handler.Spider))

	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
