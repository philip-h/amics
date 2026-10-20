// Package grading computes a student's overall grade from weighted categories.
package grading

import (
	"math"
	"strconv"
	"time"

	"github.com/philip-h/amics/internal/storage"
)

// WeightedAverage returns the overall percentage rounded to a whole number.
// ok is false when no category has a countable assignment.
func WeightedAverage(assignments []*storage.AssignmentWithGrade, categories []*storage.Category, now time.Time) (percent float64, ok bool) {
	weights := make(map[int64]int, len(categories))
	for _, c := range categories {
		weights[int64(c.Id)] = c.Weight
	}

	type totals struct{ earned, possible float64 }
	byCategory := make(map[int64]*totals)

	for _, a := range assignments {
		if !a.Visible || !a.CategoryId.Valid {
			continue
		}
		if _, known := weights[a.CategoryId.Int64]; !known {
			continue
		}
		// Ungraded work only counts (as 0) once it is past due.
		if !a.Grade.Valid && !now.After(a.DueDate) {
			continue
		}
		t := byCategory[a.CategoryId.Int64]
		if t == nil {
			t = &totals{}
			byCategory[a.CategoryId.Int64] = t
		}
		if a.Grade.Valid {
			t.earned += float64(a.Grade.Int64)
		}
		t.possible += float64(a.Points)
	}

	var weighted, weightSum float64
	for id, t := range byCategory {
		if t.possible == 0 {
			continue
		}
		w := float64(weights[id])
		weighted += w * (t.earned / t.possible)
		weightSum += w
	}
	if weightSum == 0 {
		return 0, false
	}
	return math.Round(weighted / weightSum * 100), true
}

// Format renders a WeightedAverage result for display.
func Format(percent float64, ok bool) string {
	if !ok {
		return "N/A"
	}
	return strconv.FormatFloat(percent, 'f', 0, 64) + "%"
}
