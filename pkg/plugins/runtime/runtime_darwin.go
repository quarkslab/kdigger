package runtime

import (
	"errors"

	"github.com/quarkslab/kdigger/pkg/bucket"
)

func (n Bucket) Run() (bucket.Results, error) {
	return bucket.Results{}, errors.New("runtime detection is not supported on macOS x86")
}
