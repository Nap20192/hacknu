package api

import (
	"time"

	"github.com/Nap20192/hacknu/gen/sqlc"
	"github.com/gofiber/fiber/v2"
)

func registerRoutes(app *fiber.App, q *sqlc.Queries) {
	v1 := app.Group("/api/v1")

	// ── Metric definitions ───────────────────────────────────────────────────
	locos := v1.Group("/locomotives")
	locos.Get("/:id/alerts", listActiveAlerts(q))
	locos.Get("/:id/alerts/history", listAlertsHistory(q))
	v1.Post("/alerts/:alertId/acknowledge", acknowledgeAlert(q))
}

// ── Alert handlers ────────────────────────────────────────────────────────────

// listActiveAlerts godoc
//
//	@Summary		List active (unresolved) alerts for a locomotive
//	@Tags			alerts
//	@Produce		json
//	@Param			id	path		string	true	"Locomotive ID"
//	@Success		200	{object}	PagedResponse[AlertDTO]
//	@Failure		500	{object}	Response[any]
//	@Router			/api/v1/locomotives/{id}/alerts [get]
func listActiveAlerts(q *sqlc.Queries) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := q.ListActiveAlertsByLocomotive(c.Context(), c.Params("id"))
		if err != nil {
			return c.Status(500).JSON(Response[any]{Error: err.Error()})
		}
		dtos := make([]AlertDTO, len(rows))
		for i, r := range rows {
			dtos[i] = alertToDTO(r)
		}
		return c.JSON(PagedResponse[AlertDTO]{Success: true, Data: dtos, Total: len(dtos)})
	}
}

// listAlertsHistory godoc
//
//	@Summary		List alerts in a time range
//	@Tags			alerts
//	@Produce		json
//	@Param			id		path		string	true	"Locomotive ID"
//	@Param			from	query		string	false	"RFC3339 start time"
//	@Param			to		query		string	false	"RFC3339 end time"
//	@Success		200		{object}	PagedResponse[AlertDTO]
//	@Failure		400		{object}	Response[any]
//	@Failure		500		{object}	Response[any]
//	@Router			/api/v1/locomotives/{id}/alerts/history [get]
func listAlertsHistory(q *sqlc.Queries) fiber.Handler {
	return func(c *fiber.Ctx) error {
		from, to, _, err := parseTimeRange(c)
		if err != nil {
			return c.Status(400).JSON(Response[any]{Error: err.Error()})
		}
		rows, err := q.ListAlertsByLocomotiveRange(c.Context(), sqlc.ListAlertsByLocomotiveRangeParams{
			LocomotiveID:  c.Params("id"),
			TriggeredAt:   from,
			TriggeredAt_2: to,
		})
		if err != nil {
			return c.Status(500).JSON(Response[any]{Error: err.Error()})
		}
		dtos := make([]AlertDTO, len(rows))
		for i, r := range rows {
			dtos[i] = alertToDTO(r)
		}
		return c.JSON(PagedResponse[AlertDTO]{Success: true, Data: dtos, Total: len(dtos)})
	}
}

// acknowledgeAlert godoc
//
//	@Summary		Acknowledge an alert by ID
//	@Tags			alerts
//	@Produce		json
//	@Param			alertId	path		int		true	"Alert ID"
//	@Success		200		{object}	Response[AlertDTO]
//	@Failure		400		{object}	Response[any]
//	@Failure		500		{object}	Response[any]
//	@Router			/api/v1/alerts/{alertId}/acknowledge [post]
func acknowledgeAlert(q *sqlc.Queries) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("alertId")
		if err != nil {
			return c.Status(400).JSON(Response[any]{Error: "invalid alert id"})
		}
		row, err := q.AcknowledgeAlert(c.Context(), int64(id))
		if err != nil {
			return c.Status(500).JSON(Response[any]{Error: err.Error()})
		}
		return c.JSON(Response[AlertDTO]{Success: true, Data: alertToDTO(row)})
	}
}

func alertToDTO(r sqlc.Alert) AlertDTO {
	return AlertDTO{
		ID:             r.ID,
		LocomotiveID:   r.LocomotiveID,
		TriggeredAt:    r.TriggeredAt,
		ResolvedAt:     r.ResolvedAt,
		Severity:       r.Severity,
		Code:           r.Code,
		MetricName:     r.MetricName,
		MetricValue:    r.MetricValue,
		Threshold:      r.Threshold,
		Message:        r.Message,
		Recommendation: r.Recommendation,
		Acknowledged:   r.Acknowledged,
	}
}

// parseTimeRange extracts from/to/limit from query params with sensible defaults.
func parseTimeRange(c *fiber.Ctx) (from, to time.Time, limit int32, err error) {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	limitInt := c.QueryInt("limit", 0)
	limit = int32(limitInt)

	if fromStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return from, to, limit, fiber.NewError(400, "invalid 'from' format, use RFC3339")
		}
	} else {
		from = time.Now().Add(-24 * time.Hour)
	}

	if toStr != "" {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			return from, to, limit, fiber.NewError(400, "invalid 'to' format, use RFC3339")
		}
	} else {
		to = time.Now()
	}

	return from, to, limit, nil
}
