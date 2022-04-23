package api

import (
	"log"
	"os"
)

var logError *log.Logger = log.New(os.Stdout, "api\t\t=> error\t\t=> ", log.LstdFlags)
var logInfo *log.Logger = log.New(os.Stdout, "api\t\t=> info\t\t=> ", log.LstdFlags)