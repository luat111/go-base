package middlewares

import (
	"go-base/pkg/logger"
	"go-base/pkg/tracing"
	"net/http"
	"strings"
	"time"
)

type RequestLog struct {
	CorrelationId string        `json:"correlationId"`
	StartTime     string        `json:"start_time,omitempty"`
	ResponseTime  time.Duration `json:"response_time,omitempty"`
	Method        string        `json:"method,omitempty"`
	UserAgent     string        `json:"user_agent,omitempty"`
	IP            string        `json:"ip,omitempty"`
	URI           string        `json:"uri,omitempty"`
	Status        int           `json:"status,omitempty"`
}

type StatusResponseWriter struct {
	http.ResponseWriter
	status int
}

func Logging(logger logger.ILogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			start := time.Now()
			srw := &StatusResponseWriter{ResponseWriter: rw}

			defer func(res *StatusResponseWriter, req *http.Request) {
				l := RequestLog{
					CorrelationId: tracing.FromContext(r.Context()),
					StartTime:     start.Format("2006-01-02T15:04:05"),
					ResponseTime:  time.Since(start),
					Method:        req.Method,
					URI:           req.RequestURI,
					IP:            getIPAddress(req),
					UserAgent:     req.UserAgent(),
					Status:        res.status,
				}

				if logger != nil {
					if res.status >= http.StatusInternalServerError {
						logger.Error("HTTP", "Message", l)
					} else {
						logger.Log("HTTP", "Message", l)
					}
				}
			}(srw, r)

			defer recover()

			next.ServeHTTP(srw, r)
		})
	}
}

func getIPAddress(r *http.Request) string {
	ips := strings.Split(r.Header.Get("X-Forwarded-For"), ",")

	ipAddress := ips[0]

	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	return strings.TrimSpace(ipAddress)
}
