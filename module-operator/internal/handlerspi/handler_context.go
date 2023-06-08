// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package handlerspi

import (
	"github.com/verrazzano/verrazzano-modules/pkg/vzlog"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// HandlerContext contains the handler contexts for the API handler methods
type HandlerContext struct {
	ctrlclient.Client
	Log    vzlog.VerrazzanoLogger
	DryRun bool
	CR     interface{}
	HelmInfo
}
