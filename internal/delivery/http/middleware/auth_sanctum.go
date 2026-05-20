package middleware

import (
	"strings"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

// AuthSanctum enforces a bearer token issued via AuthUseCase.Register/Login.
// If the token is valid, the user + token row are stored in fiber.Locals.
func AuthSanctum(auth *usecase.AuthUseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := strings.TrimSpace(c.Get("Authorization"))
		if !strings.HasPrefix(raw, "Bearer ") {
			return usecase.ErrUnauthorized
		}
		token, user, err := auth.VerifyToken(c.Context(), strings.TrimPrefix(raw, "Bearer "))
		if err != nil {
			return err
		}
		setUser(c, user)
		setToken(c, token)
		return c.Next()
	}
}

// OptionalAuthSanctum sets the user/token in locals only if a valid token is supplied.
// Used by the GET /api/user endpoint, which mirrors Laravel's `fn ($r) => $r->user()`.
func OptionalAuthSanctum(auth *usecase.AuthUseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := strings.TrimSpace(c.Get("Authorization"))
		if strings.HasPrefix(raw, "Bearer ") {
			token, user, err := auth.VerifyToken(c.Context(), strings.TrimPrefix(raw, "Bearer "))
			if err == nil {
				setUser(c, user)
				setToken(c, token)
			}
		}
		return c.Next()
	}
}
