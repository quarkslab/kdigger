package usernamespace

import (
	"errors"

	"github.com/quarkslab/kdigger/pkg/bucket"
)

func (n UserNamespaceBucket) Run() (bucket.Results, error) {
	return bucket.Results{}, errors.New("usernamespace detection is not supported on macOS x86")
}
