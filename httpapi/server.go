package httpapi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"time"

	session "github.com/go-session/session/v3"
	"github.com/google/uuid"
	"github.com/rbretecher/go-postman-collection"
	"gorm.io/gorm"

	"github.com/unstoppablego/framework/cache"
	"github.com/unstoppablego/framework/config"
	"github.com/unstoppablego/framework/db"
	"github.com/unstoppablego/framework/logs"
	"github.com/unstoppablego/framework/tool"
	"github.com/unstoppablego/framework/validation"
)

type ServerProvider struct {
	InternalMux *HttpApiRoute
	server      *http.Server
	Middleware  []MiddlewareX
}

func (sp *ServerProvider) RunServer(Addr string, xtls *tls.Config) {
	if sp.server == nil {
		if len(sp.Middleware) == 0 {
			// var xss XSSMiddleWare
			logs.Info(config.Cfg.Http.SessionName)
			session.InitManager(
				session.SetCookieName(config.Cfg.Http.SessionName),
				session.SetEnableSIDInHTTPHeader(true),
				session.SetSessionNameInHTTPHeader(""),
			)

			if config.Cfg.Http.SqlInjectMiddleWare {
				var sqlc SqlInjectMiddleWare
				sp.Middleware = append(sp.Middleware, sqlc)
			}

			// sp.Middleware = append(sp.Middleware, xss)
			// var sessionx SessionMiddleWare
			// sp.Middleware = append(sp.Middleware, sessionx)
		}

		if sp.InternalMux == nil {
			logs.Warn("server handle is nil ")
		}

		logs.Info("Run Server", Addr)

		if xtls == nil {
			sp.server = &http.Server{
				Addr:    Addr,
				Handler: sp.InternalMux,
			}
			log.Fatal(sp.server.ListenAndServe(), " my error")
		} else {
			sp.server = &http.Server{
				Addr:         Addr,
				Handler:      sp.InternalMux,
				TLSConfig:    xtls,
				TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
			}
			log.Fatal(sp.server.ListenAndServeTLS("", ""), " my error")
		}
	}
}

func (sp *ServerProvider) AddMux(internalMux *HttpApiRoute) {
	sp.InternalMux = internalMux
}

var internalServerProvider *ServerProvider

func Provider() *ServerProvider {
	return internalServerProvider
}

