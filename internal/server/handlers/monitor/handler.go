package monitor

import (
	"github.com/005-bot/monitor-go/internal/scheduler"
	"github.com/go-core-fx/fiberfx/handler"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

const errKey = "error"

type Handler struct {
	handler.Base

	schedulerSvc *scheduler.Service
	logger       *zap.Logger
}

func NewHandler(schedulerSvc *scheduler.Service, validator *validator.Validate, logger *zap.Logger) *Handler {
	return &Handler{
		Base: handler.Base{
			Validator: validator,
		},
		schedulerSvc: schedulerSvc,
		logger:       logger,
	}
}

func (h *Handler) Register(r fiber.Router) {
	r = r.Group("/monitor")

	r.Get("/status", h.status)
	r.Post("/run", h.run)
}

type statusResponse struct {
	Running   bool    `json:"running"`
	LastRunAt *string `json:"last_run_at"`
}

func (h *Handler) status(c *fiber.Ctx) error {
	status := h.schedulerSvc.Status()

	var lastRunAt *string
	if !status.LastRunAt.IsZero() {
		formatted := status.LastRunAt.Format("2006-01-02T15:04:05-07:00")
		lastRunAt = &formatted
	}

	return c.JSON(statusResponse{
		Running:   status.Running,
		LastRunAt: lastRunAt,
	})
}

type runResponse struct {
	Success bool `json:"success"`
}

func (h *Handler) run(c *fiber.Ctx) error {
	if err := h.schedulerSvc.Run(c.Context()); err != nil {
		h.logger.Error("monitor run failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{errKey: err.Error()})
	}

	return c.JSON(runResponse{Success: true})
}
