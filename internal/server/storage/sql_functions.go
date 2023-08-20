package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

var (
	gaugeTableName   = "gauges"   // table name in database
	counterTableName = "counters" // table name in database
)

type (
	// sqlColumns map with database columns names.
	// Column type determinates by the value type
	sqlColumns map[string]any

	// SQL interface.
	SQLQueryInterface interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
		ExecContext(context.Context, string, ...any) (sql.Result, error)
		QueryRowContext(context.Context, string, ...any) *sql.Row
	}
)

// sqlTablesMaps is private func. Creates sqlColumns for create tables.
func sqlTablesMaps() *map[string]sqlColumns {
	counters, gauges := make(sqlColumns), make(sqlColumns)
	counters["ID"] = 0
	counters["name"] = "50"
	counters["value"] = 0

	gauges["ID"] = 0
	gauges["name"] = "50"
	gauges["value"] = 0.0

	result := make(map[string]sqlColumns)
	result[counterTableName] = counters
	result[gaugeTableName] = gauges
	return &result
}

// createTable is private func. Checks if the table with name exist in database.
func createTable(
	ctx context.Context,
	name string,
	values map[string]any,
	sql *sql.DB,
) error {
	items := make([]string, 0)
	for key, val := range values {
		switch val.(type) {
		case int, int16, int32, int64, uint16, uint32, uint64:
			items = append(items, fmt.Sprintf("%s bigserial", key))
		case string:
			items = append(items, fmt.Sprintf("%s varchar(%s) UNIQUE", key, val))
		case bool:
			items = append(items, fmt.Sprintf("%s boolean", key))
		case time.Time:
			items = append(items, fmt.Sprintf("%s timestamp", key))
		case float32, float64:
			items = append(items, fmt.Sprintf("%s double precision", key))
		}
	}
	context, cansel := context.WithTimeout(ctx, 10*time.Second)
	defer cansel()
	query := fmt.Sprintf("CREATE TABLE  IF NOT EXISTS %s  (%s);", name, strings.Join(items, ","))
	_, err := sql.ExecContext(context, query)
	if err != nil {
		return fmt.Errorf("create new table ('%s') error: %w ", name, err)
	}
	return err
}

// checkDatabaseStructure is private func. Checks structure accoding to sqlTablesMaps.
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
		err := createTable(ctx, key, table, db)
		if err != nil {
			return err
		}
	}
	return nil
}

// getCounter is private func. Returns counter value from database.
func (ms *SQLStorage) getCounter(ctx context.Context, name string) (*int64, error) {
	rows, err := ms.con.QueryContext(ctx, "Select value from counters where name=$1;", name)
	if err != nil {
		return nil, fmt.Errorf("select value error: %w", err)
	}
	defer rows.Close()
	if rows.Err() != nil {
		return nil, fmt.Errorf("get counter metric rows error: %w", err)
	}
	if !rows.Next() {
		value := int64(0)
		return &value, fmt.Errorf("counter value (%s) is absent", name)
	}
	var value int64
	err = rows.Scan(&value)
	if err != nil {
		return nil, fmt.Errorf("scan counter value (%s) error: %w", name, err)
	}
	return &value, nil
}

// getGauge is private func. Returns gauge value from database.
func (ms *SQLStorage) getGauge(ctx context.Context, name string) (*float64, error) {
	rows, err := ms.con.QueryContext(ctx, "Select value from gauges where name=$1;", name)
	if err != nil {
		return nil, fmt.Errorf("select value error: %w", err)
	}
	defer rows.Close()
	if rows.Err() != nil {
		return nil, fmt.Errorf("get gauge metric rows error: %w", err)
	}
	if !rows.Next() {
		value := float64(0.0)
		return &value, fmt.Errorf("gauge value (%s) is absent", name)
	}
	var value float64
	err = rows.Scan(&value)
	if err != nil {
		return nil, fmt.Errorf("scan gauge value (%s) error: %w", name, err)
	}
	return &value, nil
}

// updateCounter is private func. Updates or creates counter value in database.
func (ms *SQLStorage) updateCounter(
	ctx context.Context,
	name string,
	value int64,
	connect SQLQueryInterface,
) (*int64, error) {

	query := `INSERT INTO counters(name, value) values($1, $2) ON CONFLICT (name) DO 
	UPDATE SET value=EXCLUDED.value+counters.value;`
	_, err := connect.ExecContext(ctx, query, name, value)
	if err != nil {
		return &value, fmt.Errorf("counters update error:%s %d: %w", name, value, err)
	}
	return &value, err
}

// updateGauge is private func. Updates or creates gauge value in database.
func (ms *SQLStorage) updateGauge(
	ctx context.Context,
	name string,
	value float64,
	connect SQLQueryInterface,
) (*float64, error) {

	_, err := connect.ExecContext(ctx,
		`INSERT INTO gauges(name, value) values($1, $2) 
		ON CONFLICT (name) DO UPDATE SET value=EXCLUDED.value;`, name, value)
	if err != nil {
		return &value, fmt.Errorf("gauges update error: %w", err)
	}
	return &value, err
}

func scanValue(table string, rows *sql.Rows) (string, error) {
	var err error
	var name string
	strValue := ""
	if table == gaugeTableName {
		var value float64
		err = rows.Scan(&name, &value)
		strValue = fmt.Sprintf("'%s' = %f", name, value)
	} else {
		var value int64
		err = rows.Scan(&name, &value)
		strValue = fmt.Sprintf("'%s' = %d", name, value)
	}
	if err != nil {
		return strValue, fmt.Errorf("scan gauge value error: %w", err)
	}
	return strValue, nil
}

func (ms *SQLStorage) getAllMetricOfType(ctx context.Context, table string) (*[]string, error) {
	values := make([]string, 0)

	query := "Select name, value from counters order by name;"
	if table == gaugeTableName {
		query = "Select name, value from gauges order by name;"
	}
	rows, err := ms.con.QueryContext(ctx, query)
	if err != nil {
		return &values, fmt.Errorf("get all metrics query error: %w", err)
	}
	defer rows.Close()
	if rows.Err() != nil {
		return &values, fmt.Errorf("get all metrics rows error: %w", err)
	}

	for rows.Next() {
		val, err := scanValue(table, rows)
		if err != nil {
			return &values, fmt.Errorf("scan value error: %w", err)
		}
		values = append(values, val)
	}
	return &values, nil
}
