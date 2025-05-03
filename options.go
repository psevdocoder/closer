package closer

import (
	"os"
	"time"
)

// Option задаёт опции для Closer
// WithSignals и WithTimeout позволяют конфигурировать Closer при создании.
type Option func(c *Closer)

// WithSignals задаёт набор сигналов (по умолчанию SIGINT и SIGTERM).
func WithSignals(sig ...os.Signal) Option {
	return func(c *Closer) {
		if len(sig) > 0 {
			c.signals = sig
		}
	}
}

// WithTimeout задаёт максимальное время ожидания функций graceful shutdown.
// Если d <= 0 — ожидание бесконечное.
func WithTimeout(d time.Duration) Option {
	return func(c *Closer) {
		c.timeout = d
	}
}
