package config

import (
	"log"
	"os"
)

var logError *log.Logger = log.New(os.Stdout, "config\t\t=> error\t\t=> ", log.LstdFlags)
var logInfo *log.Logger = log.New(os.Stdout, "config\t\t=> info\t\t=> ", log.LstdFlags)