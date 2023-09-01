package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// StorageSetter is interface for set data in storage.
type (
	StorageSetter interface {
		Update(context.Context, string, string, string) error
		UpdateJSON(context.Context, []byte) ([]byte, error)
		UpdateJSONSlice(context.Context, []byte) ([]byte, error)
		Save() error
	}

	// Saver interface.
	Saver interface {
		Save() error
	}
	// StorageGetter is interface for get data from storage.
	StorageGetter interface {
		GetMetric(context.Context, string, string) (string, error)
		GetMetricJSON(context.Context, []byte) ([]byte, error)
		GetMetricsHTML(context.Context) (string, error)
	}

	// StorageDB is additions storage work interface.
	StorageDB interface {
		PingDB(context.Context) error
		Clear(context.Context) error
		Stop() error
	}

	// Private type. Repeate funcs type.
	fbe func(context.Context, []byte) ([]byte, error)

	// Private type. Repeate funcs type.
	fse func(context.Context) (string, error)

	// Private type. Repeate funcs type.
	fsse func(context.Context, string, string) (string, error)

	// Private type. Repeate funcs type.
	fssse func(context.Context, string, string, string) error

	// Private interface. Is using for args number insreace.
	getMetricsArgs struct {
		mType string
		mName string
	}

	// Private interface. Is using for args number insreace.
	updateMetricsArgs struct {
		base   getMetricsArgs
		mValue string
	}
)

const (
	contextErrType = iota
	updateMetricErrorType
	pingErrorType
)

// Private func for get default type errors.
func getError(errType any, values ...any) error {
	switch errType {
	case contextErrType:
		return fmt.Errorf("context error: %w", values...)
	case updateMetricErrorType:
		return fmt.Errorf("update metric error: %w", values...)
	case pingErrorType:
		return fmt.Errorf("ping error: %w", values...)
	default:
		return fmt.Errorf("error undefined: %w", values...)
	}
}

// Private func.
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

// Private func.
func bytesErrorRepeater(ctx context.Context, f fbe, data []byte) ([]byte, error) {
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
				value, err = f(ctx, data)
				if err == nil {
					return value, nil
				}
			}
		}
	}
	return value, err
}

// Private func.
func seRepeater(ctx context.Context, f fse) (string, error) {
	value, err := f(ctx)
	if err != nil {
		waitTime := 1
		for i := 0; i < 3; i++ {
			select {
			case <-ctx.Done():
				return "", getError(contextErrType, ctx.Err())
			default:
				if !isRepeat(err, &waitTime) {
					return value, err
				}
				value, err = f(ctx)
				if err == nil {
					return value, nil
				}
			}
		}
	}
	return value, err
}

// Private func.
func sseRepeater(ctx context.Context, f fsse, t string, n string) (string, error) {
	value, err := f(ctx, t, n)
	if err != nil {
		waitTime := 1
		for i := 0; i < 3; i++ {
			select {
			case <-ctx.Done():
				return "", getError(contextErrType, ctx.Err())
			default:
				if !isRepeat(err, &waitTime) {
					return value, err
				}
				value, err = f(ctx, t, n)
				if err == nil {
					return value, nil
				}
			}
		}
	}
	return value, err
}

// Private func.
func ssseRepeater(ctx context.Context, f fssse, t string, n string, v string) error {
	err := f(ctx, t, n, v)
	if err != nil {
		waitTime := 1
		for i := 0; i < 3; i++ {
			select {
			case <-ctx.Done():
				return getError(contextErrType, ctx.Err())
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

// Update is processing an update metric request.
func Update(
	ctx context.Context,
	storage StorageSetter,
	metric updateMetricsArgs,
) (int, error) {
	err := ssseRepeater(ctx, storage.Update, metric.base.mType, metric.base.mName, metric.mValue)
	if err != nil {
		return http.StatusBadRequest, getError(updateMetricErrorType, err)
	}
	return http.StatusOK, nil
}

// GetMetric is processing an get one metric request.
func GetMetric(
	ctx context.Context,
	storage StorageGetter,
	metric getMetricsArgs,
) ([]byte, error) {
	body, err := sseRepeater(ctx, storage.GetMetric, metric.mType, metric.mName)
	if err != nil {
		return nil, fmt.Errorf("metric not found error: %w", err)
	}
	return []byte(body), nil
}

// GetAllMetrics is processing an get all metrics request.
func GetAllMetrics(
	ctx context.Context,
	storage StorageGetter,
) ([]byte, error) {
	body, err := seRepeater(ctx, storage.GetMetricsHTML)
	if err != nil {
		return nil, fmt.Errorf("get metrics in storage error: %w", err)
	}
	return []byte(body), nil
}

// UpdateJSON is processing an update metric by JSON request.
func UpdateJSON(
	ctx context.Context,
	body []byte,
	storage StorageSetter,
) ([]byte, error) {
	data, err := bytesErrorRepeater(ctx, storage.UpdateJSON, body)
	if err != nil {
		return nil, getError(updateMetricErrorType, err)
	}
	return data, nil
}

// GetMetricJSON is processing an get metric by JSON request.
func GetMetricJSON(
	ctx context.Context,
	storage StorageGetter,
	body []byte,
) ([]byte, int, error) {
	data, err := bytesErrorRepeater(ctx, storage.GetMetricJSON, body)
	if err != nil {
		if data != nil {
			return nil, http.StatusNotFound, fmt.Errorf("metric not found error")
		} else {
			return nil, http.StatusBadRequest, err
		}
	}
	return data, http.StatusOK, nil
}

// Ping is checks connection to database.
func Ping(
	ctx context.Context,
	storage StorageDB,
) (int, error) {
	err := storage.PingDB(ctx)
	if err != nil {
		return http.StatusInternalServerError, getError(pingErrorType, err)
	}
	return http.StatusOK, nil
}

// Clear is processing storage clearing.
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

// UpdateJSONSLice is processing an update metrics by JSON slice request.
func UpdateJSONSLice(
	ctx context.Context,
	data []byte,
	storage StorageSetter,
) ([]byte, error) {
	body, err := bytesErrorRepeater(ctx, storage.UpdateJSONSlice, data)
	if err != nil {
		return nil, fmt.Errorf("update metrics list error: %w", err)
	}
	return body, nil
}
