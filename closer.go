package closer

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Closer управляет graceful shutdown
type Closer struct {
	mu      sync.Mutex
	once    sync.Once
	done    chan struct{}
	funcs   []func() error
	signals []os.Signal
	timeout time.Duration
}

// New создаёт новый Closer и запускает слушатель сигналов.
func New(opts ...Option) *Closer {
	c := &Closer{
		done:    make(chan struct{}),
		signals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
	}
	for _, opt := range opts {
		opt(c)
	}
	// Запускаем слушатель ОС-сигналов
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, c.signals...)
	go func() {
		if sig, ok := <-ch; ok {
			slog.Info("Received shutdown signal", slog.String("signal", sig.String()))
			c.CloseAll()
		}
	}()
	return c
}

// Add метод экземпляра Closer
func (c *Closer) Add(f ...func() error) {
	c.mu.Lock()
	c.funcs = append(c.funcs, f...)
	c.mu.Unlock()
}

// Wait метод экземпляра Closer — альтернатива <-c.done
func (c *Closer) Wait() {
	<-c.done
}

// CloseAll запускает все функции ровно один раз.
// После прихода сигнала начинается отсчёт таймаута (если он >0);
// ждём завершения всех функций или истечения времени.
func (c *Closer) CloseAll() {
	c.once.Do(func() {
		slog.Info("Graceful shutdown started")
		defer close(c.done)

		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		errCh := make(chan error, len(funcs))

		// создаём контекст с таймаутом
		var ctx context.Context
		var cancel context.CancelFunc
		if c.timeout > 0 {
			ctx, cancel = context.WithTimeout(context.Background(), c.timeout)
		} else {
			ctx, cancel = context.WithCancel(context.Background())
		}
		defer cancel()

		// запускаем все функции параллельно
		for _, fn := range funcs {
			go func(f func() error) {
				errCh <- f()
			}(fn)
		}

		// ждём их завершения или таймаута
		for i := 0; i < cap(errCh); i++ {
			select {
			case err := <-errCh:
				if err != nil {
					slog.Error("Error in shutdown function", slog.Any("error", err))
				}
			case <-ctx.Done():
				slog.Error("Graceful shutdown timed out", slog.Duration("timeout", c.timeout))
				return
			}
		}

		slog.Info("Graceful shutdown completed successfully")
	})
}
