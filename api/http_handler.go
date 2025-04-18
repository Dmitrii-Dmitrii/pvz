package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"net/http"
	"pvz/internal/generated"
	"pvz/internal/models/custom_errors"
	"pvz/internal/models/user_model"
	"pvz/internal/services/product_service"
	"pvz/internal/services/pvz_service"
	"pvz/internal/services/reception_service"
	"pvz/internal/services/user_service"
)

type Handler struct {
	pvzService       pvz_service.IPvzService
	receptionService reception_service.IReceptionService
	productService   product_service.IProductService
	userService      user_service.IUserService
}

func NewHandler(pvzService pvz_service.IPvzService, receptionService reception_service.IReceptionService, productService product_service.IProductService, userService user_service.IUserService) *Handler {
	return &Handler{
		pvzService:       pvzService,
		receptionService: receptionService,
		productService:   productService,
		userService:      userService,
	}
}

func (h *Handler) PostDummyLogin(c *gin.Context) {
	var req generated.PostDummyLoginJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request to dummy login: " + err.Error()})
		return
	}

	roleDto := generated.UserRole(req.Role)
	if roleDto != generated.UserRoleEmployee && roleDto != generated.UserRoleModerator {
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
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Dummy login error: " + userErr.Error()})
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
}

func (h *Handler) PostLogin(c *gin.Context) {
	var req generated.PostLoginJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request to login: " + err.Error()})
		return
	}

	token, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request to login:" + userErr.Error()})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Login error" + err.Error()})
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
}

func (h *Handler) PostProducts(c *gin.Context) {
	var productReq generated.PostProductsJSONRequestBody
	if err := c.ShouldBindJSON(&productReq); err != nil {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to create product: " + err.Error()})
		return
	}

	productResp, err := h.productService.CreateProduct(c.Request.Context(), productReq.PvzId, productReq.Type)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to create product: " + userErr.Error()})
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Create product error" + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, productResp)
}

func (h *Handler) GetPvz(c *gin.Context, params generated.GetPvzParams) {
	pvzResp, err := h.pvzService.GetPvz(c.Request.Context(), params)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to get pvz: " + err.Error()})
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Get pvz error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, pvzResp)
}

func (h *Handler) PostPvz(c *gin.Context) {
	var pvzReq generated.PostPvzJSONRequestBody
	if err := c.ShouldBindJSON(&pvzReq); err != nil {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to create pvz: " + err.Error()})
		return
	}

	pvzResp, err := h.pvzService.CreatePvz(c.Request.Context(), pvzReq)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to create pvz: " + userErr.Error()})
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Create pvz error: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pvzResp)
}

func (h *Handler) PostPvzPvzIdCloseLastReception(c *gin.Context, pvzId openapi_types.UUID) {
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
}

func (h *Handler) PostPvzPvzIdDeleteLastProduct(c *gin.Context, pvzId openapi_types.UUID) {
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

	c.JSON(http.StatusOK, nil)
}

func (h *Handler) PostReceptions(c *gin.Context) {
	var pvzIdReq generated.PostReceptionsJSONRequestBody
	if err := c.ShouldBindJSON(&pvzIdReq); err != nil {
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
}

func (h *Handler) PostRegister(c *gin.Context) {
	var req generated.PostRegisterJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to register: " + err.Error()})
		return
	}

	roleDto := generated.UserRole(req.Role)
	if roleDto != generated.UserRoleEmployee && roleDto != generated.UserRoleModerator {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid role"})
		return
	}

	userResp, token, err := h.userService.Register(c.Request.Context(), req.Email, req.Password, roleDto)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to register: " + userErr.Error()})
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
}
