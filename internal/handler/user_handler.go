package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/talhag3/go-api-boilerplate/internal/domain"
	"github.com/talhag3/go-api-boilerplate/internal/repository"
	"github.com/talhag3/go-api-boilerplate/internal/service"
)

// UserHandler wires HTTP routes to the user service.
type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Register attaches routes to a fiber.Router.
func (h *UserHandler) Register(r fiber.Router) {
	r.Post("/users", h.Create)
	r.Get("/users", h.List)
	r.Get("/users/:id", h.GetByID)
	r.Put("/users/:id", h.Update)
	r.Delete("/users/:id", h.Delete)
}

// ---- Create ----
func (h *UserHandler) Create(c fiber.Ctx) error {
	var in domain.CreateUserInput

	// Fiber v3 uses c.Bind().JSON() instead of c.BodyParser()
	if err := c.Bind().JSON(&in); err != nil {
		return respondError(c, fiber.StatusBadRequest, "INVALID_BODY", "malformed JSON")
	}

	// Fiber v3 uses c.Context() directly
	user, err := h.svc.Create(c.Context(), in)
	if err != nil {
		return mapServiceErr(c, err)
	}
	return respondJSON(c, fiber.StatusCreated, user)
}

// ---- List ----
func (h *UserHandler) List(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	// Fiber v3 uses c.Context() directly
	users, err := h.svc.List(c.Context(), page, pageSize)
	if err != nil {
		return mapServiceErr(c, err)
	}
	return respondJSON(c, fiber.StatusOK, users)
}

// ---- GetByID ----
func (h *UserHandler) GetByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respondError(c, fiber.StatusBadRequest, "INVALID_ID", "id must be a UUID")
	}

	// Fiber v3 uses c.Context() directly
	user, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		return mapServiceErr(c, err)
	}
	return respondJSON(c, fiber.StatusOK, user)
}

// ---- Update ----
func (h *UserHandler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respondError(c, fiber.StatusBadRequest, "INVALID_ID", "id must be a UUID")
	}

	var in domain.UpdateUserInput
	if err := c.Bind().JSON(&in); err != nil {
		return respondError(c, fiber.StatusBadRequest, "INVALID_BODY", "malformed JSON")
	}

	// Fiber v3 uses c.Context() directly
	user, err := h.svc.Update(c.Context(), id, in)
	if err != nil {
		return mapServiceErr(c, err)
	}
	return respondJSON(c, fiber.StatusOK, user)
}

// ---- Delete ----
func (h *UserHandler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respondError(c, fiber.StatusBadRequest, "INVALID_ID", "id must be a UUID")
	}

	// Fiber v3 uses c.Context() directly
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return mapServiceErr(c, err)
	}
	return respondJSON(c, fiber.StatusNoContent, nil)
}

// ---- Helpers ----

func respondJSON(c fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(envelope{Success: true, Data: data})
}

func respondError(c fiber.Ctx, status int, code, msg string) error {
	return c.Status(status).JSON(envelope{
		Success: false,
		Error:   &errBody{Code: code, Message: msg},
	})
}

// mapServiceErr translates service-layer errors to HTTP status codes.
func mapServiceErr(c fiber.Ctx, err error) error {
	var inv service.ErrInvalidInput
	switch {
	case errors.As(err, &inv):
		return respondError(c, fiber.StatusBadRequest, "INVALID_INPUT", inv.Error())
	case errors.Is(err, repository.ErrUserNotFound):
		return respondError(c, fiber.StatusNotFound, "USER_NOT_FOUND", err.Error())
	case errors.Is(err, repository.ErrEmailTaken):
		return respondError(c, fiber.StatusConflict, "EMAIL_TAKEN", err.Error())
	default:
		return respondError(c, fiber.StatusInternalServerError, "INTERNAL", "internal server error")
	}
}
