// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/gardener/gardener/cmd/utils"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/gardener/ext-authz-server/cmd/ext-authz-server/app"
)

func main() {
	utils.DeduplicateWarnings()

	if err := app.NewCommand().ExecuteContext(signals.SetupSignalHandler()); err != nil {
		panic(err)
	}
}
