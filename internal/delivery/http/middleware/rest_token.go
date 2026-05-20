package middleware

import (
	"strings"
	"time"

	"github.com/Matrosovdream/formidable-storage-app-golang/internal/repository"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

// RestToken validates a bearer token against site_tokens and resolves the site into locals.
func RestToken(sites *repository.SiteRepository, siteTokens *repository.SiteTokenRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := strings.TrimSpace(c.Get("Authorization"))
		if !strings.HasPrefix(raw, "Bearer ") {
			return usecase.ErrUnauthorized
		}
		token := strings.TrimPrefix(raw, "Bearer ")
		st, err := siteTokens.FindByToken(c.Context(), nil, token)
		if err != nil {
			return err
		}
		if st == nil {
			return usecase.ErrUnauthorized
		}
		if st.ValidUntil.Valid && st.ValidUntil.Time.Before(time.Now()) {
			return usecase.ErrUnauthorized
		}
		s, err := sites.FindByID(c.Context(), nil, st.SiteID)
		if err != nil {
			return err
		}
		if s == nil {
			return usecase.ErrUnauthorized
		}
		setSite(c, s)
		return c.Next()
	}
}
