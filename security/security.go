package security

import (
	"encoding/json"
	"html"
	"regexp"
	"strings"

	"github.com/unstoppablego/framework/logs"

	// "github.com/gin-gonic/gin"

	"github.com/microcosm-cc/bluemonday"
)

// func Decimal(value float64) float64 {
// 	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
// 	return value
// }

// 过滤xss 过滤sql注入
// func SafeMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		body, err := ioutil.ReadAll(r.Body)
// 		if err != nil {
// 			w.WriteHeader(400)
// 			logs.Error(err)
// 			return
// 		}

// 		getfilter := "'|\\b(alert|confirm|prompt)\\b|<[^>]*?>|^\\+\\/v(8|9)|\\b(and|or)\\b.+?(>|<|=|\\bin\\b|\\blike\\b)|\\/\\*.+?\\*\\/|<\\s*script\\b|\\bEXEC\\b|UNION.+?SELECT|UPDATE.+?SET|INSERT\\s+INTO.+?VALUES|(SELECT|DELETE).+?FROM|(CREATE|ALTER|DROP|TRUNCATE)\\s+(TABLE|DATABASE)"

// 		rc, _ := regexp.Compile("/" + getfilter + "/is")
// 		matched := rc.MatchString(string(body))
// 		if matched {
// 			logs.Error("sqlcheck fail", string(body))
// 			w.WriteHeader(400)
// 			return
// 		}
// 		postfilter := "^\\+\\/v(8|9)|\\b(and|or)\\b.{1,6}?(=|>|<|\\bin\\b|\\blike\\b)|\\/\\*.+?\\*\\/|<\\s*script\\b|\\bEXEC\\b|UNION.+?SELECT|UPDATE.+?SET|INSERT\\s+INTO.+?VALUES|(SELECT|DELETE).+?FROM|(CREATE|ALTER|DROP|TRUNCATE)\\s+(TABLE|DATABASE)"

// 		r1, _ := regexp.Compile("/" + postfilter + "/is")
// 		matched1 := r1.MatchString(string(body))
// 		if matched1 {
// 			logs.Error("sqlcheck fail", string(body))
// 			w.WriteHeader(400)
// 			return
// 		}

// 		cookiefilter := "\\b(and|or)\\b.{1,6}?(=|>|<|\\bin\\b|\\blike\\b)|\\/\\*.+?\\*\\/|<\\s*script\\b|\\bEXEC\\b|UNION.+?SELECT|UPDATE.+?SET|INSERT\\s+INTO.+?VALUES|(SELECT|DELETE).+?FROM|(CREATE|ALTER|DROP|TRUNCATE)\\s+(TABLE|DATABASE)"
// 		r2, _ := regexp.Compile("/" + cookiefilter + "/is")
// 		matched2 := r2.MatchString(string(body))
// 		if matched2 {
// 			logs.Error("sqlcheck fail", string(body))
// 			w.WriteHeader(400)
// 			return
// 		}

// 		p := bluemonday.UGCPolicy()
// 		if len(body) != 0 {
// 			sanitizedBody, err := XSSFilterJSON(p, string(body))
// 			if err != nil {
// 				logs.Error(err)
// 				w.WriteHeader(400)
// 				return
// 			}
// 			r.Body = ioutil.NopCloser(strings.NewReader(sanitizedBody))
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }

// 用于解码Center API 数据
// func SafeMiddlewareForDecodeAPI(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		body, err := ioutil.ReadAll(r.Body)
// 		if err != nil {
// 			w.WriteHeader(400)
// 			logs.Error(err)
// 			return
// 		}

// 		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
// 		next.ServeHTTP(w, r)

// 	})
// }

// func JsonXSS(c *gin.Context) {
// 	p := bluemonday.UGCPolicy()
// 	body, err := c.GetRawData()
// 	if err != nil {
// 		logs.Error(err)
// 		c.Abort()
// 		return
// 	}
// 	if len(body) == 0 {
// 		c.Next()
// 		return
// 	}
// 	sanitizedBody, err := XSSFilterJSON(p, string(body))
// 	if err != nil {
// 		logs.Error(err)
// 		c.Abort()
// 		return
// 	}
// 	c.Request.Body = ioutil.NopCloser(strings.NewReader(sanitizedBody))

