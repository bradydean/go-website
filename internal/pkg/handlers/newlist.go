package handlers

import (
	"fmt"
	"net/http"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/bradydean/go-website/internal/pkg/components"
	"github.com/bradydean/go-website/internal/pkg/profile"
	todo "github.com/bradydean/go-website/internal/pkg/todo_api/todo/table"
)

type newListHandler struct {
	db *pgxpool.Pool
}

func NewNewListHandler(db *pgxpool.Pool) newListHandler {
	return newListHandler{db: db}
}

func (h newListHandler) Handler(c echo.Context) error {
	var title, description string

	if err := echo.FormFieldBinder(c).MustString("title", &title).MustString("description", &description).BindError(); err != nil {
		return err
	}

	titleLen := utf8.RuneCountInString(title)

	if titleLen == 0 || titleLen > 50 {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "title must be between 1 and 50 characters")
	}

	descriptionLen := utf8.RuneCountInString(description)

	if descriptionLen == 0 || descriptionLen > 50 {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "description must be between 1 and 50 characters")
	}

	profile, err := profile.MustGet(c)

	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	query, args := todo.Lists.INSERT(
		todo.Lists.Title,
		todo.Lists.Description,
		todo.Lists.UserID,
	).
		VALUES(
			title,
			description,
			profile.UserID,
		).
		RETURNING(todo.Lists.ListID).
		Sql()

	rows, _ := h.db.Query(c.Request().Context(), query, args...)
	listID, err := pgx.CollectOneRow(rows, pgx.RowTo[int64])

	if err != nil {
		return fmt.Errorf("failed to create list: %w", err)
	}

	c.Response().Header().Set("HX-Replace-Url", fmt.Sprintf("/lists/%d", listID))

	if c.Request().Header.Get("HX-Boosted") != "" {
		return components.Render(c, http.StatusOK, components.Items(profile, title, nil))
	}

	layout := components.Layout(title, components.Items(profile, title, nil))
	return components.Render(c, http.StatusOK, layout)
}
