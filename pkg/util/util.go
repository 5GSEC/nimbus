// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package util

import (
	"github.com/go-logr/logr"
	"runtime"
	"runtime/debug"
)

func LogBuildInfo(logger logr.Logger) {
	info, _ := debug.ReadBuildInfo()
	vcsRev := ""
	vcsTime := ""
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" {
			vcsRev = s.Value
		} else if s.Key == "vcs.time" {
			vcsTime = s.Value
		}
	}
	logger.Info("Build info", "git.revision", vcsRev,
		"build.time", vcsTime,
		"build.version", runtime.Version(),
		"GOOS", runtime.GOOS,
		"GOARCH", runtime.GOARCH,
	)
}
