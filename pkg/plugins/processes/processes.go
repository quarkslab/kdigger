package processes

import (
	"fmt"

	"github.com/mitchellh/go-ps"
	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "processes"
	bucketDescription = "Processes analyses the running processes in your PID namespace"
)

var bucketAliases = []string{"process", "ps"}

type ProcessesBucket struct{}

func (n ProcessesBucket) Run() (bucket.Results, error) {
	res := bucket.NewResults(bucketName)

	processes, err := ps.Processes()
	if err != nil {
		return *res, err
	}

	res.SetHeaders([]string{"pid", "ppid", "name"})
	var isSystemdFirst bool
	for _, proc := range processes {
		if proc.Pid() == 1 && proc.Executable() == "systemd" {
			isSystemdFirst = true
		}
		res.AddContent([]interface{}{proc.Pid(), proc.PPid(), proc.Executable()})
	}

	comment := fmt.Sprintf("%d processes running", len(processes))
	if isSystemdFirst {
		comment += ", systemd is the first process"
	} else {
		comment += ", systemd not found as the first process"
	}
	res.SetComment(comment)

	return *res, nil
}

// Register registers a plugin
func Register(b *bucket.Buckets) {
	b.Register(bucketName, bucketAliases, bucketDescription, false, func(config bucket.Config) (bucket.Interface, error) {
		return NewProcessesBucket(config)
	})
}

func NewProcessesBucket(config bucket.Config) (*ProcessesBucket, error) {
	return &ProcessesBucket{}, nil
}
