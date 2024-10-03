package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/unstoppablego/framework/config"
	"github.com/unstoppablego/framework/logs"
	"github.com/unstoppablego/framework/tool"
)

var UploadFilePath = config.Cfg.Http.UploadDir
var UploadFilePermission fs.FileMode = 0777

// 用于替换匹配路劲
func RespUpload(urlPath string) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		getpath := r.URL.Path
		// parts := strings.Split(r.URL.Path, "/")
		getpath = strings.Replace(getpath, urlPath, "", -1)

		filePatha := UploadFilePath + "/upload" + getpath
		logs.Info(filePatha)
		filename := filepath.Base(filePatha)
		// 打开文件
		file, err := os.Open(filePatha)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		defer file.Close()

		// 读取文件内容
		buffer := make([]byte, 512) // 读取文件头部 512 字节的内容
		n, err := file.Read(buffer)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}

		// 猜测文件的 MIME 类型
		contentType := http.DetectContentType(buffer[:n])
		w.Header().Set("Content-Type", contentType)

		// 获取文件信息
		fileInfo, err := file.Stat()
		if err != nil {
			http.Error(w, "Error getting file info", http.StatusInternalServerError)
			return
		}

		// 返回文件内容
		file.Seek(0, 0) // 重置文件读取位置
		http.ServeContent(w, r, filename, fileInfo.ModTime(), file)
	}
}

type BookResp struct {
	Name string
	Url  string
}

// 分为2部分 第一部分 需要能够注入 权限鉴定  第二部分需要能够 在上传完成后 执行后续动作
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

	} else if r.Method == http.MethodOptions {
		xdomain, err := url.Parse(r.Referer())
		if err != nil {
			logs.Error(err)
		}
		crosmain := xdomain.Scheme + "://" + xdomain.Host
		logs.Info(crosmain)
		w.Header().Set("Access-Control-Allow-Credentials", "true") //前端js也需要开启跨域请求
		if crosmain != "://" {
			w.Header().Set("Access-Control-Allow-Origin", crosmain) //来源网站
		}
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Referer, User-Agent, X-Requested-With,access-control-allow-origin, access-control-allow-headers, withCredentials, "+config.Cfg.Http.SessionName)
		w.Header().Set("Access-Control-Expose-Headers", config.Cfg.Http.SessionName)
		return
	} else {
		//处理跨域请求

		// store, err := session.Start(context.Background(), w, r)
		// if err != nil {
		// 	fmt.Fprint(w, err)
		// 	return
		// }
		// store.Set("sessionstart", true)
		// fmt.Println(store.Get("user"))
		// fmt.Println(r.Method)

		xdomain, err := url.Parse(r.Referer())
		if err != nil {
			logs.Error(err)
			// return
		}

		crosmain := xdomain.Scheme + "://" + xdomain.Host
		logs.Info(crosmain)
		w.Header().Set("Access-Control-Allow-Credentials", "true") //前端js也需要开启跨域请求
		if crosmain != "://" {
			w.Header().Set("Access-Control-Allow-Origin", crosmain) //来源网站
		}
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Referer, User-Agent, X-Requested-With,access-control-allow-origin, access-control-allow-headers, withCredentials, "+config.Cfg.Http.SessionName)
		w.Header().Set("Access-Control-Expose-Headers", config.Cfg.Http.SessionName)

		r.ParseMultipartForm(32 << 20)
		id := r.FormValue("ID")
		fmt.Println("ID:", id)

		// 获取上传的文件
		// file, _, err := r.FormFile("file")

		file, handler, err := r.FormFile("file")
		logs.Info()
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		md5src, err := io.ReadAll(file)
		if err != nil {
			logs.Error(err)
		}
		md5sum := tool.Md5b(md5src)

		timePath := time.Now().Format("2006-01-02") + "/"
		var filePath = config.Cfg.Custom["uploadfilepath"] + "/upload/" + timePath
		if err := os.MkdirAll(filePath, UploadFilePermission); !os.IsNotExist(err) {
			logs.Error(err)
			return
		}

		filesuffix := path.Ext(handler.Filename)
		fileName := md5sum + filesuffix
		copyFilePath := filePath + fileName
		copyFileSrcName := handler.Filename
		f, err := os.OpenFile(filePath+fileName, os.O_WRONLY|os.O_CREATE, UploadFilePermission)
		if err != nil {
			logs.Error(err)
			return
		}
		defer f.Close()

		// var Uid int
		// store, err := session.Start(context.Background(), w, r)
		// if err != nil {
		// 	logs.Error(err)
		// } else {
		// 	// ctxa.Session = store
		// 	if user, ok := store.Get("user"); ok {
		// 		logs.Info(user)
		// 		if usera, ok := user.(model.User); ok {

		// 			Uid = int(usera.UID)
		// 		}
		// 	}
		// }
		kid, err := strconv.Atoi(id)
		if err != nil {
			logs.Error(err)
		}
		logs.Info(md5sum, kid)
		// var uf model.Uploadfile
		// db.DB()
		// uf := model.GetUploadfileByPK_md5(md5sum, Uid, kid, db.DB())
		// if uf == nil { //如果没有找到文件，则开始上传
		// uf = &model.Uploadfile{}
		// uf.Filename = copyFileSrcName
		// uf.Filepath = filePath + md5sum + filesuffix
		// uf.Md5 = md5sum

		// uf.KnowID = kid
		// uf.Uid = Uid

		// if err := db.DB().Save(&uf).Error; err != nil {
		// 	logs.Error(err)
		// }

		_, err = file.Seek(0, 0)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		io.Copy(f, file)

		//使用当前传输处理TXT文本
		// if Uid != 0 {
		// 	services.UploadFileEmbeddingChannel <- *uf
		// }
		//发送数据vctor 处理

		logs.Info(copyFilePath, copyFileSrcName)
		ResponseCentera.Code = "200"
		var BookRespa BookResp
		BookRespa.Name = copyFileSrcName
		BookRespa.Url = config.Cfg.Custom["filedomain"] + config.Cfg.Http.UploadUrl + "/" + timePath + fileName
		ResponseCentera.Data = BookRespa

		// } else {
		// ResponseCentera.Code = "200"
		// var BookRespa BookResp
		// BookRespa.Name = uf.Filename

		// // 定义正则表达式
		// re := regexp.MustCompile(`/(\d{4}-\d{2}-\d{2})/([^/]+)$`)

		// // 匹配正则表达式
		// matches := re.FindStringSubmatch(uf.Filepath)

		// // 提取日期和文件名
		// if len(matches) == 3 {
		// 	datex := matches[1]
		// 	fileName := matches[2]
		// 	fmt.Println("日期:", datex)
		// 	fmt.Println("文件名:", fileName)
		// 	BookRespa.Url = config.Cfg.Custom["filedomain"] + "/rimg/" + datex + "/" + fileName
		// } else {
		// 	logs.Error("not find file", uf.Filepath)
		// }

		// ResponseCentera.Data = BookRespa
		// }

	}
}
