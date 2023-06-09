package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type sqlStorage struct {
	ConnectDBString string             `json:"-"`
	Logger          *zap.SugaredLogger `json:"-"`
}

func (ms *sqlStorage) getCounter(ctx context.Context, name string, connect *sql.DB) (*int64, error) {
	var db *sql.DB
	var err error

	if connect == nil {
		db, err = sql.Open("pgx", ms.ConnectDBString)
		if err != nil {
			return nil, fmt.Errorf("connect database error: %v", err)
		}
		defer db.Close()
	} else {
		db = connect
	}

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, "Select value from counters where name=$1;", name)
	if err != nil {
		return nil, fmt.Errorf("select value error: %v", err)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("get counter metric rows error: %v", err)
	}

	if !rows.Next() {
		value := int64(0)
		return &value, fmt.Errorf("counter value (%s) is absent", name)
	}
	var value int64
	err = rows.Scan(&value)
	if err != nil {
		return nil, fmt.Errorf("scan counter value (%s) error: %v", name, err)
	}
	return &value, nil
}

func (ms *sqlStorage) getGauge(ctx context.Context, name string, connect *sql.DB) (*float64, error) {
	var db *sql.DB
	var err error

	if connect == nil {
		db, err = sql.Open("pgx", ms.ConnectDBString)
		if err != nil {
			return nil, fmt.Errorf("connect database error: %v", err)
		}
		defer db.Close()
	} else {
		db = connect
	}

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, "Select value from gauges where name=$1;", name)
	if err != nil {
		return nil, fmt.Errorf("select value error: %v", err)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("get gauge metric rows error: %v", err)
	}

	if !rows.Next() {
		value := float64(0.0)
		return &value, fmt.Errorf("gauge value (%s) is absent", name)
	}
	var value float64
	err = rows.Scan(&value)
	if err != nil {
		return nil, fmt.Errorf("scan gauge value (%s) error: %v", name, err)
	}
	return &value, nil
}

func (ms *sqlStorage) updateCounter(ctx context.Context, name string, value int64) (*int64, error) {
	db, err := sql.Open("pgx", ms.ConnectDBString)
	if err != nil {
		return nil, fmt.Errorf("connect database error: %v", err)
	}
	defer db.Close()

	val, err := ms.getCounter(ctx, name, db)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err == nil {
		value += *val
		ms.Logger.Debugf("update counter '%s' = '%d'", name, value)
		_, err = db.ExecContext(ctx, "Update counters set value=$2 where name=$1;", name, value)
		return &value, err
	} else if val != nil {
		row := db.QueryRowContext(ctx, "Select max(id) from counters;")
		maxID := 1
		if err := row.Scan(&maxID); err == nil {
			maxID += 1
		}
		ms.Logger.Debugf("new counter '%s' = '%d'", name, value)
		_, err = db.ExecContext(ctx, "Insert into counters (id, name, value) values($3, $1, $2);", name, value, maxID)
		return &value, err
	}
	return nil, err
}

func (ms *sqlStorage) updateGauge(ctx context.Context, name string, value float64) (*float64, error) {
	db, err := sql.Open("pgx", ms.ConnectDBString)
	if err != nil {
		return nil, fmt.Errorf("connect database error: %v", err)
	}
	defer db.Close()

	val, err := ms.getGauge(ctx, name, db)
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err == nil {
		ms.Logger.Debugf("update gauge '%s' = '%v'", name, value)
		_, err = db.ExecContext(ctx, "Update gauges set value=$2 where name=$1;", name, value)
		return &value, err
	} else if val != nil {
		row := db.QueryRowContext(ctx, "Select max(id) from gauges;")
		maxID := 1
		if err := row.Scan(&maxID); err == nil {
			maxID += 1
		}
		ms.Logger.Debugf("new gauge '%s' = '%d'", name, value)
		_, err = db.ExecContext(ctx, "Insert into gauges (id, name, value) values($3, $1, $2);", name, value, maxID)
		return &value, err
	}
	return nil, err
}

