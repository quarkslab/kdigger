package runtime

import (
	"fmt"

	"github.com/genuinetools/bpfd/proc"
	"github.com/mtardy/kdigger/pkg/bucket"
)

const (
	bucketName        = "runtime"
	bucketDescription = "Runtime finds clues to identify which container runtime is running the container."
)

var runtimeAliases = []string{"runtimes", "rt"}

type RuntimeBucket struct{}

func (n RuntimeBucket) Run() (bucket.Results, error) {
	runtime := proc.GetContainerRuntime(0, 0)
	res := bucket.NewResults(bucketName)
	res.SetComment(fmt.Sprintf("Container runtime %q", runtime))
	return *res, nil
}

// Register registers a plugin
func Register(b *bucket.Buckets) {
	b.Register(bucketName, runtimeAliases, bucketDescription, false, func(config bucket.Config) (bucket.Interface, error) {
		return NewRuntimeBucket(config)
	})
}

func NewRuntimeBucket(config bucket.Config) (*RuntimeBucket, error) {
	return &RuntimeBucket{}, nil
}
