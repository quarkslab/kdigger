package node

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "node"
	bucketDescription = "Node retrieves various information in /proc about the current host."
)

var bucketAliases = []string{"nodes", "n"}

type Bucket struct{}

func (n Bucket) Run() (bucket.Results, error) {
	cpuinfo, err := readCPUInfo()
	if err != nil {
		return bucket.Results{}, err
	}

	meminfo, err := readMemInfo()
	if err != nil {
		return bucket.Results{}, err
	}

	kernelInfo, err := readKernelVersion()
	if err != nil {
		return bucket.Results{}, err
	}

	res := bucket.NewResults(bucketName)
	res.SetHeaders([]string{"cpuModel", "cpuCores", "memTotal", "memUsed", "kernelVersion", "kernelDetails"})
	res.AddContent([]interface{}{
		cpuinfo[0]["model name"],
		cpuinfo[0]["cpu cores"],
		toHuman(meminfo.Total),
		// for the following line, see `man free` for more explanation
		toHuman(meminfo.Total - meminfo.Free - meminfo.Buffer - meminfo.Cached - meminfo.SReclaimable),
		kernelInfo.Version,
		kernelInfo.Details,
	})
	return *res, nil
}

func Register(b *bucket.Buckets) {
	b.Register(bucket.Bucket{
		Name:        bucketName,
		Description: bucketDescription,
		Aliases:     bucketAliases,
		Factory: func(config bucket.Config) (bucket.Interface, error) {
			return NewNodeBucket(config)
		},
		SideEffects:   false,
		RequireClient: false,
	})
}

func NewNodeBucket(_ bucket.Config) (*Bucket, error) {
	return &Bucket{}, nil
}

type KernelVersion struct {
	Version string
	Details string
}

func readKernelVersion() (*KernelVersion, error) {
	// this file is a oneliner
	version, err := os.ReadFile("/proc/version")
	if err != nil {
		return nil, err
	}
	kernelVersion := &KernelVersion{}

	sVersion := strings.SplitN(string(version), " ", 4)
	if len(sVersion) < 4 {
		return nil, errors.New("error in /proc/version format, missing entries")
	}
	kernelVersion.Version = sVersion[2]
	kernelVersion.Details = strings.TrimSpace(sVersion[3])

	return kernelVersion, nil
}

// toHuman converts a value in kB from a uint64 to the most appropriate value
// the rounding is false but we don't really care
func toHuman(i uint64) string {
	units := []string{"Ki", "Mi", "Gi", "Ti", "Pi", "Ei"}
	f := float64(i)
	division := 0
	for ; f > 1024 && division < 5; division++ {
		f /= 1024
	}
	if f > 10 {
		return fmt.Sprintf("%.0f", f) + units[division]
	}
	return fmt.Sprintf("%.1f", f) + units[division]
}

type Meminfo struct {
	Total        uint64
	Free         uint64
	Available    uint64
	Buffer       uint64
	Cached       uint64
	SReclaimable uint64
}

func readMemInfo() (*Meminfo, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	meminfo := &Meminfo{}

	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ":")
		if len(line) < 2 {
			return nil, errors.New("error in /proc/meminfo format, missing colons")
		}
		value, err := strconv.ParseUint(strings.TrimSpace(strings.TrimSuffix(line[1], " kB")), 10, 64)
		if err != nil {
			return nil, err
		}
		switch strings.TrimSpace(line[0]) {
		case "MemTotal":
			meminfo.Total = value
		case "MemFree":
			meminfo.Free = value
		case "MemAvailable":
			meminfo.Available = value
		case "Buffer":
			meminfo.Buffer = value
		case "Cached":
			meminfo.Cached = value
		case "SReclaimable":
			meminfo.SReclaimable = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return meminfo, nil
}

func readCPUInfo() ([]map[string]string, error) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	cpus := []map[string]string{}
	cpu := map[string]string{}

	for scanner.Scan() {
		if scanner.Text() == "" {
			cpus = append(cpus, cpu)
			cpu = map[string]string{}
			continue
		}
		line := strings.Split(scanner.Text(), ":")
		if len(line) < 2 {
			return nil, errors.New("error in /proc/cpuinfo format, missing colons")
		}
		cpu[strings.TrimSpace(line[0])] = strings.TrimSpace(line[1])
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return cpus, nil
}
