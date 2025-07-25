package couchbase

import (
	"context"
	"microservice/domain"
	"time"

	"github.com/couchbase/gocb/v2"
	"go.uber.org/zap"
)

type CouchbaseRepository struct {
	cluster *gocb.Cluster
	bucket  *gocb.Bucket
}

func NewCouchbaseRepository() *CouchbaseRepository {

	cluster, err := gocb.Connect("couchbase://127.0.0.1", gocb.ClusterOptions{
		TimeoutsConfig: gocb.TimeoutsConfig{
			ConnectTimeout: 3 * time.Second,
			KVTimeout:      3 * time.Second,
			QueryTimeout:   3 * time.Second,
		},
		Authenticator: gocb.PasswordAuthenticator{
			Username: "Administrator",
			Password: "123456789",
		},
		Transcoder: gocb.NewJSONTranscoder(),
	})
	if err != nil {
		zap.L().Fatal("failed connect to couchbase", zap.Error(err))
	}
	bucket := cluster.Bucket("products")
	bucket.WaitUntilReady(3*time.Second, &gocb.WaitUntilReadyOptions{})

	return &CouchbaseRepository{
		cluster: cluster,
		bucket:  bucket,
	}
}

func (r *CouchbaseRepository) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	data, err := r.bucket.DefaultCollection().Get(id, &gocb.GetOptions{
		Timeout: 3 * time.Second,
		Context: ctx,
	})
	if err != nil {
		return nil, err
	}

	var product domain.Product
	if err := data.Content(&product); err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *CouchbaseRepository) CreateProduct(ctx context.Context, product *domain.Product) error {
	_, err := r.bucket.DefaultCollection().Insert(product.ID, product, &gocb.InsertOptions{
		Timeout: 3 * time.Second,
		Context: ctx,
	})
	return err
}
