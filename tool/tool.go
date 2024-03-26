package tool

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/unstoppablego/framework/logs"

	uuid "github.com/nu7hatch/gouuid"
)

// ipList := []string{"192.168.1.1/24", "fd04:3e42:4a4e:3381::/64"}
//     for i := 0; i < len(ipList); i += 1 {
//         ip, ipnet, err := net.ParseCIDR(ipList[i])
//         if err != nil {
//             fmt.Println("Error", ipList[i], err)
//             continue
//         }
//         fmt.Println(ipList[i], "-> ip:", ip, " net:", ipnet)
//     }

func GetLocalIp() net.IP {
	resp, _ := http.Get("http://icanhazip.com")
	ipstring, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(ipstring))
	return net.ParseIP(string(ipstring))
}

func VerifyEmailFormat(email string) bool {
	// pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	pattern := `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

func CreateUUID() string {
	v4, err := uuid.NewV4()
	if err != nil {
		logs.Info(err)
	}
	return v4.String()
}

func CreateHTTPClient() *http.Client {
	const (
		MaxIdleConns        int = 300
		MaxIdleConnsPerHost int = 300
		IdleConnTimeout     int = 90
	)
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
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 45 * time.Second,
	}
}

func RunUntilReturnTrue(a func(b string, c string) bool, arg1 string, arg2 string) {
	for {
		if a(arg1, arg2) {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

// [0,max)
func Rand(num int64) *big.Int {
	result, _ := rand.Int(rand.Reader, big.NewInt(num))
	return result
}

func GetMacAddrs() (macAddrs []string) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("fail to get net interfaces: %v", err)
		return macAddrs
	}

	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		if len(macAddr) == 0 {
			continue
		}

		macAddrs = append(macAddrs, macAddr)
	}
	return macAddrs
}

func GetMacAddrsString() (macAddrs string) {
	addr := GetMacAddrs()
	var stra string
	if len(addr) > 0 {
		for _, v := range addr {
			stra = stra + v
		}
	}
	return stra
}

// Keys generates a new P256 ECDSA public private key pair for TLS.
// It returns a bytes buffer for the PEM encoded private key and certificate.
func Keys(validFor time.Duration) (cert, key *bytes.Buffer, fingerprint [32]byte, err error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
		return nil, nil, fingerprint, err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
		return nil, nil, fingerprint, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Let's Encrypt"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		// IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("101.32.208.3")},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}

	// template.IPAddresses = append(template.IPAddresses, GetLocalIp())

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
		return nil, nil, fingerprint, err
	}

	// Encode and write certificate and key to bytes.Buffer
	cert = bytes.NewBuffer([]byte{})
	pem.Encode(cert, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	key = bytes.NewBuffer([]byte{})
	pem.Encode(key, pemBlockForKey(privKey))

	fingerprint = sha256.Sum256(derBytes)

	return cert, key, fingerprint, nil //TODO: maybe return a struct instead of 4 multiple return items
}

func pemBlockForKey(key *ecdsa.PrivateKey) *pem.Block {
	b, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
		os.Exit(2)
	}
	return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
}

// 定时执行函数
func Crontab(t time.Duration, f func(), fname string) {
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				logs.Error(r, fname)
				var buf [4096]byte
				n := runtime.Stack(buf[:], false)
				fmt.Println(string(buf[:n]))
			}
			Crontab(t, f, fname)
		}()
		for {
			f()
			time.Sleep(t)
		}
	}()
}

func Rsa() {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal("Private key cannot be created.", err.Error())
	}

	// Generate a pem block with the private key
	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	tml := x509.Certificate{
		// you can add any attr that you need
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(5, 0, 0),
		// you have to generate a different serial number each execution
		SerialNumber: big.NewInt(123123),
		Subject: pkix.Name{
			CommonName:   "New Name",
			Organization: []string{"New Org."},
		},
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	cert, err := x509.CreateCertificate(rand.Reader, &tml, &tml, &key.PublicKey, key)
	if err != nil {
		log.Fatal("Certificate cannot be created.", err.Error())
	}

	// Generate a pem block with the certificate
	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})
	log.Println(certPem, keyPem)
}

// package main

// import (
// 	"bytes"
// 	"crypto/rand"
// 	"crypto/rsa"
// 	"crypto/tls"
// 	"crypto/x509"
// 	"crypto/x509/pkix"
// 	"encoding/pem"
// 	"fmt"
// 	"io/ioutil"
// 	"math/big"
// 	"net"
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"time"
// )

func main2() {
	// get our ca and server certificate
	serverTLSConf, clientTLSConf, err := Certsetup()
	if err != nil {
		panic(err)
	}

	// set up the httptest.Server using our certificate signed by our CA
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "success!")
	}))
	server.TLS = serverTLSConf
	server.StartTLS()
	defer server.Close()

	// communicate with the server using an http.Client configured to trust our CA
	transport := &http.Transport{
		TLSClientConfig: clientTLSConf,
	}
	http := http.Client{
		Transport: transport,
	}
	resp, err := http.Get(server.URL)
	if err != nil {
		panic(err)
	}

	// verify the response
	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	body := strings.TrimSpace(string(respBodyBytes[:]))
	if body == "success!" {
		fmt.Println(body)
	} else {
		panic("not successful!")
	}
}

