package mhttp

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"net/textproto"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	// "vclinet/m/v2/pkg/logs"

	"github.com/fatih/color"
	"github.com/unstoppablego/framework/logs"
	"golang.org/x/net/http/httpguts"
)

//client pool ?
//https://stackoverflow.com/questions/58228560/is-there-a-way-to-use-a-specific-tcpconn-to-make-http-requests
//特定的tcp
//

type ConnectManager struct {
	PreConnect          int   //15s 开始生成密钥
	TcpConnectTimeTotal int64 //15ms 计算常规往返时间 = meth not find 5次
	TcpGatewayTimeTotal int64
	TcpConnectTimeAvg   int64
	TcpGatewayTimeAvg   int64
	KeyAry              []string //密钥
	ClientNum           int
	CodeRunTime         int64
	CodeRunTimeSuccess  int64
	CodeRunTimeAvg      int64
	ClientAry           []*http.Client //for bill
	ClientTcp           []*http.Client
	Cookies             map[string]http.Cookie
	TlsCon              []*tls.Conn
	TlsMux              *sync.Mutex
	TlsUsed             int
	Ip                  string
	Yz_log_seqn         int
	ConnectIp           map[string]int
	// Client2             *http.Client
	// Client3             *http.Client
	// Client4             *http.Client
	// Client5             *http.Client
	// PreCalcTimeTotal int //5s 计算未开售时间1次 同时产生预处理连接
	// PreCalcCodeTimes int //计算成功次数
	// Cookie 目前待定
	Mux *sync.Mutex
}

func GetConnectManager(PreNum int) ConnectManager {
	var h ConnectManager
	h.PreConnect = PreNum
	h.ClientNum = 300
	for i := 0; i < 5; i++ {
		h.ClientTcp = append(h.ClientTcp, CreateHTTPClient())
	}
	for i := 0; i < h.ClientNum; i++ {
		h.ClientAry = append(h.ClientAry, CreateHTTPClient())
	}
	h.Mux = &sync.Mutex{}
	h.TlsMux = &sync.Mutex{}
	h.Cookies = make(map[string]http.Cookie)
	h.ConnectIp = make(map[string]int)
	return h
}

func GetHttpRespBody(initResp *http.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(initResp.Body)
	initResp.Body.Close()
	return data, err
}

func (cm *ConnectManager) GetCookieString() string {
	var cookies string
	for _, v := range cm.Cookies {
		cookies += fmt.Sprintf(" %s=%s;", v.Name, v.Value)
	}
	cookies = cookies[1 : len(cookies)-1]
	return cookies
}

func (cm *ConnectManager) GetCookieStringForNewTime() string {
	var cookies string
	for _, v := range cm.Cookies {
		if v.Name == "yz_log_ftime" {
			v.Value = strconv.Itoa(int(time.Now().UnixNano() / 1000))
		}
		if v.Name == "yz_log_seqb" {
			v.Value = strconv.Itoa(int(time.Now().UnixNano()/1000 - 3052925))
		}
		if v.Name == "open_token" || v.Name == "loc_dfp" || v.Name == "trace_sdk_context_dc_ps" || v.Name == "trace_sdk_context_dc_ps_utime" {
			continue
		}
		if v.Name == "yz_log_seqn" {
			num, err := strconv.Atoi(v.Value)
			if err != nil {
				logs.Info(err)
			} else {
				cm.Yz_log_seqn += 2
				v.Value = strconv.Itoa(num + cm.Yz_log_seqn)
			}

		}

		cookies += fmt.Sprintf(" %s=%s;", v.Name, v.Value)
	}
	cookies = cookies[1 : len(cookies)-1]
	return cookies
}

func (cm ConnectManager) GetCookieValue(name string) string {
	if v, ok := cm.Cookies[name]; ok {
		return v.Value
	}
	return ""
}

