package lib

import (
	"context"
	"fmt"

	"github.com/sebrandon1/openshift-preflight/certification"
	"github.com/sebrandon1/openshift-preflight/certification/engine"
	"github.com/sebrandon1/openshift-preflight/certification/formatters"
	"github.com/sebrandon1/openshift-preflight/certification/policy"
	"github.com/sebrandon1/openshift-preflight/certification/runtime"
)

// CheckContainerRunner contains all of the components necessary to run checkContainer.
type CheckContainerRunner struct {
	Cfg       *runtime.Config
	Pc        PyxisClient
	Eng       engine.CheckEngine
	Formatter formatters.ResponseFormatter
	Rw        ResultWriter
	Rs        ResultSubmitter
}

func NewCheckContainerRunner(ctx context.Context, cfg *runtime.Config, submit bool) (*CheckContainerRunner, error) {
	cfg.Policy = policy.PolicyContainer
	cfg.Submit = submit

	pyxisClient := NewPyxisClient(ctx, cfg.ReadOnly())
	// If we have a pyxisClient, we can query for container policy exceptions.
	if pyxisClient != nil {
		policy, err := GetContainerPolicyExceptions(ctx, pyxisClient)
		if err != nil {
			return nil, err
		}

		cfg.Policy = policy
	}

	engine, err := engine.NewForConfig(ctx, cfg.ReadOnly())
	if err != nil {
		return nil, err
	}

	fmttr, err := formatters.NewForConfig(cfg.ReadOnly())
	if err != nil {
		return nil, err
	}

	rs := ResolveSubmitter(pyxisClient, cfg.ReadOnly())

	return &CheckContainerRunner{
		Cfg:       cfg,
		Pc:        pyxisClient,
		Eng:       engine,
		Formatter: fmttr,
		Rw:        &runtime.ResultWriterFile{},
		Rs:        rs,
	}, nil
}

// checkOperatorRunner contains all of the components necessary to run checkOperator.
type CheckOperatorRunner struct {
	Cfg       *runtime.Config
	Eng       engine.CheckEngine
	Formatter formatters.ResponseFormatter
	Rw        ResultWriter
}

// newCheckOperatorRunner returns a checkOperatorRunner containing all of the tooling necessary
// to run checkOperator.
func NewCheckOperatorRunner(ctx context.Context, cfg *runtime.Config) (*CheckOperatorRunner, error) {
	cfg.Policy = policy.PolicyOperator
	cfg.Submit = false // there's no such thing as submitting for operators today.

	engine, err := engine.NewForConfig(ctx, cfg.ReadOnly())
	if err != nil {
		return nil, err
	}

	fmttr, err := formatters.NewForConfig(cfg.ReadOnly())
	if err != nil {
		return nil, err
	}

	return &CheckOperatorRunner{
		Cfg:       cfg,
		Eng:       engine,
		Formatter: fmttr,
		Rw:        &runtime.ResultWriterFile{},
	}, nil
}

// resolveSubmitter will build out a resultSubmitter if the provided pyxisClient, pc, is not nil.
// The pyxisClient is a required component of the submitter. If pc is nil, then a noop submitter
// is returned instead, which does nothing.
func ResolveSubmitter(pc PyxisClient, cfg certification.Config) ResultSubmitter {
	if pc != nil {
		return &ContainerCertificationSubmitter{
			CertificationProjectID: cfg.CertificationProjectID(),
			Pyxis:                  pc,
			DockerConfig:           cfg.DockerConfig(),
			PreflightLogFile:       cfg.LogFile(),
		}
	}
	return NewNoopSubmitter(true, "", nil)
}

// GetContainerPolicyExceptions will query Pyxis to determine if
// a given project has a certification excemptions, such as root or scratch.
// This will then return the corresponding policy.
//
// If no policy exception flags are found on the project, the standard
// container policy is returned.
func GetContainerPolicyExceptions(ctx context.Context, pc PyxisClient) (policy.Policy, error) {
	certProject, err := pc.GetProject(ctx)
	if err != nil {
		return "", fmt.Errorf("could not retrieve project: %w", err)
	}
	// log.Debugf("Certification project name is: %s", certProject.Name)
	if certProject.Container.Type == "scratch" {
		return policy.PolicyScratch, nil
	}

	// if a partner sets `Host Level Access` in connect to `Privileged`, enable RootExceptionContainerPolicy checks
	if certProject.Container.Privileged {
		return policy.PolicyRoot, nil
	}
	return policy.PolicyContainer, nil
}
