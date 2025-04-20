package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"net/http/httptest"
	"pvz/api"
	"pvz/internal/models/custom_errors"
	"pvz/internal/models/reception_model"
	"pvz/internal/models/user_model"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pvz/internal/generated"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) DummyLogin(ctx context.Context, roleDto generated.UserRole) (string, error) {
	args := m.Called(ctx, roleDto)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

func (m *MockUserService) Register(ctx context.Context, emailDto openapi_types.Email, password string, roleDto generated.UserRole) (*generated.User, string, error) {
	args := m.Called(ctx, emailDto, password, roleDto)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*generated.User), args.String(1), args.Error(2)
}

func (m *MockUserService) Login(ctx context.Context, emailDto openapi_types.Email, password string) (string, error) {
	args := m.Called(ctx, emailDto, password)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

func (m *MockUserService) ValidateToken(ctx context.Context, token string) (*user_model.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user_model.User), args.Error(1)
}

type MockPvzService struct {
	mock.Mock
}

func (m *MockPvzService) CreatePvz(ctx context.Context, pvzDto generated.PVZ) (*generated.PVZ, error) {
	args := m.Called(ctx, pvzDto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*generated.PVZ), args.Error(1)
}

func (m *MockPvzService) GetPvz(ctx context.Context, params generated.GetPvzParams) ([]map[string]interface{}, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

type MockReceptionService struct {
	mock.Mock
}

func (m *MockReceptionService) CreateReception(ctx context.Context, pvzIdDto openapi_types.UUID) (*generated.Reception, error) {
	args := m.Called(ctx, pvzIdDto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*generated.Reception), args.Error(1)
}

func (m *MockReceptionService) CloseReception(ctx context.Context, pvzIdDto openapi_types.UUID) (*generated.Reception, error) {
	args := m.Called(ctx, pvzIdDto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*generated.Reception), args.Error(1)
}

func (m *MockReceptionService) GetLastReceptionStatus(ctx context.Context, pvzId pgtype.UUID) (*reception_model.ReceptionStatus, error) {
	args := m.Called(ctx, pvzId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception_model.ReceptionStatus), args.Error(1)
}

type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) CreateProduct(ctx context.Context, pvzIdDto openapi_types.UUID, productTypeJson generated.PostProductsJSONBodyType) (*generated.Product, error) {
	args := m.Called(ctx, pvzIdDto, productTypeJson)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*generated.Product), args.Error(1)
}

func (m *MockProductService) DeleteLastProduct(ctx context.Context, pvzIdDto openapi_types.UUID) error {
	args := m.Called(ctx, pvzIdDto)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func setupTestEnv() (*gin.Engine, *MockUserService, *MockPvzService, *MockProductService, *MockReceptionService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockUserService := new(MockUserService)
	mockPvzService := new(MockPvzService)
	mockProductService := new(MockProductService)
	mockReceptionService := new(MockReceptionService)

	return router, mockUserService, mockPvzService, mockProductService, mockReceptionService
}

func TestPostDummyLogin(t *testing.T) {
	t.Run("Dummy login", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		loginReq := generated.PostDummyLoginJSONRequestBody{
			Role: "employee",
		}
		jsonData, _ := json.Marshal(loginReq)

		mockUserService.On("DummyLogin", mock.Anything, generated.UserRoleEmployee).Return("valid_token", nil).Once()

		req, _ := http.NewRequest("POST", "/dummyLogin", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/dummyLogin", handler.PostDummyLogin)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUserService.AssertExpectations(t)

		cookies := w.Result().Cookies()
		found := false
		for _, cookie := range cookies {
			if cookie.Name == "auth_token" {
				found = true
				assert.Equal(t, "valid_token", cookie.Value)
				assert.Equal(t, "/", cookie.Path)
				assert.True(t, cookie.HttpOnly)
			}
		}
		assert.True(t, found)
	})

	t.Run("Dummy login with invalid role", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		loginReq := generated.PostDummyLoginJSONRequestBody{
			Role: "invalid role",
		}
		jsonData, _ := json.Marshal(loginReq)

		req, _ := http.NewRequest("POST", "/dummyLogin", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/dummyLogin", handler.PostDummyLogin)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid role", response.Message)
	})

	t.Run("Dummy login with user error", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		loginReq := generated.PostDummyLoginJSONRequestBody{
			Role: "employee",
		}
		jsonData, _ := json.Marshal(loginReq)

		userError := custom_errors.UserError{Message: "user error"}
		mockUserService.On("DummyLogin", mock.Anything, generated.UserRoleEmployee).Return("", &userError).Once()

		req, _ := http.NewRequest("POST", "/dummyLogin", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/dummyLogin", handler.PostDummyLogin)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Invalid request to dummy login")
	})

	t.Run("Dummy login with internal error", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		loginReq := generated.PostDummyLoginJSONRequestBody{
			Role: "employee",
		}
		jsonData, _ := json.Marshal(loginReq)

		mockUserService.On("DummyLogin", mock.Anything, generated.UserRoleEmployee).Return("", errors.New("internal error")).Once()

		req, _ := http.NewRequest("POST", "/dummyLogin", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/dummyLogin", handler.PostDummyLogin)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Dummy login error")
	})
}

