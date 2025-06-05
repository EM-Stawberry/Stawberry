package repository

import (
	"context"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/repository/model"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type AuditRepository struct {
	db *sqlx.DB
}

func NewAuditRepository(db *sqlx.DB) *AuditRepository {
	return &AuditRepository{db: db}
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
			model.JSONBMap(ae.ReqBody),
			model.JSONBMap(ae.RespBody)).
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

func (ar *AuditRepository) GetLogs(
	ctx context.Context,
	fromT,
	toT time.Time,
	uid uint,
	limit,
	offset int,
) ([]entity.AuditEntry, int, error) {
	query := squirrel.Select(
		"method",
		"url",
		"resp_status",
		"user_ip",
		"user_id",
		"user_role",
		"received_at",
		"req_body",
		"resp_body",
		"count (*) over () as total_count",
	).From("audit_logs").
		Where(squirrel.And{
			squirrel.GtOrEq{"received_at": fromT},
			squirrel.LtOrEq{"received_at": toT},
		}).
		OrderBy("received_at DESC")

	if uid != 0 {
		query = query.Where(squirrel.Eq{"user_id": uid})
	}

	query = query.Limit(uint64(limit)).Offset(uint64(offset))

	sqlQuery, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, 0, err
	}

	var entries []model.AuditEntryMeta
	err = ar.db.SelectContext(ctx, &entries, sqlQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	logEntities, totalCount := model.ConvertAuditEntriesToEntity(entries)
	return logEntities, totalCount, nil
}
