package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	facturacion "github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
)

type Rol string

const (
	RolAdministrador Rol = "administrador"
	RolCajero        Rol = "cajero"
)

type TipoSerie string

const (
	TipoBoleta  TipoSerie = "boleta"
	TipoFactura TipoSerie = "factura"
)

var (
	ErrNoEncontrado = errors.New("recurso no encontrado")
	ErrConflicto    = errors.New("recurso duplicado")
	ErrInvalido     = errors.New("datos invalidos")

	formatoEmail = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	formatoSerie = regexp.MustCompile(`^[BF][A-Z0-9]{3}$`)
)

type Empresa struct {
	ID              string
	RUC             string
	RazonSocial     string
	DomicilioFiscal string
	Regimen         string
	Activa          bool
	CreadoEn        time.Time
}

type Usuario struct {
	ID           string
	EmpresaID    string
	Email        string
	PasswordHash string
	Nombre       string
	Rol          Rol
	Activo       bool
	CreadoEn     time.Time
}

type Serie struct {
	EmpresaID string
	Tipo      TipoSerie
	Serie     string
	Siguiente int64
	Activa    bool
}

const (
	TipoDocDNI = "DNI"
	TipoDocRUC = "RUC"
)

type Cliente struct {
	ID            string
	EmpresaID     string
	TipoDoc       string
	Numero        string
	RazonSocial   string
	Direccion     string
	Correo        string
	Activo        bool
	CreadoEn      time.Time
	ActualizadoEn time.Time
}

func ValidarCliente(c Cliente) error {
	switch c.TipoDoc {
	case TipoDocDNI:
		if len(c.Numero) != 8 || !soloDigitos(c.Numero) {
			return fmt.Errorf("%w: DNI invalido", ErrInvalido)
		}
	case TipoDocRUC:
		if err := facturacion.ValidarRUC(c.Numero); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalido, err)
		}
	default:
		return fmt.Errorf("%w: tipo_doc invalido", ErrInvalido)
	}
	if strings.TrimSpace(c.RazonSocial) == "" {
		return fmt.Errorf("%w: razon social requerida", ErrInvalido)
	}
	if strings.TrimSpace(c.Direccion) == "" {
		return fmt.Errorf("%w: direccion requerida", ErrInvalido)
	}
	correo := strings.TrimSpace(c.Correo)
	if correo != "" {
		if err := ValidarEmail(correo); err != nil {
			return err
		}
	}
	return nil
}

func soloDigitos(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func ValidarEmpresa(empresa Empresa) error {
	if err := facturacion.ValidarRUC(empresa.RUC); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalido, err)
	}
	if strings.TrimSpace(empresa.RazonSocial) == "" {
		return fmt.Errorf("%w: razon social requerida", ErrInvalido)
	}
	return nil
}

func NormalizarEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func ValidarEmail(email string) error {
	if !formatoEmail.MatchString(NormalizarEmail(email)) {
		return fmt.Errorf("%w: email invalido", ErrInvalido)
	}
	return nil
}

func ValidarRol(rol Rol) error {
	if rol != RolAdministrador && rol != RolCajero {
		return fmt.Errorf("%w: rol invalido", ErrInvalido)
	}
	return nil
}

func ValidarSerie(tipo TipoSerie, serie string) error {
	serie = strings.ToUpper(strings.TrimSpace(serie))
	if !formatoSerie.MatchString(serie) {
		return fmt.Errorf("%w: serie invalida", ErrInvalido)
	}
	if (tipo == TipoBoleta && serie[0] != 'B') || (tipo == TipoFactura && serie[0] != 'F') {
		return fmt.Errorf("%w: serie incompatible con tipo", ErrInvalido)
	}
	if tipo != TipoBoleta && tipo != TipoFactura {
		return fmt.Errorf("%w: tipo de serie invalido", ErrInvalido)
	}
	return nil
}
