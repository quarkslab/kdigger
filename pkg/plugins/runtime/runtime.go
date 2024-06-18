package runtime

import (
	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "runtime"
	bucketDescription = "Runtime finds clues to identify which container runtime is running the container."
)

var bucketAliases = []string{"runtimes", "rt"}

type Bucket struct{}

func Register(b *bucket.Buckets) {
	b.Register(bucket.Bucket{
		Name:        bucketName,
		Description: bucketDescription,
		Aliases:     bucketAliases,
		Factory: func(config bucket.Config) (bucket.Interface, error) {
			return NewRuntimeBucket(config)
		},
		SideEffects:   false,
		RequireClient: false,
	})
}

func NewRuntimeBucket(_ bucket.Config) (*Bucket, error) {
	return &Bucket{}, nil
}
