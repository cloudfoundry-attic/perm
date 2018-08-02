package main

const (
	starting = "starting"
	finished = "finished"

	success = "success"
	failed  = "failed"

	failedToConnectToStatsD  = "failed-to-connect-to-statsd"
	failedToReadCertificate  = "failed-to-read-certificate"
	failedToAppendCertToPool = "failed-to-append-cert-to-pool"
	failedToCreatePermClient = "failed-to-create-perm-client"
	failedToGRPCDial         = "failed-to-grpc-dial"
	failedToSendMetric       = "failed-to-send-metric"

	probeQuantileMetricPrefix = "perm.probe.responses.timing"
	probeCorrect              = "perm.probe.runs.correct"
	probeSuccess              = "perm.probe.runs.success"
	probeFailedToObserve      = "perm.probe.failed.to.observe"
	probeAPIErrored           = "perm.probe.api.errored"

	alwaysSend = 1
)
