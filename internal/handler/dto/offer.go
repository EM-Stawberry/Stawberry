package dto

import (
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type PostOfferReq struct {
	ProductID uint    `json:"product_id" binding:"required"`
	ShopID    uint    `json:"shop_id" binding:"required"`
	Price     float64 `json:"price" binding:"required,gte=0"`
	Currency  string  `json:"currency" binding:"required,iso4217"`
}

type PostOfferResp struct {
	ID uint `json:"id"`
}

func (po *PostOfferReq) ConvertToEntity() entity.Offer {
	return entity.Offer{
		Price:     po.Price,
		Currency:  po.Currency,
		ShopID:    po.ShopID,
		ProductID: po.ProductID,
	}
}

func ConvertToPostOfferResp(o entity.Offer) PostOfferResp {
	return PostOfferResp{ID: o.ID}
}

type PatchOfferStatusReq struct {
	Status string `json:"status" binding:"required"`
}

type PatchOfferStatusResp struct {
	NewStatus string `json:"new_status"`
}

func (p *PatchOfferStatusReq) ConvertToEntity() entity.Offer {
	return entity.Offer{
		Status: p.Status,
	}
}

func ConvertToPatchOfferStatusResp(o entity.Offer) PatchOfferStatusResp {
	return PatchOfferStatusResp{NewStatus: o.Status}
}
