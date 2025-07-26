package product

import (
	"context"
	"errors"
	"microservice/domain"

	"github.com/couchbase/gocb/v2"
)

type GetProductRequest struct {
	ID string `json:"id" param:"id"`
}

type GetProductResponse struct {
	Product *domain.Product `json"product"`
}

type GetProductHandler struct {
	repository Repository
}

func NewGetProductHandler(repository Repository) *GetProductHandler {
	return &GetProductHandler{
		repository: repository,
	}
}

func (h *GetProductHandler) Handle(ctx context.Context, req *GetProductRequest) (*GetProductResponse, error) {
	product, err := h.repository.GetProduct(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gocb.ErrDocumentNotFound) {
			return nil, errors.New("key not found")
		}
		return nil, err
	}
	return &GetProductResponse{Product: product}, nil
}
