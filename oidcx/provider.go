package oidcx

import oidc "github.com/coreos/go-oidc"

//go:generate counterfeiter . Provider

type Provider interface {
	Verifier(config *oidc.Config) *oidc.IDTokenVerifier
}
