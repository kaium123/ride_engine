package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLocationRepository is a mock implementation of the location repository
type MockLocationRepository struct {
	mock.Mock
}

func (m *MockLocationRepository) UpdateDriverLocation(ctx context.Context, driverID int64, lat, lng float64) error {
	args := m.Called(ctx, driverID, lat, lng)
	return args.Error(0)
}

func (m *MockLocationRepository) FindNearestDrivers(ctx context.Context, lat, lng float64, maxDistance float64, limit int) ([]int64, error) {
	args := m.Called(ctx, lat, lng, maxDistance, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockLocationRepository) GetDriverLocation(ctx context.Context, driverID int64) (lat, lng float64, updatedAt *time.Time, err error) {
	args := m.Called(ctx, driverID)
	return args.Get(0).(float64), args.Get(1).(float64), args.Get(2).(*time.Time), args.Error(3)
}

func TestLocationService_UpdateDriverLocation(t *testing.T) {
	mockRepo := new(MockLocationRepository)
	service := &LocationService{
		repo: mockRepo,
	}

	ctx := context.Background()
	driverID := int64(456)
	lat := 23.8100
	lng := 90.4120

	mockRepo.On("UpdateDriverLocation", ctx, driverID, lat, lng).Return(nil)

	err := service.UpdateDriverLocation(ctx, driverID, lat, lng)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLocationService_UpdateDriverLocation_Error(t *testing.T) {
	mockRepo := new(MockLocationRepository)
	service := &LocationService{
		repo: mockRepo,
	}

	ctx := context.Background()
	driverID := int64(456)
	lat := 23.8100
	lng := 90.4120

	mockRepo.On("UpdateDriverLocation", ctx, driverID, lat, lng).Return(errors.New("database error"))

	err := service.UpdateDriverLocation(ctx, driverID, lat, lng)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestLocationService_FindNearestDrivers(t *testing.T) {
	mockRepo := new(MockLocationRepository)
	service := &LocationService{
		repo: mockRepo,
	}

	ctx := context.Background()
	lat := 23.8100
	lng := 90.4120
	maxDistance := 5000.0
	limit := 10

	expectedDrivers := []int64{1, 2, 3, 4, 5}

	mockRepo.On("FindNearestDrivers", ctx, lat, lng, maxDistance, limit).Return(expectedDrivers, nil)

	drivers, err := service.FindNearestDrivers(ctx, lat, lng, maxDistance, limit)

	assert.NoError(t, err)
	assert.NotNil(t, drivers)
	assert.Len(t, drivers, 5)
	assert.Equal(t, expectedDrivers, drivers)
	mockRepo.AssertExpectations(t)
}

func TestLocationService_FindNearestDrivers_NoDriversNearby(t *testing.T) {
	mockRepo := new(MockLocationRepository)
	service := &LocationService{
		repo: mockRepo,
	}

	ctx := context.Background()
	lat := 23.8100
	lng := 90.4120
	maxDistance := 5000.0
	limit := 10

	emptyDrivers := []int64{}

	mockRepo.On("FindNearestDrivers", ctx, lat, lng, maxDistance, limit).Return(emptyDrivers, nil)

	drivers, err := service.FindNearestDrivers(ctx, lat, lng, maxDistance, limit)

	assert.NoError(t, err)
	assert.NotNil(t, drivers)
	assert.Len(t, drivers, 0)
	mockRepo.AssertExpectations(t)
}

func TestLocationService_FindNearestDrivers_Error(t *testing.T) {
	mockRepo := new(MockLocationRepository)
	service := &LocationService{
		repo: mockRepo,
	}

	ctx := context.Background()
	lat := 23.8100
	lng := 90.4120
	maxDistance := 5000.0
	limit := 10

	mockRepo.On("FindNearestDrivers", ctx, lat, lng, maxDistance, limit).Return(nil, errors.New("query error"))

	drivers, err := service.FindNearestDrivers(ctx, lat, lng, maxDistance, limit)

	assert.Error(t, err)
	assert.Nil(t, drivers)
	assert.Contains(t, err.Error(), "query error")
	mockRepo.AssertExpectations(t)
}

func TestLocationService_GetDriverLocation(t *testing.T) {
	mockRepo := new(MockLocationRepository)
	service := &LocationService{
		repo: mockRepo,
	}

	ctx := context.Background()
	driverID := int64(456)
	expectedLat := 23.8105
	expectedLng := 90.4125
	now := time.Now()

	mockRepo.On("GetDriverLocation", ctx, driverID).Return(expectedLat, expectedLng, &now, nil)

	lat, lng, updatedAt, err := service.GetDriverLocation(ctx, driverID)

	assert.NoError(t, err)
	assert.Equal(t, expectedLat, lat)
	assert.Equal(t, expectedLng, lng)
	assert.NotNil(t, updatedAt)
	assert.Equal(t, now, *updatedAt)
	mockRepo.AssertExpectations(t)
}

func TestLocationService_GetDriverLocation_NotFound(t *testing.T) {
	mockRepo := new(MockLocationRepository)
	service := &LocationService{
		repo: mockRepo,
	}

	ctx := context.Background()
	driverID := int64(999)

	mockRepo.On("GetDriverLocation", ctx, driverID).Return(0.0, 0.0, (*time.Time)(nil), errors.New("driver location not found"))

	lat, lng, updatedAt, err := service.GetDriverLocation(ctx, driverID)

	assert.Error(t, err)
	assert.Equal(t, 0.0, lat)
	assert.Equal(t, 0.0, lng)
	assert.Nil(t, updatedAt)
	assert.Contains(t, err.Error(), "driver location not found")
	mockRepo.AssertExpectations(t)
}

func TestLocationService_FindNearestDrivers_WithDifferentLimits(t *testing.T) {
	mockRepo := new(MockLocationRepository)
	service := &LocationService{
		repo: mockRepo,
	}

	ctx := context.Background()
	lat := 23.8100
	lng := 90.4120
	maxDistance := 5000.0

	testCases := []struct {
		name          string
		limit         int
		expectedCount int
	}{
		{
			name:          "Limit 1",
			limit:         1,
			expectedCount: 1,
		},
		{
			name:          "Limit 5",
			limit:         5,
			expectedCount: 5,
		},
		{
			name:          "Limit 10",
			limit:         10,
			expectedCount: 10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			drivers := make([]int64, tc.expectedCount)
			for i := 0; i < tc.expectedCount; i++ {
				drivers[i] = int64(i + 1)
			}

			mockRepo.On("FindNearestDrivers", ctx, lat, lng, maxDistance, tc.limit).Return(drivers, nil).Once()

			result, err := service.FindNearestDrivers(ctx, lat, lng, maxDistance, tc.limit)

			assert.NoError(t, err)
			assert.Len(t, result, tc.expectedCount)
		})
	}

	mockRepo.AssertExpectations(t)
}
