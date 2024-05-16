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

type deleteItemHandler struct {
	db *pgxpool.Pool
}

func NewDeleteItemHandler(db *pgxpool.Pool) deleteItemHandler {
	return deleteItemHandler{db: db}
}

func (h deleteItemHandler) Handler(c echo.Context) error {
	var itemID, listID int64

	if err := echo.PathParamsBinder(c).MustInt64("list_id", &listID).MustInt64("item_id", &itemID).BindError(); err != nil {
		return err
	}

	profile, err := profile.MustGet(c)

	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	query, args := todo.Items.
		DELETE().
		USING(todo.Lists).
		WHERE(
			todo.Items.ItemID.EQ(pg.Int(itemID)).
				AND(todo.Items.ListID.EQ(pg.Int(listID))).
				AND(todo.Lists.UserID.EQ(pg.String(profile.UserID))),
		).
		Sql()

	if _, err := h.db.Exec(c.Request().Context(), query, args...); err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return components.Render(c, http.StatusOK, components.Empty())
}