// https://tech.my-netsol.com/?p=179 关于mac的证书标准要求
func Certsetup() (serverTLSConf *tls.Config, clientTLSConf *tls.Config, err error) {
	// set up our CA certificate
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Country:      []string{"US"},
			Organization: []string{"Test INC."},
			Province:     []string{""},
			// SerialNumber: big.NewInt(2019).String(),
			// CommonName:   "CommonName",
			// Locality:     []string{"San Francisco"},
			// StreetAddress: []string{"Golden Gate Bridge"},
			// PostalCode:    []string{"94016"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(2, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// create our private and public key
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	// pem encode
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	f, err := os.Create("./111.crt")
	f.Write(caPEM.Bytes())
	fmt.Println(caPEM.String(), string(base64.StdEncoding.EncodeToString(caBytes)))
	caPEM.Bytes()

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	// set up our server certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{"mactest inc"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94016"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback, net.ParseIP("101.32.208.3")},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(2, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	keyBytes := x509.MarshalPKCS1PublicKey(&certPrivKey.PublicKey)
	keyHash := sha256.Sum256(keyBytes)
	ski := keyHash[:]
	cert.SubjectKeyId = ski

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())
	if err != nil {
		return nil, nil, err
	}

	serverTLSConf = &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}

	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(caPEM.Bytes())
	clientTLSConf = &tls.Config{
		RootCAs: certpool,
	}

	return
}

func Md5(str string) string {
	data := []byte(str)
	return fmt.Sprintf("%x", md5.Sum(data))
}

func Md5b(str []byte) string {
	// data := []byte(str)
	return fmt.Sprintf("%x", md5.Sum(str))
}

