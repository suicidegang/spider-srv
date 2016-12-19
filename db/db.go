package db

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/suicidegang/spider-srv/db/selector"
	"github.com/suicidegang/spider-srv/db/sitemap"
	"github.com/suicidegang/spider-srv/db/url"
	"gopkg.in/redis.v5"
	//"github.com/micro/go-micro/errors"
)

var (
	DbUrl    string = "root:root@tcp(127.0.0.1:5432)/user"
	RedisUrl string = "redis://127.0.0.1:6379"
	Db       *gorm.DB
	Redis    *redis.Client
)

func Init() {
	var err error

	// Prefix micro-service tables
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return "spider_" + defaultTableName
	}

	Db, err = gorm.Open("postgres", DbUrl)
	if err != nil {
		log.Fatal(err)
	}

	Db.AutoMigrate(&sitemap.Sitemap{}, &url.Url{}, &selector.Selector{})
	Db.Model(&url.Url{}).AddUniqueIndex("idx_url_params", "url", "query_params")
	Db.LogMode(true)

	options, err := redis.ParseURL(RedisUrl)
	if err != nil {
		log.Fatal(err)
	}

	Redis = redis.NewClient(options)
	_, err = Redis.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}
}

/*
func Create(user *proto.User, password string) error {
	u := User{
		Email:    user.Email,
		Username: user.Username,
		Name:     user.Username,
		Password: password,
	}

	db.Create(&u)

	return db.Error
}

func Read(id uint64) (*proto.User, error) {
	var row User
	if err := db.First(&row, uint(id)).Error; err != nil {
		return nil, err
	}
	return transformUser(row), nil
}

func Search(query string, within []uint64) ([]*proto.User, error) {
	var rows []User
	q := db

	if len(query) == 0 && len(within) == 0 {
		return make([]*proto.User, 0), errors.BadRequest("sg.micro.srv.user.Search", "Empty params.")
	}

	if len(query) > 0 {
		like := "%" + query + "%"
		q = q.Where("email LIKE ? OR name LIKE ? OR username LIKE ?", like, like, like)
	}

	if len(within) > 0 {
		q = q.Where("id IN (?)", within)
	}

	if err := q.Find(&rows).Error; err != nil {
		return make([]*proto.User, 0), err
	}

	var users []*proto.User
	for _, usr := range rows {
		users = append(users, transformUser(usr))
	}

	return users, nil
}

func transformUser(user User) *proto.User {
	return &proto.User{
		Id:       uint64(user.ID),
		Username: user.Username,
		Email:    user.Email,
		Created:  user.CreatedAt.String(),
		Updated:  user.UpdatedAt.String(),
	}
}*/
