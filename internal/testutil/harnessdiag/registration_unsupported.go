//go:build !linux

package harnessdiag

func clockCompatible(object) bool { return false }
