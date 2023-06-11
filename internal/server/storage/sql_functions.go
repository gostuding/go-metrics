package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type sqlColumns map[string]any

func sqlTablesMaps() *map[string]sqlColumns {
	counters, gauges := make(sqlColumns), make(sqlColumns)
	counters["ID"] = 0
	counters["name"] = "50"
	counters["value"] = 0

	gauges["id"] = 0
	gauges["name"] = "50"
	gauges["value"] = 0.0

	result := make(map[string]sqlColumns)
	result["counters"] = counters
	result["gauges"] = gauges
	return &result
}

func createTable(ctx context.Context, name string, values map[string]any, sql *sql.DB) error {
	items := make([]string, 0)
	for key, val := range values {
		switch val.(type) {
		case int, int16, int32, int64, uint16, uint32, uint64:
			items = append(items, fmt.Sprintf("%s bigserial", key))
		case string:
			items = append(items, fmt.Sprintf("%s varchar(%s)", key, val))
		case bool:
			items = append(items, fmt.Sprintf("%s boolean", key))
		case time.Time:
			items = append(items, fmt.Sprintf("%s timestamp", key))
		case float32, float64:
			items = append(items, fmt.Sprintf("%s double precision", key))
		}
	}
	context, cansel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cansel()
	query := fmt.Sprintf("Create table %s (%s);", name, strings.Join(items, ","))
	_, err := sql.ExecContext(context, query)
	if err != nil {
		return fmt.Errorf("create new table ('%s') error: %w ", name, err)
	}
	return err
}

func isTableExist(ctx context.Context, name string, sql *sql.DB) error {
	context, cansel := context.WithTimeout(ctx, 1*time.Second)
	defer cansel()
	query := "Select * from INFORMATION_SCHEMA.TABLES where TABLE_NAME = $1;"
	rows, err := sql.QueryContext(context, query, name)
	if rows.Err() != nil {
		return fmt.Errorf("get table info error in rowa: %w", err)
	}
	if err != nil {
		return fmt.Errorf("get table info error: %s, ERROR: %w", name, err)
	} else if !rows.Next() {
		err = fmt.Errorf("table not exist: %s ", name)
	}
	return err
}

func checkDatabaseStructure(connectionString string) error {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return fmt.Errorf("connect database error: %w", err)
	}
	defer db.Close()
	// проверка структуры БД не должна превышать 3 секунды
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return fmt.Errorf("check database ping error: %w", err)
	}
	for key, table := range *sqlTablesMaps() {
		if isTableExist(ctx, key, db) != nil {
			err := createTable(ctx, key, table, db)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//----------------------------------------------------------------------------------------------------
// функции для работы с хранилищем
//----------------------------------------------------------------------------------------------------

type SQLQueryInterface interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (ms *SQLStorage) getCounter(ctx context.Context, name string) (*int64, error) {
	rows, err := ms.con.QueryContext(ctx, "Select value from counters where name=$1;", name)
	if err != nil {
		return nil, fmt.Errorf("select value error: %w", err)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("get counter metric rows error: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		value := int64(0)
		return &value, fmt.Errorf("counter value (%s) is absent", name)
	}
	defer rows.Close()
	var value int64
	err = rows.Scan(&value)
	if err != nil {
		return nil, fmt.Errorf("scan counter value (%s) error: %w", name, err)
	}
	return &value, nil
}

func (ms *SQLStorage) getGauge(ctx context.Context, name string) (*float64, error) {
	rows, err := ms.con.QueryContext(ctx, "Select value from gauges where name=$1;", name)
	if err != nil {
		return nil, fmt.Errorf("select value error: %w", err)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("get gauge metric rows error: %w", err)
	}

	if !rows.Next() {
		value := float64(0.0)
		return &value, fmt.Errorf("gauge value (%s) is absent", name)
	}
	defer rows.Close()

	var value float64
	err = rows.Scan(&value)
	if err != nil {
		return nil, fmt.Errorf("scan gauge value (%s) error: %w", name, err)
	}
	return &value, nil
}

func (ms *SQLStorage) updateCounter(ctx context.Context, name string, value int64, connect SQLQueryInterface) (*int64, error) {
	val, err := ms.getCounter(ctx, name)

	if err == nil {
		value += *val
		_, err = connect.ExecContext(ctx, "Update counters set value=$2 where name=$1;", name, value)
		return &value, err
	} else if val != nil {
		_, err = connect.ExecContext(ctx, "Insert into counters (name, value) values($1, $2);", name, value)
		return &value, err
	}
	return nil, err
}

func (ms *SQLStorage) updateGauge(ctx context.Context, name string, value float64, connect SQLQueryInterface) (*float64, error) {
	val, err := ms.getGauge(ctx, name)

	if err == nil {
		_, err = connect.ExecContext(ctx, "Update gauges set value=$2 where name=$1;", name, value)
		return &value, err
	} else if val != nil {
		_, err = connect.ExecContext(ctx, "Insert into gauges (name, value) values($1, $2);", name, value)
		return &value, err
	}
	return nil, err
}

func (ms *SQLStorage) getAllMetricOfType(ctx context.Context, table string) (*[]string, error) {
	values := make([]string, 0)
	rows, err := ms.con.QueryContext(ctx, fmt.Sprintf("Select name, value from %s order by name;", table))
	if err != nil {
		return &values, fmt.Errorf("get all metrics query error: %w", err)
	}
	if rows.Err() != nil {
		return &values, fmt.Errorf("get all metrics rows error: %w", err)
	}
	defer rows.Close()
	if table == "gauges" {
		for rows.Next() {
			var name string
			var value float64
			err = rows.Scan(&name, &value)
			if err != nil {
				return &values, err
			}
			values = append(values, fmt.Sprintf("'%s' = %f", name, value))
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
