package dataset

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/redis.v5"

	"log"
)

type PrepareJob struct {
	UrlID      uint
	SelectorID uint
	DB         *gorm.DB
	R          *redis.Client
}

func (job PrepareJob) Work() {

	ds, err := Prepare(job.DB, job.R, job.SelectorID, job.UrlID)
	if err != nil {
		log.Printf("[err] %v [%+v]", err, job)
		return
	}

	log.Printf("[datasets] Prepared dataset %v", ds.ID)
}
