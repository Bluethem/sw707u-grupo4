package auth

import "errors"

type Action string

const (
	ActionGestionarUsuarios Action = "gestionar_usuarios"
	ActionGestionarSeries   Action = "gestionar_series"
	ActionActualizarEmpresa Action = "actualizar_empresa"
)

var ErrRequireAdmin = errors.New("se requiere rol administrador")

func Allow(rol string, action Action) bool {
	switch action {
	case ActionGestionarUsuarios, ActionGestionarSeries, ActionActualizarEmpresa:
		return rol == RolAdministrador
	default:
		return false
	}
}

func RequireAdmin(rol string) error {
	if rol != RolAdministrador {
		return ErrRequireAdmin
	}
	return nil
}
