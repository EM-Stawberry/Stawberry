package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/EM-Stawberry/Stawberry/config"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/internal/handler/helpers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	maxBodySize = 10 * 1024
	tickrate    = time.Second * 5
)

type AuditMiddleware struct {
	cfg             *config.AuditConfig
	toFlushChan     chan []entity.AuditEntry
	logChan         chan entity.AuditEntry
	closeSignalChan chan struct{}
	service         AuditService
	wg              *sync.WaitGroup
	mutex           *sync.Mutex
	buffer          []entity.AuditEntry
	backupBuffer    []entity.AuditEntry
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
	Log(entries []entity.AuditEntry) error
}

func NewAuditMiddleware(cfg *config.AuditConfig, as AuditService) *AuditMiddleware {
	am := &AuditMiddleware{
		cfg:             cfg,
		toFlushChan:     make(chan []entity.AuditEntry), // unbuffered on purpose
		logChan:         make(chan entity.AuditEntry, cfg.QueueSize),
		closeSignalChan: make(chan struct{}, 1),
		service:         as,
		wg:              &sync.WaitGroup{},
		mutex:           &sync.Mutex{},
		buffer:          make([]entity.AuditEntry, 0, cfg.QueueSize),
		backupBuffer:    make([]entity.AuditEntry, 0, cfg.QueueSize),
	}

	am.wg.Add(cfg.WorkerPoolSize)
	for range cfg.WorkerPoolSize {
		go am.worker()
	}

	go am.flusher()

	return am
}

func (am *AuditMiddleware) swapAndFlush() {
	am.buffer, am.backupBuffer = am.backupBuffer[:0], am.buffer
	am.toFlushChan <- am.backupBuffer
}

func (am *AuditMiddleware) storeLogs(entries []entity.AuditEntry) {
	if err := am.service.Log(entries); err != nil {
		zap.L().Error("Failed to log audit entries", zap.Error(err))
	}
}

func (am *AuditMiddleware) flusher() {
	ticker := time.NewTicker(tickrate)
	defer ticker.Stop()

	for {
		select {
		case toFlush := <-am.toFlushChan:
			am.storeLogs(toFlush)
		case <-ticker.C:
			if len(am.buffer) > 0 {
				am.buffer, am.backupBuffer = am.backupBuffer[:0], am.buffer
				am.storeLogs(am.backupBuffer)
			}
		case <-am.closeSignalChan:
			return
		}
	}
}

func (am *AuditMiddleware) worker() {
	defer am.wg.Done()
	for entry := range am.logChan {
		sanitizeSensitiveData(entry.ReqBody)
		sanitizeSensitiveData(entry.RespBody)
		am.mutex.Lock()
		if len(am.buffer) > int(0.9*float64(am.cfg.QueueSize)) {
			zap.L().Warn("Audit log buffer full", zap.Int("size", len(am.buffer)))
			am.swapAndFlush()
		} else {
			am.buffer = append(am.buffer, entry)
		}
		am.mutex.Unlock()
	}
}

func (am *AuditMiddleware) Close() {
	close(am.logChan)
	am.wg.Wait()
	close(am.closeSignalChan)

	if len(am.buffer) > 0 {
		am.storeLogs(am.buffer)
	}
	if len(am.backupBuffer) > 0 {
		am.storeLogs(am.backupBuffer)
	}
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
	if data == nil {
		return
	}
	sensitiveFields := []string{"password", "fingerprint", "refresh_token", "access_token"}
	for _, field := range sensitiveFields {
		if _, ok := data[field]; ok {
			data[field] = "[REDACTED]"
		}
	}
}
