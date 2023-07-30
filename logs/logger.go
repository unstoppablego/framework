package logs

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

func init() {
	customFormatter := new(logrus.TextFormatter)

	// func(*runtime.Frame) (function string, file string)
	customFormatter.CallerPrettyfier = func(f *runtime.Frame) (string, string) {

		// s := strings.Split(f.Function, ".")
		// funcname := s[len(s)-1]
		filenum := ""
		for i := 5; ; i++ {
			_, file, line, ok := runtime.Caller(i)
			if ok {
				if !strings.Contains(file, "logrus") && !strings.Contains(file, "logger.go") {
					filenum = fmt.Sprintf("%s:%d", file, line) //file + ":" + line
					break
				}
			} else {
				break
			}
		}
		return "", filenum
	}
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
	// log.SetFormatter(&log.TextFormatter{})

	logrus.SetReportCaller(true)
}

func Trace(args ...interface{}) {

	logrus.Trace(args...)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...interface{}) {
	logrus.Debug(args...)
}

// Print logs a message at level Info on the standard logger.
func Print(args ...interface{}) {
	logrus.Print(args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	// data := make(map[string]interface{})
	// for i, v := range args {
	// 	data[strconv.Itoa(i)] = v
	// }
	var argsa []interface{}

	for _, v := range args {
		argsa = append(argsa, v, " ")
	}

	logrus.Info(argsa...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	logrus.Warn(args...)
}

// Warning logs a message at level Warn on the standard logger.
func Warning(args ...interface{}) {
	logrus.Warning(args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	logrus.Error(args...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(args ...interface{}) {
	logrus.Panic(args...)
}

// Fatal logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatal(args ...interface{}) {
	logrus.Fatal(args...)
}

func SetLevel(level logrus.Level) {
	logrus.SetLevel(level)
}
