package devices

import (
	"fmt"
	"os"
	"time"

	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "devices"
	bucketDescription = "Devices shows the list of devices available in the container."
)

var bucketAliases = []string{"device", "dev"}

type Bucket struct{}

func (n Bucket) Run() (bucket.Results, error) {
	// executes here the code of your plugin
	devs, err := readDev()
	if err != nil {
		return bucket.Results{}, err
	}

	res := bucket.NewResults(bucketName)
	res.SetHeaders([]string{"mode", "isDir", "modTime", "name"})
	for _, devDir := range devs {
		dev, err := devDir.Info()
		if err != nil {
			// file was removed or renamed, ignore this edge case
			continue
		}
		res.AddContent([]interface{}{dev.Mode().String(), dev.IsDir(), dev.ModTime().Format(time.RFC3339), dev.Name()})
	}
	res.AddComment(fmt.Sprintf("%d devices are available.", len(devs)))
	return *res, nil
}

func Register(b *bucket.Buckets) {
	b.Register(bucket.Bucket{
		Name:        bucketName,
		Description: bucketDescription,
		Aliases:     bucketAliases,
		Factory: func(config bucket.Config) (bucket.Interface, error) {
			return NewDevicesBucket(config)
		},
		SideEffects:   false,
		RequireClient: false,
	})
}

func NewDevicesBucket(_ bucket.Config) (*Bucket, error) {
	return &Bucket{}, nil
}

func readDev() ([]os.DirEntry, error) {
	files, err := os.ReadDir("/dev")
	if err != nil {
		return nil, err
	}
	return files, nil
}
