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
		res.AddComment("User namespace is active.")
	} else {
		res.AddComment("User namespace is not active.")
	}

	return *res, nil
}

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

func NewUserNamespaceBucket(config bucket.Config) (*UserNamespaceBucket, error) {
	return &UserNamespaceBucket{}, nil
}
