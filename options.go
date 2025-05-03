package closer

import (
	"os"
	"time"
)

// ExecutionMode определяет порядок и способ выполнения shutdown-функций.
type ExecutionMode int

const (
	// LIFOSequential — последовательное выполнение в обратном порядке (по умолчанию).
	LIFOSequential ExecutionMode = iota

	// FIFOSequential — последовательное выполнение в порядке добавления.
	FIFOSequential

	// Parallel — параллельное выполнение всех функций.
	// Порядок запуска не гарантируется.
	Parallel
)

// String возвращает строковое представление режима исполнения.
func (m *ExecutionMode) String() string {
	switch *m {
	case LIFOSequential:
		return "LIFO Sequential (Default)"
	case FIFOSequential:
		return "FIFO Sequential"
	case Parallel:
		return "FIFO Parallel"
	default:
		return "Unknown"
	}
}

// Option задаёт опции конфигурации Closer при создании через New или Init.
// Используется с функциями вроде WithTimeout, WithSignals и др.
type Option func(c *Closer)

// WithSignals задаёт список сигналов ОС, на которые должен реагировать Closer.
//
// По умолчанию используются syscall.SIGINT и syscall.SIGTERM.
// Если передать пустой список, сигналы не будут слушаться.
func WithSignals(sig ...os.Signal) Option {
	return func(c *Closer) {
		if len(sig) > 0 {
			c.signals = sig
		}
	}
}

// WithTimeout задаёт максимальное время, отведённое на выполнение всех shutdown-функций.
//
// Если shutdown-функции не завершатся за это время — выполнение будет прервано,
// и Closer завершит shutdown с логированием таймаута.
//
// Значение <= 0 означает отсутствие таймаута (ожидание до завершения всех функций).
func WithTimeout(d time.Duration) Option {
	return func(c *Closer) {
		c.timeout = d
	}
}

// WithExecutionMode задаёт режим выполнения shutdown-функций.
//
// Доступные значения:
//   - closer.LIFOSequential (по умолчанию)
//   - closer.FIFOSequential
//   - closer.Parallel
func WithExecutionMode(m ExecutionMode) Option {
	return func(c *Closer) {
		c.execMode = m
	}
}