/*
Post 参数定义如下 不区分POST GET 等相关问题

path string

next func(w http.ResponseWriter, r *http.Request, respData []byte, obj reqModel) (data interface{}, err error)

enableValidate bool 默认开启
*/
func Post[reqModel any](path string, next func(ctx *Context, req reqModel) (data interface{}, err error), enableCache bool) {

	if next == nil {
		panic("http: nil handler")
	}

	var enableValidate = true

	if internalServerProvider == nil {
		internalServerProvider = &ServerProvider{}
		internalServerProvider.AddMux(New())
	}

	hxxx := func(w http.ResponseWriter, r *http.Request) {
		r.Body = io.NopCloser(ReusableReader(r.Body))

		if config.Cfg.Http.CrossDomain == "all" {

			xdomain, err := url.Parse(r.Referer())
			if err != nil {
				logs.Error(err)
			}
			crosmain := xdomain.Scheme + "://" + xdomain.Host
			logs.Info(crosmain)
			w.Header().Set("Access-Control-Allow-Credentials", "true") //前端js也需要开启跨域请求
			w.Header().Set("Access-Control-Allow-Origin", crosmain)    //来源网站
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, access-control-allow-origin, access-control-allow-headers, withCredentials, "+config.Cfg.Http.SessionName)
			w.Header().Set("Access-Control-Expose-Headers", config.Cfg.Http.SessionName)
		} else if config.Cfg.Http.CrossDomain != "false" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")                 //前端js也需要开启跨域请求
			w.Header().Set("Access-Control-Allow-Origin", config.Cfg.Http.CrossDomain) //来源网站
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, access-control-allow-origin, access-control-allow-headers, withCredentials, "+config.Cfg.Http.SessionName)
			w.Header().Set("Access-Control-Expose-Headers", config.Cfg.Http.SessionName)
		} else {
			w.Header().Set("Access-Control-Expose-Headers", config.Cfg.Http.SessionName)
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(200)
			return
		}

		if enableCache {
			xbody, err := io.ReadAll(r.Body)
			if err != nil {
				logs.Error(err)
			}
			mapKey := string(r.URL.RawQuery) + string(xbody)
			if data, ok := cache.Get[string, []byte](tool.Md5(mapKey)); ok {
				w.WriteHeader(200)
				w.Write(data)
				return
			}
		}

		//init ctx
		var ctxa Context
		ctxa.W = w
		ctxa.R = r

		store, err := session.Start(context.Background(), w, r)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		ctxa.Session = store
		ctxa.Tx = db.DB()
		store.Set("sessionstart", true)
		err = store.Save()
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		if RunMiddlewareX(internalServerProvider.Middleware, &ctxa) {
			return
		}

		defer tool.HandleRecover()
		var ResponseCentera ResponseDataProvider
		ResponseCentera.Code = "40000"

		// defer func() {
		// 	data, err := json.Marshal(ResponseCentera)
		// 	if err != nil {
		// 		logs.Error(err)
		// 		return
		// 	}
		// 	w.WriteHeader(200)
		// 	w.Write(data)
		// }()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			logs.Error(err)
			return
		}

		// logs.Info(string(body))

		// data, err := base64.NewEncoding(base64Key).DecodeString(string(body))
		// if err != nil {
		// 	logs.Error(err)
		// 	return
		// }

		// logs.Info(string(data))
		// data := body

		var xm reqModel
		err = json.Unmarshal(body, &xm)
		if err != nil {
			ResponseCentera.Msg = err.Error()
			logs.Error(err)
			RetCode(w, &ResponseCentera)
			return
		}
		if enableValidate {
			err = validation.ValidateStruct(xm)
			if err != nil {
				ResponseCentera.Msg = err.Error()
				logs.Error(err)
				RetCode(w, &ResponseCentera)
				return
			}
		}

		retdata, err := next(&ctxa, xm)
		if err != nil {
			ResponseCentera.Msg = err.Error()
			logs.Error(err)
			RetCode(w, &ResponseCentera)
			return
		}

		// jsondata, err := json.Marshal(retdata)
		// if err != nil {
		// 	logs.Error(err)
		// 	return
		// }

		// rdata := base64.NewEncoding(base64Key).EncodeToString(jsondata)
		// logs.Info(rdata)

		// ResponseCentera.Code = "200"
		// ResponseCentera.Data = retdata

		ResponseCentera.Code = "200"
		ResponseCentera.Data = retdata

		// ret code
		data, err := json.Marshal(ResponseCentera)
		if err != nil {
			logs.Error(err)
			return
		}
		w.WriteHeader(200)
		w.Write(data)

		if enableCache {
			mapKey := string(r.URL.RawQuery) + string(body)
			cache.Set[string, []byte](tool.Md5(mapKey), data, cache.WithExpiration(5*time.Second))

		}
	}

	//----------api doc -------------
	if config.Cfg.Http.Doc {
		var data reqModel
		//TODO: 进一步优化 用于生成文档数据
		postdata, _ := json.Marshal(data)

		req := postman.Request{
			URL: &postman.URL{
				Raw: "{{site}}" + path,
			},
			Method: postman.Post,
			Body: &postman.Body{
				Mode: "raw",
				Raw:  string(postdata),
				Options: &postman.BodyOptions{
					Raw: postman.BodyOptionsRaw{
						Language: "json",
					},
				},
			},
		}
		item := postman.CreateItem(postman.Item{
			Name:    path,
			Request: &req,
		})
		internalServerProvider.InternalMux.AppendApiDoc(item)
	}
	//----------api doc end-------------

	internalServerProvider.InternalMux.HandleFunc(path, hxxx)
}

