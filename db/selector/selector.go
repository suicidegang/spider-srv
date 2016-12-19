package selector

import (
	"github.com/jinzhu/gorm"
)

type Selector struct {
	gorm.Model
	Name     string
	Patterns string `sql:"type:JSONB NOT NULL DEFAULT '{}'::JSONB"`
	Active   bool
}
