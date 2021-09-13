package usernamespace

import (
	"github.com/genuinetools/bpfd/proc"
	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "usernamespace"
	bucketDescription = "UserNamespace analyses the user namespace configuration."
)

var bucketAliases = []string{"usernamespaces", "userns"}

type UserNamespaceBucket struct{}

func (n UserNamespaceBucket) Run() (bucket.Results, error) {
	userNS, userMapping := proc.GetUserNamespaceInfo(0)

	res := bucket.NewResults(bucketName)
	res.SetHeaders([]string{"containerID", "hostID", "range"})
	for _, m := range userMapping {
		res.AddContent([]interface{}{m.ContainerID, m.HostID, m.Range})
	}
	if userNS {
		res.SetComment("user namespace is active")
	} else {
		res.SetComment("user namespace is not active")
	}

	return *res, nil
}

// Register registers a bucket
func Register(b *bucket.Buckets) {
	b.Register(bucketName, bucketAliases, bucketDescription, false, func(config bucket.Config) (bucket.Interface, error) {
		return NewUserNamespaceBucket(config)
	})
}

func NewUserNamespaceBucket(config bucket.Config) (*UserNamespaceBucket, error) {
	return &UserNamespaceBucket{}, nil
}