/*
AddGetRoute 将丢弃BODY数据,只返回URL query 数据
*/
func Get[reqModel any](path string, next func(ctx *Context, query reqModel) (interface{}, error), enableCache bool) {
	if next == nil {
		panic("http: nil handler")
	}

	var enableValidate = true

	if internalServerProvider == nil {
		internalServerProvider = &ServerProvider{}
		internalServerProvider.AddMux(New())
	}

	hxxx := func(w http.ResponseWriter, r *http.Request) {

		if config.Cfg.Http.CrossDomain == "all" {
			logs.Info(r.Host)
			xdomain, err := url.Parse(r.Referer())
			if err != nil {
				logs.Error(err)
			}
			crosmain := xdomain.Scheme + "://" + xdomain.Host
			logs.Info(crosmain)
			w.Header().Set("Access-Control-Allow-Credentials", "true") //前端js也需要开启跨域请求
			w.Header().Set("Access-Control-Allow-Origin", crosmain)    //来源网站
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, access-control-allow-origin, access-control-allow-headers, withCredentials, "+config.Cfg.Http.SessionName)
			w.Header().Set("Access-Control-Expose-Headers", config.Cfg.Http.SessionName)

		} else if config.Cfg.Http.CrossDomain != "false" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")                 //前端js也需要开启跨域请求
			w.Header().Set("Access-Control-Allow-Origin", config.Cfg.Http.CrossDomain) //来源网站
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, access-control-allow-origin, access-control-allow-headers, withCredentials, "+config.Cfg.Http.SessionName)
			w.Header().Set("Access-Control-Expose-Headers", config.Cfg.Http.SessionName)
		} else {
			w.Header().Set("Access-Control-Expose-Headers", config.Cfg.Http.SessionName)
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(200)
			return
		}

		r.Body = io.NopCloser(ReusableReader(r.Body))
		// var enableCache bool
		if enableCache {
			// xurlVals := r.URL.Query()
			// var req reqModel
			// var reqMap = make(map[string]interface{})
			// getType := reflect.TypeOf(req)
			// for i := 0; i < getType.NumField(); i++ {
			// 	fieldType := getType.Field(i)

			// 	tag := getType.Field(i).Tag.Get("json")
			// 	if tag == "" || tag == "-" {
			// 		reqMap[fieldType.Name] = xurlVals.Get(fieldType.Name)
			// 	} else {
			// 		reqMap[tag] = xurlVals.Get(tag)
			// 	}
			// }
			// reqdata, err := json.Marshal(reqMap)
			// if err != nil {
			// 	return
			// }

			if data, ok := cache.Get[string, []byte](string(r.URL.RawQuery)); ok {
				w.WriteHeader(200)
				w.Write(data)
				return
			}
		}
		var ctxa Context
		ctxa.W = w
		ctxa.R = r

		store, err := session.Start(context.Background(), w, r)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		ctxa.Session = store
		ctxa.Tx = db.DB()

		store.Set("sessionstart", true)
		err = store.Save()
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		if RunMiddlewareX(internalServerProvider.Middleware, &ctxa) {

			return
		}

		defer tool.HandleRecover()
		var ResponseCentera ResponseDataProvider
		ResponseCentera.Code = "40000"

		// defer func() {
		// 	data, err := json.Marshal(ResponseCentera)
		// 	if err != nil {
		// 		logs.Error(err)
		// 		return
		// 	}

		// 	w.WriteHeader(200)
		// 	w.Write(data)
		// }()

		xurlVals := r.URL.Query()
		// logs.Info(xurlVals)

		var req reqModel
		var reqMap = make(map[string]interface{})
		getType := reflect.TypeOf(req)
		for i := 0; i < getType.NumField(); i++ {
			fieldType := getType.Field(i)

			tag := getType.Field(i).Tag.Get("json")
			if tag == "" || tag == "-" {
				reqMap[fieldType.Name] = xurlVals.Get(fieldType.Name)
			} else {
				reqMap[tag] = xurlVals.Get(tag)
			}
		}

		// body, err := ioutil.ReadAll(r.Body)
		// if err != nil {
		// 	logs.Error(err)
		// 	return
		// }

		// logs.Info(string(body))

		// data, err := base64.NewEncoding(base64Key).DecodeString(string(body))
		// if err != nil {
		// 	logs.Error(err)
		// 	return
		// }

		// logs.Info(string(data))
		// data := body
		reqdata, err := json.Marshal(reqMap)
		if err != nil {
			// ret code
			RetCode(w, &ResponseCentera)
			return
		}

		// var xm reqModel
		err = json.Unmarshal(reqdata, &req)
		if err != nil {
			ResponseCentera.Msg = err.Error()
			logs.Error(err)
			// ret code
			RetCode(w, &ResponseCentera)
			return
		}
		if enableValidate {
			err = validation.ValidateStruct(req)
			if err != nil {
				ResponseCentera.Msg = err.Error()
				logs.Error(err)
				// ret code
				RetCode(w, &ResponseCentera)
				return
			}
		}

		retdata, err := next(&ctxa, req)
		if err != nil {
			ResponseCentera.Msg = err.Error()
			logs.Error(err)
			// ret code
			RetCode(w, &ResponseCentera)
			return
		}

		// jsondata, err := json.Marshal(retdata)
		// if err != nil {
		// 	logs.Error(err)
		// 	return
		// }

		// rdata := base64.NewEncoding(base64Key).EncodeToString(jsondata)
		// logs.Info(rdata)

		ResponseCentera.Code = "200"
		ResponseCentera.Data = retdata

		// ret code

		data, err := json.Marshal(ResponseCentera)
		if err != nil {
			logs.Error(err)
			return
		}
		//ret code add cache
		if enableCache {
			cache.Set[string, []byte](string(r.URL.RawQuery), data, cache.WithExpiration(5*time.Second))
		}
		w.WriteHeader(200)
		w.Write(data)

	}

	if config.Cfg.Http.Doc {
		//----------api doc -------------

		var data reqModel
		//TODO: 进一步优化 用于生成文档数据
		postdata, _ := json.Marshal(data)

		req := postman.Request{
			URL: &postman.URL{
				Raw: "{{site}}" + path,
			},
			Method: postman.Post,
			Body: &postman.Body{
				Mode: "raw",
				Raw:  string(postdata),
				Options: &postman.BodyOptions{
					Raw: postman.BodyOptionsRaw{
						Language: "json",
					},
				},
			},
		}

		item := postman.CreateItem(postman.Item{
			Name:    path,
			Request: &req,
		})

		internalServerProvider.InternalMux.AppendApiDoc(item)
	}
	//----------api doc end-------------

	internalServerProvider.InternalMux.HandleFunc(path, hxxx)
}

