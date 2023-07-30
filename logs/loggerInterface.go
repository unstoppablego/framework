package logs

type ILogger interface {
	SetLevel(level uint32)
	Trace(args ...interface{})
	Debug(args ...interface{})
	Print(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Panic(args ...interface{})
	Fatal(args ...interface{})
}
