// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package result

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// TestBuildShortDelay tests the ShortDelay builder method
// GIVEN a builder
// WHEN ShortDelay is called followed by Builder
// THEN result is returned with correct settings
func TestBuildShortDelay(t *testing.T) {
	asserts := assert.New(t)
	b := NewBuilder()
	b.ShortDelay()
	r := b.Build()
	asserts.True(r.ShouldRequeue())
	asserts.GreaterOrEqual(r.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(1) * time.Second).Seconds())
	asserts.LessOrEqual(r.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(2) * time.Second).Seconds())
}
