package webserver

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

func KeepAliveWithSignals(ctx context.Context, logger *otelzap.SugaredLogger, serversShutdown func()) {
	quit := make(chan os.Signal, 1)

	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	var osSignalMsg string

	switch <-quit {
	case syscall.SIGINT:
		osSignalMsg = "SIGINT (signal interrupt)"

	case syscall.SIGTERM:
		osSignalMsg = "SIGTERM (signal termination)"
	}

	logger.Ctx(ctx).Debugw("µ-service terminated with", "signal", osSignalMsg)
	logger.Ctx(ctx).Infof("stopping µ-service servers")
	serversShutdown()
	logger.Ctx(ctx).Infow("µ-service servers stopped")
}
