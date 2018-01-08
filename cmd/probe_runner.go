package cmd

import (
	"context"
	"time"

	"code.cloudfoundry.org/lager"
	guuid "github.com/satori/go.uuid"
)

//go:generate counterfeiter . QueryProbe

type QueryProbe interface {
	Cleanup(context.Context, lager.Logger, string) error
	Setup(context.Context, lager.Logger, string) error
	Run(context.Context, lager.Logger, string) (bool, []time.Duration, error)
}

func RunQueryProbe(ctx context.Context, logger lager.Logger, probe QueryProbe, timeout time.Duration) (correct bool, durations []time.Duration, err error) {
	uuid := guuid.NewV4()

	defer func() {
		cleanupErr := probe.Cleanup(ctx, logger.Session("cleanup"), uuid.String())
		if err == nil && cleanupErr != nil {
			err = cleanupErr
		}
	}()
	err = probe.Setup(ctx, logger.Session("setup"), uuid.String())
	if err != nil {
		return
	}

	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	correct, durations, err = probe.Run(cctx, logger.Session("run"), uuid.String())
	return
}
