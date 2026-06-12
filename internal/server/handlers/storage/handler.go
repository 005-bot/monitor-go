package storage

import (
	"errors"

	"github.com/005-bot/monitor-go/internal/domain"
	"github.com/005-bot/monitor-go/internal/storage"
	"github.com/go-core-fx/fiberfx/handler"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

const errKey = "error"

type Handler struct {
	handler.Base

	storageSvc *storage.Service
	logger     *zap.Logger
}

func NewHandler(storageSvc *storage.Service, validator *validator.Validate, logger *zap.Logger) *Handler {
	return &Handler{
		Base: handler.Base{
			Validator: validator,
		},

		storageSvc: storageSvc,
		logger:     logger,
	}
}

func (h *Handler) Register(r fiber.Router) {
	r = r.Group("/storage")

	r.Get("/etag", h.getEtag)
	r.Post("/etag", h.postEtag)
	r.Post("/diff", h.postDiff)
	r.Post("/commit", h.postCommit)
}

type etagRequest struct {
	Etag string `json:"etag" validate:"required"`
}

type etagResponse struct {
	Changed bool `json:"changed"`
}

func (h *Handler) postEtag(c *fiber.Ctx) error {
	var req etagRequest
	if err := h.BodyParserValidator(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{errKey: err.Error()})
	}

	changed, err := h.storageSvc.IsEtagChanged(c.Context(), req.Etag)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		h.logger.Error("etag check failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{errKey: err.Error()})
	}

	return c.JSON(etagResponse{Changed: changed})
}

type currentEtagResponse struct {
	Etag string `json:"etag"`
}

func (h *Handler) getEtag(c *fiber.Ctx) error {
	val, err := h.storageSvc.GetEtag(c.Context())
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{errKey: "etag not set"})
	}

	return c.JSON(currentEtagResponse{Etag: val})
}

type diffRequest struct {
	Records []domain.ParsedRecord `json:"records"`
}

type diffResponse struct {
	Changed []domain.ParsedRecord `json:"changed"`
	Count   int                   `json:"count"`
}

func (h *Handler) postDiff(c *fiber.Ctx) error {
	var req diffRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{errKey: err.Error()})
	}

	changed, err := h.storageSvc.Diff(c.Context(), req.Records)
	if err != nil {
		h.logger.Error("diff failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{errKey: err.Error()})
	}

	if changed == nil {
		changed = []domain.ParsedRecord{}
	}

	return c.JSON(diffResponse{Changed: changed, Count: len(changed)})
}

type commitRequest struct {
	Records []domain.ParsedRecord `json:"records"`
}

type commitResponse struct {
	Committed int `json:"committed"`
}

func (h *Handler) postCommit(c *fiber.Ctx) error {
	var req commitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{errKey: err.Error()})
	}

	if err := h.storageSvc.Commit(c.Context(), req.Records); err != nil {
		h.logger.Error("commit failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{errKey: err.Error()})
	}

	return c.JSON(commitResponse{Committed: len(req.Records)})
}
