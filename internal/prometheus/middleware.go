package prometheus

import (
	"errors"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

func MetricsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		TotalRequests.Inc()
		if strings.HasSuffix(ctx.Path(), "/compose") {
			ComposeRequests.Inc()
		}
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(ctx.Path()))
		defer timer.ObserveDuration()
		return next(ctx)
	}
}

func StatusMiddleware(subsystem string) func(next echo.HandlerFunc) echo.HandlerFunc {
	counter := StatusRequestsCounter(subsystem)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {

			// call the next handler to see if
			// an error occurred, see:
			// - https://github.com/labstack/echo/issues/1837#issuecomment-816399630
			// - https://github.com/labstack/echo/discussions/1820#discussioncomment-529428
			err := next(ctx)

			path := pathLabel(ctx.Path())
			method := ctx.Request().Method
			status := ctx.Response().Status

			httpErr := new(echo.HTTPError)
			if errors.As(err, &httpErr) {
				status = httpErr.Code
			}

			counter.WithLabelValues(
				method,
				path,
				strconv.Itoa(status),
				subsystem,
			).Inc()

			return err
		}
	}
}
