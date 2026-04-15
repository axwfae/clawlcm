package logger

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type defaultLogger struct{}

func New() Logger {
	return &defaultLogger{}
}

func (l *defaultLogger) Debug(msg string) {
	println("[DEBUG]", msg)
}

func (l *defaultLogger) Info(msg string) {
	println("[INFO]", msg)
}

func (l *defaultLogger) Warn(msg string) {
	println("[WARN]", msg)
}

func (l *defaultLogger) Error(msg string) {
	println("[ERROR]", msg)
}

type nilLogger struct{}

func NewNil() Logger {
	return &nilLogger{}
}

func (l *nilLogger) Debug(msg string) {}
func (l *nilLogger) Info(msg string)  {}
func (l *nilLogger) Warn(msg string)  {}
func (l *nilLogger) Error(msg string) {}
