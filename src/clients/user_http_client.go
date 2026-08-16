package clients

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go-helpme-booking/src/config"
	"go-helpme-booking/src/helpers"
)

// UserServiceClient makes HTTP calls to the User Service microservice.
// It's a thin wrapper over helpers.HTTPClient — the base URL and timeout come
// from config.App.Services.UserService; default headers, retry, and
// non-success request/response logging are all handled by helpers.HTTPClient.
type UserServiceClient struct {
	*helpers.HTTPClient
}

// NewUserServiceClient builds a client using config.App.Services.UserService (base URL + timeout).
func NewUserServiceClient() *UserServiceClient {
	cfg := config.App.Services.UserService
	return &UserServiceClient{
		HTTPClient: helpers.NewHTTPClient("user_service_client", cfg.BaseURL, time.Duration(cfg.Timeout)*time.Second),
	}
}

// User is the subset of the User Service's user representation the booking service needs.
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

// userEnvelope mirrors the User Service's helpers.Response{Data: models.User} JSON shape.
type userEnvelope struct {
	Success bool `json:"success"`
	Data    User `json:"data"`
}

// ErrUserNotFound is returned when the User Service responds 404 for a given ID.
var ErrUserNotFound = errors.New("user not found")

// GetByID calls GET /api/v1/users/{id} on the User Service and unwraps the response envelope.
func (c *UserServiceClient) GetByID(ctx context.Context, userID string) (*User, error) {
	var envelope userEnvelope
	resp, err := c.Get(ctx, "/api/v1/users/"+userID, &envelope)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &envelope.Data, nil
}
