// SPDX-License-Identifier: Apache-2.0
// Copyright 2023 Authors of Nimbus

package transformer

import (
	v1 "github.com/5GSEC/nimbus/api/v1"

	kubearmorv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorController/api/security.kubearmor.com/v1"
)

func handleProcessPolicy(rule v1.Rule, category string) (kubearmorv1.ProcessType, error) {
	processType := kubearmorv1.ProcessType{
		MatchPaths:       []kubearmorv1.ProcessPathType{},
		MatchDirectories: []kubearmorv1.ProcessDirectoryType{},
		MatchPatterns:    []kubearmorv1.ProcessPatternType{},
	}

	switch category {
	case "paths":
		for _, matchPath := range rule.MatchPaths {
			if matchPath.Path != "" {
				processType.MatchPaths = append(processType.MatchPaths, kubearmorv1.ProcessPathType{
					Path: kubearmorv1.MatchPathType(matchPath.Path),
				})
			}
		}

	case "dirs":
		for _, matchDir := range rule.MatchDirectories {
			var fromSources []kubearmorv1.MatchSourceType
			for _, source := range matchDir.FromSource {
				fromSources = append(fromSources, kubearmorv1.MatchSourceType{
					Path: kubearmorv1.MatchPathType(source.Path),
				})
			}
			if matchDir.Directory != "" || len(fromSources) > 0 {
				processType.MatchDirectories = append(processType.MatchDirectories, kubearmorv1.ProcessDirectoryType{
					Directory:  kubearmorv1.MatchDirectoryType(matchDir.Directory),
					FromSource: fromSources,
				})
			}
		}

	case "patterns":
		for _, matchPattern := range rule.MatchPatterns {
			if matchPattern.Pattern != "" {
				processType.MatchPatterns = append(processType.MatchPatterns, kubearmorv1.ProcessPatternType{
					Pattern: matchPattern.Pattern,
				})
			}
		}
	}

	// Set empty slices if fields are empty
	if len(processType.MatchPaths) == 0 {
		processType.MatchPaths = []kubearmorv1.ProcessPathType{}
	}
	if len(processType.MatchDirectories) == 0 {
		processType.MatchDirectories = []kubearmorv1.ProcessDirectoryType{}
	}
	if len(processType.MatchPatterns) == 0 {
		processType.MatchPatterns = []kubearmorv1.ProcessPatternType{}
	}

	return processType, nil
}

func handleFilePolicy(rule v1.Rule, category string) (kubearmorv1.FileType, error) {
	fileType := kubearmorv1.FileType{
		MatchPaths:       []kubearmorv1.FilePathType{},
		MatchDirectories: []kubearmorv1.FileDirectoryType{},
		MatchPatterns:    []kubearmorv1.FilePatternType{},
	}

	switch category {
	case "paths":
		for _, matchPath := range rule.MatchPaths {
			if matchPath.Path != "" {
				fileType.MatchPaths = append(fileType.MatchPaths, kubearmorv1.FilePathType{
					Path: kubearmorv1.MatchPathType(matchPath.Path),
				})
			}
		}
	case "dirs":
		for _, matchDir := range rule.MatchDirectories {
			var fromSources []kubearmorv1.MatchSourceType
			for _, source := range matchDir.FromSource {
				fromSources = append(fromSources, kubearmorv1.MatchSourceType{
					Path: kubearmorv1.MatchPathType(source.Path),
				})
			}
			if matchDir.Directory != "" || len(fromSources) > 0 {
				fileType.MatchDirectories = append(fileType.MatchDirectories, kubearmorv1.FileDirectoryType{
					Directory:  kubearmorv1.MatchDirectoryType(matchDir.Directory),
					FromSource: fromSources,
				})
			}
		}
	case "patterns":
		for _, matchPattern := range rule.MatchPatterns {
			if matchPattern.Pattern != "" {
				fileType.MatchPatterns = append(fileType.MatchPatterns, kubearmorv1.FilePatternType{
					Pattern: matchPattern.Pattern,
				})
			}
		}
	}

	// Set empty slices if fields are empty
	if len(fileType.MatchPaths) == 0 {
		fileType.MatchPaths = []kubearmorv1.FilePathType{}
	}
	if len(fileType.MatchDirectories) == 0 {
		fileType.MatchDirectories = []kubearmorv1.FileDirectoryType{}
	}
	if len(fileType.MatchPatterns) == 0 {
		fileType.MatchPatterns = []kubearmorv1.FilePatternType{}
	}

	return fileType, nil
}

