// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package catalog

type Catalog struct {
	Modules []Module `json:"modules"`
}

type Module struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
