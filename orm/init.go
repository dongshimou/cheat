package orm

import (
	"cheat/logger"
	"cheat/model"
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

var db *gorm.DB

func Get() *gorm.DB {
	return db
}
func init() {
	cfg := model.ConfigDatabase{}

	logger.Info(cfg)
	args := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	var err error
	db, err = gorm.Open("mysql", args)
	if err != nil {
		panic(err)
	}

	db.DB().SetMaxOpenConns(cfg.MaxConn)

	db.DB().SetMaxIdleConns(cfg.MaxIdle)

	db.DB().SetConnMaxLifetime(time.Minute * 10)

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return defaultTableName
	}

	db = db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8 auto_increment=1")

	db = db.AutoMigrate(
		&model.User{},
	)

	db.LogMode(true)
}
