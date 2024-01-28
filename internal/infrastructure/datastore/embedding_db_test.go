package datastore_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/yonisaka/similarity/internal/di"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	var code int

	defer func() {
		os.Exit(code)
	}()

	_ = os.Setenv("APP_ENV", "test")
	_ = os.Setenv("IS_REPLICA", "false")

	code = m.Run()
}

func TestEmbeddingRepo_ListEmbeddingByScope(t *testing.T) {
	type args struct {
		ctx   context.Context
		scope string
	}

	type test struct {
		args       args
		want       int
		wantErr    error
		beforeFunc func(*testing.T)
		afterFunc  func(*testing.T)
	}

	tests := map[string]func(t *testing.T) test{
		"Given valid query of Get List Embedding, When query executed successfully, Return no error": func(t *testing.T) test {

			args := args{
				ctx:   context.Background(),
				scope: "lelang",
			}

			return test{
				args:    args,
				want:    5,
				wantErr: nil,
			}
		},
	}

	for name, fn := range tests {
		t.Run(name, func(t *testing.T) {
			tt := fn(t)

			if tt.beforeFunc != nil {
				tt.beforeFunc(t)
			}

			if tt.afterFunc != nil {
				defer tt.afterFunc(t)
			}

			sut := di.GetEmbeddingRepo()

			got, err := sut.ListEmbeddingByScope(tt.args.ctx, tt.args.scope)

			if !assert.ErrorIs(t, err, tt.wantErr) {
				return
			}

			assert.Equal(t, tt.want, len(got))
		})
	}
}
