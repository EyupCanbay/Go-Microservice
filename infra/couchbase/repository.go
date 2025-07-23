package couchbase

import (
	"context"
	"microservice/app/product"

	"github.com/couchbase/gocb/v2"
)

type CouchbaseRepository struct {
}

func NewCouchbaseRepository() *CouchbaseRepository {

	cluster, err := gocb.Connect("couchbasehttp://localhost",gocb.ClusterOpinions{
		Authanticator: gocb.Authenticator{
			Username: ""
		}
	})
	return &CouchbaseRepository{}
}

func (r *CouchbaseRepository) GetProduct(ctx context.Context, id string) (*product.Product, error)