package server

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gostuding/go-metrics/internal/server/mocks"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func Test_seRepeater(t *testing.T) {
	type args struct {
		f   fse
		ctx context.Context
	}

	ctxTimeout, cansel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cansel()

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Без ошибок",
			args: args{
				f: func(context.Context) (string, error) {
					return "success", nil
				},
				ctx: context.Background(),
			},
			want:    "success",
			wantErr: false,
		},
		{
			name: "Ошибка подключения",
			args: args{
				f: func(context.Context) (string, error) {
					pgErr := &pgconn.PgError{}
					pgErr.Code = pgerrcode.ConnectionException
					return "pgErr", pgErr
				},
				ctx: context.Background(),
			},
			want:    "pgErr",
			wantErr: true,
		},
		{
			name: "Контекст отмены операции",
			args: args{
				f: func(context.Context) (string, error) {
					pgErr := &pgconn.PgError{}
					pgErr.Code = pgerrcode.ConnectionException
					return "pgErr", pgErr
				},
				ctx: ctxTimeout,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Ошибка без повторов",
			args: args{
				f: func(ctx context.Context) (string, error) {
					return "error", errors.New("not repeate error")
				},
				ctx: context.Background(),
			},
			want:    "error",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := seRepeater(tt.args.ctx, tt.args.f)
			if (err != nil) != tt.wantErr {
				t.Errorf("seRepeater() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("seRepeater() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bytesErrorRepeater(t *testing.T) {
	type args struct {
		f    fbe
		ctx  context.Context
		data []byte
	}

	ctxTimeout, cansel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cansel()

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Без ошибок",
			args: args{
				f: func(ctx context.Context, data []byte) ([]byte, error) {
					return data, nil
				},
				ctx:  context.Background(),
				data: []byte("success"),
			},
			want:    []byte("success"),
			wantErr: false,
		},
		{
			name: "Ошибка подключения",
			args: args{
				f: func(ctx context.Context, data []byte) ([]byte, error) {
					pgErr := &pgconn.PgError{Code: pgerrcode.ConnectionException}
					return data, pgErr
				},
				ctx:  context.Background(),
				data: []byte("pgErr"),
			},
			want:    []byte("pgErr"),
			wantErr: true,
		},
		{
			name: "Контекст отмены операции",
			args: args{
				f: func(ctx context.Context, data []byte) ([]byte, error) {
					pgErr := &pgconn.PgError{Code: pgerrcode.ConnectionException}
					return data, pgErr
				},
				data: []byte("pgErr"),
				ctx:  ctxTimeout,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Ошибка без повторов",
			args: args{
				f: func(ctx context.Context, data []byte) ([]byte, error) {
					return data, errors.New("not repeate error")
				},
				ctx:  context.Background(),
				data: []byte("error"),
			},
			want:    []byte("error"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := bytesErrorRepeater(tt.args.ctx, tt.args.f, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("bytesErrorRepeater() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bytesErrorRepeater() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sseRepeater(t *testing.T) {
	type args struct {
		f   fsse
		ctx context.Context
		t   string
		n   string
	}
	ctxTimeout, cansel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cansel()

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Без ошибок",
			args: args{
				f: func(ctx context.Context, t, n string) (string, error) {
					return t, nil
				},
				ctx: context.Background(),
				t:   "success",
				n:   "",
			},
			want:    "success",
			wantErr: false,
		},
		{
			name: "Ошибка подключения",
			args: args{
				f: func(ctx context.Context, t, n string) (string, error) {
					pgErr := &pgconn.PgError{Code: pgerrcode.ConnectionException}
					return t, pgErr
				},
				ctx: context.Background(),
				t:   "pgError",
				n:   "",
			},
			want:    "pgError",
			wantErr: true,
		},
		{
			name: "Контекст отмены операции",
			args: args{
				f: func(ctx context.Context, t, n string) (string, error) {
					pgErr := &pgconn.PgError{Code: pgerrcode.ConnectionException}
					return t, pgErr
				},
				ctx: ctxTimeout,
				t:   "pgError",
				n:   "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Ошибка без повторов",
			args: args{
				f: func(ctx context.Context, t, n string) (string, error) {
					return t, errors.New("not repeate error")
				},
				ctx: context.Background(),
				t:   "error",
				n:   "",
			},
			want:    "error",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := sseRepeater(tt.args.ctx, tt.args.f, tt.args.t, tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("sseRepeater() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("sseRepeater() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ssseRepeater(t *testing.T) {
	type args struct {
		f   fssse
		ctx context.Context
		t   string
		n   string
		v   string
	}
	ctxTimeout, cansel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cansel()

	tests := []struct {
		name    string
		args    args
		wantErr bool
		err     error
	}{
		{
			name: "Без ошибок",
			args: args{
				f: func(ctx context.Context, t, n, v string) error {
					return nil
				},
				ctx: context.Background(),
				t:   "",
				n:   "",
				v:   "",
			},
			err:     nil,
			wantErr: false,
		},
		{
			name: "Ошибка подключения",
			args: args{
				f: func(ctx context.Context, t, n, v string) error {
					pgErr := &pgconn.PgError{Code: pgerrcode.ConnectionException}
					return pgErr
				},
				ctx: context.Background(),
				t:   "",
				n:   "",
				v:   "",
			},
			err:     &pgconn.PgError{Code: pgerrcode.ConnectionException},
			wantErr: true,
		},
		{
			name: "Контекст отмены операции",
			args: args{
				f: func(ctx context.Context, t, n, v string) error {
					pgErr := &pgconn.PgError{Code: pgerrcode.ConnectionException}
					return pgErr
				},
				ctx: ctxTimeout,
				t:   "",
				n:   "",
				v:   "",
			},
			err:     context.DeadlineExceeded,
			wantErr: true,
		},
		{
			name: "Ошибка без повторов",
			args: args{
				f: func(ctx context.Context, t, n, v string) error {
					return io.EOF
				},
				ctx: context.Background(),
				t:   "",
				n:   "",
				v:   "",
			},
			err:     io.EOF,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errValue := ssseRepeater(tt.args.ctx, tt.args.f, tt.args.t, tt.args.n, tt.args.v)
			if tt.wantErr && errValue == nil {
				t.Errorf("ssseRepeater() want error, but get nil")
			}
			if !tt.wantErr && errValue != nil {
				t.Errorf("ssseRepeater() don't want error, but get: '%v'", errValue)
			}
			if tt.wantErr && !errors.Is(errValue, tt.err) && tt.err.Error() != errValue.Error() {
				t.Errorf("ssseRepeater() want: '%v', get indefuned error: '%v'", tt.err, errValue)
			}
		})
	}
}

func Test_isRepeat(t *testing.T) {
	val := 1
	pgError := pgconn.PgError{Code: pgerrcode.ConnectionException}
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ошибка PSQL",
			err:  &pgError,
			want: true,
		},
		{
			name: "другая ошибка",
			err:  errors.New("OTHER"),
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := isRepeat(tt.err, &val); got != tt.want {
				t.Errorf("isRepeat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := mocks.NewMockStorage(ctrl)
	ctx := context.Background()
	errType := errors.New("type error")
	storage.EXPECT().Update(ctx, "gauge", "name", "1").Return(nil)
	storage.EXPECT().Update(ctx, "gauger", "name", "1,0").Return(errType)

	type args struct {
		ctx     context.Context
		storage StorageSetter
		metric  updateMetricsArgs
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "Update success",
			args: args{
				ctx:     ctx,
				storage: storage,
				metric:  updateMetricsArgs{base: getMetricsArgs{mType: "gauge", mName: "name"}, mValue: "1"},
			},
			want:    http.StatusOK,
			wantErr: false,
		},
		{
			name: "Update error",
			args: args{
				ctx:     ctx,
				storage: storage,
				metric:  updateMetricsArgs{base: getMetricsArgs{mType: "gauger", mName: "name"}, mValue: "1,0"},
			},
			want:    http.StatusBadRequest,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := Update(tt.args.ctx, tt.args.storage, tt.args.metric)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Update() = %v, want %v", got, tt.want)
			}
		})
	}

	storage.EXPECT().GetMetric(ctx, "gauge", "name").Return("1", nil)
	storage.EXPECT().GetMetric(ctx, "gauger", "name").Return("", errType)
	type argsGetM struct {
		ctx     context.Context
		storage StorageGetter
		metric  getMetricsArgs
	}
	testsGetM := []struct {
		name    string
		args    argsGetM
		want    []byte
		wantErr bool
	}{
		{
			name: "GetMetric success",
			args: argsGetM{
				ctx:     ctx,
				storage: storage,
				metric:  getMetricsArgs{mType: "gauge", mName: "name"},
			},
			want:    []byte("1"),
			wantErr: false,
		},
		{
			name: "GetMetric error",
			args: argsGetM{
				ctx:     ctx,
				storage: storage,
				metric:  getMetricsArgs{mType: "gauger", mName: "name"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range testsGetM {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMetric(tt.args.ctx, tt.args.storage, tt.args.metric)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMetric() = %v, want %v", got, tt.want)
			}
		})
	}

	storage.EXPECT().GetMetricsHTML(ctx).Return("all metrics", nil)
	t.Run("GetMetricsHTML", func(t *testing.T) {
		_, err := GetAllMetrics(ctx, storage)
		if err != nil {
			t.Errorf("GetAllMetrics() error = %v", err)
			return
		}
	})

	storage.EXPECT().PingDB(ctx).Return(nil)
	t.Run("Ping", func(t *testing.T) {
		_, err := Ping(ctx, storage)
		if err != nil {
			t.Errorf("PingDB() error = %v", err)
			return
		}
	})

	storage.EXPECT().Clear(ctx).Return(nil)
	t.Run("Clear", func(t *testing.T) {
		_, err := Clear(ctx, storage)
		if err != nil {
			t.Errorf("Clear() error = %v", err)
			return
		}
	})

	storage.EXPECT().UpdateJSON(ctx, []byte("success")).Return([]byte("success"), nil)
	storage.EXPECT().UpdateJSON(ctx, []byte("error")).Return(nil, errType)
	type argsUJ struct {
		ctx     context.Context
		body    []byte
		storage StorageSetter
	}
	testsUJ := []struct {
		name    string
		args    argsUJ
		want    []byte
		wantErr bool
	}{
		{
			name:    "Update JSON success",
			args:    argsUJ{ctx: ctx, body: []byte("success"), storage: storage},
			want:    []byte("success"),
			wantErr: false,
		},
		{
			name:    "Update JSON error",
			args:    argsUJ{ctx: ctx, body: []byte("error"), storage: storage},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range testsUJ {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpdateJSON(tt.args.ctx, tt.args.body, tt.args.storage)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateJSON() = %v, want %v", got, tt.want)
			}
		})
	}

	storage.EXPECT().UpdateJSONSlice(ctx, []byte("slice")).Return([]byte("ok"), nil)
	storage.EXPECT().UpdateJSONSlice(ctx, []byte("error")).Return(nil, errType)
	type argsUJS struct {
		ctx     context.Context
		data    []byte
		storage StorageSetter
	}
	testsUJS := []struct {
		name    string
		args    argsUJS
		want    []byte
		wantErr bool
	}{
		{
			name: "Update by slice success",
			args: argsUJS{
				ctx:     ctx,
				data:    []byte("slice"),
				storage: storage,
			},
			want:    []byte("ok"),
			wantErr: false,
		},
		{
			name: "Update by slice error",
			args: argsUJS{
				ctx:     ctx,
				data:    []byte("error"),
				storage: storage,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range testsUJS {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpdateJSONSLice(tt.args.ctx, tt.args.data, tt.args.storage)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateJSONSLice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateJSONSLice() = %v, want %v", got, tt.want)
			}
		})
	}
}
