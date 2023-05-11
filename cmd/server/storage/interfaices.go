package storage

// Интерфей для установки значений в объект из строки
type StorageSeter interface {
	Update(string, string, string) (int, error)
}

type Stringer interface {
	String() string
}

// Интерфейс получения значения метрики
type StorageGetter interface {
	GetMetric(string, string) (string, int)
}

// Интерфейс для вывод значений в виде HTML
type HtmlGetter interface {
	GetMetricsHTML() string
}

type Storage interface {
	StorageSeter
	StorageGetter
	Stringer
	HtmlGetter
}
