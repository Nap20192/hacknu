package api

import (
	"time"

	"github.com/Nap20192/hacknu/gen/sqlc"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// registerRoutes mounts all REST routes on the Fiber app.
func registerRoutes(app *fiber.App, q *sqlc.Queries) {
	v1 := app.Group("/api/v1")

	// ── Locomotives ──────────────────────────────────────────────────────────
	locos := v1.Group("/locomotives")
	locos.Get("/", listLocomotives(q))
	locos.Get("/:id", getLocomotive(q))

	// ── Health snapshots ─────────────────────────────────────────────────────
	locos.Get("/:id/health", getLatestHealth(q))
	locos.Get("/:id/health/history", listHealthHistory(q))

	// ── Alerts ───────────────────────────────────────────────────────────────
	locos.Get("/:id/alerts", listActiveAlerts(q))
	locos.Get("/:id/alerts/history", listAlertsHistory(q))
	v1.Post("/alerts/:alertId/acknowledge", acknowledgeAlert(q))

	// ── Metric definitions ───────────────────────────────────────────────────
	metrics := v1.Group("/metrics/definitions")
	metrics.Get("/", listMetricDefs(q))
	metrics.Put("/", upsertMetricDef(q))
}

// ── Locomotive handlers ───────────────────────────────────────────────────────

// listLocomotives godoc
//
//	@Summary		List all locomotives
//	@Tags			locomotives
//	@Produce		json
//	@Success		200	{object}	PagedResponse[LocomotiveDTO]
//	@Failure		500	{object}	Response[any]
//	@Router			/api/v1/locomotives [get]
func parseLocoID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, fiber.NewError(400, "invalid locomotive id: must be a UUID")
	}
	return id, nil
}

func listLocomotives(q *sqlc.Queries) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := q.ListLocomotives(c.Context())
		if err != nil {
			return c.Status(500).JSON(Response[any]{Error: err.Error()})
		}
		dtos := make([]LocomotiveDTO, len(rows))
		for i, r := range rows {
			dtos[i] = locoToDTO(r)
		}
		return c.JSON(PagedResponse[LocomotiveDTO]{Success: true, Data: dtos, Total: len(dtos)})
	}
}

// getLocomotive godoc
//
//	@Summary		Get locomotive by ID
//	@Tags			locomotives
//	@Produce		json
//	@Param			id	path		string	true	"Locomotive ID"
//	@Success		200	{object}	Response[LocomotiveDTO]
//	@Failure		404	{object}	Response[any]
//	@Failure		500	{object}	Response[any]
//	@Router			/api/v1/locomotives/{id} [get]
func getLocomotive(q *sqlc.Queries) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := parseLocoID(c)
		if err != nil {
			return c.Status(400).JSON(Response[any]{Error: err.Error()})
		}
		row, err := q.GetLocomotive(c.Context(), id)
		if err != nil {
			return c.Status(404).JSON(Response[any]{Error: "locomotive not found"})
		}
		return c.JSON(Response[LocomotiveDTO]{Success: true, Data: locoToDTO(row)})
	}
}

// ── Health snapshot handlers ──────────────────────────────────────────────────

// getLatestHealth godoc
//
//	@Summary		Get latest health snapshot for a locomotive
//	@Tags			health
//	@Produce		json
//	@Param			id	path		string	true	"Locomotive ID"
//	@Success		200	{object}	Response[sqlc.HealthSnapshot]
//	@Failure		404	{object}	Response[any]
//	@Router			/api/v1/locomotives/{id}/health [get]
func getLatestHealth(q *sqlc.Queries) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := parseLocoID(c)
		if err != nil {
			return c.Status(400).JSON(Response[any]{Error: err.Error()})
		}
		snap, err := q.GetLatestHealthSnapshot(c.Context(), id)
		if err != nil {
			return c.Status(404).JSON(Response[any]{Error: "no health data found"})
		}
		return c.JSON(Response[sqlc.HealthSnapshot]{Success: true, Data: snap})
	}
}

