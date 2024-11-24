# Golang closer pattern based Graceful shutdown

Closer предоставляет возможность регистрировать обработчики в разных частях приложения для правильного
завершения приложения.

## Пример использования
```go
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"git.server.home/pkg/closer"
)

const (
	shutdownTimeout = 10 * time.Second
)

func main() {
	syscallCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := NewHttpServer()
	srv.Start()

	// Wait for syscall
	<-syscallCtx.Done()
	slog.Info("received syscall", slog.Any("signal", syscallCtx.Err()))

	// Close resources with timeout
	shutdownDeadlineCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := closer.Close(shutdownDeadlineCtx); err != nil {
		slog.Error("failed to close resources", slog.Any("error", err))
		return
	}
	slog.Info("app gracefully stopped")
}

type Srv struct {
	*http.Server
}

func NewHttpServer() *Srv {
	srv := http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello, World!"))
		}),
	}
	return &Srv{&srv}
}

func (s *Srv) Start() {
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	closer.Add(func(ctx context.Context) error {
		return s.Shutdown(ctx)
	})
}

```