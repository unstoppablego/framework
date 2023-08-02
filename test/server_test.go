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

	httpapi.Get[ReqGet]("/fuck", Fuckfunc[ReqGet], true)

	httpapi.Post[ReqGet]("/fuck2", httpapi.CustomXSSMiddleWare[ReqGet](Fuckfunc[ReqGet]), true)

	httpapi.AddFileUpload("/upload")

	httpapi.Provider().RunServer("0.0.0.0:1999", nil)
}

//go:generate go run ../main.go
func Fuckfunc[ReqGet any](ctx *httpapi.Context, query ReqGet) (interface{}, error) {

	return query, nil
}

func Api() {
	httpapi.Get[ReqGet]("/fuck", func(ctx *httpapi.Context, query ReqGet) (interface{}, error) {

		return nil, nil
	}, false)

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

// func UploadFile[reqT any](ctx *httpapi.Context, req reqT) (interface{}, error) {
// 	logs.Info("Upload File")

// 	if ctx.R.Method == "GET" {

// 	} else {
// 		ctx.R.ParseMultipartForm(32 << 20)
// 		file, handler, err := ctx.R.FormFile("file")
// 		if err != nil {
// 			fmt.Println(err)
// 			return 200, err
// 		}
// 		defer file.Close()
// 		var filePath = "./upload/" + time.Now().Format("2006-01-02") + "/"
// 		if err := os.MkdirAll(filePath, 0666); !os.IsNotExist(err) {
// 			log.Println(err)
// 		}
// 		uuidWithHyphen := uuid.New()
// 		filesuffix := path.Ext(handler.Filename)
// 		fileName := uuidWithHyphen.String() + filesuffix
// 		f, err := os.OpenFile(filePath+fileName, os.O_WRONLY|os.O_CREATE, 0666)
// 		if err != nil {
// 			fmt.Println(err)
// 			return 200, err
// 		}
// 		defer f.Close()
// 		io.Copy(f, file)
// 		return 200, nil
// 	}
// 	return 200, nil
// }