func TestPostLogin(t *testing.T) {
	t.Run("Login", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		loginReq := generated.PostLoginJSONRequestBody{
			Email:    "test@example.com",
			Password: "password123",
		}
		jsonData, _ := json.Marshal(loginReq)

		mockUserService.On("Login", mock.Anything, openapi_types.Email("test@example.com"), "password123").Return("valid_token", nil).Once()

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/login", func(c *gin.Context) {
			handler.PostLogin(c)
		})
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUserService.AssertExpectations(t)

		cookies := w.Result().Cookies()
		found := false
		for _, cookie := range cookies {
			if cookie.Name == "auth_token" {
				found = true
				assert.Equal(t, "valid_token", cookie.Value)
			}
		}
		assert.True(t, found)
	})

	t.Run("Login with wrong password", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		loginReq := generated.PostLoginJSONRequestBody{
			Email:    "test@example.com",
			Password: "wrong_password",
		}
		jsonData, _ := json.Marshal(loginReq)

		userError := custom_errors.UserError{Message: "invalid credentials"}
		mockUserService.On("Login", mock.Anything, openapi_types.Email("test@example.com"), "wrong_password").Return("", &userError).Once()

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/login", func(c *gin.Context) {
			handler.PostLogin(c)
		})
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Invalid request to login")
	})

	t.Run("Post Login with invalid email", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		loginReq := generated.PostLoginJSONRequestBody{
			Email:    "testexample.com",
			Password: "wrong_password",
		}
		jsonData, _ := json.Marshal(loginReq)

		userError := custom_errors.UserError{Message: "invalid credentials"}
		mockUserService.On("Login", mock.Anything, openapi_types.Email("testexample.com"), "wrong_password").Return("", &userError).Once()

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/login", func(c *gin.Context) {
			handler.PostLogin(c)
		})
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Invalid request to login")
	})

	t.Run("Post Login with internal err0r", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		loginReq := generated.PostLoginJSONRequestBody{
			Email:    "test@example.com",
			Password: "password123",
		}
		jsonData, _ := json.Marshal(loginReq)

		internalError := errors.New("internal error")
		mockUserService.On("Login", mock.Anything, openapi_types.Email("test@example.com"), "password123").Return("", internalError).Once()

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/login", func(c *gin.Context) {
			handler.PostLogin(c)
		})
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Login error")
	})
}

