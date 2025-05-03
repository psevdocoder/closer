package closer

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Closer управляет graceful shutdown
type Closer struct {
	mu       sync.Mutex
	once     sync.Once
	done     chan struct{}
	funcs    []func() error
	signals  []os.Signal
	timeout  time.Duration
	execMode ExecutionMode
}

// New создаёт новый экземпляр Closer и запускает
// обработку системных сигналов завершения
func New(opts ...Option) *Closer {
	c := &Closer{
		done:     make(chan struct{}),
		signals:  []os.Signal{syscall.SIGINT, syscall.SIGTERM},
		execMode: LIFOSequential,
	}
	for _, opt := range opts {
		opt(c)
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, c.signals...)
	go func() {
		if sig, ok := <-ch; ok {
			slog.Info("Received shutdown signal", slog.String("signal", sig.String()))
			c.CloseAll()
		}
	}()

	slog.Info(
		"Closer initialized",
		slog.String("mode", c.execMode.String()),
		slog.Duration("timeout", c.timeout),
	)

	return c
}

// Add регистрирует одну или несколько функций, которые будут вызваны при завершении.
func (c *Closer) Add(f ...func() error) {
	c.mu.Lock()
	c.funcs = append(c.funcs, f...)
	c.mu.Unlock()
}

// Wait блокирует текущую горутину до тех пор,
// пока не завершится выполнение всех shutdown-функций, либо не пройдет тайм-аут.
func (c *Closer) Wait() {
	<-c.done
}

// CloseAll запускает все функции ровно один раз
// и ждёт их выполнения не дольше, чем c.timeout.
func (c *Closer) CloseAll() {
	c.once.Do(func() {
		slog.Info("Graceful shutdown started")
		defer close(c.done)

		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		var ctx context.Context
		var cancel context.CancelFunc
		if c.timeout > 0 {
			ctx, cancel = context.WithTimeout(context.Background(), c.timeout)
		} else {
			ctx, cancel = context.WithCancel(context.Background())
		}
		defer cancel()

		// helper для запуска функции и обработки на случай тайм-аута
		runWithTimeout := func(fn func() error) error {
			errCh := make(chan error, 1)
			go func() { errCh <- fn() }()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case err := <-errCh:
				return err
			}
		}

		switch c.execMode {
		case LIFOSequential:
			for i := len(funcs) - 1; i >= 0; i-- {
				if err := runWithTimeout(funcs[i]); err != nil {
					if errors.Is(err, context.DeadlineExceeded) {
						slog.Error("Graceful shutdown timed out", slog.Duration("timeout", c.timeout))
						return
					}
					slog.Error("Error in shutdown function", slog.Any("error", err))
				}
			}

		case FIFOSequential:
			for i := 0; i < len(funcs); i++ {
				if err := runWithTimeout(funcs[i]); err != nil {
					if errors.Is(err, context.DeadlineExceeded) {
						slog.Error("Graceful shutdown timed out", slog.Duration("timeout", c.timeout))
						return
					}
					slog.Error("Error in shutdown function", slog.Any("error", err))
				}
			}

		case Parallel:
			errCh := make(chan error, len(funcs))
			for _, fn := range funcs {
				go func(f func() error) { errCh <- f() }(fn)
			}
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
		}

		slog.Info("Graceful shutdown completed successfully")
	})
}
