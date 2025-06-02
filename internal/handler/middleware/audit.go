package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/EM-Stawberry/Stawberry/config"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/handler/helpers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

const maxBodySize = 10 * 1024

var workerPoolSize int
var queueSize int

// 1) middleware collects data and forms an audit log entry
// 2) sends it to a worker for data sanitizing and inserting into the database
// 3) graceful shutdown through Close method that's being called at the end of main()

type AuditMiddleware struct {
	logChan chan entity.AuditEntry
	done    chan struct{}
	service AuditService
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

type AuditService interface {
	Log(entry entity.AuditEntry) error
}

func NewAuditMiddleware(cfg *config.AuditConfig, as AuditService) *AuditMiddleware {
	am := &AuditMiddleware{
		logChan: make(chan entity.AuditEntry, cfg.QueueSize),
		done:    make(chan struct{}),
		service: as,
	}

	for range cfg.WorkerPoolSize {
		go am.worker()
	}

	return am
}

func (am *AuditMiddleware) worker() {
	defer close(am.done)
	for entry := range am.logChan {
		sanitizeSensitiveData(entry.ReqBody)
		sanitizeSensitiveData(entry.RespBody)
		if err := am.service.Log(entry); err != nil {
			zap.L().Error("Audit log failure",
				zap.Error(err),
				zap.String("path", entry.Url),
				zap.Int("status", entry.RespStatus))
		}
	}
}

func (am *AuditMiddleware) Close() {
	close(am.logChan)
	<-am.done
}

func (am *AuditMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
			c.Next()
			return
		}

		receivedAt := time.Now()

		bodyBytes, _ := c.GetRawData()
		// truncate request body if too big
		if len(bodyBytes) > maxBodySize {
			bodyBytes = append(bodyBytes[:maxBodySize], []byte("... [TRUNCATED]")...)
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		usrID, _ := helpers.UserIDContext(c)

		reqBody := make(map[string]interface{})
		if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
			zap.L().Error("Failed to unmarshal request body", zap.Error(err))
		}

		respBodyBytes := blw.body.Bytes()
		// truncate response body if too big
		if len(respBodyBytes) > maxBodySize {
			respBodyBytes = append(respBodyBytes[:maxBodySize], []byte("... [TRUNCATED]")...)
		}

		respBody := make(map[string]interface{})
		if err := json.Unmarshal(respBodyBytes, &respBody); err != nil {
			zap.L().Error("Failed to unmarshal response body", zap.Error(err))
		}

		logE := entity.AuditEntry{
			Method:     c.Request.Method,
			Url:        c.Request.URL.Path,
			RespStatus: c.Writer.Status(),
			UserID:     usrID,
			IP:         c.ClientIP(),
			UserRole:   getRole(c),
			ReceivedAt: receivedAt,
			ReqBody:    reqBody,
			RespBody:   respBody,
		}

		select {
		case am.logChan <- logE:
		default:
			zap.L().Warn("Audit log channel full, dropping audit log entry",
				zap.String("path", logE.Url),
				zap.Int("status", logE.RespStatus))
		}
	}
}

func getRole(c *gin.Context) string {
	isStore, _ := helpers.UserIsStoreContext(c)
	if isStore {
		return "shop"
	}
	isAdmin, _ := helpers.UserIsAdminContext(c)
	if isAdmin {
		return "admin"
	}
	return "user"
}

func sanitizeSensitiveData(data map[string]interface{}) {
	sensitiveFields := []string{"password", "fingerprint", "refresh_token", "access_token"}
	for _, field := range sensitiveFields {
		if _, ok := data[field]; ok {
			data[field] = "[REDACTED]"
		}
	}
}
