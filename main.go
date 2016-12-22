package main

import (
	"log"

	"github.com/micro/cli"
	"github.com/micro/go-micro"
	"github.com/suicidegang/spider-srv/db"
	"github.com/suicidegang/spider-srv/db/sitemap"
	"github.com/suicidegang/spider-srv/handler"
	proto "github.com/suicidegang/spider-srv/proto/spider"
)

func main() {
	service := micro.NewService(
		micro.Name("sg.micro.srv.spider"),
		micro.Version("0.1"),
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
		),
		micro.Action(func(c *cli.Context) {
			if len(c.String("redis_url")) > 0 {
				db.RedisUrl = c.String("redis_url")
			}

			if len(c.String("pgsql_url")) > 0 {
				db.DbUrl = c.String("pgsql_url")
			}
		}),
		micro.BeforeStart(func() error {
			log.Printf("[sg.micro.srv.spider] Starting service...")

			// Start the work queue dispatcher
			sitemap.Dispatcher(4)
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
