package errs

import (
	"reflect"
	"testing"
)

func TestE(t *testing.T) {
	type args struct {
		args []interface{}
	}
	tests := []struct {
		name string
		args args
		want Error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := E(tt.args.args...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("E() = %v, want %v", got, tt.want)
			}
		})
	}
}
