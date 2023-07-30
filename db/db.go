package db

import (
	"fmt"
	"sync"
	"time"

	"github.com/unstoppablego/framework/config"
	"github.com/unstoppablego/framework/tool"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// Self ORM https://zhuanlan.zhihu.com/p/439093037

var dbInstance *gorm.DB
var dblock sync.Mutex

type SException struct {
	s    string
	code int
}

func NewSException(text string, code int) *SException {
	return &SException{s: text, code: code}
}

func (e *SException) Error() string {
	return fmt.Sprintf("[%d]%s", e.code, e.s)
}

func DB() *gorm.DB {

	defer tool.HandleRecover()

	if dbInstance == nil {
		NewDB()
	}

	sqlDB, err := dbInstance.DB()
	if err != nil {
		panic(NewSException(err.Error(), 10005))
	}

	if err := sqlDB.Ping(); err != nil {
		panic(NewSException(err.Error(), 10004))
	}
	return dbInstance
}

/*
将覆盖 内置 db 变量，通常只会使用DB 函数
但是当一些不可控因素下，将会需要使用该函数
*/
func NewDB() {
	defer tool.HandleRecover()
	db, err := gorm.Open(mysql.Open(config.Cfg.DB[0].DSN()), &gorm.Config{})
	if err != nil {
		panic(NewSException(err.Error(), 10003))
	}
	dblock.Lock()
	dbInstance = db
	dblock.Unlock()
	var Sources []gorm.Dialector
	var Replicas []gorm.Dialector
	for _, v := range config.Cfg.DB {
		if v.Type == "mysql" && v.Tag == "rw" {
			Sources = append(Sources, mysql.Open(v.DSN()))
			continue
		}
		if v.Type == "mysql" && v.Tag == "r" {
			Replicas = append(Replicas, mysql.Open(v.DSN()))
			continue
		}
	}
	dbInstance.Use(
		dbresolver.Register(dbresolver.Config{
			Sources:  Sources,
			Replicas: Replicas,
			Policy:   dbresolver.RandomPolicy{},
		}).SetConnMaxIdleTime(time.Hour).SetConnMaxLifetime(time.Duration(config.Cfg.Http.SetConnMaxLifetime) * time.Hour).SetMaxIdleConns(config.Cfg.Http.SetMaxIdleConns).SetMaxOpenConns(config.Cfg.Http.SetMaxOpenConns))
}
