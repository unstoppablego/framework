package tool

import (
	"fmt"
	"runtime"

	"github.com/unstoppablego/framework/logs"
)

func HandleRecover() {
	r := recover()
	if r != nil {
		logs.Error(r)
		var buf [4096]byte
		n := runtime.Stack(buf[:], false)
		fmt.Println(string(buf[:n]))
	}
}