func TestPostProducts(t *testing.T) {
	t.Run("Create product in reception in progress", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)
		pvzId := uuid.New()
		productReq := generated.PostProductsJSONRequestBody{
			PvzId: pvzId,
			Type:  generated.PostProductsJSONBodyTypeОбувь,
		}

		jsonData, _ := json.Marshal(productReq)
		productId := uuid.New()
		dateTime := time.Now().Add(-time.Hour)
		receptionId := uuid.New()

		mockProductService.On("CreateProduct", mock.Anything, pvzId, generated.PostProductsJSONBodyTypeОбувь).Return(&generated.Product{
			Id:          &productId,
			DateTime:    &dateTime,
			ReceptionId: receptionId,
			Type:        generated.ProductTypeОбувь,
		}, nil).Once()

		req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/products", func(c *gin.Context) {
			handler.PostProducts(c)
		})
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockProductService.AssertExpectations(t)

		var response generated.Product
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, &productId, response.Id)
		assert.Equal(t, receptionId, response.ReceptionId)
		assert.Equal(t, generated.ProductTypeОбувь, response.Type)
	})

	t.Run("Create product with invalid type", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		pvzId := uuid.New()
		productReq := generated.PostProductsJSONRequestBody{
			PvzId: pvzId,
			Type:  generated.PostProductsJSONBodyType("InvalidType"),
		}
		jsonData, _ := json.Marshal(productReq)

		userError := custom_errors.UserError{Message: "invalid credentials"}
		mockProductService.On("CreateProduct", mock.Anything, pvzId, generated.PostProductsJSONBodyType("InvalidType")).Return(&generated.Product{}, &userError).Once()

		req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/products", func(c *gin.Context) {
			handler.PostProducts(c)
		})
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Invalid request format to create product")
	})

	t.Run("Post products with internal error", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)
		pvzId := uuid.New()
		productReq := generated.PostProductsJSONRequestBody{
			PvzId: pvzId,
			Type:  generated.PostProductsJSONBodyTypeОбувь,
		}
		jsonData, _ := json.Marshal(productReq)

		internalError := errors.New("internal error")
		mockProductService.On("CreateProduct", mock.Anything, pvzId, generated.PostProductsJSONBodyTypeОбувь).Return(nil, internalError).Once()

		req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/products", func(c *gin.Context) {
			handler.PostProducts(c)
		})
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Create product error")
	})
}

