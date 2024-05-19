package middleware

import (
	"log"
	"os"

	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger = slogger.New(os.Stdout, slogger.ANSIGreen, "middleware", log.Lshortfile+log.Ldate);
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "middleware", log.Lshortfile+log.Ldate);