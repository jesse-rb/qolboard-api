package config

import (
	"log"
	"os"

	slogger "github.com/jesse-rb/slogger-go"
)

// Declare some loggers
var infoLogger = slogger.New(os.Stdout, slogger.ANSIGreen, "config", log.Lshortfile+log.Ldate);
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "config", log.Lshortfile+log.Ldate);