package model

import (
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type AuditEntry struct {
	Method     string                 `db:"method"`
	Url        string                 `db:"url"`
	RespStatus int                    `db:"resp_status"`
	UserID     uint                   `db:"user_id"`
	IP         string                 `db:"user_ip"`
	UserRole   string                 `db:"user_role"`
	ReceivedAt time.Time              `db:"received_at"`
	ReqBody    map[string]interface{} `db:"req_body"`
	RespBody   map[string]interface{} `db:"resp_body"`
}

func (a *AuditEntry) toEntity() entity.AuditEntry {
	return entity.AuditEntry{
		Method:     a.Method,
		Url:        a.Url,
		RespStatus: a.RespStatus,
		UserID:     a.UserID,
		IP:         a.IP,
		UserRole:   a.UserRole,
		ReceivedAt: a.ReceivedAt,
		ReqBody:    a.ReqBody,
		RespBody:   a.RespBody,
	}
}
