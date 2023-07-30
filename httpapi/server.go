package httpapi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"

	session "github.com/go-session/session/v3"
	"github.com/rbretecher/go-postman-collection"

	"github.com/unstoppablego/framework/config"
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
			var xss XSSMiddleWare
			var sqlc SqlInjectMiddleWare
			// var sessionx SessionMiddleWare
			sp.Middleware = append(sp.Middleware, sqlc)
			sp.Middleware = append(sp.Middleware, xss)
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
// func Post[reqModel any](path string, next func(w http.ResponseWriter, r *http.Request, bodyData []byte, req reqModel, sess SesssionStore) (data interface{}, err error)) {

// 	if next == nil {
// 		panic("http: nil handler")
// 	}

// 	var enableValidate = true

// 	if internalServerProvider == nil {
// 		internalServerProvider = &ServerProvider{}
// 		internalServerProvider.AddMux(New())
// 	}

// 	hxxx := func(w http.ResponseWriter, r *http.Request) {

// 		if RunMiddlewareX(internalServerProvider.Middleware, w, r) {
// 			return
// 		}

// 		store, err := session.Start(context.Background(), w, r)
// 		if err != nil {
// 			fmt.Fprint(w, err)
// 			return
// 		}
// 		store.Set("sessionstart", true)
// 		err = store.Save()
// 		if err != nil {
// 			fmt.Fprint(w, err)
// 			return
// 		}

// 		defer tool.HandleRecover()
// 		var ResponseCentera ResponseCenter
// 		ResponseCentera.Code = "40000"

// 		defer func() {
// 			data, err := json.Marshal(ResponseCentera)
// 			if err != nil {
// 				logs.Error(err)
// 				return
// 			}
// 			w.WriteHeader(200)
// 			w.Write(data)
// 		}()

// 		body, err := ioutil.ReadAll(r.Body)
// 		if err != nil {
// 			logs.Error(err)
// 			return
// 		}

// 		// logs.Info(string(body))

// 		// data, err := base64.NewEncoding(base64Key).DecodeString(string(body))
// 		// if err != nil {
// 		// 	logs.Error(err)
// 		// 	return
// 		// }

// 		// logs.Info(string(data))
// 		// data := body

// 		var xm reqModel
// 		err = json.Unmarshal(body, &xm)
// 		if err != nil {
// 			ResponseCentera.Msg = err.Error()
// 			logs.Error(err)
// 			return
// 		}
// 		if enableValidate {
// 			err = validation.ValidateStruct(xm)
// 			if err != nil {
// 				ResponseCentera.Msg = err.Error()
// 				logs.Error(err)
// 				return
// 			}
// 		}

// 		retdata, err := next(w, r, body, xm, store)
// 		if err != nil {
// 			ResponseCentera.Msg = err.Error()
// 			logs.Error(err)
// 			return
// 		}

// 		// jsondata, err := json.Marshal(retdata)
// 		// if err != nil {
// 		// 	logs.Error(err)
// 		// 	return
// 		// }

// 		// rdata := base64.NewEncoding(base64Key).EncodeToString(jsondata)
// 		// logs.Info(rdata)

// 		ResponseCentera.Code = "200"
// 		ResponseCentera.Data = retdata
// 	}

// 	//----------api doc -------------
// 	if config.Cfg.Http.Doc {
// 		var data reqModel
// 		//TODO: 进一步优化 用于生成文档数据
// 		postdata, _ := json.Marshal(data)

// 		req := postman.Request{
// 			URL: &postman.URL{
// 				Raw: "{{site}}" + path,
// 			},
// 			Method: postman.Post,
// 			Body: &postman.Body{
// 				Mode: "raw",
// 				Raw:  string(postdata),
// 				Options: &postman.BodyOptions{
// 					Raw: postman.BodyOptionsRaw{
// 						Language: "json",
// 					},
// 				},
// 			},
// 		}
// 		item := postman.CreateItem(postman.Item{
// 			Name:    path,
// 			Request: &req,
// 		})
// 		internalServerProvider.InternalMux.AppendApiDoc(item)
// 	}
// 	//----------api doc end-------------

// 	internalServerProvider.InternalMux.HandleFunc(path, hxxx)
// }

/*
AddGetRoute 将丢弃BODY数据,只返回URL query 数据
*/
func Get[reqModel any](path string, next func(ctx *Context, query reqModel) (interface{}, error)) {
	if next == nil {
		panic("http: nil handler")
	}

	var enableValidate = true

	if internalServerProvider == nil {
		internalServerProvider = &ServerProvider{}
		internalServerProvider.AddMux(New())
	}

	hxxx := func(w http.ResponseWriter, r *http.Request) {

		var ctxa Context
		ctxa.W = w
		ctxa.R = r
		store, err := session.Start(context.Background(), w, r)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		ctxa.Session = store

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
		var ResponseCentera ResponseCenter
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

		xurlVals := r.URL.Query()
		logs.Info(xurlVals)

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
			return
		}
		// var xm reqModel
		err = json.Unmarshal(reqdata, &req)
		if err != nil {
			ResponseCentera.Msg = err.Error()
			logs.Error(err)
			return
		}
		if enableValidate {
			err = validation.ValidateStruct(req)
			if err != nil {
				ResponseCentera.Msg = err.Error()
				logs.Error(err)
				return
			}
		}

		retdata, err := next(&ctxa, req)
		if err != nil {
			ResponseCentera.Msg = err.Error()
			logs.Error(err)
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

type Context struct {
	Session SesssionStore
	W       http.ResponseWriter
	R       *http.Request
	RawBody []byte //有些时候各种中间件将处理RawBody数据 , 已知 xss 中间件会修改Body
}
