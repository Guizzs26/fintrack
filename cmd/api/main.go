package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Guizzs26/fintrack/internal/modules/ledger"
	"github.com/Guizzs26/fintrack/internal/modules/pkg/clock"
	"github.com/Guizzs26/fintrack/internal/modules/pkg/httpx"
	"github.com/Guizzs26/fintrack/internal/modules/pkg/logger"
	"github.com/Guizzs26/fintrack/internal/modules/pkg/validatorx"
	"github.com/Guizzs26/fintrack/internal/platform/config"
	"github.com/Guizzs26/fintrack/internal/platform/postgres"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config to api: %s\n", err)
		os.Exit(1)
	}

	if err := run(ctx, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg *config.Config) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logCfg := logger.SlogConfig{
		Level:     logger.LevelDebug,
		Format:    logger.FormatJSON,
		AddSource: true,
	}
	appLogger := logger.NewSlogConfig(logCfg)
	slog.SetDefault(appLogger)

	e := echo.New()
	e.Validator = validatorx.NewValidator()
	e.HTTPErrorHandler = customerErrorHandler

	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod:        true,
		LogURI:           true,
		LogStatus:        true,
		LogLatency:       true,
		LogError:         true,
		LogRequestID:     true,
		LogRemoteIP:      true,
		LogResponseSize:  true,
		LogContentLength: true,

		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			// Define the common set of attributes for both success and error logs
			commonAttrs := []slog.Attr{
				slog.String("request_id", v.RequestID),
				slog.String("remote_ip", v.RemoteIP),
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.String("latency", v.Latency.String()),
				slog.Int64("response_size", v.ResponseSize),
				slog.String("content_length", v.ContentLength),
			}

			if v.Error == nil {
				// If there was no error, log it as an INFO level event
				appLogger.LogAttrs(c.Request().Context(), slog.LevelInfo, "HTTP_REQUEST",
					commonAttrs...,
				)
			} else {
				// If an error occurred, log it as an ERROR level event,
				// and include the specific error message.
				appLogger.LogAttrs(c.Request().Context(), slog.LevelError, "HTTP_REQUEST_ERROR",
					append(commonAttrs, slog.String("error", v.Error.Error()))...,
				)
			}
			return nil
		},
	}))

	pgConn, err := postgres.NewPostgresConnection(ctx, *cfg)
	if err != nil {
		return err
	}
	defer pgConn.Close()

	clock := clock.SystemClock{}

	// ----- Ledger module dependencies ----- //

	accountRepo := ledger.NewPostgresAccountRepository(pgConn.Pool)
	ledgerSvc := ledger.NewLedgerService(accountRepo, clock)
	ledgerHandler := ledger.NewLedgerHandler(ledgerSvc)

	apiRouteGroup := e.Group("/api/v1")
	ledgerHandler.RegisterRoutes(apiRouteGroup)

	e.Logger.Fatal(e.Start(":9999"))
	return nil
}

// customErrorHandler is the centralized error handler for the entire API
// It intercepts any error returned from a handler, inspects its type, and
// formats a standardized JSON error response using our' httpx.Error structure
func customerErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	// 1. Handle custom validation errors from our validatorx package
	var valErr validatorx.ValidationError
	if errors.As(err, &valErr) {
		errResp := httpx.NewAPIError(
			"VALIDATION_ERROR",
			"One or more fields failed validation",
			valErr.Errors, // The 'Details' field will contain the slice of FieldError
		)
		httpx.SendAPIError(c, http.StatusBadRequest, errResp)
		return
	}

	// 2. Handle known domain errors from the LEDGER MODULE
	var httpStatus int
	var errResp httpx.APIError

	switch {
	case errors.Is(err, ledger.ErrAccountNotFound):
		httpStatus = http.StatusNotFound // 404
		errResp = httpx.NewAPIError("RESOURCE_NOT_FOUND", err.Error(), nil)

	case errors.Is(err, ledger.ErrAccountArchived):
		httpStatus = http.StatusForbidden // 403
		errResp = httpx.NewAPIError("FORBIDDEN", err.Error(), nil)

	case errors.Is(err, ledger.ErrAccountNameRequired),
		errors.Is(err, ledger.ErrInconsistentAmountSign),
		errors.Is(err, ledger.ErrAmountCannotBeZero):
		httpStatus = http.StatusUnprocessableEntity // 422
		errResp = httpx.NewAPIError("BUSINESS_RULE_VIOLATION", err.Error(), nil)
	}

	if httpStatus != 0 {
		httpx.SendAPIError(c, httpStatus, errResp)
		return
	}

	// 3. Handle generic Echo HTTP errors
	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		errResp = httpx.NewAPIError("HTTP_ERROR", fmt.Sprintf("%v", httpErr.Message), nil)
		httpx.SendAPIError(c, httpErr.Code, errResp)
		return
	}

	// 4. Fallback for any other unexpected error
	c.Logger().Error(err) // Log the full error for DEBUGGING
	errResp = httpx.NewAPIError(
		"INTERNAL_SERVER_ERROR",
		"An unexpected error occurred",
		nil,
	)
	httpx.SendAPIError(c, http.StatusInternalServerError, errResp) // 500
}
