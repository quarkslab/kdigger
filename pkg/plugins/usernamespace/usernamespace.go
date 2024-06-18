package usernamespace

import (
	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "usernamespace"
	bucketDescription = "UserNamespace analyses the user namespace configuration."
)

var bucketAliases = []string{"usernamespaces", "userns"}

type Bucket struct{}

func Register(b *bucket.Buckets) {
	b.Register(bucket.Bucket{
		Name:        bucketName,
		Description: bucketDescription,
		Aliases:     bucketAliases,
		Factory: func(config bucket.Config) (bucket.Interface, error) {
			return NewUserNamespaceBucket(config)
		},
		SideEffects:   false,
		RequireClient: false,
	})
}

func NewUserNamespaceBucket(_ bucket.Config) (*Bucket, error) {
	return &Bucket{}, nil
}
