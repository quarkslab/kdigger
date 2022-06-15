package apiresources

import (
	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "apiresources"
	bucketDescription = "APIResources discovers the available APIs of the cluster."
)

var bucketAliases = []string{"api", "apiresource"}

type APIResourcesBucket struct {
	config bucket.Config
}

func (n APIResourcesBucket) Run() (bucket.Results, error) {
	// executes here the code of your plugin
	res := bucket.NewResults(bucketName)

	lists, err := n.config.Client.Discovery().ServerPreferredResources()
	if err != nil {
		return bucket.Results{}, err
	}

	res.SetHeaders([]string{"kind", "apiVersion", "namespaced"})
	for _, group := range lists {
		for _, r := range group.APIResources {
			res.AddContent([]interface{}{
				r.Kind,
				group.GroupVersion,
				r.Namespaced,
			})
		}
	}

	return *res, nil
}

// Register registers a plugin
func Register(b *bucket.Buckets) {
	b.Register(bucketName, bucketAliases, bucketDescription, false, func(config bucket.Config) (bucket.Interface, error) {
		return NewAPIResourcesBucket(config)
	})
}

func NewAPIResourcesBucket(config bucket.Config) (*APIResourcesBucket, error) {
	if config.Client == nil {
		return nil, bucket.ErrMissingClient
	}
	return &APIResourcesBucket{
		config: config,
	}, nil
}
