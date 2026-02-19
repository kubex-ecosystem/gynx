package module

import (
	kbxGet "github.com/kubex-ecosystem/kbx/get"
)

func RegX() *GNyx {
	return &GNyx{
		PrintBanner: kbxGet.EnvOrType("KUBEX_PRINT_BANNER", kbxGet.EnvOrType("KUBEX_GNYX_PRINT_BANNER", false)),
	}
}
