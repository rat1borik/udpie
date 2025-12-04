package logutils

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func SetupLogger(config *LogConfig) error {
	// Setup log rotation
	var logFile *lumberjack.Logger
	if config.OutputFile {
		logFile = &lumberjack.Logger{
			Filename:   config.File,
			MaxSize:    config.MaxSize,    // megabytes
			MaxBackups: config.MaxBackups, // number of backup files
			MaxAge:     config.MaxAge,     // days
			Compress:   config.Compress,   // compress old logs .gz
		}
	}

	log := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	// Set formatter
	if config.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	// Setup multi-writer: logFile first, then stdout at the end
	// stdout at the end so it doesn't break the chain if service fails
	var writers []io.Writer

	if config.OutputFile && logFile != nil {
		writers = append(writers, logFile)
	}

	if config.OutputStd {
		writers = append(writers, os.Stdout)
	}

	var multiWriter io.Writer
	if len(writers) > 0 {
		multiWriter = io.MultiWriter(writers...)
		log.SetOutput(multiWriter)
	}

	// Set as global logger
	logrus.SetLevel(level)
	if config.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
	if multiWriter != nil {
		logrus.SetOutput(multiWriter)
	}

	return nil
}

type LogConfig struct {
	Level      string
	File       string
	Format     string
	OutputFile bool
	OutputStd  bool
	MaxSize    int  // megabytes
	MaxBackups int  // number of backup files
	MaxAge     int  // days
	Compress   bool // compress old logs
}
