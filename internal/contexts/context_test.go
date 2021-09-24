package contexts

import (
	"context"
	"reflect"
	"testing"
)

func TestBuildContext(t *testing.T) {
	type args struct {
		ctx    context.Context
		ctxFns []PartialContextFn
	}
	tests := []struct {
		name string
		args args
		want context.Context
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildContext(tt.args.ctx, tt.args.ctxFns...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
