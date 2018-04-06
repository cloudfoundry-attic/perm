package cmd

import (
	"context"
	"time"

	"code.cloudfoundry.org/lager"
	guuid "github.com/satori/go.uuid"
	"code.cloudfoundry.org/perm/pkg/monitor"
)

//go:generate counterfeiter . Probe

type Probe interface {
	Cleanup(context.Context, lager.Logger, string) error
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

	defer func() {
		cleanupErr := probe.Cleanup(ctx, logger.Session("cleanup"), uuid.String())
		if err == nil {
			err = cleanupErr
		}
	}()
	err = probe.Setup(ctx, logger.Session("setup"), uuid.String())
	if err != nil {
		return
	}

	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return probe.Run(cctx, logger.Session("run"), uuid.String())
}

func RecordProbeResults(
	ctx context.Context,
	logger lager.Logger,
	probe Probe,
	timeout time.Duration,
	statter monitor.PermStatter,
) {

	correct, durations, err := GetProbeResults(ctx, logger, probe, timeout)

	if err != nil {
		statter.SendFailedProbe(logger.Session("metrics"))
	} else if !correct {
		statter.SendIncorrectProbe(logger.Session("metrics"))
	} else {
		for _, d := range durations {
			statter.RecordProbeDuration(logger.Session("metrics"), d)
		}
		statter.SendCorrectProbe(logger.Session("metrics"))
	}
}
