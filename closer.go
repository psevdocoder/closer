package closer

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

var (
	mu    sync.Mutex
	funcs []Func
)

// Add добавляет функцию завершения в список.
func Add(f Func) {
	mu.Lock()
	defer mu.Unlock()

	funcs = append(funcs, f)
}

// Close выполняет все зарегистрированные функции завершения.
// Если выполнение завершено с ошибками, они возвращаются как одна ошибка.
// В случае отмены контекста возвращается ошибка отмены.
func Close(ctx context.Context) error {
	mu.Lock()
	defer mu.Unlock()

	var (
		msgs     = make([]string, 0, len(funcs))
		complete = make(chan struct{}, 1)
	)

	go func() {
		for _, f := range funcs {
			if err := f(ctx); err != nil {
				msgs = append(msgs, fmt.Sprintf("[!] %v", err))
			}
		}

		complete <- struct{}{}
	}()

	select {
	case <-complete:
		break
	case <-ctx.Done():
		return fmt.Errorf("shutdown cancelled: %v", ctx.Err())
	}

	if len(msgs) > 0 {
		return fmt.Errorf(
			"shutdown finished with error(s): \n%s",
			strings.Join(msgs, "\n"),
		)
	}

	return nil
}

// Func представляет функцию завершения, которая принимает контекст
// и возвращает ошибку в случае неудачного завершения.
type Func func(ctx context.Context) error
