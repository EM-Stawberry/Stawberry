package audit

import (
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type AuditRepository interface {
	LogStore(entity.AuditEntry) error
}

type AuditService struct {
	auditRepository AuditRepository
}

func NewAuditService(ar AuditRepository) *AuditService {
	return &AuditService{
		auditRepository: ar,
	}
}

func (as *AuditService) Log(ae entity.AuditEntry) error {
	return as.auditRepository.LogStore(ae)
}
