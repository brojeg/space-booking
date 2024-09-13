package database

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckLaunchpadAvailability tests the CheckLaunchpadAvailability function
func TestCheckLaunchpadAvailability(t *testing.T) {
	// Mock SpaceX API response
	mockSpaceXResponse := []map[string]interface{}{
		{
			"id":         "launch_1",
			"name":       "Test Launch",
			"date_local": "2049-12-25T00:00:00Z", // Use UTC time
			"launchpad":  "test_launchpad",
		},
	}

	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockSpaceXResponse)
	}))
	defer server.Close()

	// Override the SpaceX API URL with the mock server URL
	SpaceXAPIURL = server.URL

	// Initialize the database service
	s := &service{}

	// Test data
	launchpadID := "test_launchpad"
	launchDate := time.Date(2049, time.December, 25, 0, 0, 0, 0, time.UTC)

	// Call the method
	isAvailable, err := s.CheckLaunchpadAvailability(launchpadID, launchDate)
	assert.NoError(t, err)
	assert.False(t, isAvailable, "Expected launchpad to be unavailable due to conflicting launch")
}

// TestCheckDestinationSchedule tests the CheckDestinationSchedule function
func TestCheckDestinationSchedule(t *testing.T) {
	// Create a sqlmock database connection
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Initialize the service with the mocked db
	s := &service{
		db: db,
	}

	// Mock expected query and result for destination IDs
	mock.ExpectQuery("SELECT id FROM destinations ORDER BY id").
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).
				AddRow(int64(1)).
				AddRow(int64(2)).
				AddRow(int64(3)).
				AddRow(int64(4)).
				AddRow(int64(5)).
				AddRow(int64(6)).
				AddRow(int64(7)),
		)

	// Test data from the request
	destinationID := int64(6)
	launchpadID := "test_launchpad"
	launchDate, err := time.Parse(time.RFC3339, "2049-12-25T00:00:00Z") // December 25, 2049 (Friday)
	require.NoError(t, err)

	// Call the method to check the destination schedule
	isValid, err := s.CheckDestinationSchedule(destinationID, launchpadID, launchDate)
	assert.NoError(t, err)
	assert.True(t, isValid, "Expected destination schedule to be valid")
	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

// Additional test for when the launchpad is available
func TestCheckLaunchpadAvailability_NoConflict(t *testing.T) {
	// Mock SpaceX API response with no conflicting launches
	mockSpaceXResponse := []map[string]interface{}{
		{
			"id":         "launch_2",
			"name":       "Another Launch",
			"date_local": "2049-12-26T00:00:00Z", // Different date
			"launchpad":  "test_launchpad",
		},
	}

	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockSpaceXResponse)
	}))
	defer server.Close()

	// Override the SpaceX API URL with the mock server URL
	SpaceXAPIURL = server.URL

	// Initialize the database service
	s := &service{}

	// Test data
	launchpadID := "test_launchpad"
	launchDate := time.Date(2049, time.December, 25, 0, 0, 0, 0, time.UTC)

	// Call the method
	isAvailable, err := s.CheckLaunchpadAvailability(launchpadID, launchDate)
	assert.NoError(t, err)
	assert.True(t, isAvailable, "Expected launchpad to be available since there is no conflicting launch")
}

// Additional test for invalid dates in the SpaceX API response
func TestCheckLaunchpadAvailability_InvalidDate(t *testing.T) {
	// Save the original value of spaceXAPIURL
	origURL := SpaceXAPIURL
	defer func() { SpaceXAPIURL = origURL }()

	// Mock SpaceX API response with invalid date
	mockSpaceXResponse := []map[string]interface{}{
		{
			"id":         "launch_3",
			"name":       "Invalid Date Launch",
			"date_local": "invalid-date-format",
			"launchpad":  "test_launchpad",
		},
	}

	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockSpaceXResponse)
	}))
	defer server.Close()

	// Override the SpaceX API URL with the mock server URL
	SpaceXAPIURL = server.URL

	// Initialize the database service
	s := &service{}

	// Test data
	launchpadID := "test_launchpad"
	launchDate := time.Date(2049, time.December, 25, 0, 0, 0, 0, time.UTC)

	// Call the method
	isAvailable, err := s.CheckLaunchpadAvailability(launchpadID, launchDate)
	assert.NoError(t, err)
	assert.True(t, isAvailable, "Expected launchpad to be available since the launch date is invalid")
}
