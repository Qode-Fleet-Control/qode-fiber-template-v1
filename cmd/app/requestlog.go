package main

// KEEP THIS: one log line per request (status + duration) is what the workspace Monitoring tab charts.

import (
	"encoding/json"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

// requestLog logs every request once, in the contract shape. Fiber only turns
// a handler error into a status after the chain returns, so (like Fiber's own
// logger middleware) the error handler runs here first and the logged status
// is the one the client got.
func requestLog(c *fiber.Ctx) error {
	start := time.Now()
	if err := c.Next(); err != nil {
		if herr := c.App().ErrorHandler(c, err); herr != nil {
			_ = c.SendStatus(fiber.StatusInternalServerError)
		}
	}
	logRequest(c.Method(), c.Path(), string(c.Request().Header.UserAgent()), c.Response().StatusCode(), time.Since(start))
	return nil
}

// requestLine is the fleet's request-log contract: one JSON object on stdout per
// finished HTTP request (health checks included), with the status and the
// duration in milliseconds. Never add bodies, cookies, auth headers or the query
// string.
type requestLine struct {
	Status     int     `json:"status"`
	DurationMs float64 `json:"durationMs"`
	Method     string  `json:"method"`
	Path       string  `json:"path"`
	UA         string  `json:"ua,omitempty"`
}

// logRequest writes one request line to stdout. json.Encoder emits the object
// and its newline in a single Write, so concurrent requests do not interleave.
func logRequest(method, path, ua string, status int, took time.Duration) {
	if status == 0 {
		status = http.StatusOK // nothing written explicitly: net/http sends 200
	}
	_ = json.NewEncoder(os.Stdout).Encode(requestLine{
		Status:     status,
		DurationMs: math.Round(float64(took)/float64(time.Microsecond)) / 1000,
		Method:     method,
		Path:       path,
		UA:         ua,
	})
}
