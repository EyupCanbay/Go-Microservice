package couchbase

import (
	"context"
	"microservice/app/product"
	"time"

	"github.com/couchbase/gocb/v2"
	"go.uber.org/zap"
)

type CouchbaseRepository struct {
	cluster *gocb.Cluster
	bucket *gocb.Bucket
}

func NewCouchbaseRepository() *CouchbaseRepository {

	cluster, err := gocb.Connect("couchbasehttp://localhost",gocb.ClusterOpinions{
		TimeoutsConfig: gocb.TimeoutsConfig{
			ConnectTimeout: 3 * time.Second,
			KVTimeout: 3 * time.Second,
			QueryTimeout: 3 * time.Second,
		},
		Authanticator: gocb.PasswordAuthenticator{
			Username: "Administrator",
			Password: "password",
		},
		Transcoder: gocb.NewJSONTranscoder(),
	})
	if err != nil {
		zap.L().Fatal("failed connect to couchbase", zap.Error(err))
	}
	bucket := cluster.Bucket("products")
	bucket.WaitUntilReady(3 * time.Second, &gocb.WaitUntilReadyOptions{})


	return &CouchbaseRepository{
		cluster: cluster,
		bucket: bucket,
	}
}

func (r *CouchbaseRepository) GetProduct(ctx context.Context, id string) (*product.Product, error)