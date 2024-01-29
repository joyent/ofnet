package log

// Logger is log message writer interface
type Logger interface {
	Trace(...interface{})
	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	Fatal(...interface{})

	Tracef(string, ...interface{})
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})

	//With(string, interface{}) Logger
	//WithError(error) Logger
}

var logger Logger

func init() {
	logger = New(false)
}

func New(debug bool) Logger {
	return newLogrus(debug)
}

func SetLogger(l Logger) {
	logger = l
}

func GetLogger() Logger {
	return logger
}
