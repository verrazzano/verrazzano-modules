// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package constants

// GlobalImagePullSecName is the name of the global image pull secret
const GlobalImagePullSecName = "verrazzano-container-registry"

// ComponentAvailability identifies the availability of a Verrazzano Component.
type ComponentAvailability string

const (
	//ComponentAvailable signifies that a Component is ready for use.
	ComponentAvailable = "Available"
	//ComponentUnavailable signifies that a Verrazzano Component is not ready for use.
	ComponentUnavailable = "Unavailable"
)
