package service

import (
	"log"
	"os"

	slogger "github.com/jesse-rb/slogger-go"
)

var infoLogger = slogger.New(os.Stdout, slogger.ANSIGreen, "service", log.Lshortfile+log.Ldate);
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "service", log.Lshortfile+log.Ldate);