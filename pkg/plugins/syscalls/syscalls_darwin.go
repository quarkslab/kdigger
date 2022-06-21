package syscalls

import (
	"errors"

	"github.com/quarkslab/kdigger/pkg/bucket"
)

func (n SyscallsBucket) Run() (bucket.Results, error) {
	return bucket.Results{}, errors.New("syscall scan is not supported on macOS x86")
}
