package ssh

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/unstoppablego/framework/logs"
	"golang.org/x/crypto/ssh"
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
	CurCommand      string
	File            string
	PrivateKey      []byte
	w               io.WriteCloser
	r               io.Reader
	session         *ssh.Session
	client          *ssh.Client
}

func NewSSH(File string, PrivateKey []byte) *SSHManager {

	return &SSHManager{File: File, PrivateKey: PrivateKey}
}

/*
address ip+":22"
*/
func (s *SSHManager) Connect(user string, address string) {

	signer, err := ssh.ParsePrivateKey(s.PrivateKey)
	if err != nil {
		logs.Error(err)
	}

	clientConfig := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: func(
			hostname string,
			remote net.Addr,
			key ssh.PublicKey) error {
			// do something in call back function
			return nil
		},
	}

	client, err := ssh.Dial("tcp", address, &clientConfig)
	s.client = client
	if err != nil {
		logs.Error(err)
		return
	}

	logs.Info("SSH Connect Ok")

	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: " + err.Error())
	}
	log.Println("Get session OK")

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     //打开回显
		ssh.TTY_OP_ISPEED: 14400, //输入速率 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, //输出速率 14.4kbaud
		ssh.VSTATUS:       1,
	}

	err = session.RequestPty("xterm", 100, 100, modes)

	if err != nil {
		logs.Error(err)
	}

	s.w, _ = session.StdinPipe()
	s.r, err = session.StdoutPipe()

	if err := session.Shell(); err != nil {
		log.Fatalf("failed to start shell: " + err.Error())
	}

	if err != nil {
		logs.Error(err)
	}
	s.session = session
}

func (s *SSHManager) Close() {
	s.client.Close()
	s.session.Close()
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

func (s *SSHManager) Write(cmd string) string {
	var EndEcho = "end" + time.Now().Format(time.RFC3339) + "End"

	s.CurCommandIndex = EndEcho
	s.CurCommand = cmd
	var SuccessEcho = " && echo '" + EndEcho + CommandSuccess + "'"
	var FailedEcho = " || echo '" + EndEcho + CommandFailed + "'"

	// logs.Info(cmd + SuccessEcho + FailedEcho)

	if _, err := s.w.Write([]byte("(" + cmd + ")" + SuccessEcho + FailedEcho + " \r\n")); err != nil {
		logs.Error(err)
	} else {
		// logs.Info(n)
	}

	return EndEcho
}

func (s *SSHManager) WaitFinish() error {

	var xxxa = make([]byte, 4096)
	fetchError := false
	for {
		rl, rerr := s.r.Read(xxxa)

		if rerr != nil {
			if rerr == io.EOF {
				logs.Info("WaitFinish", rerr)
			} else {
				logs.Info("WaitFinish", rerr)
			}
		}
		// logs.Info(s.CurCommand)
		if len(s.CurCommand) > 3 && s.CurCommand[0:3] == "scp" {
			logs.Info("scp command")
			if bytes.Contains(xxxa[0:rl], []byte("Are you sure you want to continue connecting (yes/no)")) {
				s.w.Write([]byte("yes"))
				s.Write("\r\n")
			}
		}
		if rl > 0 {
			WriteFile(s.File, string(xxxa[0:rl]))
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
			logs.Info("Command Success " + s.CurCommand)
			break
		} else if s.getState(xxxa[0:rl]) == StateFailed {
			logs.Info("Command Error " + s.CurCommand)
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
		fmt.Println("Open File Failed ", err)
	}

	defer file.Close()
	write := bufio.NewWriter(file)
	write.WriteString(data)
	write.Flush()
}
