package token

import (
	"os"

	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "token"
	bucketDescription = "Token checks for the presence of a service account token in the filesystem."

	tokenPath = "/run/secrets/kubernetes.io/serviceaccount"
)

var bucketAliases = []string{"tokens", "tk"}

type Bucket struct{}

func (n Bucket) Run() (bucket.Results, error) {
	res := bucket.NewResults(bucketName)
	if tokenFolderExist() {
		res.AddComment("A service account token is mounted.")

		res.SetHeaders([]string{"namespace", "token", "CA"})

		ns, err := readMountedData("namespace")
		if err != nil {
			return bucket.Results{}, err
		}
		t, err := readMountedData("token")
		if err != nil {
			return bucket.Results{}, err
		}
		ca, err := readMountedData("ca.crt")
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

func tokenFolderExist() bool {
	_, err := os.Stat(tokenPath)
	return !os.IsNotExist(err)
}

func readMountedData(data string) (string, error) {
	b, err := os.ReadFile(tokenPath + "/" + data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
