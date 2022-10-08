package kgen

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/rand"
)

type GenerateOpts struct {
	Name        string
	Image       string
	Namespace 	string
	Command     []string
	Privileged  bool
	HostPath    bool
	HostPid     bool
	HostNetwork bool
	Tolerations bool
}

func Generate(opts GenerateOpts) *v1.Pod {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default"
		},
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

	if opts.Namespace == "" {
		pod.ObjectMeta.Namespace = opts.Namespace
	}
	if len(opts.Namespace) != 0 {
		pod.ObjectMeta.Namespace = opts.Namespace
	}

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
