package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"space-booking/internal/models"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/time/rate"
)

// RegisterRoutes sets up the router with all endpoints.
func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(rateLimitMiddleware) // Apply rate limiting middleware
	r.Get("/health", s.healthHandler)

	// Endpoints for bookings
	r.Post("/bookings", s.CreateBookingHandler)
	r.Get("/bookings", s.GetAllBookingsHandler)

	return r
}

// healthHandler provides health information.
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

// CreateBookingHandler handles booking creation.
func (s *Server) CreateBookingHandler(w http.ResponseWriter, r *http.Request) {
	var booking models.Booking
	if err := json.NewDecoder(r.Body).Decode(&booking); err != nil {
		log.Printf("Invalid booking data: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate booking
	valid, err := s.validateBooking(&booking)
	if err != nil {
		log.Printf("Error validating booking: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !valid {
		http.Error(w, "Flight is cancelled due to scheduling conflicts.", http.StatusBadRequest)
		return
	}

	// Create booking in the database
	err = s.db.CreateBooking(&booking)
	if err != nil {
		log.Printf("Error creating booking: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Return the created booking
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(booking)
}

// GetAllBookingsHandler retrieves all bookings.
func (s *Server) GetAllBookingsHandler(w http.ResponseWriter, r *http.Request) {
	bookings, err := s.db.GetAllBookings()
	if err != nil {
		log.Printf("Error retrieving bookings: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Return bookings as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}

// validateBooking checks if the booking is valid.
func (s *Server) validateBooking(booking *models.Booking) (bool, error) {
	launchDate := booking.LaunchDate
	birthday := booking.Birthday

	// For example, check if the dates are zero values
	if launchDate.IsZero() || birthday.IsZero() {
		return false, fmt.Errorf("launch date and birthday must be provided")
	}

	// Call validation functions
	isAvailable, err := s.db.CheckLaunchpadAvailability(booking.LaunchpadID, launchDate)
	if err != nil {
		return false, err
	}
	if !isAvailable {
		return false, nil
	}

	isDestinationValid, err := s.db.CheckDestinationSchedule(booking.DestinationID, booking.LaunchpadID, launchDate)
	if err != nil {
		return false, err
	}
	if !isDestinationValid {
		return false, nil
	}

	return true, nil
}

var (
	visitors = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(1, 3) // 1 request per second, burst of 3
		visitors[ip] = limiter
	}
	return limiter
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		limiter := getVisitor(ip)

		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
