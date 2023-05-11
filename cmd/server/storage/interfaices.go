package storage

// Интерфей для установки значений в объект из строки
type StorageSeter interface {
	Update(string, string, string) (int, error)
}

// уже определён в системе, но всё же
type Stringer interface {
	String() string
}

// Интерфейс получения значения метрики
type StorageGetter interface {
	GetMetric(string, string) (string, int)
}

// Интерфейс для вывод значений в виде HTML
type HTMLGetter interface {
	GetMetricsHTML() string
}

type Storage interface {
	StorageSeter
	StorageGetter
	Stringer
	HTMLGetter
}
