package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/controller"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/internal/logger"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/srvcerrors"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Handler struct {
	ctrl   controller.ControllerProvider
	logger logger.Logger
	e      *echo.Echo
}

func NewHandler(ctrl controller.ControllerProvider, logger logger.Logger) *Handler {
	e := echo.New()
	e.Use(ZapLogger(logger))
	e.HTTPErrorHandler = ErrorHandler(logger)

	h := &Handler{
		ctrl:   ctrl,
		logger: logger,
		e:      e,
	}

	h.setupRoutes()

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.e.ServeHTTP(w, r)
}

func (h *Handler) setupRoutes() {
	api := h.e.Group("/api")

	orders := api.Group("/orders")
	orders.GET("/:id", h.getOrder)
	orders.GET("/:id/items", h.getOrderItems)
}

func (h *Handler) getOrder(c echo.Context) error {
	orderID := c.Param("id")

	if orderID == "" || strings.TrimSpace(orderID) == "" {
		return srvcerrors.ErrInvalidInput
	}

	order, err := h.ctrl.GetOrderByUID(c.Request().Context(), orderID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, order)
}

func (h *Handler) getOrderItems(c echo.Context) error {
	orderID := c.Param("id")
	if orderID == "" || strings.TrimSpace(orderID) == "" {
		return srvcerrors.ErrInvalidInput
	}

	limit := 10
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)

		if err != nil || limit <= 0 {
			return srvcerrors.ErrInvalidInput
		}
		if limit > 100 {
			limit = 100
		}
	}

	lastID := 0
	if lastIDStr := c.QueryParam("last_id"); lastIDStr != "" {
		var err error
		lastID, err = strconv.Atoi(lastIDStr)

		if err != nil || lastID < 0 {
			return srvcerrors.ErrInvalidInput
		}
	}

	items, err := h.ctrl.GetItemsByOrderUID(c.Request().Context(), orderID, lastID, limit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, items)
}

func ZapLogger(logger logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			logger.Debug("handler: incoming request",
				zap.String("method", c.Request().Method),
				zap.String("path", c.Request().URL.Path),
				zap.String("order_uid", c.Param("id")))

			err := next(c)

			duration := time.Since(start)
			if err != nil {
				logger.Debug("handler: request failed",
					zap.String("method", c.Request().Method),
					zap.String("path", c.Request().URL.Path),
					zap.Duration("duration", duration),
					zap.String("order_uid", c.Param("id")),
					zap.Error(err))
			} else {
				logger.Debug("handler: request completed",
					zap.String("method", c.Request().Method),
					zap.String("path", c.Request().URL.Path),
					zap.Duration("duration", duration),
					zap.String("order_uid", c.Param("id")))
			}
			return err
		}
	}
}

func ErrorHandler(logger logger.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		status := http.StatusInternalServerError
		message := "Internal server error"

		if errors.Is(err, srvcerrors.ErrNotFound) {
			status = http.StatusNotFound
			message = "Order not found"
		} else if errors.Is(err, srvcerrors.ErrDatabase) {
			status = http.StatusInternalServerError
			message = "Database error"
		} else if errors.Is(err, srvcerrors.ErrInvalidInput) {
			status = http.StatusBadRequest
			message = "Invalid request parameters"
		} else if errors.Is(err, srvcerrors.ErrKafka) {
			status = http.StatusInternalServerError
			message = "Kafka service error"
		} else {
			if he, ok := err.(*echo.HTTPError); ok {
				status = he.Code
				message = he.Message.(string)
			}
		}

		if status >= 500 {
			logger.Error("handler: request failed",
				zap.String("method", c.Request().Method),
				zap.String("path", c.Request().URL.Path),
				zap.Int("status", status),
				zap.String("order_uid", c.Param("id")),
				zap.Error(err))
		} else {
			logger.Warn("handler: client error",
				zap.String("method", c.Request().Method),
				zap.String("path", c.Request().URL.Path),
				zap.Int("status", status),
				zap.String("order_uid", c.Param("id")),
				zap.Error(err))
		}

		errorResponse := map[string]interface{}{
			"status":  status,
			"message": message,
		}
		
		if !c.Response().Committed {
			c.JSON(status, errorResponse)
		}
	}
}