func TestGetPvz(t *testing.T) {
	t.Run("Get pvz with default params", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()

		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)
		expectedResp := []map[string]interface{}{
			{"id": "1", "name": "ПВЗ 1", "address": "Адрес 1"},
			{"id": "2", "name": "ПВЗ 2", "address": "Адрес 2"},
		}

		mockPvzService.On("GetPvz", mock.Anything, generated.GetPvzParams{}).
			Return(expectedResp, nil).Once()

		req, _ := http.NewRequest("GET", "/pvz", nil)
		w := httptest.NewRecorder()

		router.GET("/pvz", func(c *gin.Context) {
			handler.GetPvz(c, generated.GetPvzParams{})
		})
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockPvzService.AssertExpectations(t)

		var response []map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, expectedResp, response)
	})

	t.Run("Get pvz with pagination and date range", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)
		page := 2
		limit := 15

		expectedResp := []map[string]interface{}{
			{"id": "3", "name": "ПВЗ 3", "address": "Адрес 3"},
			{"id": "4", "name": "ПВЗ 4", "address": "Адрес 4"},
		}

		params := generated.GetPvzParams{
			StartDate: &startDate,
			EndDate:   &endDate,
			Page:      &page,
			Limit:     &limit,
		}

		mockPvzService.On("GetPvz", mock.Anything, params).
			Return(expectedResp, nil).Once()

		req, _ := http.NewRequest("GET", "/pvz?startDate=2023-01-01T00:00:00Z&endDate=2023-12-31T23:59:59Z&page=2&limit=15", nil)
		w := httptest.NewRecorder()

		router.GET("/pvz", func(c *gin.Context) {
			c.Request = req
			handler.GetPvz(c, params)
		})
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockPvzService.AssertExpectations(t)

		var response []map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, expectedResp, response)
	})

	t.Run("Get pvz with invalid date range", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		startDate := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		params := generated.GetPvzParams{
			StartDate: &startDate,
			EndDate:   &endDate,
		}

		userErr := custom_errors.ErrDateRange
		mockPvzService.On("GetPvz", mock.Anything, params).
			Return(nil, userErr).Once()

		req, _ := http.NewRequest("GET", "/pvz?startDate=2023-12-31T00:00:00Z&endDate=2023-01-01T00:00:00Z", nil)
		w := httptest.NewRecorder()

		router.GET("/pvz", func(c *gin.Context) {
			c.Request = req
			handler.GetPvz(c, params)
		})
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockPvzService.AssertExpectations(t)

		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Invalid request format to get pvz")
	})

	t.Run("Get pvz with invalid limit", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		limit := 50

		params := generated.GetPvzParams{
			Limit: &limit,
		}

		userErr := custom_errors.ErrLimitValue
		mockPvzService.On("GetPvz", mock.Anything, params).
			Return(nil, userErr).Once()

		req, _ := http.NewRequest("GET", "/pvz?limit=50", nil)
		w := httptest.NewRecorder()

		router.GET("/pvz", func(c *gin.Context) {
			c.Request = req
			handler.GetPvz(c, params)
		})
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockPvzService.AssertExpectations(t)

		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Invalid request format to get pvz")
	})

	t.Run("Get pvz with invalid page", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		page := 0

		params := generated.GetPvzParams{
			Page: &page,
		}

		userErr := custom_errors.ErrPageValue
		mockPvzService.On("GetPvz", mock.Anything, params).
			Return(nil, userErr).Once()

		req, _ := http.NewRequest("GET", "/pvz?page=0", nil)
		w := httptest.NewRecorder()

		router.GET("/pvz", func(c *gin.Context) {
			c.Request = req
			handler.GetPvz(c, params)
		})
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockPvzService.AssertExpectations(t)

		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Invalid request format to get pvz")
	})

	t.Run("Get pvz with internal error", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		internalErr := errors.New("database connection error")
		mockPvzService.On("GetPvz", mock.Anything, generated.GetPvzParams{}).
			Return(nil, internalErr).Once()

		req, _ := http.NewRequest("GET", "/pvz", nil)
		w := httptest.NewRecorder()

		router.GET("/pvz", func(c *gin.Context) {
			handler.GetPvz(c, generated.GetPvzParams{})
		})
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockPvzService.AssertExpectations(t)

		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Get pvz error")
		assert.Contains(t, response.Message, internalErr.Error())
	})
}

func TestPostPvz(t *testing.T) {
	t.Run("Create pvz", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		pvzReq := generated.PostPvzJSONRequestBody{
			City: generated.СанктПетербург,
		}
		jsonData, _ := json.Marshal(pvzReq)
		pvzId := uuid.New()
		registrationDate := time.Now()

		mockPvzService.On("CreatePvz", mock.Anything, pvzReq).Return(&generated.PVZ{
			Id:               &pvzId,
			City:             generated.СанктПетербург,
			RegistrationDate: &registrationDate,
		}, nil).Once()

		req, _ := http.NewRequest("POST", "/pvz", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/pvz", handler.PostPvz)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockPvzService.AssertExpectations(t)

		var response generated.PVZ
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, &pvzId, response.Id)
		assert.Equal(t, generated.СанктПетербург, response.City)
	})

	t.Run("Create pvz with user error", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		pvzReq := generated.PostPvzJSONRequestBody{
			City: generated.PVZCity(""),
		}
		jsonData, _ := json.Marshal(pvzReq)

		userErr := custom_errors.UserError{}
		mockPvzService.On("CreatePvz", mock.Anything, pvzReq).Return(nil, &userErr).Once()

		req, _ := http.NewRequest("POST", "/pvz", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/pvz", handler.PostPvz)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Invalid request format to create pvz")
	})

	t.Run("Create pvz with internal error", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		pvzReq := generated.PostPvzJSONRequestBody{
			City: generated.СанктПетербург,
		}
		jsonData, _ := json.Marshal(pvzReq)

		internalErr := errors.New("database connection error")
		mockPvzService.On("CreatePvz", mock.Anything, pvzReq).Return(nil, internalErr).Once()

		req, _ := http.NewRequest("POST", "/pvz", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/pvz", handler.PostPvz)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Create pvz error")
	})
}

