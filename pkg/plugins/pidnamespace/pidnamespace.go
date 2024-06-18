package pidnamespace

import (
	"syscall"

	"github.com/mitchellh/go-ps"
	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "pidnamespace"
	bucketDescription = "PIDnamespace analyses the PID namespace of the container in the context of Kubernetes."
)

var bucketAliases = []string{"pidnamespaces", "pidns"}

type Bucket struct{}

func (n Bucket) Run() (bucket.Results, error) {
	deviceNumber, kubeletFound, pauseFound, err := getPIDNamespaceInfo()
	if err != nil {
		return bucket.Results{}, err
	}

	res := bucket.NewResults(bucketName)
	res.SetHeaders([]string{"deviceNumber", "pauseFound", "kubeletFound"})

	if pauseFound {
		res.AddComment("The pause process was found, pod might have shareProcessNamespace to true.")
	}
	if kubeletFound {
		res.AddComment("The kubelet process was found, pod might have hostPID to true.")
	}
	res.AddContent([]interface{}{deviceNumber, pauseFound, kubeletFound})

	return *res, nil
}

func Register(b *bucket.Buckets) {
	b.Register(bucket.Bucket{
		Name:        bucketName,
		Description: bucketDescription,
		Aliases:     bucketAliases,
		Factory: func(config bucket.Config) (bucket.Interface, error) {
			return NewPIDNamespaceBucket(config)
		},
		SideEffects:   false,
		RequireClient: false,
	})
}

func NewPIDNamespaceBucket(_ bucket.Config) (*Bucket, error) {
	return &Bucket{}, nil
}

func getPIDNamespaceInfo() (deviceNumber int, kubeletFound bool, pauseFound bool, err error) {
	// Get device number indicator
	file := "/proc/1/ns/pid"
	// Use Lstat to not follow the symlink.
	var info syscall.Stat_t
	if err := syscall.Lstat(file, &info); err != nil {
		return 0, false, false, err
	}

	deviceNumber = int(info.Dev)

	processes, err := ps.Processes()
	if err != nil {
		return 0, false, false, err
	}

	for i := range processes {
		kubeletFound = kubeletFound || processes[i].Executable() == "kubelet"
		pauseFound = pauseFound || processes[i].Executable() == "pause"
	}

	return
}
