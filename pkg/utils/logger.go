package utils

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"strings"
	"time"
)

type LogEvent struct {
	Datetime time.Time
	Content  string
	Level    string
}

const (
	debug = "DEBUG,INFO,WARN,ERROR"
	info  = "INFO,WARN,ERROR"
	warn  = "WARN,ERROR"
	//error = "ERROR"
)

func Logger(event LogEvent) {

	level := viper.GetString("loglevel") //strings.ToUpper(event.Level)
	logMessage := ""

	switch level {
	case "DEBUG":
		if strings.Contains(debug, event.Level) {
			logMessage = fmt.Sprintf("[%s] %s", event.Level, event.Content)
		}
	case "INFO":
		if strings.Contains(info, event.Level) {
			logMessage = fmt.Sprintf("[%s] %s", event.Level, event.Content)
		}
	case "WARN":
		if strings.Contains(warn, event.Level) {
			logMessage = fmt.Sprintf("[%s] %s", event.Level, event.Content)
		}
	case "ERROR":
		if event.Level == "ERROR" {
			logMessage = fmt.Sprintf("[%s] %s", event.Level, event.Content)
		}
	default:
		return
	}
	if len(logMessage) > 1 {
		log.Println(logMessage)
	}
}
