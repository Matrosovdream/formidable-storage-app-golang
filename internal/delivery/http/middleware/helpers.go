package middleware

import (
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/gofiber/fiber/v2"
)

const (
	ctxKeyUser  = "fsa.user"
	ctxKeyToken = "fsa.token"
	ctxKeySite  = "fsa.site"
)

func UserFromCtx(c *fiber.Ctx) *entity.User {
	v := c.Locals(ctxKeyUser)
	if v == nil {
		return nil
	}
	return v.(*entity.User)
}

func TokenFromCtx(c *fiber.Ctx) *entity.PersonalAccessToken {
	v := c.Locals(ctxKeyToken)
	if v == nil {
		return nil
	}
	return v.(*entity.PersonalAccessToken)
}

func SiteFromCtx(c *fiber.Ctx) *entity.Site {
	v := c.Locals(ctxKeySite)
	if v == nil {
		return nil
	}
	return v.(*entity.Site)
}

func setUser(c *fiber.Ctx, u *entity.User)             { c.Locals(ctxKeyUser, u) }
func setToken(c *fiber.Ctx, t *entity.PersonalAccessToken) { c.Locals(ctxKeyToken, t) }
func setSite(c *fiber.Ctx, s *entity.Site)             { c.Locals(ctxKeySite, s) }
