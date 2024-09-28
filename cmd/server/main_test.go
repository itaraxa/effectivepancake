package main

import (
	"net/http"
	"reflect"
	"testing"
)

func TestUpdateMemStorageHandler(t *testing.T) {
	type args struct {
		ms *MemStorage
	}
	tests := []struct {
		name string
		args args
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UpdateMemStorageHandler(tt.args.ms); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateMemStorageHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