// 计算出TCP连接耗时 nana
func (cm *ConnectManager) GetTcpTime(url string) int64 {
	var successTimes int64
	for i := 0; i < len(cm.ClientTcp); i++ {
		init2req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			logs.Info(err)
			continue
		}
		init2req.Header.Add("Connection", "keep-alive")

		connectTime, serverProcessTime, err, _ := CreateHttpTrace(init2req, cm.ClientTcp[i])

		if err == nil {
			cm.TcpConnectTimeTotal += connectTime.Nanoseconds()
			cm.TcpGatewayTimeTotal += serverProcessTime.Nanoseconds()
			successTimes++
		}
	}
	cm.TcpConnectTimeAvg = cm.TcpConnectTimeTotal / successTimes
	cm.TcpGatewayTimeAvg = cm.TcpGatewayTimeTotal / successTimes
	return cm.TcpConnectTimeAvg
}

func (cm ConnectManager) CreateWaitTimeNano(StartSoldTimeNano int64) (int64, int64) {
	nowtime := time.Now().UnixNano()
	waitTime := StartSoldTimeNano - nowtime - (cm.TcpConnectTimeAvg + cm.CodeRunTimeAvg)
	//200+100
	logs.Info("Wait Time", time.Unix((nowtime+waitTime)/1e9, (nowtime+waitTime)%1e9), "before ", (cm.TcpConnectTimeAvg+cm.CodeRunTimeAvg)/1e6)
	return waitTime, nowtime
}

func (cm ConnectManager) CreateWaitRealTimeNano(StartSoldTimeNano int64) (int64, int64) {
	nowtime := time.Now().UnixNano()
	waitTime := StartSoldTimeNano - nowtime
	//200+100
	logs.Info("Wait Time", time.Unix((nowtime+waitTime)/1e9, (nowtime+waitTime)%1e9), "before ", (cm.TcpConnectTimeAvg+cm.CodeRunTimeAvg)/1e6)
	return waitTime, nowtime
}

// coderun = code平均值 - 往返平均值
// 发送均值=coderun+发送均值0.8 -0.3 均值误差 需要小于10ms 总误差可在100ms 内 即100个请求
// 每隔1ms 发送 1批请求 一共发送5皮

// 计算理论距离时间
const (
	MaxIdleConns        int = 300
	MaxIdleConnsPerHost int = 300
	IdleConnTimeout     int = 90
)

// 创建特定的 HTTP Client
// var testClient *http.Client
func CreateHTTPClient() *http.Client {
	// if testClient == nil {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   15 * time.Second,
				KeepAlive: 15 * time.Second,
			}).DialContext,

			MaxIdleConns:        MaxIdleConns,
			MaxIdleConnsPerHost: MaxIdleConnsPerHost,
			IdleConnTimeout:     time.Duration(IdleConnTimeout) * time.Second,
			TLSHandshakeTimeout: 6 * time.Second,
		},
		Timeout: 45 * time.Second,
	}
}

// func Create() {
// 	c := CreateHTTPClient()
// 	// c.Transport.RoundTrip()
// 	// c.Do()s
// }

