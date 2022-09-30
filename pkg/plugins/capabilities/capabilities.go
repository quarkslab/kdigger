package capabilities

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/quarkslab/kdigger/pkg/bucket"
	"github.com/syndtr/gocapability/capability"
)

const (
	bucketName        = "capabilities"
	bucketDescription = "Capabilities lists all capabilities in all sets and displays dangerous capabilities in red."
)

var bucketAliases = []string{"capability", "cap"}

var dangerousCap = []capability.Cap{
	capability.CAP_CHOWN,
	capability.CAP_DAC_OVERRIDE,
	capability.CAP_DAC_READ_SEARCH,
	capability.CAP_SETUID,
	capability.CAP_SETGID,
	capability.CAP_NET_RAW,
	capability.CAP_SYS_ADMIN,
	capability.CAP_SYS_PTRACE,
	capability.CAP_SYS_MODULE,
	capability.CAP_FOWNER,
	capability.CAP_SETFCAP,
}

type Bucket struct{}

func (n Bucket) Run() (bucket.Results, error) {
	capabilities, err := getCapabilities(0)

	if err != nil {
		return bucket.Results{}, err
	}

	res := bucket.NewResults(bucketName)
	res.SetHeaders([]string{"set", "capabilities"})
	var colors text.Colors
	colors = append(colors, text.FgRed)
	for set, caps := range capabilities {
		var sCaps = []string{}
		for _, cap := range caps {
			if isDangerousCap(cap) {
				sCaps = append(sCaps, colors.Sprint(cap))
			} else {
				sCaps = append(sCaps, cap.String())
			}
		}
		res.AddContent([]interface{}{set.String(), sCaps})
	}

	if isPrivileged(capabilities[capability.BOUNDING]) {
		res.AddComment(fmt.Sprintf("The bounding set contains %d caps and you have CAP_SYS_ADMIN, you might be running a privileged container, check the number of devices available.", len(capabilities[capability.BOUNDING])))
	} else {
		res.AddComment(fmt.Sprintf("The bounding set contains %d caps, it seems that you are running a non-privileged container.", len(capabilities[capability.BOUNDING])))
	}

	noNewPrivs, err := readNoNewPrivsFlag()
	if err != nil {
		// this is an additional feature, do not "error" on this
		res.AddComment(fmt.Sprintf("error reading the NoNewPrivs flag: %s", err.Error()))
	} else {
		res.AddComment(fmt.Sprintf("NoNewPrivs flag is set to %t.", noNewPrivs))
	}

	return *res, nil
}

func Register(b *bucket.Buckets) {
	b.Register(bucket.Bucket{
		Name:        bucketName,
		Description: bucketDescription,
		Aliases:     bucketAliases,
		Factory: func(config bucket.Config) (bucket.Interface, error) {
			return NewCapabilitiesBucket(config)
		},
		SideEffects:   false,
		RequireClient: false,
	})
}

func NewCapabilitiesBucket(config bucket.Config) (*Bucket, error) {
	if !config.Color {
		text.DisableColors()
	}
	return &Bucket{}, nil
}

func isDangerousCap(cap capability.Cap) bool {
	for _, dCap := range dangerousCap {
		if cap == dCap {
			return true
		}
	}
	return false
}

// isPrivileged just detect CAP_SYS_ADMIN that might be characteristic of a
// privileged container. It is necessary but not sufficient.
func isPrivileged(caps []capability.Cap) bool {
	for _, c := range caps {
		if c == capability.CAP_SYS_ADMIN {
			return true
		}
	}
	return false
}

// getCapabilities returns the allowed capabilities for the process.
// If pid is less zero, it returns the capabilities for "self".
func getCapabilities(pid int) (map[capability.CapType][]capability.Cap, error) {
	allCaps := capability.List()

	caps, err := capability.NewPid2(pid)
	if err != nil {
		return nil, err
	}
	err = caps.Load()
	if err != nil {
		return nil, err
	}

	allowedCaps := map[capability.CapType][]capability.Cap{}
	allowedCaps[capability.EFFECTIVE] = []capability.Cap{}
	allowedCaps[capability.PERMITTED] = []capability.Cap{}
	allowedCaps[capability.INHERITABLE] = []capability.Cap{}
	allowedCaps[capability.BOUNDING] = []capability.Cap{}
	allowedCaps[capability.AMBIENT] = []capability.Cap{}

	for _, cap := range allCaps {
		if caps.Get(capability.EFFECTIVE, cap) {
			allowedCaps[capability.EFFECTIVE] = append(allowedCaps[capability.EFFECTIVE], cap)
		}
		if caps.Get(capability.PERMITTED, cap) {
			allowedCaps[capability.PERMITTED] = append(allowedCaps[capability.PERMITTED], cap)
		}
		if caps.Get(capability.INHERITABLE, cap) {
			allowedCaps[capability.INHERITABLE] = append(allowedCaps[capability.INHERITABLE], cap)
		}
		if caps.Get(capability.BOUNDING, cap) {
			allowedCaps[capability.BOUNDING] = append(allowedCaps[capability.BOUNDING], cap)
		}
		if caps.Get(capability.AMBIENT, cap) {
			allowedCaps[capability.AMBIENT] = append(allowedCaps[capability.AMBIENT], cap)
		}
	}

	return allowedCaps, nil
}

func readNoNewPrivsFlag() (bool, error) {
	file, err := os.Open("/proc/self/status")
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "NoNewPrivs") {
			line := strings.Split(scanner.Text(), ":")
			if len(line) < 2 {
				return false, errors.New("error in /proc/self/status format, missing colons")
			}
			return strings.TrimSpace(line[1]) == "1", nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, errors.New("flag NoNewPrivs was not found in /proc/self/status")
}
