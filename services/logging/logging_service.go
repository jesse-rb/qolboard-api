package logging

import (
	"log"
	"os"

	slogger "github.com/jesse-rb/slogger-go/v2"
)

var (
	infoLogger  = slogger.New(os.Stdout, slogger.ANSIGreen, "", log.Ldate+log.Ltime+log.Lshortfile)
	errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "", log.Ldate+log.Ltime+log.Lshortfile)
	debugLogger = slogger.New(os.Stdout, slogger.ANSIYellow, "", log.Ldate+log.Ltime+log.Lshortfile)
)

func init() {
	infoLogger.SetCalldepth(3)
	errorLogger.SetCalldepth(3)
	debugLogger.SetCalldepth(3)
}

func LogInfo(tag string, msg string, data any) {
	logGeneric(infoLogger, tag, msg, data)
}

func LogError(tag string, msg string, data any) {
	logGeneric(errorLogger, tag, msg, data)
}

func LogDebug(tag string, msg string, data any) {
	logGeneric(debugLogger, tag, msg, data)
}

func Here() {
	logGeneric(debugLogger, "HERE", "HERE", nil)
}

func logGeneric(logger *slogger.Logger, tag string, msg string, data any) {
	logger.Log(tag, msg, data)
}
