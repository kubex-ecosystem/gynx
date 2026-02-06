package module

import (
	"os"
	"strings"
)

func RegX() *GNyx {
	var printBannerV = os.Getenv("GROMPT_PRINT_BANNER")
	if printBannerV == "" {
		printBannerV = "true"
	}

	return &GNyx{
		PrintBanner: strings.ToLower(printBannerV) == "true",
	}
}
