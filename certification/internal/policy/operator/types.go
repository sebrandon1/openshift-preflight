package operator

import (
	"context"

	"github.com/sebrandon1/openshift-preflight/certification/internal/operatorsdk"
)

type operatorSdk interface {
	BundleValidate(context.Context, string, operatorsdk.OperatorSdkBundleValidateOptions) (*operatorsdk.OperatorSdkBundleValidateReport, error)
	Scorecard(context.Context, string, operatorsdk.OperatorSdkScorecardOptions) (*operatorsdk.OperatorSdkScorecardReport, error)
}
