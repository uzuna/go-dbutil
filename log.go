package dbutil

import (
	"log"
	"os"

	"github.com/hashicorp/logutils"
)

func init() {
	loglevel()
}

func loglevel() {
	s := os.Getenv("LOGLEVEL")
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("WARN"),
		Writer:   os.Stderr,
	}
	switch s {
	case "DEBUG", "INFO", "ERROR":
		filter.MinLevel = logutils.LogLevel(s)
	}
	log.SetOutput(filter)
}
