package server

import (
	"context"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"

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
					var pgErr *pgconn.PgError = &pgconn.PgError{}
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
					var pgErr *pgconn.PgError = &pgconn.PgError{}
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
			got, err := seRepeater(tt.args.f, tt.args.ctx)
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
					var pgErr *pgconn.PgError = &pgconn.PgError{Code: pgerrcode.ConnectionException}
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
					var pgErr *pgconn.PgError = &pgconn.PgError{Code: pgerrcode.ConnectionException}
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
			got, err := bytesErrorRepeater(tt.args.f, tt.args.ctx, tt.args.data)
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
					var pgErr *pgconn.PgError = &pgconn.PgError{Code: pgerrcode.ConnectionException}
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
					var pgErr *pgconn.PgError = &pgconn.PgError{Code: pgerrcode.ConnectionException}
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
			got, err := sseRepeater(tt.args.f, tt.args.ctx, tt.args.t, tt.args.n)
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
					var pgErr *pgconn.PgError = &pgconn.PgError{Code: pgerrcode.ConnectionException}
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
					var pgErr *pgconn.PgError = &pgconn.PgError{Code: pgerrcode.ConnectionException}
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
			errValue := ssseRepeater(tt.args.f, tt.args.ctx, tt.args.t, tt.args.n, tt.args.v)
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
