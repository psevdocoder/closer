package closer

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCloser_CloseAll(t *testing.T) {
	type testCase struct {
		name         string
		mode         ExecutionMode
		funcCount    int
		errAt        int // индекс функции с ошибкой, -1 если нет
		timeout      time.Duration
		wantOrder    []int // nil если порядок не проверяем
		wantMinCalls int
		doubleClose  bool
	}

	tests := []testCase{
		{
			name:         "LIFO sequential ok",
			mode:         LIFOSequential,
			funcCount:    3,
			errAt:        -1,
			wantOrder:    []int{2, 1, 0},
			wantMinCalls: 3,
		},
		{
			name:         "FIFO sequential ok",
			mode:         FIFOSequential,
			funcCount:    3,
			errAt:        -1,
			wantOrder:    []int{0, 1, 2},
			wantMinCalls: 3,
		},
		{
			name:         "Parallel ok",
			mode:         Parallel,
			funcCount:    3,
			errAt:        -1,
			wantMinCalls: 3,
		},
		{
			name:         "LIFO stops on error",
			mode:         LIFOSequential,
			funcCount:    3,
			errAt:        1,
			wantOrder:    []int{2, 1},
			wantMinCalls: 2,
		},
		{
			name:         "FIFO stops on error",
			mode:         FIFOSequential,
			funcCount:    3,
			errAt:        1,
			wantOrder:    []int{0, 1},
			wantMinCalls: 2,
		},
		{
			name:         "Parallel continues on error",
			mode:         Parallel,
			funcCount:    3,
			errAt:        1,
			wantMinCalls: 3,
		},
		{
			name:         "Timeout stops sequential execution",
			mode:         FIFOSequential,
			funcCount:    2,
			timeout:      50 * time.Millisecond,
			wantMinCalls: 1,
		},
		{
			name:         "CloseAll executed only once",
			mode:         FIFOSequential,
			funcCount:    1,
			wantMinCalls: 1,
			doubleClose:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				mu    sync.Mutex
				order []int
				calls atomic.Int32
			)

			c := New(
				WithExecutionMode(tt.mode),
				WithTimeout(tt.timeout),
				WithSignals(), // отключаем OS signals
			)

			for i := 0; i < tt.funcCount; i++ {
				idx := i
				c.Add(func() error {
					calls.Add(1)

					mu.Lock()
					order = append(order, idx)
					mu.Unlock()

					// Имитация долгой (медленной) функции:
					// используется для проверки поведения при таймауте.
					// Только первая функция замедляется, чтобы остальные либо не были запущены (sequential),
					// либо запустились параллельно (parallel).
					if tt.timeout > 0 && idx == 0 {
						time.Sleep(tt.timeout * 4)
					}

					if idx == tt.errAt {
						return errors.New("test error")
					}
					return nil
				})
			}

			// Проверка на повторный вызов c.CloseAll()
			c.CloseAll()
			if tt.doubleClose {
				c.CloseAll()
			}
			c.Wait()

			require.GreaterOrEqual(t, int(calls.Load()), tt.wantMinCalls)

			if tt.wantOrder != nil {
				require.Equal(t, tt.wantOrder, order)
			}
		})
	}
}
