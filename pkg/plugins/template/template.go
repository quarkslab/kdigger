package template

import (
	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "template"
	bucketDescription = "Template provides a template to write buckets."
)

var bucketAliases = []string{"templ", "tp"}

type Bucket struct{}

func (n Bucket) Run() (bucket.Results, error) {
	// executes here the code of your plugin
	res := bucket.NewResults(bucketName)
	return *res, nil
}

// Register registers a plugin
func Register(b *bucket.Buckets) {
	b.Register(bucket.Bucket{
		Name:        bucketName,
		Description: bucketDescription,
		Aliases:     bucketAliases,
		Factory: func(config bucket.Config) (bucket.Interface, error) {
			return NewTemplateBucket(config)
		},
		SideEffects:   false,
		RequireClient: false,
	})
}

func NewTemplateBucket(config bucket.Config) (*Bucket, error) {
	return &Bucket{}, nil
}