func CreateHttpTrace(req *http.Request, client *http.Client) (time.Duration, time.Duration, error, []*http.Cookie) {
	xurl := req.URL
	var startTime = time.Now()
	var t0, t1, t2, t3, t4, t5, t6 time.Time
	// var reused bool
	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { t0 = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { t1 = time.Now() },
		ConnectStart: func(_, _ string) {
			if t1.IsZero() {
				// connecting to IP
				t1 = time.Now()
			}
		},
		ConnectDone: func(net, addr string, err error) {
			if err != nil {
				log.Printf("unable to connect to host %v: %v\ns", addr, err)
			}
			t2 = time.Now()

			printf("\n%s%s %s\n", color.GreenString("Connected to "), color.CyanString(addr), xurl)
		},
		GotConn: func(g httptrace.GotConnInfo) {
			t3 = time.Now()
			printf("%s %v\n", color.GreenString("Connect reused:"), g.Reused)
		},
		GotFirstResponseByte: func() { t4 = time.Now() },
		TLSHandshakeStart:    func() { t5 = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { t6 = time.Now() },
	}

	req = req.WithContext(httptrace.WithClientTrace(context.Background(), trace))

	resp, err := client.Do(req)
	if err != nil {
		logs.Info(err)
		return 0, 0, err, nil
	}

	data, err := ioutil.ReadAll(resp.Body)

	cookies := ReadSetCookies(resp.Header)
	logs.Info(string(data))
	if err != nil {
		logs.Info(err)
		return 0, 0, err, nil
	}
	resp.Body.Close()

	// resp.Header.Get("cookies")

	if t0.IsZero() {
		// we skipped DNS
		t0 = startTime
	}
	if t1.IsZero() {
		t1 = startTime
	}
	if t2.IsZero() {
		t2 = startTime
	}

	t7 := time.Now()
	connecttime := t2.Sub(t1)
	serverProcess := t4.Sub(t3)
	printf(colorize(httpsTemplate),
		fmta(t1.Sub(t0)),    // dns lookup
		fmta(connecttime),   // tcp connection
		fmta(t6.Sub(t5)),    // tls handshake
		fmta(serverProcess), // server processing
		fmta(t7.Sub(t4)),    // content transfer
		fmtb(t1.Sub(t0)),    // namelookup
		fmtb(t2.Sub(t0)),    // connect
		fmtb(t3.Sub(t0)),    // pretransfer
		fmtb(t4.Sub(t0)),    // starttransfer
		fmtb(t7.Sub(t0)),    // total
	)
	logs.Info(connecttime, serverProcess)
	return connecttime, serverProcess, nil, cookies
}

const httpsTemplate = `` +
	`  DNS Lookup   TCP Connection   TLS Handshake   Server Processing   Content Transfer` + "\n" +
	`[%s  |     %s  |    %s  |        %s  |       %s  ]` + "\n" +
	`            |                |               |                   |                  |` + "\n" +
	`   namelookup:%s      |               |                   |                  |` + "\n" +
	`                       connect:%s     |                   |                  |` + "\n" +
	`                                   pretransfer:%s         |                  |` + "\n" +
	`                                                     starttransfer:%s        |` + "\n" +
	`                                                                                total:%s` + "\n"

func colorize(s string) string {
	v := strings.Split(s, "\n")
	v[0] = grayscale(16)(v[0])
	return strings.Join(v, "\n")
}

func grayscale(code color.Attribute) func(string, ...interface{}) string {
	return color.New(code + 232).SprintfFunc()
}

func fmta(d time.Duration) string {
	return color.CyanString("%7dms", int(d/time.Millisecond))
}

func fmtb(d time.Duration) string {
	return color.CyanString("%-9s", strconv.Itoa(int(d/time.Millisecond))+"ms")
}

func printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(color.Output, format, a...)
}

func CacheDNS(domain string) {
	ip, err := net.LookupHost(domain)
	if err != nil {
		logs.Info(err)
		return
	}
	cacheDns := ip[0]
	cmda := fmt.Sprintf("echo '%s %s'>>/etc/hosts", cacheDns, domain)
	out, err := exec.Command("bash", "-c", cmda).Output()
	logs.Info(out, ip, err)
}

