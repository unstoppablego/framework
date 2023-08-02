package test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/unstoppablego/framework/cache"
	"github.com/unstoppablego/framework/config"
	"github.com/unstoppablego/framework/httpapi"
	"github.com/unstoppablego/framework/logs"
)

type User struct {
	Id     int
	Name   string
	Age    int
	Gender bool
}

// ToString方法
func (u User) String() string {
	return "User[ Id " + strconv.Itoa(u.Id) + "]"
}

// 设置Name方法
func (u *User) SetName(name string) string {
	oldName := u.Name
	u.Name = name
	return oldName
}

// 年龄数+1
func (u *User) AddAge() bool {
	u.Age++
	return true
}

// 测试方法
func (u User) TestUser() {
	fmt.Println("我只是输出某些内容而已....")
}

type ReqGet struct {
	Id    string
	Hello string
	World string `json:"world"`
}

func TestServer(t *testing.T) {
	config.ReadConf("../config/")
	httpapi.Get[ReqGet]("/fuck", func(ctx *httpapi.Context, query ReqGet) (interface{}, error) {
		// if data, ok := ctx.Session.Get("sessionstart"); ok {
		// 	// logs.Info(data)
		// }
		return query, nil
	})
	//httpapi.CustomXSSMiddleWare(

	httpapi.Provider().RunServer("0.0.0.0:1999", nil)
}

func TestCache(t *testing.T) {
	var user User
	user.Id = 18
	cache.Set[string, User]("Hello", user, cache.WithExpiration(5*time.Second))
	for i := 0; i < 30; i++ {
		xu, ok := cache.Get[string, User]("Hello")
		logs.Info(xu, ok)
		time.Sleep(5 * time.Second)
	}

}