func TestPostPvzPvzIdCloseLastReception(t *testing.T) {
	t.Run("Close last reception", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		pvzId := uuid.New()
		receptionId := uuid.New()

		dateTime := time.Now()
		mockReceptionService.On("CloseReception", mock.Anything, pvzId).Return(&generated.Reception{
			Id:       &receptionId,
			PvzId:    pvzId,
			Status:   generated.Close,
			DateTime: dateTime,
		}, nil).Once()

		router.POST("/pvz/"+pvzId.String()+"/close-last-reception", func(c *gin.Context) {
			handler.PostPvzPvzIdCloseLastReception(c, pvzId)
		})

		req, _ := http.NewRequest("POST", "/pvz/"+pvzId.String()+"/close-last-reception", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockReceptionService.AssertExpectations(t)

		var response generated.Reception
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, &receptionId, response.Id)
		assert.Equal(t, generated.Close, response.Status)
	})

	t.Run("Close last reception with user error", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		pvzId := uuid.New()

		userErr := custom_errors.UserError{}
		mockReceptionService.On("CloseReception", mock.Anything, pvzId).Return(nil, &userErr).Once()

		router.POST("/pvz/"+pvzId.String()+"/close-last-reception", func(c *gin.Context) {
			handler.PostPvzPvzIdCloseLastReception(c, pvzId)
		})

		req, _ := http.NewRequest("POST", "/pvz/"+pvzId.String()+"/close-last-reception", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Invalid request format to close last reception error")
	})

	t.Run("Close last reception with internal error", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		pvzId := uuid.New()

		internalErr := errors.New("internal error")
		mockReceptionService.On("CloseReception", mock.Anything, pvzId).Return(nil, internalErr).Once()

		router.POST("/pvz/"+pvzId.String()+"/close-last-reception", func(c *gin.Context) {
			handler.PostPvzPvzIdCloseLastReception(c, pvzId)
		})

		req, _ := http.NewRequest("POST", "/pvz/"+pvzId.String()+"/close-last-reception", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Close last reception error")
	})
}

func TestPostPvzPvzIdDeleteLastProduct(t *testing.T) {
	t.Run("Delete last product", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		pvzId := uuid.New()

		mockProductService.On("DeleteLastProduct", mock.Anything, pvzId).Return(nil).Once()

		router.POST("/pvz/"+pvzId.String()+"/delete-last-product", func(c *gin.Context) {
			handler.PostPvzPvzIdDeleteLastProduct(c, pvzId)
		})

		req, _ := http.NewRequest("POST", "/pvz/"+pvzId.String()+"/delete-last-product", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockProductService.AssertExpectations(t)
	})

	t.Run("Delete last product with user error", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		pvzId := uuid.New()

		userErr := custom_errors.UserError{}
		mockProductService.On("DeleteLastProduct", mock.Anything, pvzId).Return(&userErr).Once()

		router.POST("/pvz/"+pvzId.String()+"/delete-last-product", func(c *gin.Context) {
			handler.PostPvzPvzIdDeleteLastProduct(c, pvzId)
		})

		req, _ := http.NewRequest("POST", "/pvz/"+pvzId.String()+"/delete-last-product", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Invalid request format to delete last product error")
	})

	t.Run("Delete last product with internal error", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		pvzId := uuid.New()

		internalErr := errors.New("internal error")
		mockProductService.On("DeleteLastProduct", mock.Anything, pvzId).Return(internalErr).Once()

		router.POST("/pvz/"+pvzId.String()+"/delete-last-product", func(c *gin.Context) {
			handler.PostPvzPvzIdDeleteLastProduct(c, pvzId)
		})

		req, _ := http.NewRequest("POST", "/pvz/"+pvzId.String()+"/delete-last-product", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Delete last reception error")
	})
}

