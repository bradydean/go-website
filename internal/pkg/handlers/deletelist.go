package handlers

import (
	"fmt"
	"net/http"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/bradydean/go-website/internal/pkg/components"
	"github.com/bradydean/go-website/internal/pkg/profile"
	todo "github.com/bradydean/go-website/internal/pkg/todo_api/todo/table"
)

type deleteListHandler struct {
	db *pgxpool.Pool
}

func NewDeleteListHandler(db *pgxpool.Pool) deleteListHandler {
	return deleteListHandler{db: db}
}

func (h deleteListHandler) Handler(c echo.Context) error {
	var listID int64

	if err := echo.PathParamsBinder(c).MustInt64("list_id", &listID).BindError(); err != nil {
		return err
	}

	profile, err := profile.MustGet(c)

	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	query, args := todo.Lists.
		DELETE().
		WHERE(
			todo.Lists.ListID.EQ(pg.Int(listID)).
				AND(todo.Lists.UserID.EQ(pg.String(profile.UserID))),
		).
		Sql()

	if _, err := h.db.Exec(c.Request().Context(), query, args...); err != nil {
		return fmt.Errorf("failed to delete list: %w", err)
	}

	return components.Render(c, http.StatusOK, components.Empty())
}
