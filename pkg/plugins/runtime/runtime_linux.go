package runtime

import (
	"fmt"

	"github.com/genuinetools/bpfd/proc"
	"github.com/quarkslab/kdigger/pkg/bucket"
)

func (n Bucket) Run() (bucket.Results, error) {
	runtime := proc.GetContainerRuntime(0, 0)
	res := bucket.NewResults(bucketName)
	res.AddComment(fmt.Sprintf("The container runtime seems to be %s.", runtime))
	return *res, nil
}
