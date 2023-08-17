package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

// -----------------------------------------------------------------------------------
// Определение интерфейсов
// -----------------------------------------------------------------------------------
// Интерфей для установки значений в объект из строки
type StorageSetter interface {
	Update(context.Context, string, string, string) error
	UpdateJSON(context.Context, []byte) ([]byte, error)
	UpdateJSONSlice(context.Context, []byte) ([]byte, error)
	Save() error
}

// Интерфейс получения значения метрики
type StorageGetter interface {
	GetMetric(context.Context, string, string) (string, error)
	GetMetricJSON(context.Context, []byte) ([]byte, error)
}

// интерфейс для работы с БД
type StorageDB interface {
	PingDB(context.Context) error
	Clear(context.Context) error
}

// Интерфейс для вывод значений в виде HTML
type HTMLGetter interface {
	GetMetricsHTML(context.Context) (string, error)
}

// -----------------------------------------------------------------------------------
// Функции для повторения действий при ошибках
// -----------------------------------------------------------------------------------
type fbe func(context.Context, []byte) ([]byte, error)

type fse func(context.Context) (string, error)

type fsse func(context.Context, string, string) (string, error)

type fssse func(context.Context, string, string, string) error

func isRepeat(err error, t *int) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
		time.Sleep(time.Duration(*t) * time.Second)
		*t += 2
	} else {
		return false
	}
	return true
}

func bytesErrorRepeater(f fbe, ctx context.Context, data []byte) ([]byte, error) {
	value, err := f(ctx, data)
	if err != nil {
		waitTime := 1
		for i := 0; i < 3; i++ {
			select {
			case <-ctx.Done():
				return nil, errors.New("context done error")
			default:
				if !isRepeat(err, &waitTime) {
					return value, err
				}
				rez, err := f(ctx, data)
				if err == nil {
					return rez, nil
				}
			}
		}
	}
	return value, err
}

func seRepeater(f fse, ctx context.Context) (string, error) {
	value, err := f(ctx)
	if err != nil {
		waitTime := 1
		for i := 0; i < 3; i++ {
			select {
			case <-ctx.Done():
				return "", fmt.Errorf("context error: %w", ctx.Err())
			default:
				if !isRepeat(err, &waitTime) {
					return value, err
				}
				value, err = f(ctx)
				if err == nil {
					return value, err
				}
			}
		}
	}
	return value, err
}

func sseRepeater(f fsse, ctx context.Context, t string, n string) (string, error) {
	value, err := f(ctx, t, n)
	if err != nil {
		waitTime := 1
		for i := 0; i < 3; i++ {
			select {
			case <-ctx.Done():
				return "", fmt.Errorf("context error: %w", ctx.Err())
			default:
				if !isRepeat(err, &waitTime) {
					return value, err
				}
				value, err = f(ctx, t, n)
				if err == nil {
					return value, err
				}
			}
		}
	}
	return value, err
}

func ssseRepeater(f fssse, ctx context.Context, t string, n string, v string) error {
	err := f(ctx, t, n, v)
	if err != nil {
		waitTime := 1
		for i := 0; i < 3; i++ {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context error: %w", ctx.Err())
			default:
				if !isRepeat(err, &waitTime) {
					return err
				}
				err = f(ctx, t, n, v)
				if err == nil {
					return nil
				}
			}
		}
	}
	return err
}

// -----------------------------------------------------------------------------------
// Определение функций, которые используют интерфейсы
// -----------------------------------------------------------------------------------
type getMetricsArgs struct {
	mType string
	mName string
}

type updateMetricsArgs struct {
	base   getMetricsArgs
	mValue string
}

// Обработка запроса на добавление или изменение метрики
func Update(writer http.ResponseWriter, request *http.Request, storage StorageSetter, metric updateMetricsArgs, logger *zap.SugaredLogger) {
	err := ssseRepeater(storage.Update, request.Context(), metric.base.mType, metric.base.mName, metric.mValue)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warnf("update metric error: %w", err)
		return
	}
	writer.WriteHeader(http.StatusOK)
	logger.Debugf("update metric '%s' success", metric.base.mType)
}

