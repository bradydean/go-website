package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"unicode/utf8"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/bradydean/go-website/internal/pkg/components"
	"github.com/bradydean/go-website/internal/pkg/profile"
	todo "github.com/bradydean/go-website/internal/pkg/todo_api/todo/table"
)

type patchItemHandler struct {
	db *pgxpool.Pool
}

func NewPatchItemHandler(db *pgxpool.Pool) patchItemHandler {
	return patchItemHandler{db: db}
}

func (h patchItemHandler) Handler(c echo.Context) error {
	var listID, itemID int64

	if err := echo.PathParamsBinder(c).
		MustInt64("list_id", &listID).
		MustInt64("item_id", &itemID).
		BindError(); err != nil {
		return err
	}

	var params struct {
		IsComplete *bool   `form:"is_complete"`
		Content    *string `form:"content"`
	}

	if err := c.Bind(&params); err != nil {
		return err
	}

	if params.Content != nil {
		contentLenth := utf8.RuneCountInString(*params.Content)
		if contentLenth == 0 || contentLenth > 50 {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "content must be between 1 and 50 characters")
		}
	}

	profile, err := profile.MustGet(c)

	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	listQuery, listArgs := pg.SELECT(pg.Int64(1)).
		FROM(todo.Lists).
		WHERE(
			todo.Lists.ListID.EQ(pg.Int(listID)).
				AND(todo.Lists.UserID.EQ(pg.String(profile.UserID))),
		).
		Sql()

	listRows, _ := h.db.Query(c.Request().Context(), listQuery, listArgs...)

	if _, err := pgx.CollectOneRow(listRows, pgx.RowTo[int64]); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return components.Render(c, http.StatusNotFound, components.NotFound(&profile))
		}
		return fmt.Errorf("failed to check if list exists: %w", err)
	}

	stmt := todo.Items.
		UPDATE().
		SET(todo.Items.ItemID.SET(pg.Int(itemID))).
		WHERE(
			todo.Items.ItemID.EQ(pg.Int(itemID)).
				AND(todo.Items.ListID.EQ(pg.Int(listID))),
		).
		RETURNING(
			todo.Items.ItemID,
			todo.Items.Content,
			todo.Items.IsComplete,
		)

	if params.Content != nil {
		stmt = stmt.SET(todo.Items.Content.SET(pg.String(*params.Content)))
	}

	if params.IsComplete != nil {
		stmt = stmt.SET(todo.Items.IsComplete.SET(pg.Bool(*params.IsComplete)))
	}

	query, args := stmt.Sql()

	rows, _ := h.db.Query(c.Request().Context(), query, args...)
	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[ItemsRecord])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return components.Render(c, http.StatusNotFound, components.NotFound(&profile))
		}
		return fmt.Errorf("failed to update item: %w", err)
	}

	item := components.Item{
		ItemID:     record.ItemID,
		Content:    record.Content,
		IsComplete: record.IsComplete,
		Url:        fmt.Sprintf("/lists/%d/items/%d", listID, itemID),
	}

	return components.Render(c, http.StatusOK, components.ListItem(item, c.Get("csrf").(string)))
}
