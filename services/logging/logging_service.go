package logging

import (
	"log"
	"os"

	slogger "github.com/jesse-rb/slogger-go/v2"
)

var (
	infoLogger  = slogger.New(os.Stdout, slogger.ANSIBlue, "", log.LUTC+log.Lshortfile, 3)
	errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "", log.LUTC, 3)
	debugLogger = slogger.New(os.Stdout, slogger.ANSIYellow, "", log.LUTC, 3)
)

func LogInfo(tag string, msg string, data interface{}) {
	logGeneric(infoLogger, tag, msg, data)
}

func LogError(tag string, msg string, data interface{}) {
	logGeneric(errorLogger, tag, msg, data)
}

func LogDebug(tag string, msg string, data interface{}) {
	logGeneric(debugLogger, tag, msg, data)
}

func Here() {
	logGeneric(debugLogger, "HERE", "HERE", nil)
}

func logGeneric(logger *slogger.Logger, tag string, msg string, data interface{}) {
	logger.Log(tag, msg, data)
}
