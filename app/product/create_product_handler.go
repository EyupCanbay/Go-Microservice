package product

import "microservice/infra/couchbase"

type CreateProductRequest struct {
}

type CreateProductResponse struct {
	ID string `json:"id"`
}

type CreateProductHandler struct {
}