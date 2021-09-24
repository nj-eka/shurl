package contexts

import (
	"context"
	"reflect"
	"testing"
)

func TestSetContextOperation(t *testing.T) {
	type args struct {
		op Operation
	}
	tests := []struct {
		name string
		args args
		want PartialContextFn
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetContextOperation(tt.args.op); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetContextOperation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddContextOperation(t *testing.T) {
	type args struct {
		op Operation
	}
	tests := []struct {
		name string
		args args
		want PartialContextFn
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddContextOperation(tt.args.op); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddContextOperation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetContextOperations(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want Operations
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetContextOperations(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetContextOperations() = %v, want %v", got, tt.want)
			}
		})
	}
}
