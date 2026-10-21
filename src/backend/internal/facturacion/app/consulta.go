package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	redisadapter "github.com/Bluethem/sw707u-grupo4/src/backend/adapters/redis"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
	"github.com/shopspring/decimal"
)

var (
	ErrDatosNoCoinciden  = errors.New("datos no coinciden")
	ErrConsultaExpirada  = errors.New("consulta expirada; solicite el archivo a la empresa")
	ErrConsultaRateLimit = errors.New("demasiados intentos")
)

const consultaDisponibilidadDias = 365

type ConsultaInput struct {
	RUC               string
	TipoDoc           string
	Numero            string
	Serie             string
	Fecha             string
	Monto             string
	NumeroComprobante int64
}

type ConsultaResult struct {
	ID                string
	Tipo              string
	Serie             string
	Numero            int64
	Fecha             string
	Total             decimal.Decimal
	Estado            string
	RazonSocialEmisor string
	Token             string
}

type ConsultaTokenStore interface {
	PutToken(ctx context.Context, token, empresaID, comprobanteID string) error
	GetToken(ctx context.Context, token string) (empresaID, comprobanteID string, ok bool, err error)
	IncrFail(ctx context.Context, ip string) (int64, error)
}

type Consulta struct {
	Repo   ports.ConsultaPublicaRepository
	Tokens ConsultaTokenStore
}

func (s *Consulta) Buscar(ctx context.Context, in ConsultaInput, clientIP string) (*ConsultaResult, error) {
	if s == nil || s.Repo == nil || s.Tokens == nil {
		return nil, fmt.Errorf("servicio consulta no configurado")
	}
	fecha, monto, err := parseConsultaInput(in)
	if err != nil {
		return s.fail(ctx, clientIP, ErrDatosNoCoinciden)
	}
	match, err := s.Repo.Match(ctx, ports.ConsultaPublicaCriteria{
		RUC:               strings.TrimSpace(in.RUC),
		TipoDoc:           strings.TrimSpace(in.TipoDoc),
		ClienteNumero:     strings.TrimSpace(in.Numero),
		Serie:             strings.TrimSpace(in.Serie),
		NumeroComprobante: in.NumeroComprobante,
		Fecha:             fecha,
		Monto:             monto,
	})
	if err != nil {
		return nil, err
	}
	if match == nil {
		return s.fail(ctx, clientIP, ErrDatosNoCoinciden)
	}
	limit := time.Now().UTC().AddDate(0, 0, -consultaDisponibilidadDias)
	if match.CreadoEn.Before(limit) {
		return s.fail(ctx, clientIP, ErrConsultaExpirada)
	}
	token, err := redisadapter.GenerateToken()
	if err != nil {
		return nil, err
	}
	if err := s.Tokens.PutToken(ctx, token, match.EmpresaID, match.ID); err != nil {
		return nil, err
	}
	return &ConsultaResult{
		ID:                match.ID,
		Tipo:              string(match.Tipo),
		Serie:             match.Serie,
		Numero:            match.Numero,
		Fecha:             match.CreadoEn.UTC().Format(time.DateOnly),
		Total:             match.Total,
		Estado:            string(match.Estado),
		RazonSocialEmisor: match.EmisorRazon,
		Token:             token,
	}, nil
}

func (s *Consulta) ValidarToken(ctx context.Context, token, comprobanteID string) (empresaID string, ok bool, err error) {
	if s == nil || s.Tokens == nil {
		return "", false, fmt.Errorf("servicio consulta no configurado")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", false, nil
	}
	empresaID, id, found, err := s.Tokens.GetToken(ctx, token)
	if err != nil || !found {
		return "", false, err
	}
	if id != comprobanteID {
		return "", false, nil
	}
	return empresaID, true, nil
}

func (s *Consulta) fail(ctx context.Context, clientIP string, base error) (*ConsultaResult, error) {
	count, err := s.Tokens.IncrFail(ctx, clientIP)
	if err != nil {
		return nil, err
	}
	if count > 10 {
		return nil, ErrConsultaRateLimit
	}
	return nil, base
}

func parseConsultaInput(in ConsultaInput) (time.Time, decimal.Decimal, error) {
	fecha, err := time.Parse(time.DateOnly, strings.TrimSpace(in.Fecha))
	if err != nil {
		return time.Time{}, decimal.Decimal{}, err
	}
	monto, err := decimal.NewFromString(strings.TrimSpace(in.Monto))
	if err != nil {
		return time.Time{}, decimal.Decimal{}, err
	}
	if in.NumeroComprobante < 1 {
		return time.Time{}, decimal.Decimal{}, fmt.Errorf("numero comprobante invalido")
	}
	return fecha, monto, nil
}
