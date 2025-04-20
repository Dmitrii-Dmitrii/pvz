package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"pvz/internal/generated"
	"pvz/internal/models/custom_errors"
	"pvz/internal/models/product_model"
	"pvz/internal/models/reception_model"
	"pvz/internal/services/product_service"
	"testing"
)

type MockProductDriver struct {
	mock.Mock
}

func (m *MockProductDriver) CreateProduct(ctx context.Context, product *product_model.Product, pvzId pgtype.UUID) (*pgtype.UUID, error) {
	args := m.Called(ctx, product, pvzId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pgtype.UUID), args.Error(1)
}

func (m *MockProductDriver) DeleteLastProduct(ctx context.Context, pvzId pgtype.UUID) error {
	args := m.Called(ctx, pvzId)
	return args.Error(0)
}

type MockReceptionService struct {
	mock.Mock
}

func (m *MockReceptionService) GetLastReceptionStatus(ctx context.Context, pvzId pgtype.UUID) (*reception_model.ReceptionStatus, error) {
	args := m.Called(ctx, pvzId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception_model.ReceptionStatus), args.Error(1)
}

func (m *MockReceptionService) CreateReception(ctx context.Context, pvzId uuid.UUID) (*generated.Reception, error) {
	args := m.Called(ctx, pvzId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*generated.Reception), args.Error(1)
}

func (m *MockReceptionService) CloseReception(ctx context.Context, pvzId uuid.UUID) (*generated.Reception, error) {
	args := m.Called(ctx, pvzId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*generated.Reception), args.Error(1)
}

func TestCreateProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("Create product in reception in progress", func(t *testing.T) {
		mockDriver := new(MockProductDriver)
		mockReceptionService := new(MockReceptionService)
		service := product_service.NewProductService(mockDriver, mockReceptionService)

		pvzIdDto := uuid.New()
		productTypeJson := generated.PostProductsJSONBodyType("электроника")
		status := reception_model.InProgress
		receptionId := pgtype.UUID{Bytes: uuid.New(), Valid: true}

		mockReceptionService.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(&status, nil)
		mockDriver.On("CreateProduct", ctx, mock.AnythingOfType("*product_model.Product"), mock.AnythingOfType("pgtype.UUID")).Return(&receptionId, nil)

		result, err := service.CreateProduct(ctx, pvzIdDto, productTypeJson)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, generated.ProductTypeЭлектроника, result.Type)
		assert.NotNil(t, result.Id)
		assert.NotNil(t, result.DateTime)
		assert.NotEmpty(t, result.ReceptionId)
		mockDriver.AssertExpectations(t)
		mockReceptionService.AssertExpectations(t)
	})

	t.Run("Create product without open reception", func(t *testing.T) {
		mockDriver := new(MockProductDriver)
		mockReceptionService := new(MockReceptionService)
		service := product_service.NewProductService(mockDriver, mockReceptionService)

		pvzIdDto := uuid.New()
		productTypeJson := generated.PostProductsJSONBodyType("электроника")
		status := reception_model.Close

		mockReceptionService.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(&status, nil)

		result, err := service.CreateProduct(ctx, pvzIdDto, productTypeJson)

		assert.Equal(t, custom_errors.ErrNoOpenReception, err)
		assert.Nil(t, result)
		mockReceptionService.AssertExpectations(t)
		mockDriver.AssertNotCalled(t, "CreateProduct")
	})

	t.Run("Create product with invalid product type", func(t *testing.T) {
		mockDriver := new(MockProductDriver)
		mockReceptionService := new(MockReceptionService)
		service := product_service.NewProductService(mockDriver, mockReceptionService)

		pvzIdDto := uuid.New()
		productTypeJson := generated.PostProductsJSONBodyType("неизвестный_тип")

		result, err := service.CreateProduct(ctx, pvzIdDto, productTypeJson)

		assert.Equal(t, custom_errors.ErrProductType, err)
		assert.Nil(t, result)
		mockReceptionService.AssertNotCalled(t, "GetLastReceptionStatus")
		mockDriver.AssertNotCalled(t, "CreateProduct")
	})

	t.Run("Create product with reception service error", func(t *testing.T) {
		mockDriver := new(MockProductDriver)
		mockReceptionService := new(MockReceptionService)
		service := product_service.NewProductService(mockDriver, mockReceptionService)

		pvzIdDto := uuid.New()
		productTypeJson := generated.PostProductsJSONBodyType("электроника")

		mockReceptionService.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(nil, assert.AnError)

		result, err := service.CreateProduct(ctx, pvzIdDto, productTypeJson)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockReceptionService.AssertExpectations(t)
		mockDriver.AssertNotCalled(t, "CreateProduct")
	})

	t.Run("Create product with error", func(t *testing.T) {
		mockDriver := new(MockProductDriver)
		mockReceptionService := new(MockReceptionService)
		service := product_service.NewProductService(mockDriver, mockReceptionService)

		pvzIdDto := uuid.New()
		productTypeJson := generated.PostProductsJSONBodyType("электроника")
		status := reception_model.InProgress

		mockReceptionService.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(&status, nil)
		mockDriver.On("CreateProduct", ctx, mock.AnythingOfType("*product_model.Product"), mock.AnythingOfType("pgtype.UUID")).Return(nil, assert.AnError)

		result, err := service.CreateProduct(ctx, pvzIdDto, productTypeJson)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockDriver.AssertExpectations(t)
		mockReceptionService.AssertExpectations(t)
	})
}

func TestDeleteLastProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("Delete last product in reception in progress", func(t *testing.T) {
		mockDriver := new(MockProductDriver)
		mockReceptionService := new(MockReceptionService)
		service := product_service.NewProductService(mockDriver, mockReceptionService)

		pvzIdDto := uuid.New()
		status := reception_model.InProgress

		mockReceptionService.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(&status, nil)
		mockDriver.On("DeleteLastProduct", ctx, mock.AnythingOfType("pgtype.UUID")).Return(nil)

		err := service.DeleteLastProduct(ctx, pvzIdDto)

		assert.NoError(t, err)
		mockDriver.AssertExpectations(t)
		mockReceptionService.AssertExpectations(t)
	})

	t.Run("Delete last product without open reception", func(t *testing.T) {
		mockDriver := new(MockProductDriver)
		mockReceptionService := new(MockReceptionService)
		service := product_service.NewProductService(mockDriver, mockReceptionService)

		pvzIdDto := uuid.New()
		status := reception_model.Close

		mockReceptionService.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(&status, nil)

		err := service.DeleteLastProduct(ctx, pvzIdDto)

		assert.Equal(t, custom_errors.ErrNoOpenReception, err)
		mockReceptionService.AssertExpectations(t)
		mockDriver.AssertNotCalled(t, "DeleteLastProduct")
	})

	t.Run("Delete last product with reception service error", func(t *testing.T) {
		mockDriver := new(MockProductDriver)
		mockReceptionService := new(MockReceptionService)
		service := product_service.NewProductService(mockDriver, mockReceptionService)

		pvzIdDto := uuid.New()

		mockReceptionService.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(nil, assert.AnError)

		err := service.DeleteLastProduct(ctx, pvzIdDto)

		assert.Error(t, err)
		mockReceptionService.AssertExpectations(t)
		mockDriver.AssertNotCalled(t, "DeleteLastProduct")
	})

	t.Run("Delete last product with error", func(t *testing.T) {
		mockDriver := new(MockProductDriver)
		mockReceptionService := new(MockReceptionService)
		service := product_service.NewProductService(mockDriver, mockReceptionService)

		pvzIdDto := uuid.New()
		status := reception_model.InProgress

		mockReceptionService.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(&status, nil)
		mockDriver.On("DeleteLastProduct", ctx, mock.AnythingOfType("pgtype.UUID")).Return(assert.AnError)

		err := service.DeleteLastProduct(ctx, pvzIdDto)

		assert.Error(t, err)
		mockDriver.AssertExpectations(t)
		mockReceptionService.AssertExpectations(t)
	})
}
