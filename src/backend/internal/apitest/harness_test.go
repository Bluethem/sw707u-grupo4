package apitest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres"
	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/postgres/migrate"
	redisadapter "github.com/Bluethem/sw707u-grupo4/src/backend/adapters/redis"
	gestionapp "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/app"
	gestiondomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/gestion/domain"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/httpapi"
	"github.com/jackc/pgx/v5/pgxpool"
	redislib "github.com/redis/go-redis/v9"
)

type harness struct {
	server   *httptest.Server
	pool     *pgxpool.Pool
	redis    *redislib.Client
	gestion  *gestionapp.Service
	usuarios *postgres.UsuarioRepository
	sessions *redisadapter.SessionStore
}

type seedData struct {
	EmpresaID     string
	AdminID       string
	AdminEmail    string
	AdminPassword string
	Token         string
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	databaseURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")
	if databaseURL == "" || redisURL == "" {
		t.Skip("tests API requieren DATABASE_URL y REDIS_URL")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("conectar Postgres: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping Postgres: %v", err)
	}
	if err := migrate.Up(ctx, pool); err != nil {
		t.Fatalf("migrar Postgres: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		TRUNCATE TABLE
			outbox_envios, comprobante_lineas, comprobantes, productos_stock,
			series_correlativos, usuarios, empresas
		RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("limpiar Postgres: %v", err)
	}

	redisOptions, err := redislib.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("parsear REDIS_URL: %v", err)
	}
	if redisOptions.DB == 0 {
		t.Fatalf("REDIS_URL debe apuntar a una DB de test distinta de 0 (por ejemplo /1)")
	}
	redisClient := redislib.NewClient(redisOptions)
	t.Cleanup(func() { _ = redisClient.Close() })
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Fatalf("ping Redis: %v", err)
	}
	if err := redisClient.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("limpiar Redis: %v", err)
	}

	empresas := postgres.NewEmpresaRepository(pool)
	usuarios := postgres.NewUsuarioRepository(pool)
	series := postgres.NewSeriesGestionRepository(pool)
	gestion := gestionapp.NewService(empresas, usuarios, series)
	sessions := redisadapter.NewSessionStore(redisClient, time.Hour)
	server := httptest.NewServer(httpapi.NewAPIHandler(httpapi.Dependencies{
		Gestion: gestion, Usuarios: usuarios, Sessions: sessions,
	}))
	t.Cleanup(server.Close)

	return &harness{
		server: server, pool: pool, redis: redisClient, gestion: gestion,
		usuarios: usuarios, sessions: sessions,
	}
}

func (h *harness) SeedEmpresaAdmin(t *testing.T) seedData {
	t.Helper()
	const (
		email    = "admin@empresa.test"
		password = "Admin123!"
	)
	empresa, err := h.gestion.CrearEmpresa(context.Background(), gestionapp.CrearEmpresaInput{
		RUC: "20100070970", RazonSocial: "Empresa API SAC",
		DomicilioFiscal: "Lima", Regimen: "general",
	})
	if err != nil {
		t.Fatalf("seed empresa: %v", err)
	}
	admin, err := h.gestion.CrearUsuario(context.Background(), empresa.ID, gestionapp.CrearUsuarioInput{
		Email: email, Password: password, Nombre: "Admin API", Rol: gestiondomain.RolAdministrador,
	})
	if err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	token, err := redisadapter.GenerateToken()
	if err != nil {
		t.Fatalf("generar token seed: %v", err)
	}
	if err := h.sessions.Create(context.Background(), token, redisadapter.SessionData{
		UserID: admin.ID, EmpresaID: empresa.ID, Rol: string(admin.Rol),
	}); err != nil {
		t.Fatalf("crear sesion seed: %v", err)
	}
	return seedData{
		EmpresaID: empresa.ID, AdminID: admin.ID, AdminEmail: email,
		AdminPassword: password, Token: token,
	}
}

func (h *harness) request(t *testing.T, method, path, token string, body any) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("codificar request: %v", err)
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, h.server.URL+path, reader)
	if err != nil {
		t.Fatalf("crear request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := h.server.Client().Do(req)
	if err != nil {
		t.Fatalf("ejecutar request: %v", err)
	}
	return res
}

func decodeResponse(t *testing.T, res *http.Response, target any) {
	t.Helper()
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		t.Fatalf("decodificar respuesta HTTP %d: %v", res.StatusCode, err)
	}
}

func requireStatus(t *testing.T, res *http.Response, want int) {
	t.Helper()
	if res.StatusCode == want {
		return
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	t.Fatalf("status = %d, esperado %d; body=%s", res.StatusCode, want, body)
}
