package namespaces

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/genuinetools/bpfd/proc"
	"github.com/mitchellh/go-ps"
	"github.com/mtardy/kdigger/pkg/bucket"
)

const (
	bucketName        = "namespaces"
	bucketDescription = "Namespaces analyses namespaces of the container in the context of Kubernetes."
)

var bucketAliases = []string{"namespace", "ns"}

type NamespacesBucket struct{}

type namespaceDetails struct {
	nsType  string
	active  bool
	details string
}

func (n NamespacesBucket) Run() (bucket.Results, error) {
	var details []namespaceDetails
	userNS, userMapping := proc.GetUserNamespaceInfo(0)
	details = append(details, namespaceDetails{nsType: "user", active: userNS, details: formatUserMapping(userMapping)})
	details = append(details, getPIDNamespaceInfo())

	res := bucket.NewResults(bucketName)
	res.SetHeaders([]string{"Namespace", "Active", "Details"})
	for _, d := range details {
		res.AddContent([]string{d.nsType, fmt.Sprint(d.active), d.details})
	}

	return *res, nil
}

// Register registers a bucket
func Register(b *bucket.Buckets) {
	b.Register(bucketName, bucketAliases, bucketDescription, func(config bucket.Config) (bucket.Interface, error) {
		return NewNamespacesBucket(config)
	})
}

func NewNamespacesBucket(config bucket.Config) (*NamespacesBucket, error) {
	return &NamespacesBucket{}, nil
}

func getPIDNamespaceInfo() namespaceDetails {
	// Get device number indicator
	file := "/proc/1/ns/pid"
	// Use Lstat to not follow the symlink.
	var info syscall.Stat_t
	if err := syscall.Lstat(file, &info); err != nil {
		return namespaceDetails{nsType: "pid", details: err.Error()}
	}

	res := fmt.Sprintf("%s device number is %d\n", file, info.Dev)

	processes, err := ps.Processes()
	if err != nil {
		return namespaceDetails{nsType: "pid", details: err.Error()}
	}

	kubeletFound := false
	pauseFound := false
	for i := range processes {
		kubeletFound = kubeletFound || processes[i].Executable() == "kubelet"
		pauseFound = pauseFound || processes[i].Executable() == "pause"
	}
	res += "Please note that false only means that kubelet was found in /proc"

	if pauseFound {
		res += "\nThe pause container was found, pod might have shareProcessNamespace"
	}

	return namespaceDetails{nsType: "pid", active: !kubeletFound, details: res}
}

func formatUserMapping(um []proc.UserMapping) string {
	if um == nil {
		return ""
	}
	var s string
	for _, m := range um {
		s += fmt.Sprintf("Container: %d  Host: %d  Range: %d\n", m.ContainerID, m.HostID, m.Range)
	}
	return strings.TrimSpace(s)
}