func TestPostReceptions(t *testing.T) {
	t.Run("Create reception", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		pvzId := uuid.New()
		receptionReq := generated.PostReceptionsJSONRequestBody{
			PvzId: pvzId,
		}
		jsonData, _ := json.Marshal(receptionReq)
		receptionId := uuid.New()

		dateTime := time.Now()
		mockReceptionService.On("CreateReception", mock.Anything, pvzId).Return(&generated.Reception{
			Id:       &receptionId,
			PvzId:    pvzId,
			Status:   generated.InProgress,
			DateTime: dateTime,
		}, nil).Once()

		req, _ := http.NewRequest("POST", "/receptions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/receptions", handler.PostReceptions)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockReceptionService.AssertExpectations(t)

		var response generated.Reception
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, &receptionId, response.Id)
		assert.Equal(t, generated.InProgress, response.Status)
	})

	t.Run("Create receptions with user error", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		pvzId := uuid.New()
		receptionReq := generated.PostReceptionsJSONRequestBody{
			PvzId: pvzId,
		}
		jsonData, _ := json.Marshal(receptionReq)

		userErr := custom_errors.UserError{}
		mockReceptionService.On("CreateReception", mock.Anything, pvzId).Return(nil, &userErr).Once()

		req, _ := http.NewRequest("POST", "/receptions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/receptions", handler.PostReceptions)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Invalid request format to create reception")
	})

	t.Run("Create receptions with internal error", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		pvzId := uuid.New()
		receptionReq := generated.PostReceptionsJSONRequestBody{
			PvzId: pvzId,
		}
		jsonData, _ := json.Marshal(receptionReq)

		internalErr := errors.New("internal error")
		mockReceptionService.On("CreateReception", mock.Anything, pvzId).Return(nil, internalErr).Once()

		req, _ := http.NewRequest("POST", "/receptions", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/receptions", handler.PostReceptions)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Create reception error")
	})
}

func TestPostRegister(t *testing.T) {
	t.Run("Register", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		registerReq := generated.PostRegisterJSONRequestBody{
			Email:    "test@example.com",
			Password: "secure_password",
			Role:     "employee",
		}
		jsonData, _ := json.Marshal(registerReq)
		userId := uuid.New()

		mockUserService.On("Register", mock.Anything, openapi_types.Email("test@example.com"), "secure_password", generated.UserRoleEmployee).
			Return(&generated.User{
				Id:    &userId,
				Email: "test@example.com",
				Role:  generated.UserRoleEmployee,
			}, "valid_token", nil).Once()

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/register", func(c *gin.Context) {
			handler.PostRegister(c)
		})
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockUserService.AssertExpectations(t)

		var response generated.User
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, &userId, response.Id)
		assert.Equal(t, openapi_types.Email("test@example.com"), response.Email)

		cookies := w.Result().Cookies()
		found := false
		for _, cookie := range cookies {
			if cookie.Name == "auth_token" {
				found = true
				assert.Equal(t, "valid_token", cookie.Value)
			}
		}
		assert.True(t, found)
	})

	t.Run("Register with invalid role", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		registerReq := generated.PostRegisterJSONRequestBody{
			Email:    "test@example.com",
			Password: "secure_password",
			Role:     "InvalidRole",
		}
		jsonData, _ := json.Marshal(registerReq)

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/register", handler.PostRegister)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Invalid role")
	})

	t.Run("register with internal error", func(t *testing.T) {
		router, mockUserService, mockPvzService, mockProductService, mockReceptionService := setupTestEnv()
		handler := api.NewHttpHandler(mockPvzService, mockReceptionService, mockProductService, mockUserService)

		registerReq := generated.PostRegisterJSONRequestBody{
			Email:    "test@example.com",
			Password: "secure_password",
			Role:     "employee",
		}
		jsonData, _ := json.Marshal(registerReq)

		internalErr := errors.New("internal error")
		mockUserService.On("Register", mock.Anything, openapi_types.Email("test@example.com"), "secure_password", generated.UserRoleEmployee).Return(nil, "", internalErr).Once()

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.POST("/register", handler.PostRegister)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response generated.Error
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response.Message, "Register error")
	})
}
