package http

import (
	"errors"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// ErrorHandler converts use-case errors and panics into Laravel-shaped JSON envelopes.
func ErrorHandler(log *logrus.Logger, debug bool) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// ValidationError → 422 with field-keyed errors
		var ve *usecase.ValidationError
		if errors.As(err, &ve) {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"success": false,
				"code":    422,
				"message": "The given data was invalid.",
				"errors":  ve.Fields,
			})
		}

		switch {
		case errors.Is(err, usecase.ErrNotFound):
			return c.Status(fiber.StatusNotFound).JSON(envelope(404, "Not Found"))
		case errors.Is(err, usecase.ErrUnauthorized):
			return c.Status(fiber.StatusUnauthorized).JSON(envelope(401, "Unauthenticated."))
		case errors.Is(err, usecase.ErrForbidden):
			return c.Status(fiber.StatusForbidden).JSON(envelope(403, "This action is unauthorized."))
		case errors.Is(err, usecase.ErrInvalidCredentials):
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"success": false,
				"code":    422,
				"message": "The given data was invalid.",
				"errors":  fiber.Map{"email": []string{"These credentials do not match our records."}},
			})
		}

		var fe *fiber.Error
		if errors.As(err, &fe) {
			return c.Status(fe.Code).JSON(envelope(fe.Code, fe.Message))
		}

		log.WithError(err).Error("unhandled error")
		body := envelope(500, "Server Error")
		if debug {
			body["debug"] = err.Error()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(body)
	}
}

func envelope(code int, msg string) fiber.Map {
	return fiber.Map{"success": false, "code": code, "message": msg}
}
