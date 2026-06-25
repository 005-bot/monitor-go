package scraper

import (
	"github.com/005-bot/monitor-go/internal/domain"
	"github.com/005-bot/monitor-go/internal/scraper"
	"github.com/go-core-fx/fiberfx/handler"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

const errKey = "error"

type Handler struct {
	handler.Base

	scraperSvc *scraper.Service
	logger     *zap.Logger
}

func NewHandler(scraperSvc *scraper.Service, validator *validator.Validate, logger *zap.Logger) *Handler {
	return &Handler{
		Base: handler.Base{
			Validator: validator,
		},
		scraperSvc: scraperSvc,
		logger:     logger,
	}
}

func (h *Handler) Register(r fiber.Router) {
	r = r.Group("/scrape")

	r.Get("/run", h.run)
}

type runResponse struct {
	Records []domain.Record `json:"records"`
	Count   int             `json:"count"`
}

func (h *Handler) run(c *fiber.Ctx) error {
	records, err := h.scraperSvc.Run(c.Context())
	if err != nil {
		h.logger.Error("scrape run failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{errKey: err.Error()})
	}

	if records == nil {
		records = []domain.Record{}
	}

	return c.JSON(runResponse{Records: records, Count: len(records)})
}
