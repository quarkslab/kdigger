package token

import (
	"os"
	"path/filepath"

	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "token"
	bucketDescription = "Token checks for the presence of a service account token in the filesystem."
)

var tokenPaths = []string{"/run/secrets/kubernetes.io/serviceaccount", "/var/run/secrets/kubernetes.io/serviceaccount"}
var bucketAliases = []string{"tokens", "tk"}

type Bucket struct{}

func (n Bucket) Run() (bucket.Results, error) {
	res := bucket.NewResults(bucketName)
	if tokenPath, ok := tokenFolderLocation(); ok {
		res.AddComment("A service account token is mounted.")

		res.SetHeaders([]string{"namespace", "token", "CA"})

		ns, err := readMountedData(tokenPath, "namespace")
		if err != nil {
			return bucket.Results{}, err
		}
		t, err := readMountedData(tokenPath, "token")
		if err != nil {
			return bucket.Results{}, err
		}
		ca, err := readMountedData(tokenPath, "ca.crt")
		if err != nil {
			return bucket.Results{}, err
		}

		res.AddContent([]interface{}{ns, t, ca})
	} else {
		res.AddComment("No service account token was found in the local filesystem")
	}
	return *res, nil
}

func Register(b *bucket.Buckets) {
	b.Register(bucket.Bucket{
		Name:        bucketName,
		Description: bucketDescription,
		Aliases:     bucketAliases,
		Factory: func(config bucket.Config) (bucket.Interface, error) {
			return NewTokenBucket(config)
		},
		SideEffects:   false,
		RequireClient: false,
	})
}

func NewTokenBucket(_ bucket.Config) (*Bucket, error) {
	return &Bucket{}, nil
}

func tokenFolderLocation() (string, bool) {
	for _, path := range tokenPaths {
		_, err := os.Stat(path)
		if !os.IsNotExist(err) {
			return path, true
		}
	}
	return "", false
}

func readMountedData(tokenPath string, data string) (string, error) {
	b, err := os.ReadFile(filepath.Join(tokenPath, data))
	if err != nil {
		return "", err
	}
	return string(b), nil
}
