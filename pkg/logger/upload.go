package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var UploadLog *logrus.Logger

func init() {
	UploadLog = logrus.New()

	file, err := os.OpenFile("logs/upload.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		UploadLog.SetOutput(file)
	} else {
		UploadLog.SetOutput(os.Stdout)
	}

	UploadLog.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}
