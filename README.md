# Golang closer pattern based Graceful shutdown

Closer предоставляет возможность регистрировать обработчики в разных частях приложения для правильного
завершения приложения.

## Пример использования

```go
package main

import (
	"log/slog"
	"syscall"
	"time"

	"git.server.lan/pkg/closer"
)

const (
	gracefulShutdownTimeout = time.Second * 10
)

func main() {
	// Настраиваем сигналы и таймаут для глобального closer
	closer.Init(
		closer.WithSignals(syscall.SIGHUP, syscall.SIGINT),
		closer.WithTimeout(gracefulShutdownTimeout),
	)

	closer.Add(func() error {
		time.Sleep(time.Hour * 5)
		slog.Info("done sleeping")
		return nil
	})

	// Запускаем основную логику приложения…

	// Ждём graceful shutdown
	closer.Wait()
}

```