func RetCode(w http.ResponseWriter, ret *ResponseDataProvider) {
	data, err := json.Marshal(ret)
	if err != nil {
		logs.Error(err)
		return
	}
	w.WriteHeader(200)
	w.Write(data)
}

type Context struct {
	Session SesssionStore
	W       http.ResponseWriter
	R       *http.Request
	RawBody []byte   //有些时候各种中间件将处理RawBody数据 , 已知 xss 中间件会修改Body
	Tx      *gorm.DB //
}

func AddFileUpload(path string) {
	internalServerProvider.InternalMux.HandleFunc(path, UploadFile)
}

func UploadFile(w http.ResponseWriter, r *http.Request) {

	var ResponseCentera ResponseDataProvider
	ResponseCentera.Code = "40000"

	defer func() {
		data, err := json.Marshal(ResponseCentera)
		if err != nil {
			logs.Error(err)
			return
		}
		w.WriteHeader(200)
		w.Write(data)
	}()

	if r.Method == "GET" {

	} else {

		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("file")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		var filePath = "./upload/" + time.Now().Format("2006-01-02") + "/"
		if err := os.MkdirAll(filePath, 0666); !os.IsNotExist(err) {
			log.Println(err)
		}
		uuidWithHyphen := uuid.New()
		filesuffix := path.Ext(handler.Filename)
		fileName := uuidWithHyphen.String() + filesuffix
		f, err := os.OpenFile(filePath+fileName, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
		ResponseCentera.Code = "200"
	}
}

//Create Event stream API
/*
Post 参数定义如下 不区分POST GET 等相关问题

path string

next func(w http.ResponseWriter, r *http.Request, respData []byte, obj reqModel) (data interface{}, err error)

enableValidate bool 默认开启
*/
func EventStream[reqModel any](path string, next func(ctx *Context, req reqModel, w http.ResponseWriter) (data interface{}, err error)) {

	if next == nil {
		panic("http: nil handler")
	}

	var enableValidate = true

	if internalServerProvider == nil {
		internalServerProvider = &ServerProvider{}
		internalServerProvider.AddMux(New())
	}

	hxxx := func(w http.ResponseWriter, r *http.Request) {

		_, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE not supported", http.StatusInternalServerError)
			return
		} else {
			logs.Info("SSE supported Yes")
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
		}

		r.Body = io.NopCloser(ReusableReader(r.Body))

		if config.Cfg.Http.CrossDomain == "all" {

			xdomain, err := url.Parse(r.Referer())
			if err != nil {
				logs.Error(err)
			}
			crosmain := xdomain.Scheme + "://" + xdomain.Host
			logs.Info(crosmain)
			w.Header().Set("Access-Control-Allow-Credentials", "true") //前端js也需要开启跨域请求
			w.Header().Set("Access-Control-Allow-Origin", crosmain)    //来源网站
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, access-control-allow-origin, access-control-allow-headers, withCredentials, "+config.Cfg.Http.SessionName)
			w.Header().Set("Access-Control-Expose-Headers", config.Cfg.Http.SessionName)
		} else if config.Cfg.Http.CrossDomain != "false" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")                 //前端js也需要开启跨域请求
			w.Header().Set("Access-Control-Allow-Origin", config.Cfg.Http.CrossDomain) //来源网站
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, access-control-allow-origin, access-control-allow-headers, withCredentials, "+config.Cfg.Http.SessionName)
			w.Header().Set("Access-Control-Expose-Headers", config.Cfg.Http.SessionName)
		} else {
			w.Header().Set("Access-Control-Expose-Headers", config.Cfg.Http.SessionName)
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(200)
			return
		}

		//init ctx
		var ctxa Context
		ctxa.W = w
		ctxa.R = r

		store, err := session.Start(context.Background(), w, r)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		ctxa.Session = store
		ctxa.Tx = db.DB()
		store.Set("sessionstart", true)
		err = store.Save()
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		if RunMiddlewareX(internalServerProvider.Middleware, &ctxa) {
			return
		}

		defer tool.HandleRecover()
		var ResponseCentera ResponseDataProvider
		ResponseCentera.Code = "40000"

		// defer func() {
		// 	data, err := json.Marshal(ResponseCentera)
		// 	if err != nil {
		// 		logs.Error(err)
		// 		return
		// 	}
		// 	w.WriteHeader(200)
		// 	w.Write(data)
		// }()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			logs.Error(err)
			return
		}

		// logs.Info(string(body))

		// data, err := base64.NewEncoding(base64Key).DecodeString(string(body))
		// if err != nil {
		// 	logs.Error(err)
		// 	return
		// }

		// logs.Info(string(data))
		// data := body

		var xm reqModel
		if len(body) > 0 {
			err = json.Unmarshal(body, &xm)
			if err != nil {
				ResponseCentera.Msg = err.Error()
				logs.Error(err)
				RetCode(w, &ResponseCentera)
				return
			}
			if enableValidate {
				err = validation.ValidateStruct(xm)
				if err != nil {
					ResponseCentera.Msg = err.Error()
					logs.Error(err)
					RetCode(w, &ResponseCentera)
					return
				}
			}
		}

		retdata, err := next(&ctxa, xm, w)
		if err != nil {
			ResponseCentera.Msg = err.Error()
			logs.Error(err)
			RetCode(w, &ResponseCentera)
			return
		}

		// jsondata, err := json.Marshal(retdata)
		// if err != nil {
		// 	logs.Error(err)
		// 	return
		// }

		// rdata := base64.NewEncoding(base64Key).EncodeToString(jsondata)
		// logs.Info(rdata)

		// ResponseCentera.Code = "200"
		// ResponseCentera.Data = retdata

		ResponseCentera.Code = "200"
		ResponseCentera.Data = retdata

		// ret code
		data, err := json.Marshal(ResponseCentera)
		if err != nil {
			logs.Error(err)
			return
		}
		w.WriteHeader(200)
		w.Write(data)

	}

	internalServerProvider.InternalMux.HandleFunc(path, hxxx)
}
