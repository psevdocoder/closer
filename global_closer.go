package closer

// globalCloser — глобальный экземпляр Closer.
// Инициализируется только через Init.
var globalCloser *Closer

// Init инициализирует глобальный Closer с указанными опциями.
// Вызывать нужно до регистрации Add/Wait/CloseAll.
func Init(opts ...Option) {
	globalCloser = New(opts...)
}

// Add регистрирует функции завершения для глобального Closer
// panic, если Init не был вызван
func Add(f ...func() error) {
	if globalCloser == nil {
		panic("closer: Init must be called before Add")
	}
	globalCloser.Add(f...)
}

// Wait блокирует до завершения graceful shutdown
// panic, если Init не был вызван
func Wait() {
	if globalCloser == nil {
		panic("closer: Init must be called before Wait")
	}
	<-globalCloser.done
}

// CloseAll вручную запускает graceful shutdown
// panic, если Init не был вызван
func CloseAll() {
	if globalCloser == nil {
		panic("closer: Init must be called before CloseAll")
	}
	globalCloser.CloseAll()
}
