package product

import (
	"context"
	"errors"
	"io"
	"microservice/domain"
	"net/http"

	"github.com/couchbase/gocb/v2"
	"go.uber.org/zap"
)

type GetProductRequest struct {
	ID string `json:"id" param:"id"`
}

type GetProductResponse struct {
	Product *domain.Product `json"product"`
}

type GetProductHandler struct {
	repository Repository
	httpClient http.Client
}

func NewGetProductHandler(repository Repository, httpClient *http.Client) *GetProductHandler {
	return &GetProductHandler{
		repository: repository,
		httpClient: *httpClient,
	}
}

func (h *GetProductHandler) Handle(ctx context.Context, req *GetProductRequest) (*GetProductResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.google.com", nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	zap.L().Info("google response", zap.String("body", string(body)))

	product, err := h.repository.GetProduct(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gocb.ErrDocumentNotFound) {
			return nil, errors.New("key not found")
		}
		return nil, err
	}
	return &GetProductResponse{Product: product}, nil
}
