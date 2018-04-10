package cmd

import (
	"context"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/perm/pkg/monitor"
	guuid "github.com/satori/go.uuid"
)

//go:generate counterfeiter . Probe

type Probe interface {
	Cleanup(context.Context, time.Duration, lager.Logger, string) error
	Setup(context.Context, lager.Logger, string) error
	Run(context.Context, lager.Logger, string) (bool, []time.Duration, error)
}

func GetProbeResults(
	ctx context.Context,
	logger lager.Logger,
	probe Probe,
	timeout time.Duration,
) (correct bool, durations []time.Duration, err error) {
	uuid := guuid.NewV4()

	cctx, _ := context.WithTimeout(ctx, timeout)

	defer func() {
		cleanupErr := probe.Cleanup(cctx, timeout, logger.Session("cleanup"), uuid.String())
		if err == nil {
			err = cleanupErr
		}
	}()
	err = probe.Setup(cctx, logger.Session("setup"), uuid.String())
	if err != nil {
		return
	}

	return probe.Run(cctx, logger.Session("run"), uuid.String())
}

func RecordProbeResults(
	ctx context.Context,
	logger lager.Logger,
	probe Probe,
	timeout time.Duration,
	statter monitor.PermStatter,
	acceptableQueryWindow time.Duration,
) {

	correct, durations, err := GetProbeResults(ctx, logger, probe, timeout)

	if err != nil {
		statter.SendFailedProbe(logger.Session("metrics"))
		return
	}
	if !correct {
		statter.SendIncorrectProbe(logger.Session("metrics"))
		return
	}
	failedQuery := false
	for _, d := range durations {
		if d > acceptableQueryWindow {
			failedQuery = true
		}
		statter.RecordProbeDuration(logger.Session("metrics"), d)
	}
	if failedQuery {
		statter.SendFailedProbe(logger.Session("metrics"))
	} else {
		statter.SendCorrectProbe(logger.Session("metrics"))
	}
}
