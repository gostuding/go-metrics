package main

// Интерфей для установки значений в объект из строки
type Seter interface {
	AddMetric(string) (int, error)
}

type Stringer interface {
	String() string
}

// Интерфейс для определения объекта MemStorage
type Storager interface {
	Seter
	Stringer
}
