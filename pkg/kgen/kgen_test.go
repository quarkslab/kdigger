package kgen

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func TestGenerateManualName(t *testing.T) {
	args := GenerateOpts{
		Name: "testname",
	}
	if got := Generate(args); got.ObjectMeta.Name != args.Name {
		t.Errorf("Generate() = %v, want %v", got.ObjectMeta.Name, args.Name)
	}
}

func TestGenerateAutoName(t *testing.T) {
	args := GenerateOpts{}
	// generate pseudo random
	rand.Seed(0)
	name := "digger-cbhtc"
	if got := Generate(args); got.ObjectMeta.Name != name {
		t.Errorf("Generate() = %v, want %v", got.ObjectMeta.Name, name)
	}
}

func TestGenerateManualImage(t *testing.T) {
	args := GenerateOpts{
		Image: "testimage",
	}
	if got := Generate(args); got.Spec.Containers[0].Image != args.Image {
		t.Errorf("Generate() = %v, want %v", got.Spec.Containers[0].Image, args.Image)
	}
}

func TestGenerateManualCommand(t *testing.T) {
	args := GenerateOpts{
		Command: []string{"test", "command"},
	}
	if got := Generate(args); len(got.Spec.Containers[0].Command) != 2 && got.Spec.Containers[0].Command[0] != "test" && got.Spec.Containers[0].Command[1] != "command" {
		t.Errorf("Generate() = %v, want %v", got.Spec.Containers[0].Command, args.Command)
	}
}

func TestGenerateHostPID(t *testing.T) {
	args := GenerateOpts{
		HostPid: true,
	}
	if got := Generate(args); !got.Spec.HostPID {
		t.Errorf("Generate() = %v, want %v", got.Spec.HostPID, true)
	}
	args = GenerateOpts{
		HostPid: false,
	}
	if got := Generate(args); got.Spec.HostPID {
		t.Errorf("Generate() = %v, want %v", got.Spec.HostPID, false)
	}
}

func TestGenerateHostNetwork(t *testing.T) {
	args := GenerateOpts{
		HostNetwork: true,
	}
	if got := Generate(args); !got.Spec.HostNetwork {
		t.Errorf("Generate() = %v, want %v", got.Spec.HostNetwork, true)
	}
	args = GenerateOpts{
		HostNetwork: false,
	}
	if got := Generate(args); got.Spec.HostNetwork {
		t.Errorf("Generate() = %v, want %v", got.Spec.HostNetwork, false)
	}
}

func TestGeneratePrivileged(t *testing.T) {
	args := []GenerateOpts{
		{
			Privileged: true,
		},
		{
			Privileged: false,
		},
	}
	for _, arg := range args {
		if got := Generate(arg); got.Spec.Containers[0].SecurityContext != nil && got.Spec.Containers[0].SecurityContext.Privileged == &arg.Privileged {
			t.Errorf("Generate() = %v, want %v", got.Spec.Containers[0].SecurityContext.Privileged, &arg.Privileged)
		}
	}
}

func TestGenerateHostPath(t *testing.T) {
	args := GenerateOpts{
		HostPath: true,
	}
	if got := Generate(args); len(got.Spec.Containers[0].VolumeMounts) != 1 &&
		got.Spec.Containers[0].VolumeMounts[0].Name == "host" &&
		got.Spec.Containers[0].VolumeMounts[0].MountPath == "hostfs" &&
		got.Spec.Volumes[0].Name == "host" &&
		got.Spec.Volumes[0].HostPath.Path == "/" {
		t.Errorf("Generate() = %v and %v", got.Spec.Containers[0].VolumeMounts[0], got.Spec.Volumes[0])
	}
}

func TestGenerateTolerations(t *testing.T) {
	args := GenerateOpts{
		Tolerations: true,
	}
	want := []v1.Toleration{
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
	}
	if got := Generate(args); reflect.DeepEqual(got, want) {
		t.Errorf("Generate() = %v, want %v", got, want)
	}
}
