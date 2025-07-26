package product

import (
	"context"
	"microservice/domain"
)

// dependency inversion yapıldı
// app katmanı infraya bağımlı olmamalı bu yüzden interface oluşturulup**  CreateProductHandler kısmına bak pr
type Repository interface { 
	CreateProduct(ctx context.Context, product *domain.Product) error
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
}