package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gostuding/go-metrics/internal/server"
	"github.com/gostuding/go-metrics/internal/server/storage"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type newStorage func(string, *zap.SugaredLogger) (*storage.SQLStorage, error)

func checkError(err error, t *int) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsInvalidAuthorizationSpecification(pgErr.Code) {
		time.Sleep(time.Duration(*t) * time.Second)
		*t += 2
	} else {
		return false
	}
	return true
}

func connectRepeater(f newStorage, con string, logger *zap.SugaredLogger) (*storage.SQLStorage, error) {
	strg, err := f(con, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err != nil {
		logger.Debug("database connection error")
		waitTime := 1
		for i := 0; i < 3; i++ {
			select {
			case <-ctx.Done():
				return nil, errors.New("context done error")
			default:
				if !checkError(err, &waitTime) {
					return nil, err
				} else {
					logger.Debug("database connection error")
				}
				strg, err := f(con, logger)
				if err == nil {
					return strg, nil
				}
			}
		}
	}
	return strg, err
}

func main() {
	cfg, err := server.GetFlags()
	if err != nil {
		log.Fatalln(err)
	}
	logger, err := server.InitLogger()
	if err != nil {
		log.Fatalln(err)
	}

	if cfg.ConnectDBString != "" {
		strg, err := connectRepeater(storage.NewSQLStorage, cfg.ConnectDBString, logger)
		if err != nil {
			log.Fatalln(err)
		}
		err = server.RunServer(cfg, strg, logger)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		storage, err := storage.NewMemStorage(cfg.Restore, cfg.FileStorePath, cfg.StoreInterval, cfg.ConnectDBString)
		if err != nil {
			log.Fatalln(err)
		}
		err = server.RunServer(cfg, storage, logger)
		if err != nil {
			log.Fatalln(err)
		}
	}

}
