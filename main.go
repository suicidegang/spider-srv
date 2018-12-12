package main

import (
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/subosito/gotenv"
	"github.com/suicidegang/spider-srv/db"
	"github.com/suicidegang/spider-srv/db/dataset"
	"github.com/suicidegang/spider-srv/db/sitemap"
	"github.com/suicidegang/spider-srv/routes"
)

func main() {
	// service := micro.NewService(
	// 	micro.Name("sg.micro.srv.spider"),
	// 	micro.Version("0.1"),
	// 	micro.RegisterTTL(time.Second*30),
	// 	micro.RegisterInterval(time.Second*10),
	// 	micro.Flags(
	// 		cli.StringFlag{
	// 			Name:   "redis_url",
	// 			EnvVar: "REDIS_URL",
	// 			Usage:  "Redis auth URL",
	// 		},
	// 		cli.StringFlag{
	// 			Name:   "pgsql_url",
	// 			EnvVar: "PGSQL",
	// 			Usage:  "Postgresql auth URL",
	// 		},
	// 		cli.StringFlag{
	// 			Name:   "async_workers",
	// 			EnvVar: "ASYNC_WORKERS",
	// 			Usage:  "How many workers per pool of async tasks.",
	// 		},
	// 	),
	// 	micro.Action(func(c *cli.Context) {
	// 		if len(c.String("redis_url")) > 0 {
	// 			db.RedisUrl = c.String("redis_url")
	// 		}

	// 		if len(c.String("pgsql_url")) > 0 {
	// 			db.DbUrl = c.String("pgsql_url")
	// 		}

	// 		if len(c.String("async_workers")) > 0 {
	// 			AsyncWorkers = c.Int("async_workers")
	// 		}
	// 	}),
	// 	micro.BeforeStart(func() error {
	// 		log.Printf("[sg.micro.srv.spider] Starting service...")

	// 		return nil
	// 	}),
	// )
	gotenv.Load()
	if v, exists := os.LookupEnv("PGSQL_URL"); exists {
		db.DbUrl = v
	}

	if v, exists := os.LookupEnv("REDIS_URL"); exists {
		db.RedisUrl = v
	}

	//service.Init()
	db.Init()

	// Start the work queue dispatcher
	sitemap.Dispatcher()
	dataset.PQueueDispatcher(4)

	api := echo.New()
	api.Use(middleware.CORS())
	api.GET("/sitemap/:id", routes.Sitemap)
	api.GET("/sitemap/:id/urls", routes.SitemapURLs)
	api.GET("/xml/:id", routes.SitemapFile)
	api.POST("/sitemap", routes.TrackSitemap)
	api.PUT("/url", routes.UpdateURL)

	// Run router.
	api.Logger.Fatal(api.Start(":1323"))
}
