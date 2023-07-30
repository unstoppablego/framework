package app

import (
	"github.com/unstoppablego/framework/logs"
)

var Log logs.ILogger

func init() {
	Log = &logs.Provider{}
}
