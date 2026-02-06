package main

import (
	gl "github.com/kubex-ecosystem/logz"

	"github.com/kubex-ecosystem/gnyx/internal/module"
)

func main() {
	if err := module.RegX().Command().Execute(); err != nil {
		gl.Log("fatal", err.Error())
	}
}
