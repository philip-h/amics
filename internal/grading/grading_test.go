package grading

import (
	"database/sql"
	"testing"
	"time"

	"github.com/philip-h/amics/internal/storage"
)

var now = time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)

func asg(category int64, points int, grade int64, graded bool, due time.Time, visible bool) *storage.AssignmentWithGrade {
	a := &storage.AssignmentWithGrade{}
	a.Points = points
	a.DueDate = due
	a.Visible = visible
	if category != 0 {
		a.CategoryId = sql.NullInt64{Int64: category, Valid: true}
	}
	a.Grade = sql.NullInt64{Int64: grade, Valid: graded}
	return a
}

func TestWeightedAverage(t *testing.T) {
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)
	cats := []*storage.Category{
		{Id: 1, Weight: 60},
		{Id: 2, Weight: 40},
	}

	tests := []struct {
		name        string
		assignments []*storage.AssignmentWithGrade
		categories  []*storage.Category
		wantPercent float64
		wantOk      bool
	}{
		{
			name:        "no categories is N/A",
			assignments: []*storage.AssignmentWithGrade{asg(0, 10, 10, true, past, true)},
			categories:  nil,
			wantOk:      false,
		},
		{
			name:        "only uncategorized assignments is N/A",
			assignments: []*storage.AssignmentWithGrade{asg(0, 10, 10, true, past, true)},
			categories:  cats,
			wantOk:      false,
		},
		{
			name: "weights applied across categories",
			assignments: []*storage.AssignmentWithGrade{
				asg(1, 100, 80, true, past, true), // 80% at weight 60
				asg(2, 100, 50, true, past, true), // 50% at weight 40
			},
			categories:  cats,
			wantPercent: 68, // 0.6*80 + 0.4*50
			wantOk:      true,
		},
		{
			name: "points-weighted inside a category",
			assignments: []*storage.AssignmentWithGrade{
				asg(1, 10, 10, true, past, true),
				asg(1, 90, 45, true, past, true),
			},
			categories:  cats,
			wantPercent: 55, // 55/100
			wantOk:      true,
		},
		{
			name: "empty category is renormalized away",
			assignments: []*storage.AssignmentWithGrade{
				asg(2, 100, 90, true, past, true),
			},
			categories:  cats,
			wantPercent: 90,
			wantOk:      true,
		},
		{
			name: "ungraded before due date is ignored",
			assignments: []*storage.AssignmentWithGrade{
				asg(1, 100, 80, true, past, true),
				asg(1, 100, 0, false, future, true),
			},
			categories:  cats,
			wantPercent: 80,
			wantOk:      true,
		},
		{
			name: "ungraded past due date counts as zero",
			assignments: []*storage.AssignmentWithGrade{
				asg(1, 100, 80, true, past, true),
				asg(1, 100, 0, false, past, true),
			},
			categories:  cats,
			wantPercent: 40,
			wantOk:      true,
		},
		{
			name: "graded before due date still counts",
			assignments: []*storage.AssignmentWithGrade{
				asg(1, 100, 70, true, future, true),
			},
			categories:  cats,
			wantPercent: 70,
			wantOk:      true,
		},
		{
			name: "invisible assignments are excluded",
			assignments: []*storage.AssignmentWithGrade{
				asg(1, 100, 10, true, past, false),
			},
			categories: cats,
			wantOk:     false,
		},
		{
			name: "assignment in unknown category is excluded",
			assignments: []*storage.AssignmentWithGrade{
				asg(99, 100, 100, true, past, true),
			},
			categories: cats,
			wantOk:     false,
		},
		{
			name: "rounds to nearest whole percent",
			assignments: []*storage.AssignmentWithGrade{
				asg(1, 3, 2, true, past, true), // 66.67%
			},
			categories:  cats,
			wantPercent: 67,
			wantOk:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := WeightedAverage(tt.assignments, tt.categories, now)
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && got != tt.wantPercent {
				t.Errorf("percent = %v, want %v", got, tt.wantPercent)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	if got := Format(0, false); got != "N/A" {
		t.Errorf("Format not ok = %q, want N/A", got)
	}
	if got := Format(87, true); got != "87%" {
		t.Errorf("Format = %q, want 87%%", got)
	}
}
