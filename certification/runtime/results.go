package runtime

import (
	"time"

	"github.com/sebrandon1/openshift-preflight/certification"
)

type Result struct {
	certification.Check
	ElapsedTime time.Duration
}

type Results struct {
	TestedImage       string
	PassedOverall     bool
	TestedOn          OpenshiftClusterVersion
	CertificationHash string
	Passed            []Result
	Failed            []Result
	Errors            []Result
}
