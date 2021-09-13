package authorization

import (
	"context"
	"fmt"
	"sort"

	"github.com/quarkslab/kdigger/pkg/bucket"
	v1 "k8s.io/api/authorization/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/describe"
	rbacutil "k8s.io/kubectl/pkg/util/rbac"
)

const (
	bucketName        = "authorization"
	bucketDescription = "Authorization checks your API permissions with the current context or the available token."
)

var bucketAliases = []string{"authorizations", "auth"}

type AuthorizationBucket struct {
	config bucket.Config
}

func (n AuthorizationBucket) Run() (bucket.Results, error) {
	res := bucket.NewResults(bucketName)

	// create the self subject rules review object
	obj := &v1.SelfSubjectRulesReview{
		Spec: v1.SelfSubjectRulesReviewSpec{
			Namespace: n.config.Namespace,
		},
	}

	res.SetComment(fmt.Sprintf("Checking current context/token permissions in the %q namespace.", n.config.Namespace))

	// do the actual request
	response, err := n.config.Client.AuthorizationV1().SelfSubjectRulesReviews().Create(
		context.TODO(),
		obj,
		metav1.CreateOptions{},
	)
	if err != nil {
		return bucket.Results{}, err
	}

	// format the response
	rules, err := getCompactRules(response.Status)
	if err != nil {
		return bucket.Results{}, err
	}
	res.SetHeaders([]string{"resources", "nonResourceURLs", "ressourceNames", "verbs"})
	for _, r := range rules {
		res.AddContent([]interface{}{
			describe.CombineResourceGroup(r.Resources, r.APIGroups),
			r.NonResourceURLs,
			r.ResourceNames,
			r.Verbs,
		})
	}
	return *res, nil
}

func Register(b *bucket.Buckets) {
	b.Register(bucketName, bucketAliases, bucketDescription, false, func(config bucket.Config) (bucket.Interface, error) {
		return NewAuthorizationBucket(config)
	})
}

func NewAuthorizationBucket(c bucket.Config) (*AuthorizationBucket, error) {
	if c.Client == nil {
		return nil, bucket.ErrMissingClient
	}
	return &AuthorizationBucket{
		config: c,
	}, nil
}

// partial copy of https://github.com/kubernetes/kubectl/blob/0f88fc6b598b7e883a391a477215afb080ec7733/pkg/cmd/auth/cani.go#L323
func getCompactRules(status v1.SubjectRulesReviewStatus) ([]rbacv1.PolicyRule, error) {
	if status.Incomplete {
		// TODO
		// fmt.Fprintf(o.ErrOut, "warning: the list may be incomplete: %v\n", status.EvaluationError)
	}

	breakdownRules := []rbacv1.PolicyRule{}
	for _, rule := range convertToPolicyRule(status) {
		breakdownRules = append(breakdownRules, rbacutil.BreakdownRule(rule)...)
	}

	compactRules, err := rbacutil.CompactRules(breakdownRules)
	if err != nil {
		return nil, err
	}
	sort.Stable(rbacutil.SortableRuleSlice(compactRules))

	return compactRules, nil
}

// copy of https://github.com/kubernetes/kubectl/blob/0f88fc6b598b7e883a391a477215afb080ec7733/pkg/cmd/auth/cani.go#L355
func convertToPolicyRule(status v1.SubjectRulesReviewStatus) []rbacv1.PolicyRule {
	ret := []rbacv1.PolicyRule{}
	for _, resource := range status.ResourceRules {
		ret = append(ret, rbacv1.PolicyRule{
			Verbs:         resource.Verbs,
			APIGroups:     resource.APIGroups,
			Resources:     resource.Resources,
			ResourceNames: resource.ResourceNames,
		})
	}

	for _, nonResource := range status.NonResourceRules {
		ret = append(ret, rbacv1.PolicyRule{
			Verbs:           nonResource.Verbs,
			NonResourceURLs: nonResource.NonResourceURLs,
		})
	}

	return ret
}
