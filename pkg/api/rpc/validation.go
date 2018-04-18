package rpc

import (
	"errors"
	"strings"

	"code.cloudfoundry.org/perm/protos/gen"
)

func validateActor(actor *protos.Actor) error {
	namespace := actor.GetNamespace()
	if strings.Trim(namespace, "\t \n") == "" {
		return errors.New("actor namespace cannot be empty")
	}

	return nil
}
