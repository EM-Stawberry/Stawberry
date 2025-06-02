package repository

import (
	"context"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type AuditRepository struct {
	db *sqlx.DB
}

func (ar *AuditRepository) LogStore(ae entity.AuditEntry) error {
	insertAuditLogQ, args := squirrel.Insert("audit_logs").
		Columns(
			"method",
			"url",
			"resp_status",
			"user_ip",
			"user_id",
			"user_role",
			"received_at",
			"req_body",
			"resp_body").
		Values(
			ae.Method,
			ae.Url,
			ae.RespStatus,
			ae.IP,
			ae.UserID,
			ae.UserRole,
			ae.ReceivedAt,
			ae.ReqBody,
			ae.RespBody).
		PlaceholderFormat(squirrel.Dollar).
		MustSql()

	ctxTO, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	_, err := ar.db.ExecContext(ctxTO, insertAuditLogQ, args...)
	if err != nil {
		return err
	}

	return nil
}

func NewAuditRepository(db *sqlx.DB) *AuditRepository {
	return &AuditRepository{db: db}
}
