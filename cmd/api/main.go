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
	ctxlogger "github.com/Guizzs26/fintrack/internal/modules/pkg/logger/context"
	"github.com/Guizzs26/fintrack/internal/modules/pkg/validatorx"
	"github.com/Guizzs26/fintrack/internal/platform/config"
	"github.com/Guizzs26/fintrack/internal/platform/postgres"
	"github.com/google/uuid"
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
	baseLogger := logger.NewSlogConfig(logCfg)
	slog.SetDefault(baseLogger)

	e := echo.New()
	e.HideBanner = true
	e.Validator = validatorx.NewValidator()
	e.HTTPErrorHandler = customerErrorHandler

	e.Use(middleware.Recover())
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return uuid.NewString()
		},
	}))
	e.Use(middleware.BodyLimit("2MB"))
	e.Use(ContextualLoggerMiddleware(baseLogger))
	e.Use(RequestLoggerMiddleware())

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

// ContextualLoggerMiddleware creates a request-scoped logger containing the request ID
// and injects it into the standard `context.Context` for use in downstream handlers and services
func ContextualLoggerMiddleware(baseLogger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			requestLogger := baseLogger.With(slog.String("request_id", requestID))

			ctx := c.Request().Context()
			ctxWithLogger := ctxlogger.SetLogger(ctx, requestLogger)
			c.SetRequest(c.Request().WithContext(ctxWithLogger))

			return next(c)
		}
	}
}

// RequestLoggerMiddleware configures and returns Echo's built-in request logger middleware
// It uses the contextual logger (injected by ContextualLoggerMiddleware) to ensure
// that every access log automatically includes the corresponding request ID
func RequestLoggerMiddleware() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogMethod:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogLatency:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger := ctxlogger.GetLogger(c.Request().Context())

			if v.Error == nil {
				logger.LogAttrs(c.Request().Context(), slog.LevelInfo, "HTTP_REQUEST",
					slog.String("method", v.Method),
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("latency", v.Latency.String()),
				)
			} else {
				logger.LogAttrs(c.Request().Context(), slog.LevelError, "HTTP_REQUEST_ERROR",
					slog.String("method", v.Method),
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("latency", v.Latency.String()),
					slog.String("error", v.Error.Error()),
				)
			}
			return nil
		},
	})
}

// customErrorHandler is the centralized error handler for the entire API
// It intercepts any error returned from a handler, inspects its type, and
// formats a standardized JSON error response using our' httpx.Error structure
func customerErrorHandler(err error, c echo.Context) {
	log := ctxlogger.GetLogger(c.Request().Context())
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
	log.Error("unhandled internal error", slog.String("error", err.Error()))
	errResp = httpx.NewAPIError(
		"INTERNAL_SERVER_ERROR",
		"An unexpected error occurred",
		nil,
	)
	httpx.SendAPIError(c, http.StatusInternalServerError, errResp) // 500
}
