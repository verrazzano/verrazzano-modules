// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package result

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// TestBuildDefault tests the builder Build method
// GIVEN a builder
// WHEN Build is called without any other builder calls
// THEN result has the correct settings
func TestBuildDefault(t *testing.T) {
	asserts := assert.New(t)
	b := NewBuilder()
	r := b.Build()
	asserts.False(r.IsError())
	asserts.Zero(r.GetError())
	asserts.False(r.ShouldRequeue())
	asserts.Zero(r.GetCtrlRuntimeResult().RequeueAfter.Seconds())
}

// TestBuildError tests the Error builder method
// GIVEN a builder
// WHEN Error is called followed by Build
// THEN result has the correct settings
func TestBuildError(t *testing.T) {
	asserts := assert.New(t)
	err := errors.New("error")
	b := NewBuilder()
	b.Error(err)
	r := b.Build()
	asserts.True(r.IsError())
	asserts.Equal(err, r.GetError())
	asserts.False(r.ShouldRequeue())
	asserts.Zero(r.GetCtrlRuntimeResult().RequeueAfter.Seconds())
}

// TestBuildShortDelay tests the ShortDelay builder method
// GIVEN a builder
// WHEN ShortDelay is called followed by Build
// THEN result is returned with correct settings
func TestBuildShortDelay(t *testing.T) {
	asserts := assert.New(t)
	b := NewBuilder()
	b.ShortDelay()
	r := b.Build()
	asserts.False(r.IsError())
	asserts.NoError(r.GetError())
	asserts.True(r.ShouldRequeue())
	asserts.GreaterOrEqual(r.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(200) * time.Millisecond).Seconds())
	asserts.LessOrEqual(r.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(500) * time.Millisecond).Seconds())
}

// TestBuildDelay tests the Delay builder method
// GIVEN a builder
// WHEN Delay is called followed by Build
// THEN result is returned with correct settings
func TestBuildDelay(t *testing.T) {
	asserts := assert.New(t)
	b := NewBuilder()
	b.Delay(5, 6, time.Second)
	r := b.Build()
	asserts.False(r.IsError())
	asserts.NoError(r.GetError())
	asserts.True(r.ShouldRequeue())
	asserts.GreaterOrEqual(r.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(5) * time.Second).Seconds())
	asserts.LessOrEqual(r.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(6) * time.Second).Seconds())
}

// TestBuildAll tests the builder method when all build methods are called
// GIVEN a builder
// WHEN Error, Delay, and Short are called followed by Build
// THEN result is returned with correct settings
func TestBuildAll(t *testing.T) {
	asserts := assert.New(t)
	err := errors.New("error")
	err2 := errors.New("error")

	// Delay take precedence over ShortDelay
	b := NewBuilder()
	b.Error(err).ShortDelay().Delay(4, 5, time.Minute)
	r := b.Build()
	asserts.True(r.ShouldRequeue())
	asserts.True(r.IsError())
	asserts.Equal(err, r.GetError())
	asserts.GreaterOrEqual(r.GetCtrlRuntimeResult().RequeueAfter.Minutes(), (time.Duration(4) * time.Minute).Minutes())
	asserts.LessOrEqual(r.GetCtrlRuntimeResult().RequeueAfter.Minutes(), (time.Duration(5) * time.Minute).Minutes())

	// ShortDelay take precedence over Delay
	b.Error(err2).Delay(4, 5, time.Minute).ShortDelay()
	r = b.Build()
	asserts.True(r.ShouldRequeue())
	asserts.True(r.IsError())
	asserts.Equal(err2, r.GetError())
	asserts.GreaterOrEqual(r.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(200) * time.Millisecond).Seconds())
	asserts.LessOrEqual(r.GetCtrlRuntimeResult().RequeueAfter.Seconds(), (time.Duration(500) * time.Millisecond).Seconds())
}
