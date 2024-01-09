// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

// Package adapter provides security engine adapters to use with nimbus.
package adapter

import (
	"context"

	v1 "github.com/5GSEC/nimbus/api/v1"
)

// The Adapters currently supported by nimbus.
var Adapters = []string{"kubearmor"}

// Adapter knows how to create/update and delete security-engine policies.
type Adapter interface {
	ApplyPolicy(ctx context.Context, np v1.NimbusPolicy) error
	DeletePolicy(ctx context.Context, np v1.NimbusPolicy) error
}
