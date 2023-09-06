# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).


## Документация проекта

В директории проекта выполнить:
```
godoc -http ":8080" 
```
где `":8080"` - адрес для запуска сервера с документацией.
Для открытия вкладки с документацией проекта перейти по ссылке:
```
http://localhost:8080/pkg/github.com/gostuding/go-metrics/?m=all
```
## Запуск тестов golangci-lint

Установите локально golangci-lint (см. официальную документацию)
В директории проекта выполните команду:
```
./golint_run.sh
```
Реузльтаты работы golangci-lint будут отображены в файле `./golangci-lint/report.json`

## Запуск юнит-тестов 

В пакете internal/serve/storage хранятся тесты для MemStorage и SqlStorage.
При выполнении команды ```go test ./...``` будут запущены ттесты только для MemStorage.
```
go test ./...
```
Для включения тестов для SqlStorage необходимо указать ```--tags=sql_storage```, 
а также строку подключения к БД: ```-args dsn="connection"```.
Пример запуска тестов sql_storage:
```
go test ./... --tags=sql_storage -args dsn="host=localhost user=postgres database=metrics"
``` 

## Компиляция серверной части проекта

Для компиляции серверной части проекта выполните команду:

```
go build -ldflags "-X 'main.buildVersion=VERSION' -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')'  -X 'main.buildCommit=COMMENT'" cmd/server/main.go 
```
Где ```VERSION``` -  версия сборки, а ```COMMENT``` - коментария пользователя.
При запуске серверной части проекта будет выведены версия, дата и коментарий пользователя.
В качестве примера, строка сборки может выглядеть так: 

```
go build -ldflags "-X 'main.buildVersion=v1.0.01' -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')'  -X 'main.buildCommit=INIT RELEASE'" cmd/server/main.go
```
При завуске скомпилированного исполняемого файла будет выведена информация о сборке:
```
Build version: v1.0.01
Build date: 2023/09/03 22:00:46
Build commit: INIT RELEASE
```
Если параметры не указаны, то вывод будет следующим:
```
Build version: N/A
Build date: N/A
Build commit: N/A
```

## Компиляция агента проекта

Параметры аналогичны выше описанным, за исключением пути до main.go файла. Пример команды:
```
go build -ldflags "-X 'main.buildVersion=v1.0.01' -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')'  -X 'main.buildCommit=INIT RELEASE'" cmd/agent/main.go
```

## Статические анализаторы

В проект добавлен набор основных статических анализаторов. Исходный код содержится в ```cmd/staticlint```. 
При необходимости, можно внести изменения и скомпилировать исполняемый файл командой:
```
go build -o staticlint cmd/staticlint/main.go
```
Для запуска статических анализаторов запустить собранный файл командой:
```
staticlint ./...
```

