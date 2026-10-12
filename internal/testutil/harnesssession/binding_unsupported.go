//go:build !linux

package harnesssession

import "github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"

func ClaimRecoveryOutput(string, string, string, string, string) error { return harnesslocal.ErrUnsafe }
