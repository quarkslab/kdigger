package usernamespace

import (
	"github.com/genuinetools/bpfd/proc"
	"github.com/quarkslab/kdigger/pkg/bucket"
)

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
