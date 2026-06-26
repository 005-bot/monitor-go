package parser

import (
	"github.com/005-bot/monitor-go/internal/domain"
	"github.com/005-bot/monitor-go/internal/parser"
	"github.com/005-bot/monitor-go/internal/scraper"
	"github.com/go-core-fx/fiberfx/handler"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

const errKey = "error"

type Handler struct {
	handler.Base

	parserSvc  *parser.Service
	scraperSvc *scraper.Service
	logger     *zap.Logger
}

func NewHandler(
	parserSvc *parser.Service,
	scraperSvc *scraper.Service,
	validator *validator.Validate,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		Base: handler.Base{
			Validator: validator,
		},
		parserSvc:  parserSvc,
		scraperSvc: scraperSvc,
		logger:     logger,
	}
}

func (h *Handler) Register(r fiber.Router) {
	r = r.Group("/parser")

	r.Post("/record", h.parseRecord)
	r.Post("/scrape-and-parse", h.scrapeAndParse)
}

type parseRecordRequest struct {
	Record domain.Record `json:"record" validate:"required"`
}

type parseRecordResponse struct {
	Record domain.ParsedRecord `json:"record"`
}

func (h *Handler) parseRecord(c *fiber.Ctx) error {
	var req parseRecordRequest
	if err := h.BodyParserValidator(c, &req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{errKey: err.Error()})
	}

	parsed, err := h.parserSvc.Parse(c.Context(), req.Record)
	if err != nil {
		h.logger.Error("parse failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{errKey: err.Error()})
	}

	return c.JSON(parseRecordResponse{Record: parsed})
}

type scrapeAndParseResponse struct {
	Records []domain.ParsedRecord `json:"records"`
	Count   int                   `json:"count"`
}

func (h *Handler) scrapeAndParse(c *fiber.Ctx) error {
	records, err := h.scraperSvc.Run(c.Context())
	if err != nil {
		h.logger.Error("scrape run failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{errKey: err.Error()})
	}

	if records == nil {
		records = []domain.Record{}
	}

	parsed, err := h.parserSvc.ParseBatch(c.Context(), records)
	if err != nil {
		h.logger.Error("parse batch failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{errKey: err.Error()})
	}

	return c.JSON(scrapeAndParseResponse{Records: parsed, Count: len(parsed)})
}
