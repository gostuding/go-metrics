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

type sqlStorage struct {
	ConnectDBString string             `json:"-"`
	Logger          *zap.SugaredLogger `json:"-"`
}

func NewSQLStorage(DBconnect string, logger *zap.SugaredLogger) (*sqlStorage, error) {
	storage := sqlStorage{
		ConnectDBString: DBconnect,
		Logger:          logger,
	}
	return &storage, checkDatabaseStructure(DBconnect, logger)
}

func (ms *sqlStorage) Update(ctx context.Context, mType string, mName string, mValue string) error {
	db, err := sql.Open("pgx", ms.ConnectDBString)
	if err != nil {
		return fmt.Errorf("connect database error: %v", err)
	}
	defer db.Close()

	switch mType {
	case "counter":
		counter, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return fmt.Errorf("counter value convert error: %v", err)
		}
		_, err = ms.updateCounter(ctx, mName, counter, db)
		return err
	case "gauge":
		gauges, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return fmt.Errorf("gauge value convert error: %v", err)
		}
		_, err = ms.updateGauge(ctx, mName, gauges, db)
		return err

	default:
		return errors.New("metric type incorrect. Availible types are: guage or counter")
	}
}

// Получение значения метрики по типу и имени
func (ms *sqlStorage) GetMetric(ctx context.Context, mType string, mName string) (string, error) {
	db, err := sql.Open("pgx", ms.ConnectDBString)
	if err != nil {
		return "", fmt.Errorf("connect database error: %v", err)
	}
	defer db.Close()

	switch mType {
	case "gauge":
		value, err := ms.getGauge(ctx, mName, db)
		return fmt.Sprintf("%v", *value), err
	case "counter":
		value, err := ms.getCounter(ctx, mName, db)
		return fmt.Sprintf("%d", *value), err
	default:
		return "", fmt.Errorf("metric '%s' with type '%s' not found", mName, mType)
	}
}

// Список всех метрик в html
func (ms *sqlStorage) GetMetricsHTML(ctx context.Context) (string, error) {
	db, err := sql.Open("pgx", ms.ConnectDBString)
	if err != nil {
		return "", fmt.Errorf("connect database error: %v", err)
	}
	defer db.Close()

	gauges, err := ms.getAllMetricOfType(ctx, "gauges", db)
	if err != nil {
		return "", fmt.Errorf("get gauges metrics error: %v", err)
	}
	counters, err := ms.getAllMetricOfType(ctx, "counters", db)
	if err != nil {
		return "", fmt.Errorf("get counters metrics error: %v", err)
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

func (ms *sqlStorage) updateOneMetric(ctx context.Context, m metric, connect SQLQueryInterface) (*metric, error) {
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
func (ms *sqlStorage) UpdateJSON(ctx context.Context, data []byte) ([]byte, error) {
	var metric metric
	err := json.Unmarshal(data, &metric)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %v", err)
	}

	db, err := sql.Open("pgx", ms.ConnectDBString)
	if err != nil {
		return nil, fmt.Errorf("connect database error: %v", err)
	}
	defer db.Close()

	item, err := ms.updateOneMetric(ctx, metric, db)
	if err != nil {
		return nil, err
	}

	resp, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("convert to json error: %v", err)
	}
	return resp, nil
}

// запрос метрик через json
func (ms *sqlStorage) GetMetricJSON(ctx context.Context, data []byte) ([]byte, error) {
	var metric metric
	err := json.Unmarshal(data, &metric)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %v", err)
	}

	db, err := sql.Open("pgx", ms.ConnectDBString)
	if err != nil {
		return nil, fmt.Errorf("connect database error: %v", err)
	}
	defer db.Close()

	switch metric.MType {
	case "counter":
		value, err := ms.getCounter(ctx, metric.ID, db)
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
		value, err := ms.getGauge(ctx, metric.ID, db)
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

func (ms *sqlStorage) Save() error {
	// метод - заглушка, проверка подключения к БД, т.к. все данные хранятся в БД
	return ms.PingDB(context.Background())
}

// проверка подключения к БД
func (ms *sqlStorage) PingDB(ctx context.Context) error {
	if ms.ConnectDBString == "" {
		return fmt.Errorf("connect DB string undefined")
	}

	db, err := sql.Open("pgx", ms.ConnectDBString)
	if err != nil {
		return fmt.Errorf("database connect error: %v", err)
	}
	defer db.Close()

	if err = db.PingContext(ctx); err != nil {
		return fmt.Errorf("check database ping error: %v", err)
	}
	return nil
}

// очистка БД
func (ms *sqlStorage) Clear(ctx context.Context) error {
	db, err := sql.Open("pgx", ms.ConnectDBString)
	if err != nil {
		return fmt.Errorf("connect database error: %v", err)
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, "Delete from gauges;")
	if err != nil {
		return fmt.Errorf("clear gauges table error: %v", err)
	}
	_, err = db.ExecContext(ctx, "Delete from counters;")
	if err != nil {
		return fmt.Errorf("clear counters table error: %v", err)
	}
	return nil
}

// обновление через json slice
func (ms *sqlStorage) UpdateJSONSlice(ctx context.Context, data []byte) ([]byte, error) {
	var metrics []metric
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %w", err)
	}

	db, err := sql.Open("pgx", ms.ConnectDBString)
	if err != nil {
		return nil, fmt.Errorf("connect database error: %v", err)
	}
	defer db.Close()

	transaction, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("create database transaction error: %v", err)
	}
	defer transaction.Rollback()

	gup, err := transaction.PrepareContext(ctx, "Update gauges set value=$1 where name=$2")
	if err != nil {
		return nil, fmt.Errorf("prepare query mk error: %v", err)
	}
	defer gup.Close()
	gip, err := transaction.PrepareContext(ctx, "Insert into gauges (value, name) values($1, $2)")
	if err != nil {
		return nil, fmt.Errorf("prepare query mk error: %v", err)
	}
	defer gip.Close()

	cup, err := transaction.PrepareContext(ctx, "Update counters set value=$1 where name=$2")
	if err != nil {
		return nil, fmt.Errorf("prepare query mk error: %v", err)
	}
	defer cup.Close()
	cip, err := transaction.PrepareContext(ctx, "Insert into counters (value, name) values($1, $2)")
	if err != nil {
		return nil, fmt.Errorf("prepare query mk error: %v", err)
	}
	defer cip.Close()

	for _, value := range metrics {
		switch value.MType {
		case "gauge":
			if value.Value != nil {
				val, err := ms.getGauge(ctx, value.ID, db)
				if err == nil {
					_, err = gup.Exec(&value.Value, value.ID)
				} else if val != nil {
					_, err = gip.Exec(&value.Value, value.ID)
				}
				if err != nil {
					return nil, fmt.Errorf("gauge transaction error: %v", err)
				}
			}
		case "counter":
			if value.Delta != nil {
				val, err := ms.getCounter(ctx, value.ID, db)
				if err == nil {
					delta := *value.Delta
					delta += *val
					_, err = cup.Exec(delta, value.ID)
				} else if val != nil {
					_, err = cip.Exec(&value.Delta, value.ID)
				}
				if err != nil {
					return nil, fmt.Errorf("counter transaction error: %v", err)
				}
			}
		}
	}

	err = transaction.Commit()
	if err != nil {
		return nil, fmt.Errorf("transaction commit error: %v", err)
	}
	return []byte(""), nil
}
