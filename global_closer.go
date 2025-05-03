package closer

// globalCloser — глобальный экземпляр Closer, создаётся через Init().
// Используется в Add, Wait, CloseAll. Обращение до Init вызовет панику.
var globalCloser *Closer

// Init инициализирует глобальный Closer с указанными опциями.
//
// Вызывать Init необходимо один раз в начале работы приложения,
// до любых вызовов Add, Wait или CloseAll.
//
// Пример:
//
//	func main() {
//	    closer.Init(closer.WithTimeout(5 * time.Second))
//	    closer.Add(func() error { ... })
//	    ...
//	    closer.Wait()
//	}
func Init(opts ...Option) {
	globalCloser = New(opts...)
}

// Add регистрирует одну или несколько функций завершения для глобального Closer.
//
// Все функции будут вызваны при получении сигнала завершения или при вызове CloseAll.
//
// Panic, если Init не был вызван.
func Add(f ...func() error) {
	if globalCloser == nil {
		panic("closer: Init must be called before Add")
	}
	globalCloser.Add(f...)
}

// Wait блокирует выполнение до завершения graceful shutdown.
//
// Обычно вызывается в main, чтобы не завершать приложение,
// пока все shutdown-функции не завершатся.
//
// Panic, если Init не был вызван.
func Wait() {
	if globalCloser == nil {
		panic("closer: Init must be called before Wait")
	}
	globalCloser.Wait()
}

// CloseAll вручную запускает graceful shutdown, вызывая все Add-функции.
//
// Panic, если Init не был вызван.
func CloseAll() {
	if globalCloser == nil {
		panic("closer: Init must be called before CloseAll")
	}
	globalCloser.CloseAll()
}
