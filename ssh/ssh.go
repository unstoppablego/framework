package ssh

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/unstoppablego/framework/app"
)

func Write(rw io.WriteCloser, cmd string) string {
	var EndEcho = "end" + time.Now().Format(time.RFC3339) + "End"
	rw.Write([]byte(cmd + " && echo '" + EndEcho + "' \r\n"))
	return EndEcho
}

// 用于在测试模式下检测数据的存放
var forTestLog []byte

func WaitFinish(rr io.Reader, end string, FileName string) error {

	var xxxa = make([]byte, 4096)
	// var isNeedRerun = false
	fetchError := false
	// var success = 0
	for {
		rl, rerr := rr.Read(xxxa)
		if rerr != nil {
			app.Log.Info("WaitFinish", rerr)
		}

		// 如果是测试模式

		if rl > 0 {
			// fmt.Println(string(xxxa[0:rl]))
			WriteFile(FileName, string(xxxa[0:rl]))
		}

		//如果是测试模式
		// if comm.LaunchMode == 2 && rl > 0 {
		// 	forTestLog = append(forTestLog, xxxa[0:rl]...)
		// 	var TraceAction *[]comm.ActionTrace

		// 	if comm.CheckMode == 1 {
		// 		TraceAction = &comm.ActionTraceAlipaySendMode1HaveCart
		// 	} else if comm.CheckMode == 2 {
		// 		TraceAction = &comm.ActionTraceAlipaySendMode1NoCart
		// 	} else if comm.CheckMode == 3 {
		// 		TraceAction = &comm.ActionTraceAlipaySendMode2HaveCart
		// 	} else if comm.CheckMode == 4 {
		// 		TraceAction = &comm.ActionTraceAlipaySendMode2HaveCart
		// 	}

		// 	var loopinfo = ""
		// 	var tmpsuccess = 0

		// 	for vi, v := range *TraceAction {
		// 		if bytes.Contains(forTestLog, []byte(v.Action)) {
		// 			(*TraceAction)[vi].Success = true
		// 		}

		// 		if (*TraceAction)[vi].Success {
		// 			tmpsuccess++
		// 		}
		// 		loopinfo += fmt.Sprintf("%s %v ", strings.Replace(strings.Replace(v.Action, "1635769101-", "", -1), "-success", "", -1), v.Success)
		// 	}

		// 	if tmpsuccess > success {
		// 		success = tmpsuccess
		// 		fmt.Println(loopinfo)
		// 	}

		// 	if success == len(*TraceAction) {
		// 		loopinfo = "success\n\n" + loopinfo
		// 		fmt.Println(loopinfo)
		// 	}
		// }

		if bytes.Contains(xxxa[0:rl], []byte("Could not get lock /var/lib/dpkg/lock-frontend")) {
			log.Println("find lock error")
			time.Sleep(15 * time.Second)
			return fmt.Errorf("lock error")
		}
		//
		// if bytes.Contains(xxxa[0:rl], []byte("Could not resolve host: hk1.yfwq.org")) {
		// 	log.Println("find lock error")
		// 	time.Sleep(15 * time.Second)
		// 	return fmt.Errorf("lock error")
		// }

		if bytes.Contains(xxxa[0:rl], []byte("E: Failed to fetch")) {
			fmt.Println("E: Failed to fetch")
			fetchError = true
		}

		if bytes.Contains(xxxa[0:rl], []byte(end)) {
			fmt.Println("运行结束")
			break
		}
	}
	if fetchError {
		return fmt.Errorf("fetch error")
	}
	return nil
}

func WriteFile(filePath string, data string) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
	}

	defer file.Close()
	write := bufio.NewWriter(file)
	write.WriteString(data)
	write.Flush()
}
