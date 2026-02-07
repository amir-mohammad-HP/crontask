package signals

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/amir-mohammad-HP/crontask/pkg/logger"
)

type Handler struct {
	logger logger.Logger
}

func NewHandler(logger logger.Logger) *Handler {
	return &Handler{logger: logger}
}

func (h *Handler) Handle(ctx context.Context, shutdownFunc func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	select {
	case <-ctx.Done():
		h.logger.Info("Signal handler context cancelled")
		return
	case sig := <-sigChan:
		h.logger.Info("Received signal", "signal", sig)
		shutdownFunc()
	}
}
