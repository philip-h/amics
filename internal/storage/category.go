package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

var ErrWeightsNotHundred = errors.New("category weights must total 100")

type Category struct {
	Id       int
	CourseId int
	Name     string
	Weight   int
}

type CategoryStore struct {
	db *sql.DB
}

func (s *CategoryStore) ListByCourseId(ctx context.Context, courseId int) ([]*Category, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, course_id, name, weight
		FROM assignment_category
		WHERE course_id = $1
		ORDER BY id`, courseId)
	if err != nil {
		return nil, fmt.Errorf("db(category.listByCourseId): %w", err)
	}
	defer rows.Close()

	categories := []*Category{}
	for rows.Next() {
		c := &Category{}
		if err := rows.Scan(&c.Id, &c.CourseId, &c.Name, &c.Weight); err != nil {
			return nil, fmt.Errorf("db(category.listByCourseId): %w", err)
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db(category.listByCourseId): %w", err)
	}
	return categories, nil
}

// ReplaceAll upserts the given categories (Id 0 = new) and deletes the course's
// categories that are missing from the list. An empty list clears all categories.
func (s *CategoryStore) ReplaceAll(ctx context.Context, courseId int, categories []*Category) error {
	if len(categories) > 0 {
		total := 0
		for _, c := range categories {
			total += c.Weight
		}
		if total != 100 {
			return ErrWeightsNotHundred
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db(category.replaceAll): %w", err)
	}
	defer tx.Rollback()

	keptIds := []int64{}
	for _, c := range categories {
		if c.Id == 0 {
			err = tx.QueryRowContext(ctx,
				`INSERT INTO assignment_category (course_id, name, weight)
				VALUES ($1, $2, $3) RETURNING id`,
				courseId, c.Name, c.Weight).Scan(&c.Id)
		} else {
			var res sql.Result
			res, err = tx.ExecContext(ctx,
				`UPDATE assignment_category SET name = $1, weight = $2
				WHERE id = $3 AND course_id = $4`,
				c.Name, c.Weight, c.Id, courseId)
			if err == nil {
				var n int64
				if n, err = res.RowsAffected(); err == nil && n == 0 {
					err = fmt.Errorf("category %d not in course %d", c.Id, courseId)
				}
			}
		}
		if err != nil {
			return fmt.Errorf("db(category.replaceAll): %w", err)
		}
		keptIds = append(keptIds, int64(c.Id))
	}

	// Deleting a category sets category_id to NULL on its assignments.
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM assignment_category
		WHERE course_id = $1 AND NOT (id = ANY($2))`,
		courseId, pq.Array(keptIds)); err != nil {
		return fmt.Errorf("db(category.replaceAll): %w", err)
	}

	return tx.Commit()
}
