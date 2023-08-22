package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// SQLStorage contains metrics data in database.
type SQLStorage struct {
	con *sql.DB
}

// NewSQLStorage creates SQLStorage.
func NewSQLStorage(dsn string) (*SQLStorage, error) {
	db, err := sql.Open(databaseType, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect database crate error: %w", err)
	}
	storage := SQLStorage{
		con: db,
	}
	return &storage, checkDatabaseStructure(dsn)
}

// Update creates or updates metric value in storage.
func (ms *SQLStorage) Update(
	ctx context.Context,
	mType string,
	mName string,
	mValue string,
) error {
	switch mType {
	case counterType:
		counter, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return fmt.Errorf("counter value convert error: %w", err)
		}
		_, err = ms.updateCounter(ctx, mName, counter, ms.con)
		return err
	case gaugeType:
		gauges, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return fmt.Errorf("gauge value convert error: %w", err)
		}
		_, err = ms.updateGauge(ctx, mName, gauges, ms.con)
		return err
	default:
		return errors.New("metric type incorrect. Availible types are: guage or counter")
	}
}

// GetMetric returns the metric value as string.
func (ms *SQLStorage) GetMetric(
	ctx context.Context,
	mType string,
	mName string,
) (string, error) {
	switch mType {
	case gaugeType:
		value, err := ms.getGauge(ctx, mName)
		return fmt.Sprintf("%f", *value), err
	case counterType:
		value, err := ms.getCounter(ctx, mName)
		return fmt.Sprintf("%d", *value), err
	default:
		return "", fmt.Errorf("metric '%s' with type '%s' not found", mName, mType)
	}
}

// GetMetricsHTML returns all metrics values as HTML string.
func (ms *SQLStorage) GetMetricsHTML(ctx context.Context) (string, error) {
	gauges, err := ms.getAllMetricOfType(ctx, gaugeTableName)
	if err != nil {
		return "", fmt.Errorf("get gauges metrics error: %w", err)
	}
	counters, err := ms.getAllMetricOfType(ctx, counterTableName)
	if err != nil {
		return "", fmt.Errorf("get counters metrics error: %w", err)
	}

	body := "<!doctype html> <html lang='en'> <head> <meta charset='utf-8'> <title>Список метрик</title></head>"
	body += "<body><header><h1><p>Metrics list</p></h1></header>"
	body += "<h1><p>Gauges</p></h1>"
	for index, value := range *gauges {
		body += fmt.Sprintf("<nav><p>%d. %s</p></nav>", index+1, value)
	}
	body += "<h1><p>Counters</p></h1>"
	for index, value := range *counters {
		body += fmt.Sprintf("<nav><p>%d. %s</p></nav>", index+1, value)
	}
	body += "</body></html>"
	return body, nil
}

// updateOneMetric is private func for update storage.
func (ms *SQLStorage) updateOneMetric(ctx context.Context, m metric, connect SQLQueryInterface) (*metric, error) {
	switch m.MType {
	case counterType:
		if m.Delta != nil {
			value, err := ms.updateCounter(ctx, m.ID, *m.Delta, connect)
			if err != nil {
				return nil, err
			}
			m.Delta = value
		} else {
			return nil, errors.New("metric's delta indefined")
		}
	case gaugeType:
		if m.Value != nil {
			value, err := ms.updateGauge(ctx, m.ID, *m.Value, connect)
			if err != nil {
				return nil, err
			}
			m.Value = value
		} else {
			return nil, errors.New("metric's value indefined")
		}
	default:
		return nil, errors.New("metric type error, use counter like int64 or gauge like float64")
	}
	return &m, nil
}

// UpdateJSON creates or updates metric value in storage.
func (ms *SQLStorage) UpdateJSON(ctx context.Context, data []byte) ([]byte, error) {
	var metric metric
	err := json.Unmarshal(data, &metric)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %w", err)
	}

	item, err := ms.updateOneMetric(ctx, metric, ms.con)
	if err != nil {
		return nil, err
	}

	resp, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("convert to json error: %w", err)
	}
	return resp, nil
}

