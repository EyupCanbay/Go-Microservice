package product

import (
	"context"
	"microservice/domain"
)


type CreateProductRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateProductResponse struct {
	ID string `json:"id"`
}

//** oluşuturulan inteface burdaji repository e verilir
type CreateProductHandler struct {
	repository Repository
}

func NewCreateProductHandler(repository Repository) *CreateProductHandler {
	return &CreateProductHandler{
		repository: repository,
	}
}

func (h *CreateProductHandler) Handle(ctx context.Context, req *CreateProductRequest) (*CreateProductResponse, error) {
	product := &domain.Product{
		ID:   req.ID,
		Name: req.Name,
	}

	err := h.repository.CreateProduct(ctx, product)
	if err != nil {
		return nil, err
	}

	return &CreateProductResponse{ID: product.ID}, nil
}
