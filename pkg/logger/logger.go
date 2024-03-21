package logger

type Logger struct {
	Errs []string
}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) AddErr(err string) {
	l.Errs = append(l.Errs, err)
}
