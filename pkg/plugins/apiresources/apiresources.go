package apiresources

import (
	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "apiresources"
	bucketDescription = "APIResources discovers the available APIs of the cluster."
)

var bucketAliases = []string{"api", "apiresource"}

type Bucket struct {
	config bucket.Config
}

func (n Bucket) Run() (bucket.Results, error) {
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

func Register(b *bucket.Buckets) {
	b.Register(bucket.Bucket{
		Name:        bucketName,
		Description: bucketDescription,
		Aliases:     bucketAliases,
		Factory: func(config bucket.Config) (bucket.Interface, error) {
			return NewAPIResourcesBucket(config)
		},
		SideEffects:   false,
		RequireClient: true,
	})
}

func NewAPIResourcesBucket(config bucket.Config) (*Bucket, error) {
	if config.Client == nil {
		return nil, bucket.ErrMissingClient
	}
	return &Bucket{
		config: config,
	}, nil
}
