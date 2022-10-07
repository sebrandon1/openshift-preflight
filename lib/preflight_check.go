package lib

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sebrandon1/openshift-preflight/certification/artifacts"
	"github.com/sebrandon1/openshift-preflight/certification/engine"
	"github.com/sebrandon1/openshift-preflight/certification/formatters"
	"github.com/sebrandon1/openshift-preflight/certification/runtime"

	log "github.com/sirupsen/logrus"
)

// PreflightCheck executes checks, interacts with pyxis, format output, writes, and submits results.
func PreflightCheck(
	ctx context.Context,
	cfg *runtime.Config,
	pc PyxisClient, //nolint:unparam // pyxisClient is currently unused.
	eng engine.CheckEngine,
	formatter formatters.ResponseFormatter,
	rw ResultWriter,
	rs ResultSubmitter,
) error {
	// configure the artifacts directory if the user requested a different directory.
	if cfg.Artifacts != "" {
		artifacts.SetDir(cfg.Artifacts)
	}

	// create the results file early to catch cases where we are not
	// able to write to the filesystem before we attempt to execute checks.
	resultsFilePath, err := artifacts.WriteFile(resultsFilenameWithExtension(formatter.FileExtension()), strings.NewReader(""))
	if err != nil {
		return err
	}
	resultsFile, err := rw.OpenFile(resultsFilePath)
	if err != nil {
		return err
	}
	defer resultsFile.Close()

	resultsOutputTarget := io.MultiWriter(os.Stdout, resultsFile)

	// execute the checks
	if err := eng.ExecuteChecks(ctx); err != nil {
		return err
	}
	results := eng.Results(ctx)

	// return results to the user and then close output files
	formattedResults, err := formatter.Format(ctx, results)
	if err != nil {
		return err
	}

	fmt.Fprintln(resultsOutputTarget, string(formattedResults))

	if cfg.WriteJUnit {
		if err := writeJUnit(ctx, results); err != nil {
			return err
		}
	}

	if cfg.Submit {
		if err := rs.Submit(ctx); err != nil {
			return err
		}
	}

	log.Infof("Preflight result: %s", convertPassedOverall(results.PassedOverall))

	return nil
}

func writeJUnit(ctx context.Context, results runtime.Results) error {
	var cfg runtime.Config
	cfg.ResponseFormat = "junitxml"

	junitformatter, err := formatters.NewForConfig(cfg.ReadOnly())
	if err != nil {
		return err
	}
	junitResults, err := junitformatter.Format(ctx, results)
	if err != nil {
		return err
	}

	junitFilename, err := artifacts.WriteFile("results-junit.xml", bytes.NewReader((junitResults)))
	if err != nil {
		return err
	}
	log.Tracef("JUnitXML written to %s", junitFilename)

	return nil
}

func resultsFilenameWithExtension(ext string) string {
	return strings.Join([]string{"results", ext}, ".")
}

func convertPassedOverall(passedOverall bool) string {
	if passedOverall {
		return "PASSED"
	}

	return "FAILED"
}
