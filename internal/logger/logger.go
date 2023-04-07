package logger

import (
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

func createFile(logFilePath string) (*os.File, error) {

	// Create the log file if it doesn't exist
	logFile := path.Join(logFilePath, "./qbit.log")

	// Check if the file exists
	return os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
}

func New(LogFilePath string) (*logrus.Logger, error) {
	logFile, err := createFile(LogFilePath)
	if err != nil {
		return nil, err
	}

	log := logrus.New()

	log.SetOutput(logFile)
	log.SetLevel(logrus.DebugLevel)

	return log, nil
}
