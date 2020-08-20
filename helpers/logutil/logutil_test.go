package logutil

import (
	"context"
	"testing"

	"github.com/go-chi/chi/middleware"
)

func TestGetReqID(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx context.Context
	}
	genargs := func(value string) args {
		ctx := context.Background()
		ctx = context.WithValue(ctx, middleware.RequestIDKey, value)
		return args{ctx: ctx}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1", genargs("steins/gate"), "gate"},
		{"2", genargs("idolm@ster"), "idolm@ster"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := GetReqID(tt.args.ctx); got != tt.want {
				t.Errorf("GetReqID() = %v, want %v", got, tt.want)
			}
		})
	}
}
