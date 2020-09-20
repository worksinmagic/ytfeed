package main

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/worksinmagic/ytfeed/app/ytfeed"
)

const (
	cleanupDelay = 3 * time.Second
)

func main() {
	logger := log.New()

	logger.Infoln("Starting server")

	ctx, cancel := context.WithCancel(context.Background())
	err := ytfeed.Run(ctx, logger)
	if err != nil {
		cancel()
		logger.Infoln("Server shutting down because of unexpected error, waiting 3 seconds to complete cleaning up")
		<-time.After(cleanupDelay)

		logger.Fatalf("Unexpected error: %v", err)
	}

	cancel()
	logger.Infoln("Server shutting down, waiting 3 seconds to complete cleaning up")
	<-time.After(cleanupDelay)

	logger.Infoln("Server shut down")
}
