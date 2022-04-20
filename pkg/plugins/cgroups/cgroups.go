package cgroups

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "cgroups"
	bucketDescription = "Cgroups reads the /proc/self/cgroup files that can leak information under cgroups v1."
)

var bucketAliases = []string{"cgroup", "cg"}

type CgroupsBucket struct{}

type Cgroup struct {
	HierarchyID    string
	ControllerList string
	CgroupPath     string
}

func (n CgroupsBucket) Run() (bucket.Results, error) {
	// executes here the code of your plugin
	cgroups, err := readCgroupFile()
	if err != nil {
		return bucket.Results{}, err
	}

	res := bucket.NewResults(bucketName)
	if len(cgroups) <= 1 {
		res.SetComment("This kernel might use cgroups v2, thus explaning the lack of information.")
	}
	res.SetHeaders([]string{"hierarchyID", "controllerList", "cgroupPath"})
	for _, cgroup := range cgroups {
		res.AddContent([]interface{}{cgroup.HierarchyID, cgroup.ControllerList, cgroup.CgroupPath})
	}
	return *res, nil
}

// Register registers a plugin
func Register(b *bucket.Buckets) {
	b.Register(bucketName, bucketAliases, bucketDescription, false, func(config bucket.Config) (bucket.Interface, error) {
		return NewCgroupsBucket(config)
	})
}

func NewCgroupsBucket(config bucket.Config) (*CgroupsBucket, error) {
	return &CgroupsBucket{}, nil
}

func readCgroupFile() ([]Cgroup, error) {
	file, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := []Cgroup{}
	for scanner.Scan() {
		sCgroup := strings.Split(scanner.Text(), ":")

		if len(sCgroup) < 3 {
			return nil, errors.New("format of /proc/self/cgroup file is incorrect, missing colons")
		}
		cgroup := Cgroup{}
		cgroup.HierarchyID = sCgroup[0]
		cgroup.ControllerList = sCgroup[1]
		cgroup.CgroupPath = sCgroup[2]

		lines = append(lines, cgroup)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