// 查询国家编码是否规范用,本质上没啥用
func CheckISOCode(keyword string) string {
	// ISO 3166-1
	cc := map[string]string{
		"BR": "巴西",
		"IN": "印度",
		"CI": "科特迪瓦",
		"VN": "越南",
		"BO": "玻利维亚",
		"KR": "韩国",
		"AT": "奥地利",
		"RO": "罗马尼亚",
		"ID": "印度尼西亚",
		"IR": "伊朗",
		"TH": "泰国",
		"CN": "中国",
		"RS": "塞尔维亚",
		"TW": "中国台湾",
		"MF": "法属圣马丁",
		"LT": "立陶宛",
		"MY": "马来西亚",
		"EG": "埃及",
		"RU": "俄罗斯",
		"JO": "约旦",
		"MA": "摩洛哥",
		"QA": "卡塔尔",
		"CO": "哥伦比亚",
		"KZ": "哈萨克斯坦",
		"DE": "德国",
		"GB": "英国",
		"PK": "巴基斯坦",
		"TR": "土耳其",
		"UY": "乌拉圭",
		"HU": "匈牙利",
		"IE": "爱尔兰",
		"PR": "波多黎各",
		"UA": "乌克兰",
		"IT": "意大利",
		"FR": "法国",
		"US": "美国",
		"JP": "日本",
		"MX": "墨西哥",
		"PL": "波兰",
		"AM": "亚美尼亚",
		"ES": "西班牙",
		"BD": "孟加拉",
		"NZ": "新西兰",
		"AL": "阿尔巴尼亚",
		"SG": "新加坡",
		"CY": "塞浦路斯",
		"ZA": "南非",
		"AR": "阿根廷",
		"BG": "保加利亚",
		"IL": "以色列",
		"VE": "委内瑞拉",
		"TN": "突尼斯",
		"CL": "智利",
		"AE": "阿联酋",
		"DO": "多米尼加",
		"CW": "库拉索",
		"TJ": "塔吉克斯坦",
		"PY": "巴拉圭",
		"WS": "萨摩亚",
		"OM": "阿曼",
		"AU": "澳大利亚",
		"CH": "瑞士",
		"GR": "希腊",
		"PE": "秘鲁",
		"BA": "波斯尼亚和黑塞哥维那",
		"DZ": "阿尔及利亚",
		"SI": "斯洛文尼亚",
		"NL": "荷兰",
		"HN": "洪都拉斯",
		"SA": "沙特阿拉伯",
		"LB": "黎巴嫩",
		"LK": "斯里兰卡",
		"BH": "巴林",
		"MK": "北马其顿",
		"KW": "科威特",
		"PH": "菲律宾",
		"TT": "特立尼达和多巴哥",
		"MN": "蒙古",
		"PT": "葡萄牙",
		"UZ": "乌兹别克斯坦",
		"PA": "巴拿马",
		"HR": "克罗地亚",
		"SR": "苏里南",
		"FI": "芬兰",
		"KH": "柬埔寨",
		"DM": "多米尼克",
		"SV": "萨尔瓦多",
		"MD": "摩尔多瓦",
		"LI": "列支敦士登",
		"HK": "中国香港",
		"SK": "斯洛伐克",
		"BS": "巴哈马",
		"CR": "哥斯达黎加",
		"JE": "泽西岛",
		"AZ": "阿塞拜疆",
		"SN": "塞内加尔",
		"XK": "科索沃",
		"IQ": "伊拉克",
		"GE": "格鲁吉亚",
		"CA": "加拿大",
		"BN": "文莱",
		"LC": "圣卢西亚",
		"BY": "白俄罗斯",
		"MT": "马耳他",
		"MO": "中国澳门",
		"TZ": "坦桑尼亚",
		"MU": "毛里求斯",
		"GP": "瓜德罗普",
		"BE": "比利时",
		"NP": "尼泊尔",
		"AO": "安哥拉",
		"YE": "也门",
		"KG": "吉尔吉斯斯坦",
		"EC": "厄瓜多尔",
		"SE": "瑞典",
		"ME": "黑山",
		"LV": "拉脱维亚",
		"MQ": "马提尼克",
		"RE": "留尼汪岛",
		"SD": "苏丹",
		"GY": "圭亚那",
		"CM": "喀麦隆",
		"VI": "美属维尔京群岛",
		"NG": "尼日利亚",
		"AD": "安道尔",
		"LA": "老挝",
		"CZ": "捷克",
		"JM": "牙买加",
		"KE": "肯尼亚",
		"IM": "马恩岛",
		"GT": "危地马拉",
		"PS": "巴勒斯坦",
		"GA": "加蓬",
		"DK": "丹麦",
		"NI": "尼加拉瓜",
		"GU": "关岛",
		"MV": "马尔代夫",
		"KN": "圣基茨和尼维斯",
		"SC": "塞舌尔",
		"CV": "佛得角",
		"NC": "新喀里多尼亚",
		"YT": "马约特",
		"BB": "巴巴多斯",
		"SL": "塞拉利昂",
		"FJ": "斐济",
		"BW": "博茨瓦纳",
		"BZ": "伯利兹",
		"AW": "阿鲁巴",
		"TG": "多哥",
		"EE": "爱沙尼亚",
		"GG": "根西岛",
		"MR": "毛里塔尼亚",
		"GF": "法属圭亚那",
		"MZ": "莫桑比克",
		"NO": "挪威",
		"AF": "阿富汗",
		"MP": "北马里亚纳群岛",
		"GN": "几内亚",
		"LU": "卢森堡",
		"BM": "百慕大",
		"SY": "叙利亚",
		"GD": "格林纳达",
		"VG": "英属维尔京群岛",
		"HT": "海地",
		"KY": "开曼群岛",
		"NA": "纳米比亚",
		"GH": "加纳",
		"MG": "马达加斯加",
		"VC": "圣文森特和格林纳丁斯",
		"LY": "利比亚",
		"AS": "美属萨摩亚",
		"BJ": "贝宁",
		"UG": "乌干达",
		"VU": "瓦努阿图",
		"AG": "安提瓜和巴布达",
		"TC": "特克斯和凯科斯群岛",
		"BF": "布基纳法索",
		"SM": "圣马力诺",
		"NE": "尼日尔",
		"ZW": "津巴布韦",
		"IS": "冰岛",
		"LS": "莱索托",
		"PF": "法属波利尼西亚",
		"CK": "库克群岛",
		"ZM": "赞比亚",
		"MS": "蒙特塞拉特岛",
		"ST": "圣多美和普林西比",
		"BQ": "荷兰加勒比",
		"SX": "荷属圣马丁",
		"TL": "东帝汶",
		"GL": "格陵兰",
		"WF": "瓦利斯和富图纳群岛",
		"MW": "马拉维",
		"GI": "直布罗陀",
		"MM": "缅甸",
		"SZ": "斯威士兰",
		"BT": "不丹",
		"ET": "埃塞俄比亚",
		"FK": "福克兰群岛",
		"GM": "冈比亚",
		"GQ": "赤道几内亚",
		"KM": "科摩罗",
		"BI": "布隆迪",
		"CG": "刚果共和国",
		"RW": "卢旺达",
		"SS": "南苏丹",
		"CD": "刚果民主共和国",
		"ML": "马里",
		"MC": "摩纳哥",
		"SO": "索马里",
		"LR": "利比里亚",
		"FO": "法罗群岛",
		"PW": "帕劳",
		"AI": "安圭拉",
		"TO": "汤加",
		"SB": "所罗门群岛",
		"CU": "古巴",
		"AX": "奥兰群岛",
		"MH": "马绍尔群岛",
	}

	if v, ok := cc[keyword]; ok {
		return v
	}
	return ""
}

func InSlice[T string | int | int64 | int32 | int16 | int8 | uint | uint8 | uint16 | uint32 | uint64](ary []T, sub T) bool {
	for _, v := range ary {
		if v == sub {
			return true
		}
	}
	return false
}

func SliceFilter[T any](ary []T, filter func(v T) bool) []T {
	var ret []T
	for _, v := range ary {
		if filter(v) {
			ret = append(ret, v)
		}

	}
	return ret
}
