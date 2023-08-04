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

var forTestLog []byte

func WaitFinish(rr io.Reader, end string, FileName string) error {

	var xxxa = make([]byte, 4096)
	fetchError := false
	for {
		rl, rerr := rr.Read(xxxa)
		if rerr != nil {
			app.Log.Info("WaitFinish", rerr)
		}

		if rl > 0 {
			WriteFile(FileName, string(xxxa[0:rl]))
		}

		if bytes.Contains(xxxa[0:rl], []byte("Could not get lock /var/lib/dpkg/lock-frontend")) {
			log.Println("find lock error")
			time.Sleep(15 * time.Second)
			return fmt.Errorf("lock error")
		}

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
