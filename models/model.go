package model

import (
	"log"
	"os"
	"time"

	slogger "github.com/jesse-rb/slogger-go"
	"gorm.io/gorm"
)

type Model struct {
    ID        uint64			`gorm:"primarykey" json:"id"`
    CreatedAt time.Time 		`json:"createdAt"`
    UpdatedAt time.Time 		`json:"updatedAt"`
    DeletedAt gorm.DeletedAt 	`gorm:"index" json:"deletedAt"`
}

var infoLogger = slogger.New(os.Stdout, slogger.ANSIGreen, "models", log.Lshortfile+log.Ldate);
var errorLogger = slogger.New(os.Stderr, slogger.ANSIRed, "models", log.Lshortfile+log.Ldate);

