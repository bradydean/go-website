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

type newItemHandler struct {
	db *pgxpool.Pool
}

func NewNewItemHandler(db *pgxpool.Pool) newItemHandler {
	return newItemHandler{db: db}
}

func (h newItemHandler) Handler(c echo.Context) error {
	var listID int64

	if err := echo.PathParamsBinder(c).MustInt64("list_id", &listID).BindError(); err != nil {
		return err
	}

	var content string

	if err := echo.FormFieldBinder(c).MustString("content", &content).BindError(); err != nil {
		return err
	}

	contentLength := utf8.RuneCountInString(content)

	if contentLength == 0 || contentLength > 50 {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "content must be between 1 and 50 characters")
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

	query, args := todo.Items.
		INSERT(
			todo.Items.Content,
			todo.Items.IsComplete,
			todo.Items.ListID,
		).
		VALUES(
			content,
			false,
			listID,
		).
		RETURNING(
			todo.Items.ItemID,
			todo.Items.Content,
			todo.Items.IsComplete,
		).
		Sql()

	rows, _ := h.db.Query(c.Request().Context(), query, args...)
	record, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[ItemsRecord])

	if err != nil {
		return fmt.Errorf("failed to insert item: %w", err)
	}

	item := components.Item{
		ItemID:     record.ItemID,
		Content:    record.Content,
		IsComplete: record.IsComplete,
		Url:        fmt.Sprintf("/lists/%d/items/%d", listID, record.ItemID),
	}

	return components.Render(c, http.StatusCreated, components.ListItem(item))
}
