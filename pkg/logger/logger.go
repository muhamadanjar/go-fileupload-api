package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func init() {
	Log = logrus.New()

	// Output ke stdout
	Log.SetOutput(os.Stdout)

	// Format bisa diganti ke TextFormatter atau JSON
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Level default: Info (bisa diubah jadi Debug, Error, dll)
	Log.SetLevel(logrus.InfoLevel)

	// Panic, Fatal, Error, Warn, Info, Debug, Trace
	// Log.SetLevel(logrus.DebugLevel)
}
