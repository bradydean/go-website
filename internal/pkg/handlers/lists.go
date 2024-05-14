package handlers

import (
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

type listsHandler struct {
	db *pgxpool.Pool
}

func NewListsHandler(db *pgxpool.Pool) listsHandler {
	return listsHandler{db: db}
}

type listsRecord struct {
	ListID      int64  `db:"lists.list_id"`
	Title       string `db:"lists.title"`
	Description string `db:"lists.description"`
}

func (h listsHandler) Handler(c echo.Context) error {
	profile, err := profile.MustGet(c)

	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	query, args := pg.SELECT(
		todo.Lists.ListID,
		todo.Lists.Title,
		todo.Lists.Description,
	).
		FROM(todo.Lists).
		WHERE(todo.Lists.UserID.EQ(pg.String(profile.UserID))).
		ORDER_BY(todo.Lists.ListID).
		Sql()

	rows, _ := h.db.Query(c.Request().Context(), query, args...)
	records, err := pgx.CollectRows(rows, pgx.RowToStructByName[listsRecord])

	if err != nil {
		return fmt.Errorf("failed to fetch lists: %w", err)
	}

	lists := make([]components.List, 0, len(records))

	for _, record := range records {
		lists = append(lists, components.List{
			Title:       record.Title,
			Description: record.Description,
			Url:         fmt.Sprintf("/lists/%d", record.ListID),
		})
	}

	layout := components.Layout("My Lists", &profile, components.Lists(lists))

	return components.Render(c, http.StatusOK, layout)
}