// listHealthHistory godoc
//
//	@Summary		List health snapshots in a time range
//	@Tags			health
//	@Produce		json
//	@Param			id		path		string		true	"Locomotive ID"
//	@Param			from	query		string		false	"RFC3339 start time"
//	@Param			to		query		string		false	"RFC3339 end time"
//	@Param			limit	query		int			false	"Max records (default 100)"
//	@Success		200		{object}	PagedResponse[sqlc.HealthSnapshot]
//	@Failure		400		{object}	Response[any]
//	@Failure		500		{object}	Response[any]
//	@Router			/api/v1/locomotives/{id}/health/history [get]
func listHealthHistory(q *sqlc.Queries) fiber.Handler {
	return func(c *fiber.Ctx) error {
		from, to, limit, err := parseTimeRange(c)
		if err != nil {
			return c.Status(400).JSON(Response[any]{Error: err.Error()})
		}
		id, err := parseLocoID(c)
		if err != nil {
			return c.Status(400).JSON(Response[any]{Error: err.Error()})
		}
		if limit > 0 {
			rows, err := q.ListHealthSnapshotsLatest(c.Context(), sqlc.ListHealthSnapshotsLatestParams{
				LocomotiveID: id,
				Limit:        limit,
			})
			if err != nil {
				return c.Status(500).JSON(Response[any]{Error: err.Error()})
			}
			return c.JSON(PagedResponse[sqlc.HealthSnapshot]{Success: true, Data: rows, Total: len(rows)})
		}
		rows, err := q.ListHealthSnapshotsRange(c.Context(), sqlc.ListHealthSnapshotsRangeParams{
			LocomotiveID: id,
			Ts:           from,
			Ts_2:         to,
		})
		if err != nil {
			return c.Status(500).JSON(Response[any]{Error: err.Error()})
		}
		return c.JSON(PagedResponse[sqlc.HealthSnapshot]{Success: true, Data: rows, Total: len(rows)})
	}
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
		id, err := parseLocoID(c)
		if err != nil {
			return c.Status(400).JSON(Response[any]{Error: err.Error()})
		}
		rows, err := q.ListActiveAlertsByLocomotive(c.Context(), id)
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
		id, err := parseLocoID(c)
		if err != nil {
			return c.Status(400).JSON(Response[any]{Error: err.Error()})
		}
		from, to, _, err := parseTimeRange(c)
		if err != nil {
			return c.Status(400).JSON(Response[any]{Error: err.Error()})
		}
		rows, err := q.ListAlertsByLocomotiveRange(c.Context(), sqlc.ListAlertsByLocomotiveRangeParams{
			LocomotiveID:  id,
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

// ── Metric definition handlers ────────────────────────────────────────────────

// listMetricDefs godoc
//
//	@Summary		List all metric definitions
//	@Tags			metrics
//	@Produce		json
//	@Success		200	{object}	PagedResponse[MetricDefinitionDTO]
//	@Failure		500	{object}	Response[any]
//	@Router			/api/v1/metrics/definitions [get]
func listMetricDefs(q *sqlc.Queries) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := q.ListMetricDefinitions(c.Context())
		if err != nil {
			return c.Status(500).JSON(Response[any]{Error: err.Error()})
		}
		dtos := make([]MetricDefinitionDTO, len(rows))
		for i, r := range rows {
			dtos[i] = metricDefToDTO(r)
		}
		return c.JSON(PagedResponse[MetricDefinitionDTO]{Success: true, Data: dtos, Total: len(dtos)})
	}
}

// upsertMetricDef godoc
//
//	@Summary		Create or update a metric definition
//	@Tags			metrics
//	@Accept			json
//	@Produce		json
//	@Param			body	body		MetricDefinitionDTO	true	"Metric definition"
//	@Success		200		{object}	Response[MetricDefinitionDTO]
//	@Failure		400		{object}	Response[any]
//	@Failure		500		{object}	Response[any]
//	@Router			/api/v1/metrics/definitions [put]
func upsertMetricDef(q *sqlc.Queries) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var dto MetricDefinitionDTO
		if err := c.BodyParser(&dto); err != nil {
			return c.Status(400).JSON(Response[any]{Error: "invalid request body"})
		}
		if dto.Name == "" {
			return c.Status(400).JSON(Response[any]{Error: "name is required"})
		}
		row, err := q.UpsertMetricDefinition(c.Context(), sqlc.UpsertMetricDefinitionParams{
			Name:         dto.Name,
			Display:      dto.Display,
			Description:  dto.Description,
			Unit:         dto.Unit,
			WarnAbove:    dto.WarnAbove,
			WarnBelow:    dto.WarnBelow,
			CritAbove:    dto.CritAbove,
			CritBelow:    dto.CritBelow,
			HealthWeight: dto.HealthWeight,
			EmaAlpha:     0.3, // sensible default
		})
		if err != nil {
			return c.Status(500).JSON(Response[any]{Error: err.Error()})
		}
		return c.JSON(Response[MetricDefinitionDTO]{Success: true, Data: metricDefToDTO(row)})
	}
}

// ── Mapping helpers ───────────────────────────────────────────────────────────

func locoToDTO(r sqlc.Locomotive) LocomotiveDTO {
	return LocomotiveDTO{
		ID:           r.ID,
		DisplayName:  r.DisplayName,
		LocoType:     r.LocoType,
		RegisteredAt: r.RegisteredAt,
		LastSeenAt:   r.LastSeenAt,
		Active:       r.Active,
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

func metricDefToDTO(r sqlc.MetricDefinition) MetricDefinitionDTO {
	return MetricDefinitionDTO{
		Name:         r.Name,
		Display:      r.Display,
		Description:  r.Description,
		Unit:         r.Unit,
		WarnAbove:    r.WarnAbove,
		WarnBelow:    r.WarnBelow,
		CritAbove:    r.CritAbove,
		CritBelow:    r.CritBelow,
		HealthWeight: r.HealthWeight,
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
