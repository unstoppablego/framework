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
	"github.com/unstoppablego/framework/tool"
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

	// httpapi.AddFileUpload("/upload")

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

func TestCert(t *testing.T) {
	tool.Ca()
	tool.CreateSSH_KEY()
	// tool.GenerateRsaKey(true)
}

// 0YzUnNUzPN8qbR6FaTK62oi75weJRzNRwOcbpkRbEIHP5rR+parum1MmvOAYvYkJ7jyrz5sDuBlOg1TDO6hIs8xbV8meF79ZbdXed6ezn52ze2PdhPyzWEevK9eQLMRxRUYqM9nThclxPOuPU8lASZdvnEc1z3QZmy4i5xUKh/uyM8eofojDbQF9Db16sf4x4QOF0NZfKcte87V2mRWTIDEfJysmRyanMYSQdDnSIbKlfwok89fVESo3Ypf7C7xfgTpaleYLCEAaJdumc+bjmvdr48cxUIBto0o70spY5Or9Dw8u+kFyTkcm+j3B6hFi3f8g9dADaAfOQ9TQTviK+vO7jPqECTRtkaI4F2uXPwuIAcoupQ0KPYxgyqXVaABRsiy7r4kxKI+k1SKkfzrjfBmOi3zXoJp4V4jcl4CEXBre+TL8X2VScw8B+YBj9jFQsOI4xdcW/UoBkbb9leLAzufCuieseqOKZvWvcgr6YN1g46GcNkmQrRhjOzrQv/Pmyy/5vDfVi7AUrKvK81A8qOAPDlnD4nHdeCqu2Wuvc4E8ioz1D4qCzJX2hjSyX12Abb9rR6DB+3W1DSIoDznC2A9PNSourRHczUbri5hQpMcMjZOPzAnEBQTP55hzc82pfphReB0Zzy6VgkexLC5D7wn9YR6yL9mamSzLzudA6cU=
// ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDQS++ncv7DDapxbX+GBDw3hkXG0J5HoxDJIJcri5wWwkQOVK+OoCWsiOigRdn2xvWfObvvdQiWcvP7hvVTiVs+tgxx0FRxHiQZir24D3YuoYnhxKSkArxBdzmUkcw5TB72dR2OufIc5MxsddHNGE4lGi9fbTniZIvGd6zFLNL/7Vn41gX1dXVJnRNd8bf8giSxDIKiTz7GIccuMKy9l6uhoDAwV4iXAxRmuioSVFyQQfWDaHVCZ8DIy9ghrckZ2jQXOzobz874C1/7LlaFOQlsClQPU8cYS5YGQV3BU01UGusreSLpydklE8IvDtLDfr2bsFF0Dkl+uFM38EIEy0Um1sJ51TzCzZkWpr+WKonGzLFpEkCw1vuow258H/ZaBpq25gpsxcoFv5Xx6wwjBUjQRcWOfQhmRC840jjcJO+OgCbtaEg8L6xraCvcXageJluz8Oqoa2lSUTtlyTz/tdz2j96c6cmxV0zQi6m27I9a66CyWnPBKNs7/p40c8Q2dyfi2BpRTXQwIxj0ciwo/Zq2Pv4JqmeZx+3CFMDzI2olcGDFoPIel8XS8i65rmg04B3nPMaOfaHtZRQ7m5S8R9/iBa/dhAZW3dEO5OyvUnBhcmZgsT7msU+pqgaflYpHH65iazcXjgRzFiXIOaHO0RotT3Hw8LyajpYUy7bZE1n54w== root@VM-4-8-debian
