package logging

import (
	"log"
	"os"

	slogger "github.com/jesse-rb/slogger-go"
)

var (
	infoLogger  = slogger.New(os.Stdout, slogger.ANSIGreen, "", log.LUTC)
	errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "", log.LUTC)
	debugLogger = slogger.New(os.Stdout, slogger.ANSIYellow, "", log.LUTC)
)

func LogInfo(tag string, msg string, data interface{}) {
	infoLogger.Log(tag, msg, data)
}

func LogError(tag string, msg string, data interface{}) {
	errorLogger.Log(tag, msg, data)
}

func LogDebug(tag string, msg string, data interface{}) {
	debugLogger.Log(tag, msg, data)
}
