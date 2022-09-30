package syscalls

import (
	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "syscalls"
	bucketDescription = "Syscalls scans most of the syscalls to detect which are blocked and allowed."
)

var bucketAliases = []string{"syscall", "sys"}

type Bucket struct{}

func Register(b *bucket.Buckets) {
	b.Register(bucket.Bucket{
		Name:        bucketName,
		Description: bucketDescription,
		Aliases:     bucketAliases,
		Factory: func(config bucket.Config) (bucket.Interface, error) {
			return NewSyscallsBucket(config)
		},
		SideEffects:   true,
		RequireClient: false,
	})
}

func NewSyscallsBucket(config bucket.Config) (*Bucket, error) {
	return &Bucket{}, nil
}
