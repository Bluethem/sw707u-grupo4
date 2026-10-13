package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres/migrate"
	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	platformauth "github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestGestionRepositoriesTenantIsolation(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL no configurado")
	}
	ctx := context.Background()
	adminPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("crear pool administrativo: %v", err)
	}
	t.Cleanup(adminPool.Close)
	if err := adminPool.Ping(ctx); err != nil {
		t.Fatalf("conectar postgres: %v", err)
	}
	esquema := "gestion_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+esquema); err != nil {
		t.Fatalf("crear esquema aislado: %v", err)
	}
	t.Cleanup(func() {
		_, _ = adminPool.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+esquema+" CASCADE")
	})

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("leer DATABASE_URL: %v", err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = esquema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("crear pool aislado: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := migrate.Up(ctx, pool); err != nil {
		t.Fatalf("migrar: %v", err)
	}

	empresas := NewEmpresaRepository(pool)
	usuarios := NewUsuarioRepository(pool)
	series := NewSeriesGestionRepository(pool)
	clientes := NewClienteRepository(pool)
	service := gestionapp.NewService(empresas, usuarios, series, clientes)

	sufijo := uuid.NewString()
	empresaA, err := service.CrearEmpresa(ctx, gestionapp.CrearEmpresaInput{
		RUC: rucUnico(t, sufijo+"a"), RazonSocial: "Tenant A",
	})
	if err != nil {
		t.Fatalf("crear empresa A: %v", err)
	}
	var empresaB gestiondomain.Empresa
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM series_correlativos WHERE empresa_id IN ($1, $2)`, empresaA.ID, empresaB.ID); err != nil {
			t.Errorf("limpiar series: %v", err)
		}
		if _, err := pool.Exec(context.Background(), `DELETE FROM usuarios WHERE empresa_id IN ($1, $2)`, empresaA.ID, empresaB.ID); err != nil {
			t.Errorf("limpiar usuarios: %v", err)
		}
		if _, err := pool.Exec(context.Background(), `DELETE FROM empresas WHERE id IN ($1, $2)`, empresaA.ID, empresaB.ID); err != nil {
			t.Errorf("limpiar empresas: %v", err)
		}
	})
	empresaB, err = service.CrearEmpresa(ctx, gestionapp.CrearEmpresaInput{
		RUC: rucUnico(t, sufijo+"b"), RazonSocial: "Tenant B",
	})
	if err != nil {
		t.Fatalf("crear empresa B: %v", err)
	}

	usuarioA, err := service.CrearUsuario(ctx, empresaA.ID, gestionapp.CrearUsuarioInput{
		Email: " ADMIN-" + sufijo + "@example.com ", Password: "secreto123", Nombre: "Admin A", Rol: gestiondomain.RolAdministrador,
	})
	if err != nil {
		t.Fatalf("crear usuario A: %v", err)
	}
	if usuarioA.Email != "admin-"+sufijo+"@example.com" || !platformauth.CheckPassword(usuarioA.PasswordHash, "secreto123") {
		t.Fatal("usuario no normalizo email o no almaceno un hash valido")
	}
	if _, err := service.CrearUsuario(ctx, empresaB.ID, gestionapp.CrearUsuarioInput{
		Email: "admin-b-" + sufijo + "@example.com", Password: "secreto123", Rol: gestiondomain.RolAdministrador,
	}); err != nil {
		t.Fatalf("crear usuario B: %v", err)
	}

	listaA, err := service.ListarUsuarios(ctx, empresaA.ID)
	if err != nil || len(listaA) != 1 || listaA[0].EmpresaID != empresaA.ID {
		t.Fatalf("usuarios tenant A = %#v, err = %v", listaA, err)
	}
	nombre := "Intruso"
	if _, err := service.ActualizarUsuario(ctx, empresaB.ID, usuarioA.ID, gestionapp.ActualizarUsuarioInput{Nombre: &nombre}); !errors.Is(err, gestiondomain.ErrNoEncontrado) {
		t.Fatalf("actualizacion cross-tenant debio ocultar usuario: %v", err)
	}

	if _, err := service.CrearSerie(ctx, empresaA.ID, gestionapp.CrearSerieInput{
		Tipo: gestiondomain.TipoBoleta, Serie: "b001",
	}); err != nil {
		t.Fatalf("crear serie A: %v", err)
	}
	if _, err := service.CrearSerie(ctx, empresaB.ID, gestionapp.CrearSerieInput{
		Tipo: gestiondomain.TipoFactura, Serie: "F001", Siguiente: 10,
	}); err != nil {
		t.Fatalf("crear serie B: %v", err)
	}
	seriesA, err := service.ListarSeries(ctx, empresaA.ID)
	if err != nil || len(seriesA) != 1 || seriesA[0].Serie != "B001" || seriesA[0].Siguiente != 1 {
		t.Fatalf("series tenant A = %#v, err = %v", seriesA, err)
	}
	seriesB, err := service.ListarSeries(ctx, empresaB.ID)
	if err != nil || len(seriesB) != 1 || seriesB[0].Serie != "F001" {
		t.Fatalf("series tenant B = %#v, err = %v", seriesB, err)
	}
}

func rucUnico(t *testing.T, semilla string) string {
	t.Helper()
	hash := sha256.Sum256([]byte(semilla))
	base := []byte("2000000000")
	for i := 1; i < len(base); i++ {
		base[i] = '0' + hash[i]%10
	}
	for digito := byte('0'); digito <= '9'; digito++ {
		candidato := string(base) + string(digito)
		if gestiondomain.ValidarEmpresa(gestiondomain.Empresa{RUC: candidato, RazonSocial: "x"}) == nil {
			return candidato
		}
	}
	t.Fatalf("no se pudo construir RUC valido desde %q", semilla)
	return ""
}
