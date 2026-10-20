package model

import (
	"strconv"
	"strings"

	"github.com/philip-h/amics/internal/storage"
)

type CategoryRow struct {
	Id     string
	Name   string
	Weight string
}

type CategoriesForm struct {
	CourseId    string
	Rows        []CategoryRow
	ServerError string
}

func NewCategoriesFormFromStorage(courseId string, categories []*storage.Category) *CategoriesForm {
	rows := make([]CategoryRow, len(categories))
	for i, c := range categories {
		rows[i] = CategoryRow{
			Id:     strconv.Itoa(c.Id),
			Name:   c.Name,
			Weight: strconv.Itoa(c.Weight),
		}
	}
	return &CategoriesForm{CourseId: courseId, Rows: rows}
}

// NewCategoriesFormPost binds parallel form values; short slices are padded with blanks.
func NewCategoriesFormPost(courseId string, ids, names, weights []string) *CategoriesForm {
	rows := make([]CategoryRow, len(names))
	for i := range names {
		rows[i].Name = names[i]
		if i < len(ids) {
			rows[i].Id = ids[i]
		}
		if i < len(weights) {
			rows[i].Weight = weights[i]
		}
	}
	return &CategoriesForm{CourseId: courseId, Rows: rows}
}

// Total sums the weights that parse as integers.
func (f *CategoriesForm) Total() int {
	total := 0
	for _, r := range f.Rows {
		if w, err := strconv.Atoi(strings.TrimSpace(r.Weight)); err == nil {
			total += w
		}
	}
	return total
}

// Validate returns human-readable problems; no rows is valid and clears the course's categories.
func (f *CategoriesForm) Validate() []string {
	problems := []string{}
	seen := make(map[string]bool)
	for i, r := range f.Rows {
		n := strconv.Itoa(i + 1)
		name := strings.TrimSpace(r.Name)
		if name == "" {
			problems = append(problems, "Category "+n+": name is required")
		} else if seen[strings.ToLower(name)] {
			problems = append(problems, "Category "+n+": duplicate name \""+name+"\"")
		}
		seen[strings.ToLower(name)] = true

		w, err := strconv.Atoi(strings.TrimSpace(r.Weight))
		if err != nil {
			problems = append(problems, "Category "+n+": weight must be a whole number")
		} else if w < 0 || w > 100 {
			problems = append(problems, "Category "+n+": weight must be between 0 and 100")
		}
	}
	if len(f.Rows) > 0 && f.Total() != 100 {
		problems = append(problems, "Weights must add up to 100 (currently "+strconv.Itoa(f.Total())+")")
	}
	return problems
}

// ToCategories converts a validated form; unparseable ids are returned as an error flag.
func (f *CategoriesForm) ToCategories(courseId int) ([]*storage.Category, bool) {
	categories := make([]*storage.Category, len(f.Rows))
	for i, r := range f.Rows {
		id := 0
		if strings.TrimSpace(r.Id) != "" {
			parsed, err := strconv.Atoi(r.Id)
			if err != nil {
				return nil, false
			}
			id = parsed
		}
		weight, _ := strconv.Atoi(strings.TrimSpace(r.Weight))
		categories[i] = &storage.Category{
			Id:       id,
			CourseId: courseId,
			Name:     strings.TrimSpace(r.Name),
			Weight:   weight,
		}
	}
	return categories, true
}
