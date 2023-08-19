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

type (
	StorageSetter interface {
		Update(context.Context, string, string, string) error
		UpdateJSON(context.Context, []byte) ([]byte, error)
		UpdateJSONSlice(context.Context, []byte) ([]byte, error)
		Save() error
	}

	// Интерфейс получения значения метрики
	StorageGetter interface {
		GetMetric(context.Context, string, string) (string, error)
		GetMetricJSON(context.Context, []byte) ([]byte, error)
	}

	// интерфейс для работы с БД
	StorageDB interface {
		PingDB(context.Context) error
		Clear(context.Context) error
	}

	// Интерфейс для вывод значений в виде HTML
	HTMLGetter interface {
		GetMetricsHTML(context.Context) (string, error)
	}

	// -----------------------------------------------------------------------------------
	// Функции для повторения действий при ошибках
	// -----------------------------------------------------------------------------------
	fbe func(context.Context, []byte) ([]byte, error)

	fse func(context.Context) (string, error)

	fsse func(context.Context, string, string) (string, error)

	fssse func(context.Context, string, string, string) error

	// -----------------------------------------------------------------------------------
	// Определение функций, которые используют интерфейсы
	// -----------------------------------------------------------------------------------
	getMetricsArgs struct {
		mType string
		mName string
	}

	updateMetricsArgs struct {
		base   getMetricsArgs
		mValue string
	}
)

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

// Обработка запроса на добавление или изменение метрики
func Update(
	ctx context.Context,
	storage StorageSetter,
	metric updateMetricsArgs,
) (int, error) {
	err := ssseRepeater(storage.Update, ctx, metric.base.mType, metric.base.mName, metric.mValue)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("update metric error: %w", err)
	}
	return http.StatusOK, nil
}

// Обработка запроса значения метрики
func GetMetric(
	ctx context.Context,
	storage StorageGetter,
	metric getMetricsArgs,
) ([]byte, error) {
	body, err := sseRepeater(storage.GetMetric, ctx, metric.mType, metric.mName)
	if err != nil {
		return nil, fmt.Errorf("metric not found error: %w", err)
	}
	return []byte(body), nil
}

// Запрос всех метрик в html
func GetAllMetrics(
	ctx context.Context,
	storage HTMLGetter,
) ([]byte, error) {
	body, err := seRepeater(storage.GetMetricsHTML, ctx)
	if err != nil {
		return nil, fmt.Errorf("get metrics in storage error: %w", err)
	}
	return []byte(body), nil
}

// обновление в JSON формате
func UpdateJSON(
	ctx context.Context,
	body []byte,
	storage StorageSetter,
) ([]byte, error) {
	data, err := bytesErrorRepeater(storage.UpdateJSON, ctx, body)
	if err != nil {
		return nil, fmt.Errorf("update metric error: %w", err)
	}
	return data, nil
}

// получение метрики в JSON формате
func GetMetricJSON(
	writer http.ResponseWriter,
	request *http.Request,
	storage StorageGetter,
	logger *zap.SugaredLogger,
	key []byte) {
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
func Ping(
	ctx context.Context,
	storage StorageDB,
) (int, error) {
	err := storage.PingDB(ctx)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("ping error: %w", err)
	}
	return http.StatusOK, nil
}

// очистка storage
func Clear(
	ctx context.Context,
	storage StorageDB,
) (int, error) {
	err := storage.Clear(ctx)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("clear storage error: %w", err)
	}
	return http.StatusOK, nil
}

// обновление списком json
func UpdateJSONSLice(
	ctx context.Context,
	data []byte,
	storage StorageSetter,
) ([]byte, error) {
	body, err := bytesErrorRepeater(storage.UpdateJSONSlice, ctx, data)
	if err != nil {
		return nil, fmt.Errorf("update metrics list error: %w", err)
	}
	return body, nil
}
