package preflight

import (
	"context"

	"github.com/redhat-openshift-ecosystem/openshift-preflight/certification/policy"
	"github.com/redhat-openshift-ecosystem/openshift-preflight/certification/runtime"
	"github.com/redhat-openshift-ecosystem/openshift-preflight/internal/lib"
)

type operatorCheckOption = func(*operatorCheck)

// NewOperatorCheck is a check that runs preflight's Operator Policy.
func NewOperatorCheck(image, kubeconfig, indeximage string, opts ...operatorCheckOption) *operatorCheck {
	c := &operatorCheck{
		image:      image,
		kubeconfig: kubeconfig,
		indeximage: indeximage,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Run executes the check and returns the results.
func (c operatorCheck) Run(ctx context.Context) (runtime.Results, error) {
	if c.image == "" {
		return runtime.Results{}, ErrImageEmpty
	}

	cfg := runtime.Config{
		Image:          c.image,
		Policy:         policy.PolicyOperator,
		ResponseFormat: "json",
		Bundle:         true,
		IndexImage:     c.indeximage,
		Kubeconfig:     c.kubeconfig,
	}

	runner, err := lib.NewCheckOperatorRunner(ctx, &cfg)
	if err != nil {
		return runtime.Results{}, err
	}

	if err := runner.Eng.ExecuteChecks(ctx); err != nil {
		return runtime.Results{}, err
	}

	res := runner.Eng.Results(ctx)
	return res, nil
}

type operatorCheck struct {
	image      string
	kubeconfig string
	indeximage string
	// formatter formatters.ResponseFormatter
}
