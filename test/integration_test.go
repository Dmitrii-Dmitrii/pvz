package test

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"pvz/api"
	"pvz/internal/drivers/product_driver"
	"pvz/internal/drivers/pvz_driver"
	"pvz/internal/drivers/reception_driver"
	"pvz/internal/generated"
	"pvz/internal/services/product_service"
	"pvz/internal/services/pvz_service"
	"pvz/internal/services/reception_service"
	"pvz/test/drivers"
	"testing"
)

func TestIntegration(t *testing.T) {
	pool, cleanup := drivers.SetupPostgresContainer(t)
	defer cleanup()

	pvzDriver := pvz_driver.NewPvzDriver(pool)
	receptionDriver := reception_driver.NewReceptionDriver(pool)
	productDriver := product_driver.NewProductDriver(pool)

	pvzService := pvz_service.NewPvzService(pvzDriver)
	receptionService := reception_service.NewReceptionService(receptionDriver)
	productService := product_service.NewProductService(productDriver, receptionService)

	handler := api.NewHttpHandler(pvzService, receptionService, productService, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	pvzReq := generated.PostPvzJSONRequestBody{
		City: generated.СанктПетербург,
	}

	jsonData, _ := json.Marshal(pvzReq)

	req, _ := http.NewRequest("POST", "/pvz", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.POST("/pvz", func(c *gin.Context) {
		handler.PostPvz(c)
	})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var pvzResp generated.PVZ
	json.Unmarshal(w.Body.Bytes(), &pvzResp)
	assert.Equal(t, generated.СанктПетербург, pvzResp.City)

	pvzId := *pvzResp.Id

	receptionReq := generated.PostReceptionsJSONRequestBody{
		PvzId: pvzId,
	}

	jsonData, _ = json.Marshal(receptionReq)

	req, _ = http.NewRequest("POST", "/receptions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.POST("/receptions", func(c *gin.Context) {
		handler.PostReceptions(c)
	})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var receptionResp generated.Reception
	json.Unmarshal(w.Body.Bytes(), &receptionResp)
	assert.Equal(t, generated.InProgress, receptionResp.Status)

	router.POST("/products", func(c *gin.Context) {
		handler.PostProducts(c)
	})
	for i := 0; i < 50; i++ {
		productType := generated.PostProductsJSONBodyTypeОбувь
		if i%3 == 1 {
			productType = generated.PostProductsJSONBodyTypeОдежда
		} else if i%3 == 2 {
			productType = generated.PostProductsJSONBodyTypeЭлектроника
		}

		productReq := generated.PostProductsJSONRequestBody{
			PvzId: *pvzResp.Id,
			Type:  productType,
		}

		jsonData, _ = json.Marshal(productReq)

		req, _ = http.NewRequest("POST", "/products", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var productResp generated.Product
		json.Unmarshal(w.Body.Bytes(), &productResp)
		assert.Equal(t, *receptionResp.Id, productResp.ReceptionId)
		assert.Equal(t, generated.ProductType(productType), productResp.Type)
	}

	router.POST("/pvz/"+pvzId.String()+"/close-last-reception", func(c *gin.Context) {
		handler.PostPvzPvzIdCloseLastReception(c, pvzId)
	})

	req, _ = http.NewRequest("POST", "/pvz/"+pvzId.String()+"/close-last-reception", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response generated.Reception
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, receptionResp.Id, response.Id)
	assert.Equal(t, generated.Close, response.Status)
}