// 	c.Next()
// }

func XSSFilterJSON(p *bluemonday.Policy, s string) (string, error) {
	var data interface{}
	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		logs.Info("Not JSON Data", s)
		return s, nil
	}

	b := strings.Builder{}
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	err = e.Encode(xssFilterJSONData(p, data))
	if err != nil {
		return "", err
	}
	// use `TrimSpace` to trim newline char add by `Encode`.
	return strings.TrimSpace(b.String()), nil
}

func xssFilterJSONData(p *bluemonday.Policy, data interface{}) interface{} {
	if s, ok := data.([]interface{}); ok {
		for i, v := range s {
			s[i] = xssFilterJSONData(p, v)
		}
		return s
	} else if m, ok := data.(map[string]interface{}); ok {
		for k, v := range m {
			m[k] = xssFilterJSONData(p, v)
		}
		return m
	} else if str, ok := data.(string); ok {
		return XSSFilterPlain(p, str)
	}
	return data
}

func XSSFilterPlain(p *bluemonday.Policy, s string) string {
	sanitized := p.Sanitize(s)
	return html.UnescapeString(sanitized)
}

func SqlInjectCheck(body []byte, query string) (safe bool) {

	// logs.Info(query)
	// logs.Info(string(body))

	var getfilter = `'|\b(alert|confirm|prompt)\b|<[^>]*?>|^\+\/v(8|9)|\b(and|or)\b.+?(>|<|=|\bin\b|\blike\b)|\/\*.+?\*\/|<\s*script\b|\bEXEC\b|UNION.+?SELECT|UPDATE.+?SET|INSERT\s+INTO.+?VALUES|(SELECT|DELETE).+?FROM|(CREATE|ALTER|DROP|TRUNCATE)\s+(TABLE|DATABASE)`
	var postfilter = `^\+\/v(8|9)|\b(and|or)\b.{1,6}?(=|>|<|\bin\b|\blike\b)|\/\*.+?\*\/|<\s*script\b|\bEXEC\b|UNION.+?SELECT|UPDATE.+?SET|INSERT\s+INTO.+?VALUES|(SELECT|DELETE).+?FROM|(CREATE|ALTER|DROP|TRUNCATE)\s+(TABLE|DATABASE)`

	// var cookiefilter = `\b(and|or)\b.{1,6}?(=|>|<|\bin\b|\blike\b)|\/\*.+?\*\/|<\s*script\b|\bEXEC\b|UNION.+?SELECT|UPDATE.+?SET|INSERT\s+INTO.+?VALUES|(SELECT|DELETE).+?FROM|(CREATE|ALTER|DROP|TRUNCATE)\s+(TABLE|DATABASE)`

	if len(query) > 0 {
		r, err := regexp.Compile("(?is)" + getfilter + "")
		if err != nil {
			logs.Error(err)
			return true
		}
		matched := r.MatchString(string(query))
		if matched {
			logs.Error("QUERY SQL Inject:", string(query))
			return
		}
	}

	if len(body) > 0 {
		r1, err := regexp.Compile("(?is)" + postfilter + "")
		if err != nil {
			logs.Error(err)
			return true
		}
		matched1 := r1.MatchString(string(body))
		if matched1 {
			logs.Error("BODY SQL Inject:", string(body))
			return
		}
	}

	return true
}

// func CheckString(c *gin.Context) {
// 	// body, err := c.GetRawData()
// 	// if err != nil {
// 	// 	logs.Error(err)
// 	// 	c.Abort()
// 	// 	return
// 	// }
// 	// if len(body) == 0 {
// 	// 	c.Next()
// 	// 	return
// 	// }
// 	// needFind := []string{"\\", "'", "/", "..", "../", "./", "//", ";"}
// 	// for _, v := range needFind {
// 	// 	checkResult := strings.Contains(string(body), v)
// 	// 	if checkResult {
// 	// 		logs.Error("string check fail", string(body))
// 	// 		c.Abort()
// 	// 		return
// 	// 	}
// 	// }
// 	// c.Request.Body = ioutil.NopCloser(strings.NewReader(string(body)))
// 	// c.Next()
// }

