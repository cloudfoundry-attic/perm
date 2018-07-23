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
	Cleanup(context.Context, time.Duration, lager.Logger, string) ([]monitor.LabeledDuration, error)
	Setup(context.Context, lager.Logger, string) ([]monitor.LabeledDuration, error)
	Run(context.Context, lager.Logger, string) (bool, []monitor.LabeledDuration, error)
}

func GetProbeResults(logger lager.Logger, probe Probe, timeout time.Duration) (correct bool, durations []monitor.LabeledDuration, err error) {
	uuid := guuid.NewV4()

	defer func() {
		ctx, _ := context.WithTimeout(context.Background(), timeout)

		cleanupDurations, cleanupErr := probe.Cleanup(ctx, timeout, logger.Session("cleanup"), uuid.String())
		if err == nil {
			err = cleanupErr
		}

		durations = append(durations, cleanupDurations...)
	}()

	ctx, _ := context.WithTimeout(context.Background(), timeout)

	durations, err = probe.Setup(ctx, logger.Session("setup"), uuid.String())
	if err != nil {
		return
	}

	correct, runDurations, err := probe.Run(ctx, logger.Session("run"), uuid.String())
	durations = append(durations, runDurations...)
	return
}

func RecordProbeResults(
	logger lager.Logger,
	probe Probe,
	statter monitor.PermStatter,
	requestDuration time.Duration,
	timeout time.Duration,
) {
	correct, durations, err := GetProbeResults(logger, probe, timeout)

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
		if d.Duration > requestDuration {
			failedQuery = true
		}
		statter.RecordProbeDuration(logger.Session("metrics"), d.Duration)
	}
	if failedQuery {
		statter.SendFailedProbe(logger.Session("metrics"))
	} else {
		statter.SendCorrectProbe(logger.Session("metrics"))
	}
}
