// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package result

import (
	"k8s.io/apimachinery/pkg/util/rand"
	"time"
)

type Builder interface {
	Delay(min int, max int, units time.Duration) Builder
	ShortDelay() Builder
	Error(err error) Builder
	Build() Result
}

type builder struct {
	controllerResult
}

var _ Builder = &builder{}

func NewBuilder() Builder {
	return &builder{}
}

func (b *builder) Delay(min int, max int, units time.Duration) Builder {
	var random = rand.IntnRange(min, max)
	delayNanos := time.Duration(random) * units
	b.controllerResult.result.RequeueAfter = delayNanos
	b.controllerResult.result.Requeue = true
	return b
}

func (b *builder) ShortDelay() Builder {
	return b.Delay(1, 2, time.Second)
}

func (b *builder) Error(err error) Builder {
	b.controllerResult.err = err
	return b
}

func (b *builder) Build() Result {
	return b.controllerResult
}
