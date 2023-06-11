package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type SQLStorage struct {
	ConnectDBString string
	con             *sql.DB
	Logger          *zap.SugaredLogger
}

func NewSQLStorage(DBconnect string, logger *zap.SugaredLogger) (*SQLStorage, error) {
	db, err := sql.Open("pgx", DBconnect)
	if err != nil {
		return nil, fmt.Errorf("connect database crate error: %w", err)
	}
	storage := SQLStorage{
		con:             db,
		ConnectDBString: DBconnect,
		Logger:          logger,
	}
	return &storage, checkDatabaseStructure(DBconnect)
}

func (ms *SQLStorage) Update(ctx context.Context, mType string, mName string, mValue string) error {
	switch mType {
	case "counter":
		counter, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return fmt.Errorf("counter value convert error: %w", err)
		}
		_, err = ms.updateCounter(ctx, mName, counter, ms.con)
		return err
	case "gauge":
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

// Получение значения метрики по типу и имени
func (ms *SQLStorage) GetMetric(ctx context.Context, mType string, mName string) (string, error) {
	switch mType {
	case "gauge":
		value, err := ms.getGauge(ctx, mName)
		return fmt.Sprintf("%f", *value), err
	case "counter":
		value, err := ms.getCounter(ctx, mName)
		return fmt.Sprintf("%d", *value), err
	default:
		return "", fmt.Errorf("metric '%s' with type '%s' not found", mName, mType)
	}
}

// Список всех метрик в html
func (ms *SQLStorage) GetMetricsHTML(ctx context.Context) (string, error) {
	gauges, err := ms.getAllMetricOfType(ctx, "gauges")
	if err != nil {
		return "", fmt.Errorf("get gauges metrics error: %w", err)
	}
	counters, err := ms.getAllMetricOfType(ctx, "counters")
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

func (ms *SQLStorage) updateOneMetric(ctx context.Context, m metric, connect SQLQueryInterface) (*metric, error) {
	switch m.MType {
	case "counter":
		if m.Delta != nil {
			value, err := ms.updateCounter(ctx, m.ID, *m.Delta, connect)
			if err != nil {
				return nil, err
			}
			m.Delta = value
		} else {
			return nil, errors.New("metric's delta indefined")
		}
	case "gauge":
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

// обновление через json
func (ms *SQLStorage) UpdateJSON(ctx context.Context, data []byte) ([]byte, error) {
	var metric metric
	err := json.Unmarshal(data, &metric)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %w", err)
	}

	db, err := sql.Open("pgx", ms.ConnectDBString)
	if err != nil {
		return nil, fmt.Errorf("connect database error: %w", err)
	}
	defer db.Close()

	item, err := ms.updateOneMetric(ctx, metric, db)
	if err != nil {
		return nil, err
	}

	resp, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("convert to json error: %w", err)
	}
	return resp, nil
}

// запрос метрик через json
func (ms *SQLStorage) GetMetricJSON(ctx context.Context, data []byte) ([]byte, error) {
	var metric metric
	err := json.Unmarshal(data, &metric)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %w", err)
	}

	switch metric.MType {
	case "counter":
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
			return nil, err
		}
		return resp, nil
	case "gauge":
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
			return nil, err
		}
		return resp, nil
	default:
		return nil, fmt.Errorf("metric type ('%s') error, use counter like int64 or gauge like float64", metric.MType)
	}
}

func (ms *SQLStorage) Save() error {
	// метод - заглушка, проверка подключения к БД, т.к. все данные хранятся в БД
	return ms.PingDB(context.Background())
}

// проверка подключения к БД
func (ms *SQLStorage) PingDB(ctx context.Context) error {
	if ms.ConnectDBString == "" {
		return fmt.Errorf("connect DB string undefined")
	}

	if err := ms.con.PingContext(ctx); err != nil {
		return fmt.Errorf("check database ping error: %w", err)
	}
	return nil
}

// очистка БД
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

// обновление через json slice
func (ms *SQLStorage) UpdateJSONSlice(ctx context.Context, data []byte) ([]byte, error) {
	var metrics []metric
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %w", err)
	}

	db, err := sql.Open("pgx", ms.ConnectDBString)
	if err != nil {
		return nil, fmt.Errorf("connect database error: %w", err)
	}
	defer db.Close()

	for _, item := range metrics {
		switch item.MType {
		case "counter":
			_, err = ms.updateCounter(ctx, item.ID, *item.Delta, db)
			if err != nil {
				return nil, err
			}
		case "gauge":
			_, err = ms.updateGauge(ctx, item.ID, *item.Value, db)
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}
