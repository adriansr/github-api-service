package config

import (
	"reflect"
	"testing"
)

func TestLoadRaw(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name:    "Error on empty contents",
			args:    args{[]byte{}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Empty JSON",
			args:    args{[]byte("{}")},
			want:    &Config{},
			wantErr: false,
		},
		{
			name: "Credentials",
			args: args{[]byte(`{
					    "github_credentials": {
        					"username": "some_user",
        					"password": "some_pass"
						}
				}`)},
			want:    &Config{Credentials: GitHubCredentials{"some_user", "some_pass"}},
			wantErr: false,
		},
		{
			name: "HTTP client timeout",
			args: args{[]byte(`{
					    "client": {
        					"timeout": "1s500ms"
						}
				}`)},
			// expect a duration of 1.5s, here in nanos:
			want:    &Config{Client: HTTPClientConfig{Duration{1500000000}, ""}},
			wantErr: false,
		},
		{
			name: "Invalid duration 1",
			args: args{[]byte(`{
					    "client": {
        					"timeout": ""
						}
				}`)},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid duration 2",
			args: args{[]byte(`{
					    "client": {
        					"timeout": 15
						}
				}`)},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid duration 3",
			args: args{[]byte(`{
					    "client": {
        					"timeout": "
						}
				}`)},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Full config",
			args: args{[]byte(`{
						"github_credentials": {
        					"username": "user",
        					"password": "password"
						},
					    "client": {
							"timeout": "500ms",
							"api_url": "https://api.github.com"
						},
						"server": {
							"listen": "1.2.3.4:8080"
						}
				}`)},

			want: &Config{GitHubCredentials{"user", "password"},
				HTTPClientConfig{Duration{500000000}, "https://api.github.com"},
				HTTPServerConfig{"1.2.3.4:8080"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadRaw(tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadRaw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadRaw() = %v, want %v", got, tt.want)
			}
		})
	}
}
