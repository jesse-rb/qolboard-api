package controller

import (
	"log"
	"os"

	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger = slogger.New(os.Stdout, slogger.ANSIGreen, "controllers", log.Lshortfile+log.Ldate);
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "controllers", log.Lshortfile+log.Ldate);