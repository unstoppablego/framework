package ssh

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/unstoppablego/framework/logs"
)

const (
	CommandSuccess = "success"
	CommandFailed  = "failed"
)
const (
	StateSuccess = 1
	StateFailed  = 0
	StateRun     = 2
)

type SSHManager struct {
	CurCommandIndex string
	File            string
}

func NewSSH(File string) *SSHManager {

	return &SSHManager{File: File}
}

func (s *SSHManager) getState(data []byte) int {

	if bytes.Contains(data, []byte(s.CurCommandIndex+CommandSuccess)) {
		return StateSuccess
	}
	if bytes.Contains(data, []byte(s.CurCommandIndex+CommandFailed)) {
		return StateFailed
	}
	return StateRun
}

func (s *SSHManager) Write(rw io.WriteCloser, cmd string) string {
	var EndEcho = "end" + time.Now().Format(time.RFC3339) + "End"

	s.CurCommandIndex = EndEcho

	var SuccessEcho = " && echo " + EndEcho + CommandSuccess
	var FailedEcho = " || echo" + EndEcho + CommandFailed

	rw.Write([]byte(cmd + SuccessEcho + FailedEcho + "' \r\n"))

	return EndEcho
}

func (s *SSHManager) WaitFinish(rr io.Reader, FileName string) error {

	var xxxa = make([]byte, 4096)
	fetchError := false
	for {
		rl, rerr := rr.Read(xxxa)

		if rerr != nil {
			logs.Info("WaitFinish", rerr)
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

		if s.getState(xxxa[0:rl]) == StateSuccess {

			break
		} else if s.getState(xxxa[0:rl]) == StateFailed {
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
