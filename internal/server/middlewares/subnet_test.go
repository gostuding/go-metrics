package middlewares

import (
	"net"
	"net/http"
	"testing"
)

func Test_checkSubnet(t *testing.T) {
	var localMask = "127.0.0.1/32"
	type args struct {
		subnet string
		host   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Correct subnet address",
			args:    args{subnet: localMask, host: "127.0.0.1"},
			wantErr: false,
		},
		{
			name:    "Error subnet address",
			args:    args{subnet: localMask, host: "127.0.0.2"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, subnet, err := net.ParseCIDR(tt.args.subnet)
			if err != nil {
				t.Errorf("parse subnet (%s) error: %v", tt.args.subnet, err)
				return
			}
			r, err := http.NewRequest(http.MethodGet, "", nil)
			if err != nil {
				t.Errorf("create rewuest error: %v", err)
				return
			}
			r.Header.Add(ipHeaderName, tt.args.host)
			if err := checkSubnet(subnet, r); (err != nil) != tt.wantErr {
				t.Errorf("checkSubnet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
