package api

import (
	"context"
	"github.com/gin-gonic/gin"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"net/http"
	"pvz/internal/generated"
	"pvz/internal/services/products"
	"pvz/internal/services/pvzs"
	"pvz/internal/services/receptions"
)

type Handler struct {
	pvzService       pvzs.IPvzService
	receptionService receptions.IReceptionService
	productService   products.IProductService
}

func NewHandler(pvzService pvzs.IPvzService, receptionService receptions.IReceptionService, productService products.IProductService) *Handler {
	return &Handler{
		pvzService:       pvzService,
		receptionService: receptionService,
		productService:   productService,
	}
}

func (h *Handler) PostDummyLogin(c *gin.Context) {}

func (h *Handler) PostLogin(c *gin.Context) {}

func (h *Handler) PostProducts(c *gin.Context) {}

func (h *Handler) GetPvz(c *gin.Context, params generated.GetPvzParams) {
	pvzResp, err := h.pvzService.GetPvz(context.Background(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Get pvz error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, pvzResp)
}

func (h *Handler) PostPvz(c *gin.Context) {
	var pvzReq generated.PostPvzJSONRequestBody
	if err := c.ShouldBindJSON(&pvzReq); err != nil {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format: " + err.Error()})
		return
	}

	pvzResp, err := h.pvzService.CreatePvz(context.Background(), pvzReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, generated.Error{Message: "Create pvz error: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pvzResp)
}

func (h *Handler) PostPvzPvzIdCloseLastReception(c *gin.Context, pvzId openapi_types.UUID) {
	receptionResp, err := h.receptionService.CloseReception(context.Background(), pvzId)
	if err != nil {
		c.JSON(http.StatusBadRequest, generated.Error{Message: "Close reception error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, receptionResp)
}

func (h *Handler) PostPvzPvzIdDeleteLastProduct(c *gin.Context, pvzId openapi_types.UUID) {}

func (h *Handler) PostReceptions(c *gin.Context) {}

func (h *Handler) PostRegister(c *gin.Context) {}
