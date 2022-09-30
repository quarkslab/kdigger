package containerdetect

import (
	"bytes"
	"fmt"
	"os"
	"syscall"

	"github.com/mitchellh/go-ps"
	"github.com/quarkslab/kdigger/pkg/bucket"
	"github.com/quarkslab/kdigger/pkg/plugins/mount"
)

// this bucket follows a discussion on twitter
// https://twitter.com/g3rzi/status/1564594977220562945

const (
	bucketName        = "containerdetect"
	bucketDescription = "ContainerDetect retrieves hints that the process is running inside a typical container."
)

var bucketAliases = []string{"container", "cdetect"}

type Bucket struct{}

func (n Bucket) Run() (bucket.Results, error) {
	res := bucket.NewResults(bucketName)
	res.SetHeaders([]string{"hint", "result"})

	// there is a systemd with pid 1 on the host
	systemdFirstPID, err := isProcessPID("systemd", 1)
	if err != nil {
		return *res, err
	}
	res.AddContent([]interface{}{"systemd is not PID 1", !systemdFirstPID})

	// there is a kthreadd with pid 2 on the host
	kthreadSecondPID, err := isProcessPID("kthreadd", 2)
	if err != nil {
		return *res, err
	}
	res.AddContent([]interface{}{"kthreadd is not PID 2", !kthreadSecondPID})

	// the inode of root on the host should be 2
	root, err := os.Stat("/")
	if err != nil {
		return *res, err
	}
	if stat, ok := root.Sys().(*syscall.Stat_t); ok {
		res.AddContent([]interface{}{"inode number of root is not 2", !(stat.Ino == 2)})
	}

	// root as an overlay filesystem might imply running in a container
	mnts, err := mount.Mounts()
	if err != nil {
		return *res, err
	}
	isRootOverlayFS := false
	for _, mnt := range mnts {
		if mnt.Filesystem == "overlay" && mnt.Path == "/" {
			isRootOverlayFS = true
		}
	}
	res.AddContent([]interface{}{"root is an overlay fs", isRootOverlayFS})

	// /etc/fstab might be empty in a container
	isFstabEmpty, err := isFstabEmpty()
	if err != nil {
		return *res, err
	}
	res.AddContent([]interface{}{"/etc/fstab is empty", isFstabEmpty})

	// /boot might be empty in a container
	isBootEmpty, err := isBootFolderEmpty()
	if err != nil {
		return *res, err
	}
	res.AddContent([]interface{}{"/boot is empty", isBootEmpty})

	res.AddComment("A majority of true hints might imply running in a container.")

	return *res, nil
}

func isProcessPID(process string, pid int) (bool, error) {
	p, err := ps.FindProcess(pid)
	if err != nil {
		return false, fmt.Errorf("failed to find process %d: %w", pid, err)
	}
	if p != nil && p.Executable() == process {
		return true, nil
	}
	return false, nil
}

func isFstabEmpty() (bool, error) {
	file, err := os.ReadFile("/etc/fstab")
	if err != nil {
		return false, err
	}
	lines := bytes.Split(file, []byte("\n"))
	for _, line := range lines {
		if len(line) != 0 && line[0] != '#' {
			return false, nil
		}
	}
	return true, nil
}

func isBootFolderEmpty() (bool, error) {
	files, err := os.ReadDir("/boot")
	if err != nil {
		return false, err
	}
	return len(files) == 0, nil
}

// Register registers a plugin
func Register(b *bucket.Buckets) {
	b.Register(bucket.Bucket{
		Name:        bucketName,
		Description: bucketDescription,
		Aliases:     bucketAliases,
		Factory: func(config bucket.Config) (bucket.Interface, error) {
			return NewContainerDetectBucket(config)
		},
		SideEffects:   false,
		RequireClient: false,
	})
}

func NewContainerDetectBucket(config bucket.Config) (*Bucket, error) {
	return &Bucket{}, nil
}
