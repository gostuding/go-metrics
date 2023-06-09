package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
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
		return fmt.Errorf("get table info error in rowa: %v", err)
	}
	if err != nil {
		return fmt.Errorf("get table info error: %s, ERROR: %v", name, err)
	} else if !rows.Next() {
		err = fmt.Errorf("table not exist: %s ", name)
	}
	return err
}

func checkDatabaseStructure(connectionString string, logger *zap.SugaredLogger) error {
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
		logger.Debugf("Check table: %s", key)
		if isTableExist(ctx, key, db) != nil {
			err := createTable(ctx, key, table, db)
			if err != nil {
				return err
			}
		}
	}
	logger.Debug("database ctructure checked")
	return nil
}
