package services

import (
	"fmt"
	"net"

	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "services"
	bucketDescription = "Services uses CoreDNS wildcards feature to discover every service available in the cluster."
)

var bucketAliases = []string{"service", "svc"}

type Bucket struct{}

func (n Bucket) Run() (bucket.Results, error) {
	res := bucket.NewResults(bucketName)
	_, hosts, err := net.LookupSRV("", "", "any.any.svc.cluster.local")
	if err != nil {
		return *res, fmt.Errorf("error when requesting coreDNS: %w", err)
	}
	res.SetHeaders([]string{"service", "port"})
	for _, svc := range hosts {
		res.AddContent([]interface{}{svc.Target, svc.Port})
	}
	return *res, nil
}

func Register(b *bucket.Buckets) {
	b.Register(bucket.Bucket{
		Name:        bucketName,
		Description: bucketDescription,
		Aliases:     bucketAliases,
		Factory: func(config bucket.Config) (bucket.Interface, error) {
			return NewServicesBucket(config)
		},
		SideEffects:   false,
		RequireClient: false,
	})
}

func NewServicesBucket(_ bucket.Config) (*Bucket, error) {
	return &Bucket{}, nil
}