// GetMetricJSON returns the metric value as string.
func (ms *SQLStorage) GetMetricJSON(ctx context.Context, data []byte) ([]byte, error) {
	var metric metric
	err := json.Unmarshal(data, &metric)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %w", err)
	}

	switch metric.MType {
	case counterType:
		value, err := ms.getCounter(ctx, metric.ID)
		if err != nil {
			if value != nil {
				return []byte(""), err
			}
			return nil, err
		}
		metric.Delta = value
		resp, err := json.Marshal(metric)
		if err != nil {
			return nil, fmt.Errorf("marshal counter metric error: %w", err)
		}
		return resp, nil
	case gaugeType:
		value, err := ms.getGauge(ctx, metric.ID)
		if err != nil {
			if value != nil {
				return []byte(""), err
			}
			return nil, err
		}
		metric.Value = value
		resp, err := json.Marshal(metric)
		if err != nil {
			return nil, fmt.Errorf("marshal gauge metric error: %w", err)
		}
		return resp, nil
	default:
		return nil, fmt.Errorf("metric type ('%s') error, use counter like int64 or gauge like float64", metric.MType)
	}
}

// Save doesn't have mean. Used to satisfy the interface.
func (ms *SQLStorage) Save() error {
	return ms.PingDB(context.Background())
}

// PingDB checks connection to database server.
func (ms *SQLStorage) PingDB(ctx context.Context) error {
	if err := ms.con.PingContext(ctx); err != nil {
		return fmt.Errorf("check database ping error: %w", err)
	}
	return nil
}

// Clear deletes all metrics data from the database.
func (ms *SQLStorage) Clear(ctx context.Context) error {
	_, err := ms.con.ExecContext(ctx, "Delete from gauges;")
	if err != nil {
		return fmt.Errorf("clear gauges table error: %w", err)
	}
	_, err = ms.con.ExecContext(ctx, "Delete from counters;")
	if err != nil {
		return fmt.Errorf("clear counters table error: %w", err)
	}
	return nil
}

func sliceInsert(ctx context.Context, sqtx *sql.Tx, tbl string, mp map[string]string, excl string) error {
	if len(mp) == 0 {
		return nil
	}
	rs := make([]string, 0)
	values := make([]any, 0)
	for key, val := range mp {
		rs = append(rs, fmt.Sprintf("($%d, $%d)", len(rs)*2+1, len(rs)*2+2)) //nolint:gomnd //<-def values
		values = append(values, key)
		values = append(values, val)
	}
	query := "INSERT INTO " + tbl + " (name, value) values " + strings.Join(rs, ",") +
		" ON CONFLICT (name) DO UPDATE SET value=EXCLUDED.value" + excl + ";"
	_, err := sqtx.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("sliceInsert error: %w", err)
	}
	return nil
}

func mkMetricsMaps(metrics []metric) (map[string]string, map[string]string) {
	countersLst := make(map[string]int64)
	gaugeLst := make(map[string]string)
	for _, item := range metrics {
		switch item.MType {
		case counterType:
			if item.Delta == nil {
				continue
			}
			countersLst[item.ID] += *item.Delta
		case gaugeType:
			if item.Value == nil {
				continue
			}
			gaugeLst[item.ID] = strconv.FormatFloat(*item.Value, 'f', -1, 64)
		}
	}
	countersString := make(map[string]string)
	for key, value := range countersLst {
		countersString[key] = fmt.Sprintf("%d", value)
	}
	return countersString, gaugeLst
}

// UpdateJSONSlice updates the repository with metrics that are obtained
// by translating the received JSON into a list of metrics.
func (ms *SQLStorage) UpdateJSONSlice(
	ctx context.Context,
	data []byte,
) ([]byte, error) {
	var metrics []metric
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %w", err)
	}

	// запись данных в БД
	sqtx, err := ms.con.Begin()
	defer sqtx.Rollback() //nolint:errcheck //<-senselessly
	if err != nil {
		return nil, fmt.Errorf("transaction create error: %w", err)
	}

	counters, gauges := mkMetricsMaps(metrics)
	err = sliceInsert(ctx, sqtx, counterTableName, counters, fmt.Sprintf("+%s.value", counterTableName))
	if err != nil {
		return nil, fmt.Errorf("insert counters slice error: %w", err)
	}
	err = sliceInsert(ctx, sqtx, gaugeTableName, gauges, "")
	if err != nil {
		return nil, fmt.Errorf("insert gauges slice error: %w", err)
	}

	err = sqtx.Commit()
	if err != nil {
		return nil, fmt.Errorf("transaction commit error: %w", err)
	}
	return nil, nil
}

// Stop is closing connection to database.
func (ms *SQLStorage) Stop() error {
	err := ms.con.Close()
	if err != nil {
		return fmt.Errorf("database connection close error: %w", err)
	}
	return nil
}
