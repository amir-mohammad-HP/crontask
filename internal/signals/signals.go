package signals

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/amir-mohammad-HP/crontask/pkg/logger"
)

type Handler struct {
	logger *logger.StdLogger
}

func NewHandler(logger *logger.StdLogger) *Handler {
	return &Handler{logger: logger}
}

func (h *Handler) Handle(ctx context.Context, shutdownFunc func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	select {
	case <-ctx.Done():
		h.logger.Info("signal handler | Signal handler context cancelled")
		return
	case sig := <-sigChan:
		h.logger.Info("signal handler | Received signal %s", sig.String())
		shutdownFunc()
	}
}
