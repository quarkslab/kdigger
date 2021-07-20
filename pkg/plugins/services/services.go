package services

import (
	"fmt"
	"net"

	"github.com/mtardy/kdigger/pkg/bucket"
)

const (
	bucketName        = "services"
	bucketDescription = "Services uses CoreDNS wildcards feature to discover every service available in the cluster."
)

var bucketAliases = []string{"service", "svc"}

type ServicesBucket struct{}

func (n ServicesBucket) Run() (bucket.Results, error) {
	res := bucket.NewResults(bucketName)
	_, hosts, err := net.LookupSRV("", "", "any.any.svc.cluster.local")
	if err != nil {
		res.SetComment(fmt.Sprintf("error when requesting DNS: %s", err.Error()))
		return *res, nil
	}
	res.SetHeaders([]string{"service", "port"})
	for _, svc := range hosts {
		res.AddContent([]interface{}{svc.Target, svc.Port})
	}
	return *res, nil
}

// Register registers a plugin
func Register(b *bucket.Buckets) {
	b.Register(bucketName, bucketAliases, bucketDescription, func(config bucket.Config) (bucket.Interface, error) {
		return NewServicesBucket(config)
	})
}

func NewServicesBucket(config bucket.Config) (*ServicesBucket, error) {
	return &ServicesBucket{}, nil
}