// func XXX() {
// 	// var getfilter = `'|\b(alert|confirm|prompt)\b|<[^>]*?>|^\\+\/v(8|9)|\\b(and|or)\\b.+?(>|<|=|\\bin\\b|\\blike\\b)|\\/\\*.+?\\*\\/|<\\s*script\\b|\\bEXEC\\b|UNION.+?SELECT|UPDATE.+?SET|INSERT\\s+INTO.+?VALUES|(SELECT|DELETE).+?FROM|(CREATE|ALTER|DROP|TRUNCATE)\\s+(TABLE|DATABASE)`
// 	// // $postfilter="^\\+\/v(8|9)|\\b(and|or)\\b.{1,6}?(=|>|<|\\bin\\b|\\blike\\b)|\\/\\*.+?\\*\\/|<\\s*script\\b|<\\s*img\\b|\\bEXEC\\b|UNION.+?SELECT|UPDATE.+?SET|INSERT\\s+INTO.+?VALUES|(SELECT|DELETE).+?FROM|(CREATE|ALTER|DROP|TRUNCATE)\\s+(TABLE|DATABASE)";
// 	// var postfilter = `^\\+\/v(8|9)|\\b(and|or)\\b.{1,6}?(=|>|<|\\bin\\b|\\blike\\b)|\\/\\*.+?\\*\\/|<\\s*script\\b|\\bEXEC\\b|UNION.+?SELECT|UPDATE.+?SET|INSERT\\s+INTO.+?VALUES|(SELECT|DELETE).+?FROM|(CREATE|ALTER|DROP|TRUNCATE)\\s+(TABLE|DATABASE)`
// 	// var cookiefilter = `\\b(and|or)\\b.{1,6}?(=|>|<|\\bin\\b|\\blike\\b)|\\/\\*.+?\\*\\/|<\\s*script\\b|\\bEXEC\\b|UNION.+?SELECT|UPDATE.+?SET|INSERT\\s+INTO.+?VALUES|(SELECT|DELETE).+?FROM|(CREATE|ALTER|DROP|TRUNCATE)\\s+(TABLE|DATABASE)`

// 	var getfilter = `'|\b(alert|confirm|prompt)\b|<[^>]*?>|^\+\/v(8|9)|\b(and|or)\b.+?(>|<|=|\bin\b|\blike\b)|\/\*.+?\*\/|<\s*script\b|\bEXEC\b|UNION.+?SELECT|UPDATE.+?SET|INSERT\s+INTO.+?VALUES|(SELECT|DELETE).+?FROM|(CREATE|ALTER|DROP|TRUNCATE)\s+(TABLE|DATABASE)`
// 	var postfilter = `^\+\/v(8|9)|\b(and|or)\b.{1,6}?(=|>|<|\bin\b|\blike\b)|\/\*.+?\*\/|<\s*script\b|\bEXEC\b|UNION.+?SELECT|UPDATE.+?SET|INSERT\s+INTO.+?VALUES|(SELECT|DELETE).+?FROM|(CREATE|ALTER|DROP|TRUNCATE)\s+(TABLE|DATABASE)`
// 	var cookiefilter = `\b(and|or)\b.{1,6}?(=|>|<|\bin\b|\blike\b)|\/\*.+?\*\/|<\s*script\b|\bEXEC\b|UNION.+?SELECT|UPDATE.+?SET|INSERT\s+INTO.+?VALUES|(SELECT|DELETE).+?FROM|(CREATE|ALTER|DROP|TRUNCATE)\s+(TABLE|DATABASE)`

// 	fmt.Println(getfilter)
// 	fmt.Println(postfilter)
// 	fmt.Println(cookiefilter)
// 	logs.Info(SqlCheck([]byte(` select 1,2,TABLE_NAME from information_schema.TABLES where TABLE_SCHEMA = 'news' #`), ""))

// }