func ReadSetCookies(h http.Header) []*http.Cookie {
	cookieCount := len(h["Set-Cookie"])
	if cookieCount == 0 {
		return []*http.Cookie{}
	}
	cookies := make([]*http.Cookie, 0, cookieCount)
	for _, line := range h["Set-Cookie"] {
		parts := strings.Split(textproto.TrimString(line), ";")
		if len(parts) == 1 && parts[0] == "" {
			continue
		}
		parts[0] = textproto.TrimString(parts[0])
		j := strings.Index(parts[0], "=")
		if j < 0 {
			continue
		}
		name, value := parts[0][:j], parts[0][j+1:]
		if !isCookieNameValid(name) {
			continue
		}
		value, ok := parseCookieValue(value, true)
		if !ok {
			continue
		}
		c := &http.Cookie{
			Name:  name,
			Value: value,
			Raw:   line,
		}
		for i := 1; i < len(parts); i++ {
			parts[i] = textproto.TrimString(parts[i])
			if len(parts[i]) == 0 {
				continue
			}

			attr, val := parts[i], ""
			if j := strings.Index(attr, "="); j >= 0 {
				attr, val = attr[:j], attr[j+1:]
			}
			lowerAttr := strings.ToLower(attr)
			val, ok = parseCookieValue(val, false)
			if !ok {
				c.Unparsed = append(c.Unparsed, parts[i])
				continue
			}
			switch lowerAttr {
			case "samesite":
				lowerVal := strings.ToLower(val)
				switch lowerVal {
				case "lax":
					c.SameSite = http.SameSiteLaxMode
				case "strict":
					c.SameSite = http.SameSiteStrictMode
				case "none":
					c.SameSite = http.SameSiteNoneMode
				default:
					c.SameSite = http.SameSiteDefaultMode
				}
				continue
			case "secure":
				c.Secure = true
				continue
			case "httponly":
				c.HttpOnly = true
				continue
			case "domain":
				c.Domain = val
				continue
			case "max-age":
				secs, err := strconv.Atoi(val)
				if err != nil || secs != 0 && val[0] == '0' {
					break
				}
				if secs <= 0 {
					secs = -1
				}
				c.MaxAge = secs
				continue
			case "expires":
				c.RawExpires = val
				exptime, err := time.Parse(time.RFC1123, val)
				if err != nil {
					exptime, err = time.Parse("Mon, 02-Jan-2006 15:04:05 MST", val)
					if err != nil {
						c.Expires = time.Time{}
						break
					}
				}
				c.Expires = exptime.UTC()
				continue
			case "path":
				c.Path = val
				continue
			}
			c.Unparsed = append(c.Unparsed, parts[i])
		}
		cookies = append(cookies, c)
	}
	return cookies
}

func parseCookieValue(raw string, allowDoubleQuote bool) (string, bool) {
	// Strip the quotes, if present.
	if allowDoubleQuote && len(raw) > 1 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1 : len(raw)-1]
	}
	for i := 0; i < len(raw); i++ {
		if !validCookieValueByte(raw[i]) {
			return "", false
		}
	}
	return raw, true
}

func validCookieValueByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
}

func isCookieNameValid(raw string) bool {
	if raw == "" {
		return false
	}
	return strings.IndexFunc(raw, isNotToken) < 0
}

func isNotToken(r rune) bool {
	return !httpguts.IsTokenRune(r)
}

// To system Cookie?
type GetCookie struct {
	Name         string `json:"name"`
	Value        string `json:"value"`
	Domain       string `json:"domain"`
	Path         string `json:"path"`
	Expires      int    `json:"expires"`
	Size         int    `json:"size"`
	HTTPOnly     bool   `json:"httpOnly"`
	Secure       bool   `json:"secure"`
	Session      bool   `json:"session"`
	SameParty    bool   `json:"sameParty"`
	SourceScheme string `json:"sourceScheme"`
	SourcePort   int    `json:"sourcePort"`
}

func (c GetCookie) GetCookieToCookie() http.Cookie {
	var sc http.Cookie
	sc.Name = c.Name
	sc.Value = c.Value
	sc.Domain = c.Domain
	sc.Secure = c.Secure
	sc.HttpOnly = c.HTTPOnly
	return sc
}

