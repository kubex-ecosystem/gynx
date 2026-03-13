package main

import (
	gl "github.com/kubex-ecosystem/logz"

	"github.com/kubex-ecosystem/gnyx/internal/module"

	kbxMod "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
)

func main() {
	kbxMod.InitArgsDefaults()

	if err := module.RegX().Command().Execute(); err != nil {
		gl.Log("fatal", err.Error())
	}
}
