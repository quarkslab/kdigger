package mount

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "mount"
	bucketDescription = "Mount shows all mounted devices in the container."

	mountPath = "/proc/mounts"
)

var bucketAliases = []string{"mounts", "mn"}

type Bucket struct{}

func (m Bucket) Run() (bucket.Results, error) {
	values, err := Mounts()
	if err != nil {
		return bucket.Results{}, err
	}
	res := bucket.NewResults(bucketName)
	res.SetHeaders([]string{"device", "path", "filesystem", "flags"})
	for _, m := range values {
		res.AddContent([]interface{}{m.Device, m.Path, m.Filesystem, m.Flags})
	}
	res.AddComment(fmt.Sprintf("%d devices are mounted.", len(values)))
	return *res, nil
}

func Register(b *bucket.Buckets) {
	b.Register(bucket.Bucket{
		Name:        bucketName,
		Description: bucketDescription,
		Aliases:     bucketAliases,
		Factory: func(config bucket.Config) (bucket.Interface, error) {
			return NewMountBucket(config)
		},
		SideEffects:   false,
		RequireClient: false,
	})
}

func NewMountBucket(_ bucket.Config) (*Bucket, error) {
	return &Bucket{}, nil
}

type Mount struct {
	Device     string
	Path       string
	Filesystem string
	Flags      string
}

func Mounts() ([]Mount, error) {
	file, err := os.Open(mountPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var mounts []Mount
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		scanner.Text()
		parts := strings.SplitN(scanner.Text(), " ", 5)
		if len(parts) != 5 {
			return nil, syscall.EIO
		}
		mounts = append(mounts, Mount{parts[0], parts[1], parts[2], parts[3]})
	}
	return mounts, nil
}
