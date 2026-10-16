package sunat

import (
	"context"
	"errors"
	"fmt"
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
	return f.applyMode(comprobanteID, "comprobante", func(id string) ports.SunatResult {
		return ports.SunatResult{
			Aceptado: true,
			Codigo:   "0",
			Mensaje:  "aceptado (fake SUNAT)",
			CDR:      []byte("CDR-OK:" + id),
		}
	})
}

func (f *Fake) EnviarBaja(ctx context.Context, comprobanteID, motivo string) (ports.SunatResult, error) {
	_ = ctx
	_ = motivo
	return f.applyMode(comprobanteID, "baja", func(id string) ports.SunatResult {
		return ports.SunatResult{
			Aceptado: true,
			Codigo:   "0",
			Mensaje:  "baja aceptada (fake SUNAT)",
			CDR:      []byte("CDR-BAJA:" + id),
		}
	})
}

func (f *Fake) EnviarResumen(ctx context.Context, resumenID string, payload []byte) (ports.SunatResult, error) {
	_ = ctx
	_ = payload
	return f.applyMode(resumenID, "resumen", func(id string) ports.SunatResult {
		return ports.SunatResult{
			Aceptado: true,
			Codigo:   "0",
			Mensaje:  "resumen aceptado (fake SUNAT)",
			CDR:      []byte("CDR-RESUMEN:" + id),
		}
	})
}

func (f *Fake) applyMode(
	id, kind string,
	onAccept func(id string) ports.SunatResult,
) (ports.SunatResult, error) {
	switch f.Mode {
	case ModeReject:
		return ports.SunatResult{
			Aceptado: false,
			Codigo:   "FAKE-REJECT",
			Mensaje:  fmt.Sprintf("%s rechazado (fake SUNAT)", kind),
		}, nil
	case ModeError:
		return ports.SunatResult{}, errors.New("error de transporte fake SUNAT")
	default:
		return onAccept(id), nil
	}
}
