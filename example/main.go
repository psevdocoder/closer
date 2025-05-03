package main

import (
	"git.server.home/pkg/closer"
	"log/slog"
	"syscall"
	"time"
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
