package model

import "testing"

func TestCategoriesFormValidate(t *testing.T) {
	tests := []struct {
		name         string
		names        []string
		weights      []string
		wantProblems int
	}{
		{"no rows is valid", nil, nil, 0},
		{"sums to 100", []string{"Tests", "Labs"}, []string{"70", "30"}, 0},
		{"sums to 99", []string{"Tests", "Labs"}, []string{"70", "29"}, 1},
		{"sums to 101", []string{"Tests", "Labs"}, []string{"70", "31"}, 1},
		{"blank name", []string{"", "Labs"}, []string{"50", "50"}, 1},
		{"duplicate name ignoring case", []string{"Tests", "tests"}, []string{"50", "50"}, 1},
		{"non-numeric weight", []string{"Tests"}, []string{"abc"}, 2},
		{"negative weight", []string{"Tests", "Labs"}, []string{"-10", "110"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewCategoriesFormPost("1", nil, tt.names, tt.weights)
			if got := len(f.Validate()); got != tt.wantProblems {
				t.Errorf("got %d problems %v, want %d", got, f.Validate(), tt.wantProblems)
			}
		})
	}
}
