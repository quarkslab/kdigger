package token

import (
	"os"

	"github.com/mtardy/kdigger/pkg/bucket"
)

const (
	bucketName        = "token"
	bucketDescription = "Token checks for the presence of a service account token in the filesystem."

	tokenPath = "/run/secrets/kubernetes.io/serviceaccount"
)

var bucketAliases = []string{"tokens", "tk"}

type TokenBucket struct{}

func (n TokenBucket) Run() (bucket.Results, error) {
	res := bucket.NewResults(bucketName)
	if tokenFolderExist() {
		res.SetComment("A service account token is mounted.")

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
		res.SetComment("No service account token was found in the local file system")
	}
	return *res, nil
}

func Register(b *bucket.Buckets) {
	b.Register(bucketName, bucketAliases, bucketDescription, func(config bucket.Config) (bucket.Interface, error) {
		return NewTokenBucket(config)
	})
}

func NewTokenBucket(c bucket.Config) (*TokenBucket, error) {
	return &TokenBucket{}, nil
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
