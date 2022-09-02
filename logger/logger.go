package logger

import (
	"log"
	"os"
)

var logError *log.Logger = log.New(os.Stdout, "error\t\t=> ", log.LstdFlags)
var logInfo *log.Logger = log.New(os.Stdout, "info\t\t=> ", log.LstdFlags)

func _log(l *log.Logger, prefix string, msg string) {
	l.Println(prefix+"\t\t=>"+msg);
}

func LogInfo(prefix string, msg string) {
	_log(logInfo, prefix, msg);
}

func LogError(prefix string, msg string) {
	_log(logError, prefix, msg);
}
