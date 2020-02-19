package main

import (
	"reflect"
	"testing"
)

func TestSearch(t *testing.T) {
	type args struct {
		input   []string
		pattern string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{"first", args{[]string{"1", "2", "3", "4", "5"}, "1"}, []string{"1"}},
		{"last", args{[]string{"1", "2", "3", "4", "5"}, "5"}, []string{"5"}},
		{"middle", args{[]string{"1", "2", "3", "4", "5"}, "3"}, []string{"3"}},
		{"multi", args{[]string{"12", "21", "33", "44", "55"}, "1"}, []string{"12", "21"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Search(test.args.input, test.args.pattern)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("Not equal, want: %v, got: %v", test.want, got)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		input   []string
		pattern string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{"first", args{[]string{"1", "2", "3", "4", "5"}, "1"}, []string{"2", "3", "4", "5"}},
		{"last", args{[]string{"1", "2", "3", "4", "5"}, "5"}, []string{"1", "2", "3", "4"}},
		{"middle", args{[]string{"1", "2", "3", "4", "5"}, "3"}, []string{"1", "2", "4", "5"}},
		{"multi-1", args{[]string{"11", "11", "33", "44", "55"}, "1"}, []string{"33", "44", "55"}},
		{"multi-2", args{[]string{"11", "22", "11", "44", "55"}, "1"}, []string{"22", "44", "55"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Delete(test.args.input, test.args.pattern)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("Not equal, want: %v, got: %v", test.want, got)
			}
		})
	}
}
