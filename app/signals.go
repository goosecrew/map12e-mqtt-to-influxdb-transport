package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func CatchSigTerm(ctx context.Context, cancel context.CancelFunc) {
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case <-ctx.Done():
			return
		case <-cancelChan:
			cancel()
			return
		}
	}
}
