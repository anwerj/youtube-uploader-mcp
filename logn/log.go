package logn

import (
	"fmt"
	"os"
	"os/user"
)

const logFileName = "youtube_uploader_mcp.log"

var logFile *os.File

func init() {
	var err error
	// Create or verify the log file exists
	user, err := user.Current()
	if err != nil {
		panic("Error getting current user: " + err.Error())
	}
	logFilePath := user.HomeDir + "/" + logFileName

	if logFile, err = os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0644); err != nil {
		panic("Error creating log file: " + err.Error())
	}
}

func Info(message string, args ...any) {
	if logFile == nil {
		return
	}
	_, err := fmt.Fprintf(logFile, "[INFO] "+message+"\n", args...)
	if err != nil {
		panic("Error writing to log file: " + err.Error())
	}
}
