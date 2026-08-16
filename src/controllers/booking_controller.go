package controllers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-helpme-booking/src/clients"
	"go-helpme-booking/src/config"
	"go-helpme-booking/src/helpers"
	"go-helpme-booking/src/services"
	"go-helpme-booking/src/utils/logger"
	"go.uber.org/zap"
)

type BookingController struct {
	userClient *clients.UserServiceClient
	service    services.BookingService
}

func NewBookingController(userClient *clients.UserServiceClient, service services.BookingService) *BookingController {
	return &BookingController{userClient: userClient, service: service}
}

func requestLog(c *gin.Context) *logger.Logger {
	return logger.WithRequestID(c.GetString(config.CtxRequestID))
}

// lookupUser fetches the owning user from the User Service, forwarding the caller's
// Bearer token so the User Service's own auth middleware accepts the S2S call.
func (bc *BookingController) lookupUser(c *gin.Context, userIDStr string) (*clients.User, error) {
	ctx := helpers.WithHeaders(c.Request.Context(), map[string]string{
		config.HeaderAuthorization: c.GetHeader(config.HeaderAuthorization),
	})
	return bc.userClient.GetByID(ctx, userIDStr)
}

// List godoc
// @Summary      List bookings
// @Description  Looks up the owning user via the User Service, then lists their bookings from Postgres
// @Tags         Bookings
// @Produce      json
// @Param        user_id  query     string  true  "User UUID"
// @Success      200      {object}  helpers.Response{data=object{user=clients.User,bookings=[]models.Booking,meta=helpers.Meta}}
// @Failure      400      {object}  helpers.Response
// @Failure      404      {object}  helpers.Response
// @Failure      500      {object}  helpers.Response
// @Security     BearerAuth
// @Router       /bookings [get]
func (bc *BookingController) List(c *gin.Context) {
	log := requestLog(c)

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		helpers.BadRequest(c, "user_id is required")
		return
	}
	userID, parseErr := uuid.Parse(userIDStr)
	if parseErr != nil {
		helpers.BadRequest(c, "user_id must be a valid UUID")
		return
	}

	if err := helpers.Try(func() error {
		user, svcErr := bc.lookupUser(c, userIDStr)
		if svcErr != nil {
			if errors.Is(svcErr, clients.ErrUserNotFound) {
				helpers.NotFound(c)
			} else {
				helpers.InternalError(c, svcErr.Error())
			}
			return svcErr
		}

		page, pageSize := helpers.ParsePagination(c)
		bookings, total, dbErr := bc.service.ListByUserID(userID, page, pageSize)
		if dbErr != nil {
			helpers.InternalError(c, dbErr.Error())
			return dbErr
		}

		helpers.OK(c, "success", gin.H{
			"user":     user,
			"bookings": bookings,
			"meta":     helpers.PaginationMeta(page, pageSize, total),
		})
		return nil
	}); err != nil {
		log.Error("BookingController.List", zap.Error(err), zap.String("user_id", userIDStr))
	}
}

// Create godoc
// @Summary      Create a booking
// @Description  Creates a booking for a user, after verifying the user exists via the User Service
// @Tags         Bookings
// @Accept       json
// @Produce      json
// @Param        user_id  query     string                      true  "User UUID"
// @Param        body     body      services.CreateBookingInput true  "Booking payload"
// @Success      201      {object}  helpers.Response{data=models.Booking}
// @Failure      400      {object}  helpers.Response
// @Failure      404      {object}  helpers.Response
// @Failure      500      {object}  helpers.Response
// @Security     BearerAuth
// @Router       /bookings [post]
func (bc *BookingController) Create(c *gin.Context) {
	log := requestLog(c)

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		helpers.BadRequest(c, "user_id is required")
		return
	}
	userID, parseErr := uuid.Parse(userIDStr)
	if parseErr != nil {
		helpers.BadRequest(c, "user_id must be a valid UUID")
		return
	}

	var input services.CreateBookingInput
	if bindErr := c.ShouldBindJSON(&input); bindErr != nil {
		helpers.BadRequest(c, bindErr.Error())
		return
	}

	if err := helpers.Try(func() error {
		if _, svcErr := bc.lookupUser(c, userIDStr); svcErr != nil {
			if errors.Is(svcErr, clients.ErrUserNotFound) {
				helpers.NotFound(c)
			} else {
				helpers.InternalError(c, svcErr.Error())
			}
			return svcErr
		}

		booking, dbErr := bc.service.Create(userID, input)
		if dbErr != nil {
			helpers.InternalError(c, dbErr.Error())
			return dbErr
		}

		log.Info("booking created", zap.String("id", booking.ID.String()), zap.String("user_id", userIDStr))
		helpers.Created(c, booking)
		return nil
	}); err != nil {
		log.Error("BookingController.Create", zap.Error(err), zap.String("user_id", userIDStr))
	}
}
