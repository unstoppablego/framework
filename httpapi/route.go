package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	// "server/m/v2/config"
	"strings"

	"github.com/rbretecher/go-postman-collection"
	"github.com/unstoppablego/framework/logs"
	"github.com/unstoppablego/framework/tool"
	"github.com/unstoppablego/framework/validation"
)

type HttpApiRoute struct {
	m   *http.ServeMux
	doc []*postman.Items
}

type HttpApiRouteHandler struct {
	h http.Handler
	p string
}

func (h HttpApiRouteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.p == "" {
		logs.Info("HTTP API " + r.Method + " " + r.URL.Path)
	} else {
		logs.Info("HTTP API " + r.Method + " " + h.p)
	}

	h.h.ServeHTTP(w, r)
}

func New() *HttpApiRoute {
	return &HttpApiRoute{m: &http.ServeMux{}}
	// return &x
}

func (c *HttpApiRoute) Handle(pattern string, handler http.Handler) {
	var h = HttpApiRouteHandler{h: handler}
	c.m.Handle(pattern, h)
}

func (c *HttpApiRoute) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	hxxx := func(w http.ResponseWriter, r *http.Request) {
		logs.Info("HTTP API " + r.Method + " " + pattern)
		handler(w, r)
	}
	c.m.HandleFunc(pattern, hxxx)
}

func (c *HttpApiRoute) Handler(r *http.Request) (h http.Handler, pattern string) {
	return c.m.Handler(r)
}

func (c *HttpApiRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.m.ServeHTTP(w, r)
}

func (c *HttpApiRoute) AppendApiDoc(i *postman.Items) {
	c.doc = append(c.doc, i)
}

type ResponseDataProvider struct {
	Code string
	Msg  string
	Data interface{}
}

func (r *ResponseDataProvider) IsOk() bool {
	return r.Code == "200"
}

// 拥有 data 加密 和 字段校验的函数
func HandleApiWithEncode[reqModel any](mux *HttpApiRoute, path string, next func(w http.ResponseWriter, r *http.Request, respData []byte, obj reqModel) (data interface{}, err error), base64Key string, enableValidate bool) {
	hxxx := func(w http.ResponseWriter, r *http.Request) {
		// logs.Info("HTTP API " + r.Method + " " + path)
		defer tool.HandleRecover()
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

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logs.Error(err)
			return
		}

		// logs.Info(string(body))

		data, err := base64.NewEncoding(base64Key).DecodeString(string(body))
		if err != nil {
			logs.Error(err)
			return
		}

		// logs.Info(string(data))
		// data := body

		var xm reqModel
		err = json.Unmarshal(data, &xm)
		if err != nil {
			ResponseCentera.Msg = err.Error()
			logs.Error(err)
			return
		}
		if enableValidate {
			err = validation.ValidateStruct(xm)
			if err != nil {
				ResponseCentera.Msg = err.Error()
				logs.Error(err)
				return
			}
		}

		retdata, err := next(w, r, data, xm)
		if err != nil {
			ResponseCentera.Msg = err.Error()
			logs.Error(err)
			return
		}

		jsondata, err := json.Marshal(retdata)
		if err != nil {
			logs.Error(err)
			return
		}

		rdata := base64.NewEncoding(base64Key).EncodeToString(jsondata)
		// logs.Info(rdata)

		ResponseCentera.Code = "200"
		ResponseCentera.Data = rdata
	}

	//----------api doc -------------
	if next == nil {
		panic("http: nil handler")
	}

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

	mux.AppendApiDoc(item)
	//----------api doc end-------------

	mux.HandleFunc(path, hxxx)
}

// CallEncodeApi 调用 MyHandleMiddlewareY2 加密服务器
//
// xmodel 返回的结构体 respModel
func CallEncodeApi[respModel any](url string, postdata []byte, base64key string) (bool, *respModel) {
	httpClient := tool.CreateHTTPClient()

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(base64.NewEncoding(base64key).EncodeToString(postdata)))
	if err != nil {
		logs.Error(err)
		return false, nil
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		logs.Error(err)
		return false, nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error(err)
		return false, nil
	}
	resp.Body.Close()

	var ResponseCentera ResponseDataProvider
	err = json.Unmarshal(body, &ResponseCentera)
	if err != nil {
		logs.Error(err)
		return false, nil
	}
	logs.Info(string(body))
	if ResponseCentera.IsOk() {
		if v, ok := ResponseCentera.Data.(string); ok {
			var xm respModel
			retdata, err := base64.NewEncoding(base64key).DecodeString(v)
			if err != nil {
				logs.Error(err)
				return false, nil
			}
			err = json.Unmarshal(retdata, &xm)
			if err != nil {
				logs.Error(err)
				return false, nil
			}
			return true, &xm
		}

	}
	return false, nil
}

func (ccc *HttpApiRoute) CreateApi() {

	items := ccc.doc
	c := postman.CreateCollection("My collection", "My awesome collection")
	ver := &postman.Variable{Key: "site", Value: "https://www.baidu.com"}
	c.Variables = append(c.Variables, ver)
	for _, v := range items {
		// v.Variables = append(v.Variables, ver)
		c.AddItem(v)
	}

	file, err := os.Create("postman_collection.json")
	defer file.Close()

	if err != nil {
		panic(err)
	}

	err = c.Write(file, postman.V200)
}
