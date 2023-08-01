package test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
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

	httpapi.Get[ReqGet]("/fuck", httpapi.CustomXSSMiddleWare(func(ctx *httpapi.Context, query ReqGet) (interface{}, error) {
		// if data, ok := ctx.Session.Get("sessionstart"); ok {
		// 	// logs.Info(data)
		// }
		return query, nil
	}))

	httpapi.Provider().RunServer("0.0.0.0:1999", nil)
}

func TestServerGin(t *testing.T) {
	r := gin.Default()
	r.GET("/fuck", func(c *gin.Context) {

		logs.Info(c.Query("Id"))
		logs.Info(c.Query("Hello"))
		logs.Info(c.Query("world"))
		var req ReqGet
		if err := c.ShouldBindQuery(&req); err == nil {
			c.JSON(200, gin.H{"name": req.Id, "address": req.Hello, "world": req.World})
		} else {
			c.JSON(400, gin.H{"error": err.Error()})
		}
		// c.JSON(http.StatusOK, gin.H{
		// 	"message": "pong",
		// })
	})
	r.Run(":1999")
}
