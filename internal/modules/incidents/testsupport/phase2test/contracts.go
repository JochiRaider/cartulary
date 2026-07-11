package phase2test

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
)

type ErrorContract = contracttest.ErrorContract

type ExtensionProfileContract = contracttest.ExtensionProfileContract

func ErrorContractByCode(t testing.TB, code string) ErrorContract {
	t.Helper()
	return contracttest.ErrorContractByCode(t, code)
}

func RequireErrorContract(t testing.TB, code string, wantStatus int) {
	t.Helper()
	contracttest.RequireErrorContract(t, code, wantStatus)
}

func CurrentProfileExtensions(t testing.TB) []ExtensionProfileContract {
	t.Helper()
	return contracttest.CurrentProfileExtensions(t)
}
