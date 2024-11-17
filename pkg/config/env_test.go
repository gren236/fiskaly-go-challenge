package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type bar struct {
	Baz float64 `env:"BAR_BAZ"`
}

type test struct {
	Hello string `env:"HELLO"`
	World int    `env:"WORLD"`

	Bar bar
}

type testUnsupported struct {
	Hello string `env:"HELLO"`
	World byte   `env:"WORLD"`
}

type testUnexported struct {
	Hello string `env:"HELLO"`
	world int    `env:"WORLD"` //nolint:unused
}

func TestEnv_Set(t *testing.T) {
	type fields struct {
		source func(string) string
	}
	type args struct {
		conf any
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantConf any
		wantErr  bool
	}{
		{
			name: "struct values correctly set",
			fields: fields{
				source: func(s string) string {
					switch s {
					case "HELLO":
						return "Lorem"
					case "WORLD":
						return "42"
					case "BAR_BAZ":
						return "1.5"
					default:
						return ""
					}
				},
			},
			args: args{
				conf: &test{},
			},
			wantConf: &test{
				Hello: "Lorem",
				World: 42,
				Bar: bar{
					Baz: 1.5,
				},
			},
			wantErr: false,
		},
		{
			name: "struct has unsupported field types",
			fields: fields{
				source: func(s string) string {
					switch s {
					case "HELLO":
						return "Lorem"
					case "WORLD":
						return "42"
					default:
						return ""
					}
				},
			},
			args: args{
				conf: &testUnsupported{},
			},
			wantConf: &testUnsupported{
				Hello: "Lorem",
			},
			wantErr: true,
		},
		{
			name: "struct has unexported fields",
			fields: fields{
				source: func(s string) string {
					switch s {
					case "HELLO":
						return "Lorem"
					case "WORLD":
						return "42"
					default:
						return ""
					}
				},
			},
			args: args{
				conf: &testUnexported{},
			},
			wantConf: &testUnexported{
				Hello: "Lorem",
			},
			wantErr: true,
		},
		{
			name: "pointers passed not point to struct",
			fields: fields{
				source: func(s string) string {
					return ""
				},
			},
			args: args{
				conf: func() any {
					i := 32
					j := &i

					return &j
				}(),
			},
			wantConf: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Env{
				source: tt.fields.source,
			}

			err := e.Set(tt.args.conf)

			if tt.wantErr {
				require.Error(t, err)
			}

			if tt.wantConf != nil {
				require.Equal(t, tt.wantConf, tt.args.conf)
			}
		})
	}
}
