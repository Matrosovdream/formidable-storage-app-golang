package http

import (
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/delivery/http/middleware"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model/converter"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type AuthController struct {
	auth *usecase.AuthUseCase
}

func NewAuthController(auth *usecase.AuthUseCase) *AuthController {
	return &AuthController{auth: auth}
}

func (h *AuthController) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return usecase.NewValidationError("_", "Invalid JSON body")
	}
	resp, err := h.auth.Login(c.Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(resp)
}

func (h *AuthController) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return usecase.NewValidationError("_", "Invalid JSON body")
	}
	resp, err := h.auth.Register(c.Context(), req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *AuthController) Logout(c *fiber.Ctx) error {
	tok := middleware.TokenFromCtx(c)
	if tok == nil {
		return usecase.ErrUnauthorized
	}
	if err := h.auth.Logout(c.Context(), tok.ID); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "Logged out"})
}

// Me mirrors Laravel's `Route::get('/user', fn ($r) => $r->user())`:
// returns the user JSON if authenticated, else literal null.
func (h *AuthController) Me(c *fiber.Ctx) error {
	u := middleware.UserFromCtx(c)
	if u == nil {
		return c.JSON(nil)
	}
	return c.JSON(converter.ToUserResponse(u))
}
