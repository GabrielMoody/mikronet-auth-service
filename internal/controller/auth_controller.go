package controller

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/GabrielMoody/mikronet-auth-service/internal/dto"
	"github.com/GabrielMoody/mikronet-auth-service/internal/middleware"
	"github.com/GabrielMoody/mikronet-auth-service/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type AuthController interface {
	CreateUser(c *fiber.Ctx) error
	CreateDriver(c *fiber.Ctx) error
	LoginUser(c *fiber.Ctx) error
	SendResetPasswordLink(c *fiber.Ctx) error
	ResetPassword(c *fiber.Ctx) error
	ChangePassword(c *fiber.Ctx) error
	ResetPasswordUI(c *fiber.Ctx) error
}

type AuthControllerImpl struct {
	AuthService service.AuthService
}

func readImage(image *multipart.FileHeader) ([]byte, error) {
	f, err := image.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fileData, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return fileData, nil
}

func (a *AuthControllerImpl) ChangePassword(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	payload, _ := middleware.GetJWTPayload(token[7:], os.Getenv("JWT_SECRET"))

	ctx := c.Context()
	var user dto.ChangePasswordReq

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error",
			"errors": err.Error(),
		})
	}

	_, errService := a.AuthService.ChangePasswordService(ctx, payload["id"].(string), user)

	if errService != nil && errService.ValidationErrors != nil {
		return c.Status(errService.Code).JSON(fiber.Map{
			"status": "error",
			"errors": errService.ValidationErrors,
		})
	}

	if errService != nil && errService.Err != nil {
		return c.Status(errService.Code).JSON(fiber.Map{
			"status": "error",
			"errors": errService.Err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Berhasil mengubah password!",
	})
}

func (a *AuthControllerImpl) CreateUser(c *fiber.Ctx) error {
	var user dto.UserRegistrationsReq
	ctx := c.Context()

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error",
			"errors": err.Error(),
		})
	}

	_, errService := a.AuthService.CreateUserService(ctx, user, "user")

	if errService != nil && errService.ValidationErrors != nil {
		return c.Status(errService.Code).JSON(fiber.Map{
			"status": "error",
			"errors": errService.ValidationErrors,
		})
	}

	if errService != nil && errService.Err != nil {
		return c.Status(errService.Code).JSON(fiber.Map{
			"status": "error",
			"errors": errService.Err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "akun berhasil dibuat!",
	})
}

func (a *AuthControllerImpl) CreateDriver(c *fiber.Ctx) error {
	var driver dto.DriverRegistrationsReq
	ctx := c.Context()
	pp, err := c.FormFile("profile_picture")
	ktp, err := c.FormFile("ktp")

	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error",
			"errors": "error reading image",
		})
	}

	if err := c.BodyParser(&driver); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error",
			"errors": err.Error(),
		})
	}

	fileDataPP, err := readImage(pp)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "error",
			"errors": err.Error(),
		})
	}

	fileDataKtp, err := readImage(ktp)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "error",
			"errors": err.Error(),
		})
	}

	_, errService := a.AuthService.CreateDriverService(ctx, driver, "driver", fileDataPP, fileDataKtp)

	if errService != nil && errService.ValidationErrors != nil {
		return c.Status(errService.Code).JSON(fiber.Map{
			"status": "error",
			"errors": errService.ValidationErrors,
		})
	}

	if errService != nil && errService.Err != nil {
		return c.Status(errService.Code).JSON(fiber.Map{
			"status": "error",
			"errors": errService.Err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "akun berhasil dibuat!",
	})
}

func (a *AuthControllerImpl) LoginUser(c *fiber.Ctx) error {
	ctx := c.Context()
	var user dto.UserLoginReq

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error",
			"errors": err.Error(),
		})
	}

	res, errService := a.AuthService.LoginUserService(ctx, user)

	if errService != nil && errService.ValidationErrors != nil {
		return c.Status(errService.Code).JSON(fiber.Map{
			"status": "error",
			"errors": errService.ValidationErrors,
		})
	}

	if errService != nil && errService.Err != nil {
		return c.Status(errService.Code).JSON(fiber.Map{
			"status": "error",
			"errors": errService.Err.Error(),
		})
	}

	claims := jwt.MapClaims{
		"id":    res.ID,
		"email": res.Email,
		"role":  res.Role,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
		"iss":   os.Getenv("JWT_ISS"),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, errToken := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if errToken != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status": "error",
			"errors": "Invalid token",
		})
	}

	refreshClaims := jwt.MapClaims{
		"id":  res.ID,
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
		"iss": os.Getenv("JWT_ISS"),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	rt, errRefreshToken := refreshToken.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if errRefreshToken != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status": "error",
			"errors": "Invalid refresh token",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"access_token":  t,
			"refresh_token": rt,
		},
	})
}

func (a *AuthControllerImpl) SendResetPasswordLink(c *fiber.Ctx) error {
	var email dto.ForgotPasswordReq
	ctx := c.Context()

	if err := c.BodyParser(&email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error",
			"errors": err.Error(),
		})
	}

	res, err := a.AuthService.SendResetPasswordService(ctx, email)

	if err != nil {
		return c.Status(err.Code).JSON(fiber.Map{
			"status": "error",
			"errors": err.Err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": res,
	})
}

func (a *AuthControllerImpl) ResetPassword(c *fiber.Ctx) error {
	code := c.Params("code")
	var rp dto.ResetPasswordReq
	ctx := c.Context()

	if err := c.BodyParser(&rp); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error",
			"errors": err.Error(),
		})
	}

	_, errService := a.AuthService.ResetPassword(ctx, rp, code)

	if errService != nil && errService.ValidationErrors != nil {
		return c.Status(errService.Code).JSON(fiber.Map{
			"status": "error",
			"errors": errService.ValidationErrors,
		})
	}

	if errService != nil && errService.Err != nil {
		return c.Status(errService.Code).JSON(fiber.Map{
			"status": "error",
			"errors": errService.Err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Berhasil melakukan reset password!",
	})
}

func (a *AuthControllerImpl) ResetPasswordUI(c *fiber.Ctx) error {
	return c.SendFile("./views/reset_password.html")
}

func NewAuthController(authService service.AuthService) AuthController {
	return &AuthControllerImpl{
		AuthService: authService,
	}
}
