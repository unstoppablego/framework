package httpapi

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/microcosm-cc/bluemonday"
	"github.com/unstoppablego/framework/logs"
	"github.com/unstoppablego/framework/security"
)

/*
 */
type MiddlewareX interface {
	Handle(ctx *Context) (Abort bool)
	Enable() bool
	Name() string
}

// r.Body = io.NopCloser(ReusableReader(r.Body))
type XSSMiddleWare struct {
}

func (x XSSMiddleWare) Handle(ctx *Context) bool {

	p := bluemonday.UGCPolicy()
	body, err := ioutil.ReadAll(ctx.R.Body)

	ctx.W.Header().Set("xss", "run")
	if err != nil {
		logs.Error(err)
		return false
	}

	if len(body) == 0 {
		return false
	}
	logs.Info(string(body))
	// ctx.RawBody = body
	sanitizedBody, err := security.XSSFilterJSON(p, string(body))
	if err != nil {
		logs.Error("XSSMiddleWarex Sanitized Body Error", err)
		return true
	}

	logs.Info(sanitizedBody)

	ctx.R.Body = io.NopCloser(ReusableReader(bytes.NewBuffer([]byte(sanitizedBody))))
	return false
}

func (x XSSMiddleWare) Enable() bool {
	return true
}

func (x XSSMiddleWare) Name() string {
	return "XSSMiddleWare"
}

type SqlInjectMiddleWare struct {
}

func (x SqlInjectMiddleWare) Handle(ctx *Context) bool {

	body, err := ioutil.ReadAll(ctx.R.Body)
	if err != nil {
		return true
	}

	if !security.SqlInjectCheck(body, ctx.R.URL.RawQuery) {
		return true
	}

	return false
}

func (x SqlInjectMiddleWare) Enable() bool {
	return true
}

func (x SqlInjectMiddleWare) Name() string {
	return "SqlInjectMiddleWare"
}

type reusableReader struct {
	io.Reader
	readBuf *bytes.Buffer
	backBuf *bytes.Buffer
}

func ReusableReader(r io.Reader) io.Reader {
	readBuf := bytes.Buffer{}
	readBuf.ReadFrom(r) // error handling ignored for brevity
	backBuf := bytes.Buffer{}

	return reusableReader{
		io.TeeReader(&readBuf, &backBuf),
		&readBuf,
		&backBuf,
	}
}

func (r reusableReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if err == io.EOF {
		r.reset()
	}
	return n, err
}

func (r reusableReader) reset() {
	io.Copy(r.readBuf, r.backBuf) // nolint: errcheck
}

func RunMiddlewareX(d []MiddlewareX, ctx *Context) (Abort bool) {

	ctx.R.Body = io.NopCloser(ReusableReader(ctx.R.Body))

	body, err := ioutil.ReadAll(ctx.R.Body)
	if err != nil {
		logs.Info(err)
	}
	ctx.RawBody = body

	for i := 0; i < len(d); i++ {
		mx := d[i]
		if mx.Enable() {
			if mx.Handle(ctx) {
				return true
			}
		}
	}
	return
}

func JTWMiddleWarex(w http.ResponseWriter, r *http.Request) {

}
