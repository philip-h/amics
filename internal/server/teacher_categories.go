package server

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/philip-h/amics/internal/httpe"
	"github.com/philip-h/amics/internal/storage"
	"github.com/philip-h/amics/view/model"
	"github.com/philip-h/amics/view/teacher"
)

// categoryIdFromForm returns a NULL id for blank input; ok is false if the id is not one of the course's categories.
func categoryIdFromForm(raw string, categories []*storage.Category) (id sql.NullInt64, ok bool) {
	if raw == "" {
		return sql.NullInt64{}, true
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return sql.NullInt64{}, false
	}
	for _, c := range categories {
		if c.Id == parsed {
			return sql.NullInt64{Int64: int64(parsed), Valid: true}, true
		}
	}
	return sql.NullInt64{}, false
}

func handleTeacherCategoriesGet(store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		courseId, err := intPathValue(r, "courseId")
		if err != nil {
			return httpe.ServerError(fmt.Errorf("[teacher.categories.get] Could not convert course ID to int: %w", err), http.StatusBadRequest)
		}

		categories, err := store.Categories.ListByCourseId(r.Context(), courseId)
		if err != nil {
			return fmt.Errorf("[teacher.categories.get] Could not get categories: %w", err)
		}

		formVm := model.NewCategoriesFormFromStorage(strconv.Itoa(courseId), categories)
		return teacher.ManageCategories(formVm).Render(r.Context(), w)
	})
}

func handleTeacherCategoriesPut(logger *Logger, store *storage.Storage) http.Handler {
	return httpe.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		courseId, err := intPathValue(r, "courseId")
		if err != nil {
			return httpe.ServerError(fmt.Errorf("[teacher.categories.put] Could not convert course ID to int: %w", err), http.StatusBadRequest)
		}

		if err := r.ParseForm(); err != nil {
			return httpe.ServerError(fmt.Errorf("[teacher.categories.put] Could not parse form: %w", err), http.StatusBadRequest)
		}
		formVm := model.NewCategoriesFormPost(
			strconv.Itoa(courseId),
			r.PostForm["category-id"],
			r.PostForm["category-name"],
			r.PostForm["category-weight"],
		)

		if problems := formVm.Validate(); len(problems) > 0 {
			logger.L.Debug("[teacher.categories.put] Failed validation", slog.Any("problems", problems))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return teacher.ManageCategories(formVm).Render(r.Context(), w)
		}

		categories, ok := formVm.ToCategories(courseId)
		if !ok {
			return httpe.ServerError(errors.New("[teacher.categories.put] Invalid category id"), http.StatusBadRequest)
		}

		if err := store.Categories.ReplaceAll(r.Context(), courseId, categories); err != nil {
			logger.L.Error("[teacher.categories.put] Could not save categories", slog.String("err", err.Error()))
			formVm.ServerError = "Could not save categories. Names must be unique, and each category must belong to this course."
			w.WriteHeader(http.StatusUnprocessableEntity)
			return teacher.ManageCategories(formVm).Render(r.Context(), w)
		}

		w.Header().Set("HX-Redirect", "/teacher")
		w.Header().Set("HX-Push-Url", "/teacher")
		setFlash(w, &Flash{Message: "Successfully updated grade categories", IsError: false})
		return nil
	})
}
