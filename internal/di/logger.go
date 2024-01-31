package di

import (
	"log"

	"github.com/yonisaka/similarity/pkg/logger"
)

// GetLogger get the logger wrapper.
func GetLogger() logger.Logger {
	l, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}

	return l
}
