package api

import (
	"errors"
	"github.com/Dmitrii-Dmitrii/pvz/internal/generated"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/custom_errors"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/user_model"
	"github.com/Dmitrii-Dmitrii/pvz/internal/services/product_service"
	"github.com/Dmitrii-Dmitrii/pvz/internal/services/pvz_service"
	"github.com/Dmitrii-Dmitrii/pvz/internal/services/reception_service"
	"github.com/Dmitrii-Dmitrii/pvz/internal/services/user_service"
	"github.com/gin-gonic/gin"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog/log"
	"net/http"
)

type HttpHandler struct {
	pvzService       pvz_service.IPvzService
	receptionService reception_service.IReceptionService
	productService   product_service.IProductService
	userService      user_service.IUserService
}

func NewHttpHandler(pvzService pvz_service.IPvzService, receptionService reception_service.IReceptionService, productService product_service.IProductService, userService user_service.IUserService) *HttpHandler {
	return &HttpHandler{
		pvzService:       pvzService,
		receptionService: receptionService,
		productService:   productService,
		userService:      userService,
	}
}

func (h *HttpHandler) PostDummyLogin(c *gin.Context) {
	log.Info().Msg("dummy login started")

	var req generated.PostDummyLoginJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("failed to bind json body")
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request to dummy login: " + err.Error()})
		return
	}

	roleDto := generated.UserRole(req.Role)
	if roleDto != generated.UserRoleEmployee && roleDto != generated.UserRoleModerator {
		log.Error().Msg("invalid role")
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid role"})
		return
	}

	token, err := h.userService.DummyLogin(c.Request.Context(), roleDto)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request to dummy login: " + userErr.Error()})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Dummy login error: " + err.Error()})
		return
	}

	c.SetCookie(
		"auth_token",
		token,
		int(user_model.JwtExpiry.Seconds()),
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, token)
	log.Info().Msgf("dummy login result: %s", token)
}

func (h *HttpHandler) PostLogin(c *gin.Context) {
	log.Info().Msg("login started")

	var req generated.PostLoginJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("failed to bind json body")
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request to login: " + err.Error()})
		return
	}

	token, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		c.JSON(http.StatusUnauthorized, generated.Error{Message: "Invalid request to login:" + userErr.Error()})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Login error: " + err.Error()})
		return
	}

	c.SetCookie(
		"auth_token",
		token,
		int(user_model.JwtExpiry.Seconds()),
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, token)

	log.Info().Msgf("login result: %s", token)
}

func (h *HttpHandler) PostProducts(c *gin.Context) {
	log.Info().Msg("products started")

	var productReq generated.PostProductsJSONRequestBody
	if err := c.ShouldBindJSON(&productReq); err != nil {
		log.Error().Err(err).Msg("failed to bind json body")
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to create product: " + err.Error()})
		return
	}

	productResp, err := h.productService.CreateProduct(c.Request.Context(), productReq.PvzId, productReq.Type)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to create product: " + userErr.Error()})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Create product error: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, productResp)

	log.Info().Msgf("products result: %s", productResp)
}

func (h *HttpHandler) GetPvz(c *gin.Context, params generated.GetPvzParams) {
	log.Info().Msg("get pvz started")

	pvzResp, err := h.pvzService.GetPvzFullInfo(c.Request.Context(), params)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to get pvz: " + err.Error()})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Get pvz error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, pvzResp)

	log.Info().Msgf("pvz result: %s", pvzResp)
}

func (h *HttpHandler) PostPvz(c *gin.Context) {
	log.Info().Msg("pvz started")

	var pvzReq generated.PostPvzJSONRequestBody
	if err := c.ShouldBindJSON(&pvzReq); err != nil {
		log.Error().Err(err).Msg("failed to bind json body")
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to create pvz: " + err.Error()})
		return
	}

	pvzResp, err := h.pvzService.CreatePvz(c.Request.Context(), pvzReq)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to create pvz: " + userErr.Error()})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Create pvz error: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pvzResp)

	log.Info().Msgf("pvz result: %s", pvzResp)
}

func (h *HttpHandler) PostPvzPvzIdCloseLastReception(c *gin.Context, pvzId openapi_types.UUID) {
	log.Info().Msg("close last reception started")

	receptionResp, err := h.receptionService.CloseReception(c.Request.Context(), pvzId)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to close last reception error: " + userErr.Error()})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Close last reception error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, receptionResp)

	log.Info().Msgf("close last reception result: %s", receptionResp)
}

func (h *HttpHandler) PostPvzPvzIdDeleteLastProduct(c *gin.Context, pvzId openapi_types.UUID) {
	log.Info().Msg("delete last product started")

	err := h.productService.DeleteLastProduct(c.Request.Context(), pvzId)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to delete last product error: " + userErr.Error()})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Delete last reception error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
	log.Info().Msg("delete last product finished")
}

func (h *HttpHandler) PostReceptions(c *gin.Context) {
	log.Info().Msg("receptions started")

	var pvzIdReq generated.PostReceptionsJSONRequestBody
	if err := c.ShouldBindJSON(&pvzIdReq); err != nil {
		log.Error().Err(err).Msg("failed to bind json body")
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to create reception: " + err.Error()})
		return
	}

	receptionResp, err := h.receptionService.CreateReception(c.Request.Context(), pvzIdReq.PvzId)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to create reception: " + userErr.Error()})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Create reception error: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, receptionResp)

	log.Info().Msgf("receptions result: %s", receptionResp)
}

func (h *HttpHandler) PostRegister(c *gin.Context) {
	log.Info().Msg("register started")

	var req generated.PostRegisterJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Msg("failed to bind json body")
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to register: " + err.Error()})
		return
	}

	roleDto := generated.UserRole(req.Role)
	if roleDto != generated.UserRoleEmployee && roleDto != generated.UserRoleModerator {
		log.Error().Msg("invalid role")
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid role"})
		return
	}

	userResp, token, err := h.userService.Register(c.Request.Context(), req.Email, req.Password, roleDto)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to register: " + userErr.Error()})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Register error: " + err.Error()})
		return
	}

	c.SetCookie(
		"auth_token",
		token,
		int(user_model.JwtExpiry.Seconds()),
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusCreated, userResp)

	log.Info().Msgf("register result: %s", userResp)
}
