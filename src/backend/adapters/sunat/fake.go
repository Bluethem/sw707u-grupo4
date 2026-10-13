package sunat

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
)

type Mode string

const (
	ModeAccept Mode = "accept"
	ModeReject Mode = "reject"
	ModeError  Mode = "error"
)

type Fake struct {
	Mode Mode
}

func NewFake(mode Mode) *Fake {
	if mode == "" {
		mode = ModeFromEnv()
	}
	if mode == "" {
		mode = ModeAccept
	}
	return &Fake{Mode: mode}
}

func ModeFromEnv() Mode {
	return Mode(strings.ToLower(strings.TrimSpace(os.Getenv("SUNAT_FAKE_MODE"))))
}

func (f *Fake) Enviar(_ context.Context, comprobanteID string, _ []byte) (ports.SunatResult, error) {
	switch f.Mode {
	case ModeReject:
		return ports.SunatResult{
			Aceptado: false,
			Codigo:   "FAKE-REJECT",
			Mensaje:  "comprobante rechazado (fake SUNAT)",
		}, nil
	case ModeError:
		return ports.SunatResult{}, errors.New("error de transporte fake SUNAT")
	default:
		return ports.SunatResult{
			Aceptado: true,
			Codigo:   "0",
			Mensaje:  "aceptado (fake SUNAT)",
			CDR:      []byte("CDR-OK:" + comprobanteID),
		}, nil
	}
}
