package logger

// todo: create custom logger that puts into the file
var debugFilename = "debug-ripple"

type Logger struct {
	debugMode bool
}

func Init(debugMode bool) *Logger {
	return &Logger{
		debugMode: debugMode,
	}
}

func (l *Logger) Println(str string) {

}
