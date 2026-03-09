package types

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	ci "github.com/kubex-ecosystem/gnyx/interfaces"
	gl "github.com/kubex-ecosystem/logz"
)

type SignalManager[T chan string] struct {
	// Logger is the Logger instance for this domus instance.
	Logger *gl.LoggerZ
	// Reference is the reference ID and name.
	*Reference
	// SigChan is the channel for the signal.
	SigChan    chan os.Signal
	channelCtl T
}

// NewSignalManager creates a new SignalManager instance.
func newSignalManager[T chan string](channelCtl T, logger *gl.LoggerZ) *SignalManager[T] {
	if logger == nil {
		logger = gl.GetLoggerZ("domus")
	}
	return &SignalManager[T]{
		Logger:     logger,
		Reference:  newReference("SignalManager"),
		SigChan:    make(chan os.Signal, 1),
		channelCtl: channelCtl,
	}
}

// NewSignalManager creates a new SignalManager instance.
func NewSignalManager[T chan string](channelCtl chan string, logger *gl.LoggerZ) ci.ISignalManager[T] {
	return newSignalManager[T](channelCtl, logger)
}

// ListenForSignals sets up the signal channel to listen for specific signals.
func (sm *SignalManager[T]) ListenForSignals() (<-chan string, error) {
	signal.Notify(sm.SigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for sig := range sm.SigChan {
			gl.Printf("Sinal recebido: %s\n", sig.String())
			if sm.channelCtl != nil {
				sm.channelCtl <- fmt.Sprintf("{\"context\":\"%s\", \"message\":\"%s\"}", sm.GetName(), ""+sig.String())
			} else {
				gl.Println("Canal de controle não definido.")
			}
		}
	}()
	return sm.channelCtl, nil
}

// StopListening stops listening for signals and closes the channel.
func (sm *SignalManager[T]) StopListening() {
	signal.Stop(sm.SigChan)
	close(sm.SigChan)
	gl.Log("info", "Parando escuta de sinais")
}
