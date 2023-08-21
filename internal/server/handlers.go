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

type (
	// interface for set data in storage
	StorageSetter interface {
		Update(context.Context, string, string, string) error
		UpdateJSON(context.Context, []byte) ([]byte, error)
		UpdateJSONSlice(context.Context, []byte) ([]byte, error)
		Save() error
	}

	// interface for get data from storage
	StorageGetter interface {
		GetMetric(context.Context, string, string) (string, error)
		GetMetricJSON(context.Context, []byte) ([]byte, error)
		GetMetricsHTML(context.Context) (string, error)
	}

	// additions storage work interface
	StorageDB interface {
		PingDB(context.Context) error
		Clear(context.Context) error
		Stop() error
	}

	// ptivate type. repeate funcs type
	fbe func(context.Context, []byte) ([]byte, error)

	// ptivate type. repeate funcs type
	fse func(context.Context) (string, error)

	// ptivate type. repeate funcs type
	fsse func(context.Context, string, string) (string, error)

	// ptivate type. repeate funcs type
	fssse func(context.Context, string, string, string) error

	// private interface. Is using for args number insreace.
	getMetricsArgs struct {
		mType string
		mName string
	}

	// private interface. Is using for args number insreace.
	updateMetricsArgs struct {
		base   getMetricsArgs
		mValue string
	}
)

// private func
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

// private func
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

// private func
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

// private func
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

// private func
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

// Update is processing an update metric request.
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

// GetMetric is processing an get one metric request.
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

// GetAllMetrics is processing an get all metrics request.
func GetAllMetrics(
	ctx context.Context,
	storage StorageGetter,
) ([]byte, error) {
	body, err := seRepeater(storage.GetMetricsHTML, ctx)
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
	data, err := bytesErrorRepeater(storage.UpdateJSON, ctx, body)
	if err != nil {
		return nil, fmt.Errorf("update metric error: %w", err)
	}
	return data, nil
}

// GetMetricJSON is processing an get metric by JSON request.
func GetMetricJSON(
	ctx context.Context,
	storage StorageGetter,
	body []byte,
) ([]byte, int, error) {
	data, err := bytesErrorRepeater(storage.GetMetricJSON, ctx, body)
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
		return http.StatusInternalServerError, fmt.Errorf("ping error: %w", err)
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
	body, err := bytesErrorRepeater(storage.UpdateJSONSlice, ctx, data)
	if err != nil {
		return nil, fmt.Errorf("update metrics list error: %w", err)
	}
	return body, nil
}
