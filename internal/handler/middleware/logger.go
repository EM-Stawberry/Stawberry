package middleware

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	colorGreen  = "\033[32m"
	colorBlue   = "\033[34m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorReset  = "\033[0m"
)

func methodColor(method string) string {
	switch method {
	case "GET":
		return colorGreen + method + colorReset
	case "POST":
		return colorBlue + method + colorReset
	case "PUT":
		return colorYellow + method + colorReset
	case "DELETE":
		return colorRed + method + colorReset
	case "PATCH":
		return colorBlue + method + colorReset
	default:
		return method
	}
}

func statusCodeColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return colorGreen + fmt.Sprintf("%d", code) + colorReset
	case code >= 300 && code < 400:
		return colorBlue + fmt.Sprintf("%d", code) + colorReset
	case code >= 400 && code < 500:
		return colorYellow + fmt.Sprintf("%d", code) + colorReset
	default:
		return colorRed + fmt.Sprintf("%d", code) + colorReset
	}
}

var ginRoutesRegex = regexp.MustCompile(`(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS|CONNECT|TRACE)\s+(.+)\s+--> (.+) \((\d+) handlers\)`)

func formatGinDebugMessage(s string) string {
	if matches := ginRoutesRegex.FindStringSubmatch(s); len(matches) == 5 {
		method := matches[1]
		path := matches[2]
		handler := matches[3]

		handlerParts := strings.Split(handler, ".")
		shortHandler := handlerParts[len(handlerParts)-1]

		// Начало сообщения с методом
		prefix := fmt.Sprintf("Route: %s ", methodColor(method))
		
		// Вычисляем фактическую длину без ANSI-кодов
		visibleLength := len("Route: " + method + " " + path)
		
		// Фиксированная позиция стрелки
		arrowPosition := 50
		
		// Вычисляем сколько пробелов нужно добавить
		spacing := arrowPosition - visibleLength
		if spacing < 1 {
			spacing = 1
		}
		
		return prefix + path + strings.Repeat(" ", spacing) + "→ " + shortHandler
	}

	s = strings.TrimPrefix(s, "[GIN-debug] ")

	if strings.HasPrefix(s, "Listening and serving HTTP") {
		return "Server started: " + s
	}

	if strings.HasPrefix(s, "redirecting request") {
		return "Redirect: " + s
	}

	if strings.HasPrefix(s, "Loading HTML Templates") {
		return "Templates loaded: " + s
	}

	if strings.Contains(s, "router") {
		return "Router: " + s
	}

	return s
}

type zapWriter struct {
	logger *zap.Logger
}

func (w zapWriter) Write(p []byte) (n int, err error) {
	s := strings.TrimSpace(string(p))

	if strings.Contains(s, "[GIN-debug]") {
		message := formatGinDebugMessage(s)
		w.logger.Debug(message, zap.String("component", "gin"))
	} else if strings.Contains(s, "[GIN]") {
		message := strings.TrimSpace(strings.Replace(s, "[GIN]", "", 1))
		w.logger.Debug(message, zap.String("component", "gin"))
	} else {
		w.logger.Debug(s)
	}

	return len(p), nil
}

func SetupGinWithZap(logger *zap.Logger) {
	gin.DefaultWriter = &zapWriter{logger: logger}
	gin.DefaultErrorWriter = &zapWriter{logger: logger.WithOptions(zap.IncreaseLevel(zapcore.ErrorLevel))}
}

func ZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		status := c.Writer.Status()
		method := c.Request.Method
		ip := c.ClientIP()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if len(query) > 0 {
			path = path + "?" + query
		}

		message := fmt.Sprintf("%s %-5s %12s %s",
			methodColor(method),
			statusCodeColor(status),
			latency.String(),
			path,
		)

		fields := []zap.Field{zap.String("ip", ip)}
		
		if errorMessage != "" {
			fields = append(fields, zap.String("error", errorMessage))
		}

		switch {
		case status >= 500:
			logger.Error(message, fields...)
		case status >= 400:
			logger.Warn(message, fields...)
		default:
			logger.Info(message, fields...)
		}
	}
}

func ZapRecovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
