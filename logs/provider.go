package logs

import (
	"github.com/sirupsen/logrus"
)

type Provider struct {
}

func (p *Provider) Trace(args ...interface{}) {

	logrus.Trace(args...)
}

// Debug logs a message at level Debug on the standard logger.
func (p *Provider) Debug(args ...interface{}) {
	logrus.Debug(args...)
}

// Print logs a message at level Info on the standard logger.
func (p *Provider) Print(args ...interface{}) {
	logrus.Print(args...)
}

// Info logs a message at level Info on the standard logger.
func (p *Provider) Info(args ...interface{}) {
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
func (p *Provider) Warn(args ...interface{}) {
	logrus.Warn(args...)
}

// Warning logs a message at level Warn on the standard logger.
func (p *Provider) Warning(args ...interface{}) {
	logrus.Warning(args...)
}

// Error logs a message at level Error on the standard logger.
func (p *Provider) Error(args ...interface{}) {
	logrus.Error(args...)
}

// Panic logs a message at level Panic on the standard logger.
func (p *Provider) Panic(args ...interface{}) {
	logrus.Panic(args...)
}

// Fatal logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func (p *Provider) Fatal(args ...interface{}) {
	logrus.Fatal(args...)
}

func (p *Provider) SetLevel(level uint32) {
	logrus.SetLevel(logrus.Level(level))
}