func handleNetworkPolicy(rule v1.Rule) (kubearmorv1.NetworkType, error) {
	networkType := kubearmorv1.NetworkType{
		MatchProtocols: []kubearmorv1.MatchNetworkProtocolType{},
	}

	for _, matchProtocol := range rule.MatchProtocols {
		if matchProtocol.Protocol != "" {
			networkType.MatchProtocols = append(networkType.MatchProtocols, kubearmorv1.MatchNetworkProtocolType{
				Protocol: kubearmorv1.MatchNetworkProtocolStringType(matchProtocol.Protocol),
			})
		}
	}
	return networkType, nil
}

func handleSyscallPolicy(rule v1.Rule, category string) (kubearmorv1.SyscallsType, error) {
	// Initialize syscallType with default values
	syscallType := kubearmorv1.SyscallsType{
		MatchSyscalls: []kubearmorv1.SyscallMatchType{},
		MatchPaths:    []kubearmorv1.SyscallMatchPathType{},
	}

	switch category {
	case "syscalls":
		for _, matchSyscall := range rule.MatchSyscalls {
			syscallMatch := kubearmorv1.SyscallMatchType{
				Syscalls: []kubearmorv1.Syscall{},
			}
			for _, syscall := range matchSyscall.Syscalls {
				if syscall != "" {
					syscallMatch.Syscalls = append(syscallMatch.Syscalls, kubearmorv1.Syscall(syscall))
				}
			}
			syscallType.MatchSyscalls = append(syscallType.MatchSyscalls, syscallMatch)
		}

	case "paths":
		for _, matchSyscallPath := range rule.MatchSyscallPaths {
			syscallMatchPath := kubearmorv1.SyscallMatchPathType{
				Path:       kubearmorv1.MatchSyscallPathType(matchSyscallPath.Path),
				Recursive:  matchSyscallPath.Recursive,
				Syscalls:   []kubearmorv1.Syscall{},
				FromSource: []kubearmorv1.SyscallFromSourceType{},
			}
			for _, syscall := range matchSyscallPath.Syscalls {
				if syscall != "" {
					syscallMatchPath.Syscalls = append(syscallMatchPath.Syscalls, kubearmorv1.Syscall(syscall))
				}
			}
			for _, fromSource := range matchSyscallPath.FromSource {
				syscallFromSource := kubearmorv1.SyscallFromSourceType{
					Path: kubearmorv1.MatchPathType(fromSource.Path),
					Dir:  fromSource.Dir,
				}
				syscallMatchPath.FromSource = append(syscallMatchPath.FromSource, syscallFromSource)
			}
			syscallType.MatchPaths = append(syscallType.MatchPaths, syscallMatchPath)
		}
	}

	// Set empty slices if fields are empty
	if len(syscallType.MatchSyscalls) == 0 {
		syscallType.MatchSyscalls = []kubearmorv1.SyscallMatchType{}
	}
	// Set empty slices if fields are empty
	if len(syscallType.MatchPaths) == 0 {
		syscallType.MatchPaths = []kubearmorv1.SyscallMatchPathType{}
	}

	return syscallType, nil
}

func handleCapabilityPolicy(rule v1.Rule) (kubearmorv1.CapabilitiesType, error) {
	capabilityType := kubearmorv1.CapabilitiesType{
		MatchCapabilities: []kubearmorv1.MatchCapabilitiesType{},
	}

	for _, matchCapability := range rule.MatchCapabilities {
		if matchCapability.Capability != "" {
			capabilityType.MatchCapabilities = append(capabilityType.MatchCapabilities, kubearmorv1.MatchCapabilitiesType{
				Capability: kubearmorv1.MatchCapabilitiesStringType(matchCapability.Capability),
			})
		}
	}
	return capabilityType, nil
}
