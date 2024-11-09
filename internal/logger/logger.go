package logger

type Logger interface {
	Error(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
}
