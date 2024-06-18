package kgen

import (
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	fuzz "github.com/google/gofuzz"
	"github.com/syndtr/gocapability/capability"
)

const (
	// The probability of capabilities to be "ALL" will be 1/fuzzCapabilityAllOrEmptyChances.
	fuzzCapabilityAllOrEmptyChances = 6
	fuzzNilChances                  = .5
	// Chances of putting zero in integers used for runAsUser, runAsGroup, etc.
	fuzzIntZeroChances = .2
)

// Fuzzing will pick a uniform integer between 0 and fuzzCapabilityRandomMaxLen - 1 for capabilities list
var fuzzCapabilityRandomMaxLen = len(capability.List())

type GenerateOpts struct {
	Name        string
	Image       string
	Namespace   string
	Command     []string
	Privileged  bool
	HostPath    bool
	HostPid     bool
	HostNetwork bool
	Tolerations bool
}

var intMutator = func(e *int64, c fuzz.Continue) {
	if c.Float64() >= fuzzIntZeroChances {
		*e = int64(c.Int31())
	} else {
		*e = 0
	}
}

var seccompMutator = func(e *v1.SeccompProfile, c fuzz.Continue) {
	supportedProfileValues := [...]string{"Localhost", "RuntimeDefault", "Unconfined"}
	n := c.Intn(len(supportedProfileValues))

	e.Type = v1.SeccompProfileType(supportedProfileValues[n])
	if e.Type == v1.SeccompProfileType(supportedProfileValues[0]) {
		// spec.securityContext.seccompProfile.localhostProfile: Required value: must be set when seccomp type is Localhost
		s := c.RandString()
		e.LocalhostProfile = &s
	}
}

func FuzzPodSecurityContext(sc **v1.PodSecurityContext) {
	f := fuzz.New().NilChance(fuzzNilChances).Funcs(
		func(e *v1.Sysctl, c fuzz.Continue) {
			// must have at most 253 characters and match regex ^([a-z0-9]([-_a-z0-9]*[a-z0-9])?[\./])*[a-z0-9]([-_a-z0-9]*[a-z0-9])?$
			e.Name = sysctls[c.Intn(len(sysctls))]
			c.Fuzz(&e.Value)
		},
		func(e *v1.PodFSGroupChangePolicy, c fuzz.Continue) {
			if c.Intn(2) == 0 {
				*e = v1.FSGroupChangeAlways
			} else {
				*e = v1.FSGroupChangeOnRootMismatch
			}
		},
		func(e *[]v1.Sysctl, c fuzz.Continue) {
			if c.Float64() >= fuzzNilChances {
				uniqSysctls := map[string]v1.Sysctl{}
				for n := c.Intn(10); len(uniqSysctls) < n; {
					var sysctl v1.Sysctl
					c.Fuzz(&sysctl)
					uniqSysctls[sysctl.Name] = sysctl
				}
				for _, v := range uniqSysctls {
					*e = append(*e, v)
				}
			} else {
				*e = []v1.Sysctl{}
			}
		},
		// let's ignore Windows for now
		func(e *v1.WindowsSecurityContextOptions, _ fuzz.Continue) {
			*e = v1.WindowsSecurityContextOptions{}
		},
		// for supplementalGroups, runAsUser and runAsGroup value that must be between 0 and 2147483647, inclusive
		intMutator,
		seccompMutator,
	)

	securityContext := &v1.PodSecurityContext{}

	f.Fuzz(securityContext)

	*sc = securityContext
}

// FuzzContainerSecurityContext will override the SecurityContext with random (valid) values
func FuzzContainerSecurityContext(sc **v1.SecurityContext) {
	// add .NilChange(0) to disable nil pointers generation, by default is 0.2
	f := fuzz.New().NilChance(fuzzNilChances).Funcs(
		func(e *v1.WindowsSecurityContextOptions, _ fuzz.Continue) {
			*e = v1.WindowsSecurityContextOptions{}
		},
		func(e *v1.Capability, c fuzz.Continue) {
			caps := capability.List()
			n := c.Intn(len(caps))
			*e = v1.Capability(strings.ToUpper(caps[n].String()))
		},
		func(e *[]v1.Capability, c fuzz.Continue) {
			if r := c.Intn(fuzzCapabilityAllOrEmptyChances); r == 0 {
				*e = []v1.Capability{"ALL"}
			} else if r == 1 {
				*e = []v1.Capability{}
			} else {
				length := c.Intn(fuzzCapabilityRandomMaxLen)
				for i := 0; i < length; i++ {
					var capa v1.Capability
					c.Fuzz(&capa)
					*e = append(*e, capa)
				}
			}
		},
		// for runAsUser and runAsGroup value that must be between 0 and 2147483647, inclusive
		intMutator,
		seccompMutator,
	)
	securityContext := &v1.SecurityContext{}
	f.Fuzz(securityContext)

	// cannot set `allowPrivilegeEscalation` to false and `privileged` to true
	// there are more interdependences, like CAP_SYS_ADMIN imply privileged but
	// maybe it's too rare to be interesting
	if securityContext.AllowPrivilegeEscalation != nil && !*securityContext.AllowPrivilegeEscalation {
		b := false
		securityContext.Privileged = &b
	}

	*sc = securityContext
}

func CopyToInitAndFuzz(spec *v1.PodSpec) {
	if len(spec.Containers) > 0 {
		spec.InitContainers = append(spec.InitContainers, spec.Containers[0])
		spec.InitContainers[0].Name += "-init"
		FuzzContainerSecurityContext(&spec.InitContainers[0].SecurityContext)
	}
}

func Generate(opts GenerateOpts) *v1.Pod {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Command: []string{"sleep", "infinitely"},
					Name:    "digger",
					Image:   "busybox",
				},
			},
		},
	}

	pod.ObjectMeta.Namespace = opts.Namespace

	if opts.Name == "" {
		pod.ObjectMeta.Name = "digger-" + rand.String(5)
	} else {
		pod.ObjectMeta.Name = opts.Name
		pod.Spec.Containers[0].Name = opts.Name
	}

	if opts.Image != "" {
		pod.Spec.Containers[0].Image = opts.Image
	}

	if len(opts.Command) != 0 {
		pod.Spec.Containers[0].Command = opts.Command
	}

	pod.Spec.HostPID = opts.HostPid

	pod.Spec.HostNetwork = opts.HostNetwork

	if opts.Privileged {
		pod.Spec.Containers[0].SecurityContext = &v1.SecurityContext{
			Privileged: &opts.Privileged,
		}
	}

	if opts.HostPath {
		pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
			Name: "host",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/",
				},
			},
		})
		pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name:      "host",
			MountPath: "/hostfs",
		})
	}

	if opts.Tolerations {
		pod.Spec.Tolerations = append(pod.Spec.Tolerations, []v1.Toleration{
			{
				Key:      "NoExecute",
				Operator: "Exists",
			},
			{
				Key:      "NoSchedule",
				Operator: "Exists",
			},
			{
				Key:      "CriticalAddonsOnly",
				Operator: "Exists",
			},
		}...)
	}

	return pod
}