// Обработка запроса значения метрики
func GetMetric(writer http.ResponseWriter, request *http.Request, storage StorageGetter, metric getMetricsArgs,
	logger *zap.SugaredLogger, key []byte) {
	body, err := sseRepeater(storage.GetMetric, request.Context(), metric.mType, metric.mName)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		logger.Warn(err)
		return
	}
	_, err = writer.Write([]byte(body))
	if err != nil {
		logger.Warnf("write data to client error: %w", err)
	}
}

// Запрос всех метрик в html
func GetAllMetrics(writer http.ResponseWriter, request *http.Request, storage HTMLGetter,
	logger *zap.SugaredLogger, key []byte) {
	writer.Header().Set("Content-Type", "text/html")
	body, err := seRepeater(storage.GetMetricsHTML, request.Context())
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warnf("get metrics in html error: %w", err)
		return
	}
	_, err = writer.Write([]byte(body))
	if err != nil {
		logger.Warnf("write metrics data to client error: %w", err)
	}
}

// обновление в JSON формате
func UpdateJSON(writer http.ResponseWriter, request *http.Request, storage StorageSetter,
	logger *zap.SugaredLogger, key []byte) {
	writer.Header().Set("Content-Type", "application/json")
	data, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warnf("read request body error, %s: %w", data, err)
		return
	}
	body, err := bytesErrorRepeater(storage.UpdateJSON, request.Context(), data)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warnf("update metric error: %w", err)
		return
	}
	logger.Debug("update metric by json success")
	_, err = writer.Write(body)
	if err != nil {
		logger.Warnf("write data to client error: %w", err)
	}
}

// получение метрики в JSON формате
func GetMetricJSON(writer http.ResponseWriter, request *http.Request, storage StorageGetter,
	logger *zap.SugaredLogger, key []byte) {
	writer.Header().Set("Content-Type", "application/json")
	data, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warnf("get metric json, read request body error: %w", err)
		return
	}
	body, err := bytesErrorRepeater(storage.GetMetricJSON, request.Context(), data)
	if err != nil {
		if body != nil {
			writer.WriteHeader(http.StatusNotFound)
		} else {
			writer.WriteHeader(http.StatusBadRequest)
		}
		logger.Warnf("get metric json error: %w", err)
		return
	}
	_, err = writer.Write(body)
	if err != nil {
		logger.Warnf("get metric json, write data to client error: %w", err)
	}
}

// проверка подключения к БД
func Ping(writer http.ResponseWriter, request *http.Request, storage StorageDB, logger *zap.SugaredLogger) {
	writer.Header().Set("Content-Type", "")
	err := storage.PingDB(request.Context())
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		logger.Warnln(err)
		return
	}
	logger.Debug("database ping success")
	writer.WriteHeader(http.StatusOK)
}

// очистка storage
func Clear(writer http.ResponseWriter, request *http.Request, storage StorageDB, logger *zap.SugaredLogger) {
	writer.Header().Set("Content-Type", "")
	err := storage.Clear(request.Context())
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		logger.Warnln(err)
		return
	}
	logger.Debug("storage clear success")
	writer.WriteHeader(http.StatusOK)
}

// обновление списком json
func UpdateJSONSLice(writer http.ResponseWriter,
	request *http.Request,
	storage StorageSetter,
	logger *zap.SugaredLogger,
	key []byte) {
	writer.Header().Set("Content-Type", "text/html")
	data, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warnf("read request body error: %w", err)
		return
	}

	body, err := bytesErrorRepeater(storage.UpdateJSONSlice, request.Context(), data)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warnf("update metrics list error: %w", err)
		return
	}

	logger.Debug("update metrics by json list success")
	_, err = writer.Write(body)
	if err != nil {
		logger.Warnf("write data to client error: %w", err)
	}
}
