package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"space-booking/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

// MockDatabase is a mock implementation of the database.Service interface
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Health() map[string]string {
	return map[string]string{"status": "up"}
}

func (m *MockDatabase) Close() error {
	return nil
}

func (m *MockDatabase) CreateBooking(booking *models.Booking) error {
	args := m.Called(booking)
	return args.Error(0)
}

func (m *MockDatabase) GetAllBookings() ([]models.Booking, error) {
	args := m.Called()
	return args.Get(0).([]models.Booking), args.Error(1)
}

func (m *MockDatabase) CheckLaunchpadAvailability(launchpadID string, launchDate time.Time) (bool, error) {
	args := m.Called(launchpadID, launchDate)
	return args.Bool(0), args.Error(1)
}

func (m *MockDatabase) CheckDestinationSchedule(destinationID int64, launchpadID string, launchDate time.Time) (bool, error) {
	args := m.Called(destinationID, launchpadID, launchDate)
	return args.Bool(0), args.Error(1)
}

func TestCreateBookingHandler(t *testing.T) {
	// Setup
	db := new(MockDatabase)
	s := &Server{db: db}

	// Prepare test data
	bookingData := models.Booking{
		FirstName:     "Test",
		LastName:      "User",
		Gender:        "Non-binary",
		Birthday:      time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
		LaunchpadID:   "test_launchpad",
		DestinationID: 1,
		LaunchDate:    time.Date(2049, time.December, 25, 0, 0, 0, 0, time.UTC),
	}

	// Marshal booking data to JSON
	jsonData, err := json.Marshal(bookingData)
	assert.NoError(t, err)

	// Mock database methods
	db.On("CheckLaunchpadAvailability", bookingData.LaunchpadID, bookingData.LaunchDate).Return(true, nil)
	db.On("CheckDestinationSchedule", bookingData.DestinationID, bookingData.LaunchpadID, bookingData.LaunchDate).Return(true, nil)
	db.On("CreateBooking", mock.AnythingOfType("*models.Booking")).Return(nil)

	// Create a request to pass to our handler
	req, err := http.NewRequest("POST", "/bookings", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.CreateBookingHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect
	assert.Equal(t, http.StatusCreated, rr.Code, "Expected status code 201 Created")

	// Check the response body
	var responseBooking models.Booking
	err = json.Unmarshal(rr.Body.Bytes(), &responseBooking)
	assert.NoError(t, err)
	assert.Equal(t, bookingData.FirstName, responseBooking.FirstName)
	assert.Equal(t, bookingData.LastName, responseBooking.LastName)

	// Ensure that the mocked methods were called
	db.AssertExpectations(t)
}

func TestGetAllBookingsHandler(t *testing.T) {
	// Setup
	db := new(MockDatabase)
	s := &Server{db: db}

	// Prepare test data
	bookings := []models.Booking{
		{
			ID:            1,
			FirstName:     "Test",
			LastName:      "User",
			Gender:        "Non-binary",
			Birthday:      time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
			LaunchpadID:   "test_launchpad",
			DestinationID: 1,
			LaunchDate:    time.Date(2049, time.December, 25, 0, 0, 0, 0, time.UTC),
		},
	}

	// Mock database method
	db.On("GetAllBookings").Return(bookings, nil)

	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/bookings", nil)
	assert.NoError(t, err)

	// Record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.GetAllBookingsHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200 OK")

	// Check the response body
	var responseBookings []models.Booking
	err = json.Unmarshal(rr.Body.Bytes(), &responseBookings)
	assert.NoError(t, err)
	assert.Equal(t, len(bookings), len(responseBookings))
	assert.Equal(t, bookings[0].FirstName, responseBookings[0].FirstName)

	// Ensure that the mocked methods were called
	db.AssertExpectations(t)
}

// Reset the visitors map before each test to avoid interference between tests.
func resetVisitors() {
	mu.Lock()
	defer mu.Unlock()
	visitors = make(map[string]*rate.Limiter)
}

func TestRateLimitMiddleware(t *testing.T) {
	// Reset the visitors map
	resetVisitors()

	// Create a simple handler that returns 200 OK
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Apply the rate-limiting middleware
	rateLimitedHandler := rateLimitMiddleware(handler)

	// Create a test server
	ts := httptest.NewServer(rateLimitedHandler)
	defer ts.Close()

	client := &http.Client{}

	// Simulate requests from the same IP address
	ip := "192.0.2.1:1234" // Using a fixed IP for testing

	// Replace RemoteAddr in the request to simulate the same IP
	doRequest := func() *http.Response {
		req, err := http.NewRequest("GET", ts.URL, nil)
		require.NoError(t, err)

		// Override the RemoteAddr
		req.RemoteAddr = ip

		resp, err := client.Do(req)
		require.NoError(t, err)
		return resp
	}

	// The rate limiter allows 1 request per second with a burst of 3
	// So we can make 3 immediate requests, and then subsequent requests should be limited

	// Make 3 allowed requests
	for i := 0; i < 3; i++ {
		resp := doRequest()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200 OK on request %d", i+1)
		resp.Body.Close()
	}

	// The 4th request should be rate-limited
	resp := doRequest()
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "Expected status 429 Too Many Requests on 4th request")
	resp.Body.Close()

	// Wait for 1 second to allow the limiter to refill
	time.Sleep(1 * time.Second)

	// After waiting, we should be able to make another request
	resp = doRequest()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status 200 OK after waiting")
	resp.Body.Close()
}