func (cm *ConnectManager) InitTcpSocket(urlstr string, num int) {
	wg := sync.WaitGroup{}
	wg.Add(num)
	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		var sleepTime time.Duration = time.Duration(rand.Intn(5)) * time.Millisecond

		time.Sleep(sleepTime)
		go func() {
			defer wg.Done()
			urlx, err := url.Parse(urlstr)
			if err != nil {
				logs.Info(err)
				return
			}

			conf := &tls.Config{
				InsecureSkipVerify: true,
			}

			ipa, err := net.LookupHost(urlx.Host)
			if err != nil {
				logs.Info(err)
				return
			}

			startip := ""
			startNum := 0
			cm.TlsMux.Lock()
			for _, v := range ipa {
				if _, ok := cm.ConnectIp[v]; !ok {
					cm.ConnectIp[v] = 0
				}
			}
			for ip, num := range cm.ConnectIp {
				if startip == "" {
					startip = ip
					startNum = num
				} else {
					if startNum > num {
						startip = ip
					}
				}
			}
			cm.ConnectIp[startip] = cm.ConnectIp[startip] + 1
			cm.TlsMux.Unlock()

			cm.Ip = startip
			conn, err := tls.Dial("tcp", startip+":443", conf)
			if err != nil {
				log.Println(err)
				return
			}
			cm.TlsMux.Lock()
			cm.TlsCon = append(cm.TlsCon, conn)
			cm.TlsMux.Unlock()
		}()
	}
	wg.Wait()
}

func (cm *ConnectManager) HttpSendAndRecv(req *http.Request) ([]byte, time.Duration, string) {
	// defer func() { // 必须要先声明defer，否则不能捕获到panic异常
	// 	if errx := recover(); errx != nil {
	// 		fmt.Println(errx) // 这里的err其实就是panic传入的内容
	// 		// err = fmt.Errorf(fmt.Sprintf("%v", errx))
	// 	}
	// }()

	nowtiem := time.Now()
	var index int
	cm.TlsMux.Lock()
	index = cm.TlsUsed
	cm.TlsUsed++
	cm.TlsMux.Unlock()
	if len(cm.TlsCon)-1 < index {
		// logs.Info("no connect", index)
		return nil, time.Duration(1), ""
	}

	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		logs.Info(err)
		return nil, time.Duration(1), ""
	}

	// if len(cm.TlsCon) <= 10 {

	// } else {
	// 	index = len(cm.TlsCon) - index
	// }

	//后面创建的 TLS 比较稳定
	sl, err := cm.TlsCon[index].Write(dump)
	if err != nil {
		logs.Info(err, sl)
		return nil, time.Duration(1), ""
	}
	resp, err := http.ReadResponse(bufio.NewReader(cm.TlsCon[index]), req)
	if err != nil {
		logs.Info(err)
		return nil, time.Duration(1), ""
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Info(err)
		return nil, time.Duration(1), ""
	}
	subtime := time.Since(nowtiem)

	// fmt.Println(, subtime)
	ip := cm.TlsCon[index].RemoteAddr().String()

	cm.TlsCon[index].Close()

	return data, subtime, ip
}

func (cm *ConnectManager) HttpSendAndRecvNewTcp(req *http.Request) ([]byte, time.Duration) {
	nowtiem := time.Now()
	// var index int
	// cm.TlsMux.Lock()
	// index = cm.TlsUsed
	// cm.TlsUsed++
	// cm.TlsMux.Unlock()
	// if len(cm.TlsCon)-1 < index {
	// 	logs.Info("no connect")
	// 	return nil, time.Duration(1)
	// }

	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		logs.Info(err)
		return nil, time.Duration(1)
	}

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", cm.Ip+":443", conf)
	if err != nil {
		log.Println(err)
		return nil, time.Since(nowtiem)
	}

	sl, err := conn.Write(dump)
	if err != nil {
		logs.Info(err, sl)
		return nil, time.Duration(1)
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), req)

	if err != nil {
		logs.Info(err)
		return nil, time.Duration(1)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Info(err)
		return nil, time.Duration(1)
	}
	subtime := time.Since(nowtiem)
	conn.Close()

	return data, subtime
}
