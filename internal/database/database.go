package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"space-booking/internal/models"
	"strconv"
	"time"

	// PostgreSQL driver
	_ "github.com/jackc/pgx/v5/stdlib"

	// Migration libraries
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	// Environment variables
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	CreateBooking(booking *models.Booking) error
	GetAllBookings() ([]models.Booking, error)
	CheckLaunchpadAvailability(launchpadID string, launchDate time.Time) (bool, error)
	CheckDestinationSchedule(destinationID int64, launchpadID string, launchDate time.Time) (bool, error)
}

type service struct {
	db *sql.DB
}

var (
	database     = os.Getenv("DB_DATABASE")
	password     = os.Getenv("DB_PASSWORD")
	username     = os.Getenv("DB_USERNAME")
	port         = os.Getenv("DB_PORT")
	host         = os.Getenv("DB_HOST")
	schema       = os.Getenv("DB_SCHEMA")
	SpaceXAPIURL = os.Getenv("SPACEXAPIURL")
	dbInstance   *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s",
		username, password, host, port, database, schema,
	)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}

	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 100 { // Assuming 100 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}

func (s *service) CreateBooking(booking *models.Booking) error {
	query := `
		INSERT INTO bookings (first_name, last_name, gender, birthday, launchpad_id, destination_id, launch_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	var id int
	err := s.db.QueryRow(
		query,
		booking.FirstName,
		booking.LastName,
		booking.Gender,
		booking.Birthday,
		booking.LaunchpadID,
		booking.DestinationID,
		booking.LaunchDate,
	).Scan(&id)
	if err != nil {
		return err
	}
	booking.ID = id
	return nil
}

func (s *service) GetAllBookings() ([]models.Booking, error) {
	query := `
		SELECT id, first_name, last_name, gender, birthday, launchpad_id, destination_id, launch_date
		FROM bookings
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []models.Booking
	for rows.Next() {
		var booking models.Booking
		err := rows.Scan(
			&booking.ID,
			&booking.FirstName,
			&booking.LastName,
			&booking.Gender,
			&booking.Birthday,
			&booking.LaunchpadID,
			&booking.DestinationID,
			&booking.LaunchDate,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	return bookings, nil
}

func (s *service) CheckLaunchpadAvailability(launchpadID string, launchDate time.Time) (bool, error) {
	var launches []struct {
		Launchpad string `json:"launchpad"`
		Name      string `json:"name"`
		DateLocal string `json:"date_local"`
		ID        string `json:"id"`
	}
	resp, err := http.Get(SpaceXAPIURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(body, &launches)
	if err != nil {
		return false, err
	}
	// Log the launches for debugging
	log.Printf("Received %d launches", len(launches))
	for i, launch := range launches {
		log.Printf("Launch %d: %+v", i, launch)
	}

	for _, launch := range launches {
		dateLocal, err := time.Parse(time.RFC3339, launch.DateLocal)
		if err != nil {
			log.Printf("Error parsing dateLocal for launch %s: %v", launch.ID, err)
			continue // Skip this launch due to invalid date
		}

		if launch.Launchpad == launchpadID && dateLocal.Format("2006-01-02") == launchDate.Format("2006-01-02") {
			// Launchpad is not available on this date
			return false, nil
		}
	}

	// No conflicting launches found
	return true, nil
}

func (s *service) CheckDestinationSchedule(destinationID int64, launchpadID string, launchDate time.Time) (bool, error) {
	// Get list of destinations
	rows, err := s.db.Query(`SELECT id FROM destinations ORDER BY id`)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return false, err
	}
	defer rows.Close()

	var destinationIDs []int64
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return false, err
		}
		destinationIDs = append(destinationIDs, id)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		return false, err
	}

	log.Printf("Destination IDs: %v", destinationIDs)

	if len(destinationIDs) == 0 {
		return false, fmt.Errorf("no destinations available")
	}

	weekday := (int(launchDate.Weekday()) + 6) % 7 // Monday=1, ..., Sunday=7
	index := weekday % len(destinationIDs)
	expectedDestinationID := destinationIDs[index]

	log.Printf("Weekday: %d, Index: %d, Expected Destination ID: %d", weekday, index, expectedDestinationID)
	log.Printf("destinationID: %d, expectedDestinationID: %d", destinationID, expectedDestinationID)

	if destinationID != expectedDestinationID {
		return false, nil
	}

	return true, nil
}
