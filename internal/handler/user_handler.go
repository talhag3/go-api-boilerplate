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

// UserHandler connects HTTP requests to our user service.
type UserHandler struct {
	svc service.UserService // The service layer where the real business logic is!
}

// NewUserHandler is a constructor function to create a new UserHandler.
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Register maps our endpoints (routes) to the Fiber router.
func (h *UserHandler) Register(r fiber.Router) {
	r.Post("/users", h.Create)      // Create a user
	r.Get("/users", h.List)          // List users with pagination
	r.Get("/users/:id", h.GetByID)   // Get a specific user by their ID
	r.Put("/users/:id", h.Update)    // Update a user's details
	r.Delete("/users/:id", h.Delete) // Delete a user
}

// ---- Create ----
// Create handles the POST /users request.
func (h *UserHandler) Create(c fiber.Ctx) error {
	var in domain.CreateUserInput

	// Bind the incoming JSON request body to our input struct.
	// Fiber v3 uses c.Bind().JSON() which is different from Fiber v2's BodyParser!
	if err := c.Bind().JSON(&in); err != nil {
		// If the JSON is malformed (like missing a comma), return a 400 Bad Request
		return respondError(c, fiber.StatusBadRequest, "INVALID_BODY", "malformed JSON")
	}

	// Call the service layer to create the user.
	// We pass c.Context() because it's a context.Context that Fiber provides!
	user, err := h.svc.Create(c.Context(), in)
	if err != nil {
		// Map the service error to the correct HTTP status code
		return mapServiceErr(c, err)
	}
	// Return 201 Created and the user data
	return respondJSON(c, fiber.StatusCreated, user)
}

// ---- List ----
// List handles GET /users request for listing all users.
func (h *UserHandler) List(c fiber.Ctx) error {
	// Get query parameters "page" and "page_size" from the URL.
	// They are strings, so we convert them to integers using strconv.Atoi.
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	// Call the service to list users
	users, err := h.svc.List(c.Context(), page, pageSize)
	if err != nil {
		return mapServiceErr(c, err)
	}
	// Return 200 OK and the list of users
	return respondJSON(c, fiber.StatusOK, users)
}

// ---- GetByID ----
// GetByID handles GET /users/:id to fetch a single user.
func (h *UserHandler) GetByID(c fiber.Ctx) error {
	// Parse the :id parameter from the URL. It must be a valid UUID.
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		// If it's not a valid UUID (e.g. /users/123), return a 400 Bad Request
		return respondError(c, fiber.StatusBadRequest, "INVALID_ID", "id must be a UUID")
	}

	// Call the service to get the user
	user, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		return mapServiceErr(c, err)
	}
	return respondJSON(c, fiber.StatusOK, user)
}

// ---- Update ----
// Update handles PUT /users/:id to update user fields.
func (h *UserHandler) Update(c fiber.Ctx) error {
	// Validate the ID from URL parameters
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respondError(c, fiber.StatusBadRequest, "INVALID_ID", "id must be a UUID")
	}

	var in domain.UpdateUserInput
	// Parse the updates from JSON body
	if err := c.Bind().JSON(&in); err != nil {
		return respondError(c, fiber.StatusBadRequest, "INVALID_BODY", "malformed JSON")
	}

	// Call the service to perform the update
	user, err := h.svc.Update(c.Context(), id, in)
	if err != nil {
		return mapServiceErr(c, err)
	}
	return respondJSON(c, fiber.StatusOK, user)
}

// ---- Delete ----
// Delete handles DELETE /users/:id to remove a user.
func (h *UserHandler) Delete(c fiber.Ctx) error {
	// Validate the ID from URL parameters
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return respondError(c, fiber.StatusBadRequest, "INVALID_ID", "id must be a UUID")
	}

	// Call the service to delete the user
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return mapServiceErr(c, err)
	}
	// Return 204 No Content because there's no data to send back after deletion!
	return respondJSON(c, fiber.StatusNoContent, nil)
}

// ---- Helpers ----

// respondJSON is a helper to write a success JSON response wrapper.
func respondJSON(c fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(envelope{Success: true, Data: data})
}

// respondError is a helper to write an error JSON response wrapper.
func respondError(c fiber.Ctx, status int, code, msg string) error {
	return c.Status(status).JSON(envelope{
		Success: false,
		Error:   &errBody{Code: code, Message: msg},
	})
}

// mapServiceErr translates our custom service/repository errors into proper HTTP status codes.
// This is nice so we don't leak database errors directly to our API clients!
func mapServiceErr(c fiber.Ctx, err error) error {
	var inv service.ErrInvalidInput
	switch {
	// If the user sent bad input data (validated in the service)
	case errors.As(err, &inv):
		return respondError(c, fiber.StatusBadRequest, "INVALID_INPUT", inv.Error())
	// If the user was not found in the database
	case errors.Is(err, repository.ErrUserNotFound):
		return respondError(c, fiber.StatusNotFound, "USER_NOT_FOUND", err.Error())
	// If the email is already taken by another user
	case errors.Is(err, repository.ErrEmailTaken):
		return respondError(c, fiber.StatusConflict, "EMAIL_TAKEN", err.Error())
	// Default case: something went wrong on our end (database connection lost, etc.)
	default:
		return respondError(c, fiber.StatusInternalServerError, "INTERNAL", "internal server error")
	}
}