func (ms *sqlStorage) getAllMetricOfType(ctx context.Context, table string) (*[]string, error) {
	values := make([]string, 0)
	db, err := sql.Open("pgx", ms.ConnectDBString)
	if err != nil {
		return &values, fmt.Errorf("connect database error: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, fmt.Sprintf("Select name, value from %s order by name;", table))
	if err != nil {
		return &values, fmt.Errorf("get all metrics query error: %v", err)
	}
	if rows.Err() != nil {
		return &values, fmt.Errorf("get all metrics rows error: %v", err)
	}

	if table == "gauges" {
		for rows.Next() {
			var name string
			var value float64
			err = rows.Scan(&name, &value)
			if err != nil {
				return &values, err
			}
			values = append(values, fmt.Sprintf("'%s' = %v", name, value))
		}
		return &values, nil
	}

	for rows.Next() {
		var name string
		var value int64
		err = rows.Scan(&name, &value)
		if err != nil {
			return &values, err
		}
		values = append(values, fmt.Sprintf("'%s' = %d", name, value))
	}
	return &values, nil
}

func NewSQLStorage(DBconnect string, logger *zap.SugaredLogger) (*sqlStorage, error) {
	storage := sqlStorage{
		ConnectDBString: DBconnect,
		Logger:          logger,
	}
	return &storage, checkDatabaseStructure(DBconnect, logger)
}

func (ms *sqlStorage) Update(ctx context.Context, mType string, mName string, mValue string) error {
	switch mType {
	case "counter":
		counter, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return fmt.Errorf("counter value convert error: %v", err)
		}
		_, err = ms.updateCounter(ctx, mName, counter)
		return err
	case "gauge":
		gauges, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return fmt.Errorf("gauge value convert error: %v", err)
		}
		_, err = ms.updateGauge(ctx, mName, gauges)
		return err

	default:
		return errors.New("metric type incorrect. Availible types are: guage or counter")
	}
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
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return fmt.Errorf("check database ping error: %v", err)
	}
	return nil
}

// Получение значения метрики по типу и имени
func (ms *sqlStorage) GetMetric(ctx context.Context, mType string, mName string) (string, error) {
	switch mType {
	case "gauge":
		value, err := ms.getGauge(ctx, mName, nil)
		return fmt.Sprintf("%v", *value), err
	case "counter":
		value, err := ms.getCounter(ctx, mName, nil)
		return fmt.Sprintf("%d", *value), err
	default:
		return "", fmt.Errorf("metric '%s' with type '%s' not found", mName, mType)
	}
}

// Список всех метрик в html
func (ms *sqlStorage) GetMetricsHTML(ctx context.Context) string {
	gauges, err := ms.getAllMetricOfType(ctx, "gauges")
	if err != nil {
		ms.Logger.Warnf("get gauges metrics error: %v", err)
	}
	counters, err := ms.getAllMetricOfType(ctx, "counters")
	if err != nil {
		ms.Logger.Warnf("get counters metrics error: %v", err)
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
	return body
}

// обновление через json
func (ms *sqlStorage) UpdateJSON(ctx context.Context, data []byte) ([]byte, error) {
	var metric metric
	err := json.Unmarshal(data, &metric)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %v", err)
	}

	switch metric.MType {
	case "counter":
		if metric.Delta != nil {
			value, err := ms.updateCounter(ctx, metric.ID, *metric.Delta)
			if err != nil {
				return nil, err
			}
			metric.Delta = value
		} else {
			return nil, errors.New("metric's delta indefined")
		}
	case "gauge":
		if metric.Value != nil {
			value, err := ms.updateGauge(ctx, metric.ID, *metric.Value)
			if err != nil {
				return nil, err
			}
			metric.Value = value
		} else {
			return nil, errors.New("metric's value indefined")
		}
	default:
		return nil, errors.New("metric type error, use counter like int64 or gauge like float64")
	}

	resp, err := json.Marshal(metric)
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

	switch metric.MType {
	case "counter":
		value, err := ms.getCounter(ctx, metric.ID, nil)
		if err != nil {
			return nil, err
		}
		metric.Delta = value
		resp, err := json.Marshal(metric)
		if err != nil {
			return []byte(""), err
		}
		return resp, nil
	case "gauge":
		value, err := ms.getGauge(ctx, metric.ID, nil)
		if err != nil {
			return nil, err
		}
		metric.Value = value
		resp, err := json.Marshal(metric)
		if err != nil {
			return []byte(""), err
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
