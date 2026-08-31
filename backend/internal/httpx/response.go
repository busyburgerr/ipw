// Package httpx contains HTTP-layer plumbing shared by all feature handlers:
// a consistent JSON envelope, error mapping, and middleware.
package httpx

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// APIError is the single error shape every endpoint returns.
type APIError struct {
	Code    string `json:"code"`    // stable machine-readable slug, e.g. "not_found"
	Message string `json:"message"` // human-readable, safe to show the user
}

type errorEnvelope struct {
	Error APIError `json:"error"`
}

// DomainError is returned by the service layer to signal a specific failure
// class. Handlers do not need to know HTTP status codes; ErrorHandler maps them.
type DomainError struct {
	Code    string
	Message string
	Status  int
}

func (e *DomainError) Error() string { return e.Message }

func NewDomainError(status int, code, message string) *DomainError {
	return &DomainError{Code: code, Message: message, Status: status}
}

// Common constructors.
func ErrBadRequest(msg string) *DomainError {
	return NewDomainError(fiber.StatusBadRequest, "bad_request", msg)
}
func ErrUnauthorized(msg string) *DomainError {
	return NewDomainError(fiber.StatusUnauthorized, "unauthorized", msg)
}
func ErrForbidden(msg string) *DomainError {
	return NewDomainError(fiber.StatusForbidden, "forbidden", msg)
}
func ErrNotFound(msg string) *DomainError {
	return NewDomainError(fiber.StatusNotFound, "not_found", msg)
}
func ErrConflict(msg string) *DomainError {
	return NewDomainError(fiber.StatusConflict, "conflict", msg)
}

// ErrorHandler is Fiber's central error handler. It never leaks internal detail
// in production: unknown errors become a generic 500.
func ErrorHandler(c *fiber.Ctx, err error) error {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return c.Status(domainErr.Status).JSON(errorEnvelope{APIError{domainErr.Code, domainErr.Message}})
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(errorEnvelope{APIError{"http_error", fiberErr.Message}})
	}

	return c.Status(fiber.StatusInternalServerError).
		JSON(errorEnvelope{APIError{"internal", "internal server error"}})
}

// OK writes a 200 with the payload as-is (already a DTO, never a GORM model).
func OK(c *fiber.Ctx, payload any) error {
	return c.Status(fiber.StatusOK).JSON(payload)
}

// Created writes a 201.
func Created(c *fiber.Ctx, payload any) error {
	return c.Status(fiber.StatusCreated).JSON(payload)
}
