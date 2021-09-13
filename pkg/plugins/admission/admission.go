package admission

import (
	"context"
	"errors"
	"reflect"
	"sync"

	"github.com/quarkslab/kdigger/pkg/bucket"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	bucketName        = "admission"
	bucketDescription = "Admission scans the admission controller chain by creating specific pods to find what is prevented or not."
)

var bucketAliases = []string{"admissions", "adm"}

var currentNamespace string

// AdmissionBucket implements Bucket
type AdmissionBucket struct {
	client kubernetes.Interface

	podFactoryChain []podFactory
	podsToClean     []*v1.Pod
	cleaningLock    *sync.Mutex

	config bucket.Config
}

type admissionResult struct {
	pod     string
	success bool
	err     error
}

// Register registers a bucket
func Register(b *bucket.Buckets) {
	b.Register(bucketName, bucketAliases, bucketDescription, true, func(config bucket.Config) (bucket.Interface, error) {
		return NewAdmissionBucket(config)
	})
}

// Run runs the admission test.
func (a *AdmissionBucket) Run() (bucket.Results, error) {
	res := bucket.NewResults(bucketName)
	if !a.config.AdmForce && !a.CanIDelete() {
		return *res, errors.New("cannot delete pod, will not be able to clean the scan artifacts, force creation --admission-force")
	}
	a.initialize()
	c := make(chan admissionResult, len(a.podFactoryChain))

	for _, f := range a.podFactoryChain {
		go func(a *AdmissionBucket, f podFactory, c chan admissionResult) {
			err := a.use(f)
			if err != nil {
				// if kerrors.IsForbidden(err) {
				c <- admissionResult{
					pod:     reflect.TypeOf(f).Name(),
					success: false,
					err:     err,
				}
				return
				// }
			}
			c <- admissionResult{
				pod:     reflect.TypeOf(f).Name(),
				success: true,
				err:     nil,
			}
		}(a, f, c)
	}

	var results []admissionResult
	for i := 0; i < cap(c); i++ {
		results = append(results, <-c)
	}

	res.SetHeaders([]string{"pod", "success", "error"})
	for _, r := range results {
		if r.err != nil {
			res.AddContent([]interface{}{r.pod, r.success, r.err})
		} else {
			res.AddContent([]interface{}{r.pod, r.success, ""})
		}
	}

	err := a.Cleanup()
	if a.config.AdmForce {
		return *res, nil
	} else {
		return *res, err
	}
}

func (a *AdmissionBucket) use(f podFactory) error {
	pod := f.NewPod()
	pod, err := a.client.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return err
	} else {
		a.cleaningLock.Lock()
		a.podsToClean = append(a.podsToClean, pod)
		a.cleaningLock.Unlock()
	}
	return nil
}

// initialize initiliazes the pod factory chain to use during the scan.
func (a *AdmissionBucket) initialize() {
	a.podFactoryChain = []podFactory{
		privilegedPod{},
		hostPathPod{},
		hostPIDPod{},
		hostNetworkPod{},
		runAsRootPod{},
		privilegeEscalationPod{},
	}
}

func (a AdmissionBucket) CanIDelete() bool {
	err := a.client.CoreV1().Pods(currentNamespace).Delete(context.TODO(), "delete-test", metav1.DeleteOptions{})
	return !kerrors.IsForbidden(err)
}

// Cleanup deletes side effects pods that were successfuly created during the scan.
// TODO parallelize maybe?
func (a AdmissionBucket) Cleanup() error {
	for _, p := range a.podsToClean {
		err := a.client.CoreV1().Pods(p.Namespace).Delete(context.TODO(), p.Name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

// NewAdmissionBucket creates a new admission bucket with the kubernetes client.
func NewAdmissionBucket(cf bucket.Config) (*AdmissionBucket, error) {
	if cf.Client == nil {
		return nil, bucket.ErrMissingClient
	}
	currentNamespace = cf.Namespace
	return &AdmissionBucket{
		client:       cf.Client,
		cleaningLock: &sync.Mutex{},
		config:       cf,
	}, nil
}

// getGenericPod creates a generic pod.
func getGenericPod() *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    currentNamespace,
			GenerateName: "admission-bucket-",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "foo",
					Image: "ThisImageDoesNotExist",
				},
			},
		},
	}
}

// podFactory should be implemented by every particular pod creator to test admission.
type podFactory interface {
	NewPod() *v1.Pod
}

// hostPathPod implements podFactory
type hostPathPod struct{}

// NewPod creates a pod with the whole host filesystem mounted.
func (p hostPathPod) NewPod() *v1.Pod {
	pod := getGenericPod()
	pod.Spec.Volumes = []v1.Volume{{
		Name: "rootfs",
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: "/",
			},
		},
	}}
	return pod
}

// privilegedPod implements podFactory
type privilegedPod struct{}

// NewPod creates a pod with the privileged flag set to true.
func (p privilegedPod) NewPod() *v1.Pod {
	pod := getGenericPod()
	privileged := true
	pod.Spec.Containers[0].SecurityContext = &v1.SecurityContext{
		Privileged: &privileged,
	}
	return pod
}

// hostNetworkPod implements podFactory
type hostNetworkPod struct{}

// NewPod creates a pod with host network flag set to true.
func (p hostNetworkPod) NewPod() *v1.Pod {
	pod := getGenericPod()
	pod.Spec.HostNetwork = true
	return pod
}

// hostNetworkPod implements podFactory
type hostPIDPod struct{}

// NewPod creates a pod with host network flag set to true.
func (p hostPIDPod) NewPod() *v1.Pod {
	pod := getGenericPod()
	pod.Spec.HostPID = true
	return pod
}

// runAsRootPod implements podFactory and create a pod
type runAsRootPod struct{}

// NewPod creates a container running as root
func (p runAsRootPod) NewPod() *v1.Pod {
	pod := getGenericPod()
	runAsNonRoot := false // this is the default value
	pod.Spec.Containers[0].SecurityContext = &v1.SecurityContext{
		RunAsNonRoot: &runAsNonRoot,
	}
	return pod
}

// privilegeEscalationPod implements podFactory
type privilegeEscalationPod struct{}

// privilegeEscalationPod creates a container with allowPrivilegeEscalation to true
func (p privilegeEscalationPod) NewPod() *v1.Pod {
	pod := getGenericPod()
	allowPrivilegeEscalation := true
	pod.Spec.Containers[0].SecurityContext = &v1.SecurityContext{
		AllowPrivilegeEscalation: &allowPrivilegeEscalation,
	}
	return pod
}
