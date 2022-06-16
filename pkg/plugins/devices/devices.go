package devices

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "devices"
	bucketDescription = "Devices shows the list of devices available in the container."
)

var bucketAliases = []string{"device", "dev"}

type DevicesBucket struct{}

func (n DevicesBucket) Run() (bucket.Results, error) {
	// executes here the code of your plugin
	devs, err := readDev()
	if err != nil {
		return bucket.Results{}, err
	}

	res := bucket.NewResults(bucketName)
	res.SetHeaders([]string{"mode", "isDir", "modTime", "name"})
	for _, dev := range devs {
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

func NewDevicesBucket(config bucket.Config) (*DevicesBucket, error) {
	return &DevicesBucket{}, nil
}

func readDev() ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir("/dev")
	if err != nil {
		return nil, err
	}
	return files, nil
}
