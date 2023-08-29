package main

import (
	"errors"
	"expvar"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"url_shortner/internal/data"
	"url_shortner/internal/validator"

	"github.com/felixge/httpsnoop"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

// performs rate limiting on incoming requests.
func (app *application) rateLimit(next http.HandlerFunc) http.HandlerFunc {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Periodically clean up the clients map to remove stale entries.
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only carry out the rate limiting check if it's enabled in the application configuration.
		if app.config.limiter.enabled {
			ip := realip.FromRequest(r)

			mu.Lock()
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}
			clients[ip].lastSeen = time.Now()

			// Check if the client's request rate exceeds the rate limit.
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}
			mu.Unlock()
		}

		// If rate limiting is not triggered, pass the request to the next handler.
		next.ServeHTTP(w, r)
	})
}

// recovers from panics and sends a server error response.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// records metrics for api response
func (app *application) metrics(next http.Handler) http.Handler {
	totalRequestsRecieved := expvar.NewInt("total_request_recieved")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_μs")
	totalResponsesSentByStatus := expvar.NewMap("total_responses_sent_by_status")
	averageResponseTimeMicroSeconds := expvar.NewInt("average_processing_time_μs")
	minResponseTimeMicroSeconds := expvar.NewFloat("minimum_processing_time_μs")
	maxResponseTimeMicroSeconds := expvar.NewFloat("maximum_processing_time_μs")

	minResponseTimeMicroSeconds.Set(math.MaxFloat64)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == "/debug/vars" {
			next.ServeHTTP(w, r) // Forward the request directly
			return
		}

		totalRequestsRecieved.Add(1)
		metrics := httpsnoop.CaptureMetrics(next, w, r)

		responseTime := metrics.Duration.Microseconds()

		totalResponsesSent.Add(1)
		totalProcessingTimeMicroseconds.Add(responseTime)
		averageResponseTimeMicroSeconds.Set(totalProcessingTimeMicroseconds.Value() / totalResponsesSent.Value())
		minResponseTimeMicroSeconds.Set(math.Min(float64(minResponseTimeMicroSeconds.Value()), float64(responseTime)))
		maxResponseTimeMicroSeconds.Set(math.Max(float64(maxResponseTimeMicroSeconds.Value()), float64(responseTime)))
		totalResponsesSentByStatus.Add(strconv.Itoa(metrics.Code), 1)

	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.setUserInContext(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]
		v := validator.New()
		data.ValidateTokenPlainText(v, token)
		if !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.Model.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.setUserInContext(r, user)
		r = app.setAuthTokenPlaintextInContext(r, token)
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.getUserFromContext(r)
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
