package version

import (
	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "version"
	bucketDescription = "Version dumps the API server version informations."
)

var bucketAliases = []string{"versions", "v"}

type VersionBucket struct {
	config bucket.Config
}

func (n VersionBucket) Run() (bucket.Results, error) {
	res := bucket.NewResults(bucketName)
	v, err := n.config.Client.Discovery().ServerVersion()
	if err != nil {
		return bucket.Results{}, err
	}
	res.SetHeaders([]string{"version", "buildDate", "platform", "goVersion"})
	res.AddContent([]interface{}{v.GitVersion, v.BuildDate, v.Platform, v.GoVersion})
	return *res, nil
}

// Register registers a plugin
func Register(b *bucket.Buckets) {
	b.Register(bucketName, bucketAliases, bucketDescription, false, func(config bucket.Config) (bucket.Interface, error) {
		return NewVersionBucket(config)
	})
}

func NewVersionBucket(config bucket.Config) (*VersionBucket, error) {
	if config.Client == nil {
		return nil, bucket.ErrMissingClient
	}
	return &VersionBucket{
		config: config,
	}, nil
}
