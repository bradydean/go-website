package handlers

import (
	"errors"
	"fmt"
	"net/http"

	pg "github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/bradydean/go-website/internal/pkg/components"
	"github.com/bradydean/go-website/internal/pkg/profile"
	todo "github.com/bradydean/go-website/internal/pkg/todo_api/todo/table"
)

type itemsHandler struct {
	db *pgxpool.Pool
}

func NewItemsHandler(db *pgxpool.Pool) itemsHandler {
	return itemsHandler{db: db}
}

type ItemsRecord struct {
	ItemID     int64  `db:"items.item_id"`
	Content    string `db:"items.content"`
	IsComplete bool   `db:"items.is_complete"`
}

func (h itemsHandler) Handler(c echo.Context) error {
	var listID int64

	if err := echo.PathParamsBinder(c).MustInt64("list_id", &listID).BindError(); err != nil {
		return err
	}

	profile, err := profile.MustGet(c)

	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	listQuery, listArgs := pg.SELECT(
		todo.Lists.ListID,
		todo.Lists.Title,
		todo.Lists.Description,
	).
		FROM(todo.Lists).
		WHERE(
			todo.Lists.ListID.EQ(pg.Int(listID)).
				AND(todo.Lists.UserID.EQ(pg.String(profile.UserID))),
		).
		Sql()

	type ListRecord struct {
		ListID      int64  `db:"lists.list_id"`
		Title       string `db:"lists.title"`
		Description string `db:"lists.description"`
	}

	listRows, _ := h.db.Query(c.Request().Context(), listQuery, listArgs...)
	listRecord, err := pgx.CollectOneRow(listRows, pgx.RowToStructByName[ListRecord])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if c.Request().Header.Get("HX-Boosted") != "" {
				return components.Render(c, http.StatusNotFound, components.Boost("List Not Found", components.NotFound(&profile)))
			}

			layout := components.Layout("List Not Found", c.Get("csrf").(string), components.NotFound(&profile))
			return components.Render(c, http.StatusNotFound, layout)
		}

		return fmt.Errorf("failed to check if list exists: %w", err)
	}

	itemsQuery, itemsArgs := pg.SELECT(
		todo.Items.ItemID,
		todo.Items.Content,
		todo.Items.IsComplete,
	).
		FROM(todo.Items).
		WHERE(todo.Items.ListID.EQ(pg.Int(listID))).
		ORDER_BY(todo.Items.ItemID).
		Sql()

	itemRows, _ := h.db.Query(c.Request().Context(), itemsQuery, itemsArgs...)
	itemRecords, err := pgx.CollectRows(itemRows, pgx.RowToStructByName[ItemsRecord])

	if err != nil {
		return fmt.Errorf("failed to fetch items: %w", err)
	}

	list := components.List{
		ListID:      listRecord.ListID,
		Title:       listRecord.Title,
		Description: listRecord.Description,
		Url:         fmt.Sprintf("/lists/%d", listRecord.ListID),
	}

	items := make([]components.Item, 0, len(itemRecords))

	for _, record := range itemRecords {
		items = append(items, components.Item{
			ItemID:     record.ItemID,
			Content:    record.Content,
			IsComplete: record.IsComplete,
			Url:        fmt.Sprintf("/lists/%d/items/%d", listID, record.ItemID),
		})
	}

	if c.Request().Header.Get("HX-Boosted") != "" {
		return components.Render(c, http.StatusOK, components.Boost(listRecord.Title, components.Items(profile, list, items)))
	}

	layout := components.Layout(listRecord.Title, c.Get("csrf").(string), components.Items(profile, list, items))
	return components.Render(c, http.StatusOK, layout)
}
