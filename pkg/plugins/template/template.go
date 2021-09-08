package template

import (
	"github.com/mtardy/kdigger/pkg/bucket"
)

const (
	bucketName        = "template"
	bucketDescription = "Template provides a template to write buckets."
)

var bucketAliases = []string{"templ", "tp"}

type TemplateBucket struct{}

func (n TemplateBucket) Run() (bucket.Results, error) {
	// executes here the code of your plugin
	res := bucket.NewResults(bucketName)
	return *res, nil
}

// Register registers a plugin
func Register(b *bucket.Buckets) {
	b.Register(bucketName, bucketAliases, bucketDescription, false, func(config bucket.Config) (bucket.Interface, error) {
		return NewTemplateBucket(config)
	})
}

func NewTemplateBucket(config bucket.Config) (*TemplateBucket, error) {
	return &TemplateBucket{}, nil
}
