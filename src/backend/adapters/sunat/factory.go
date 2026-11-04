package sunat

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/config"
)

// GatewaySelection describe qué adaptador se eligió (sin secretos).
type GatewaySelection struct {
	Gateway ports.SunatGateway
	Mode    string // fake | beta | production
	URL     string
}

// NewFromConfig elige Fake o Client SOAP según credenciales y SUNAT_ENV.
func NewFromConfig(cfg config.Config) (GatewaySelection, error) {
	mode, url, err := cfg.ResolveSunatEndpoint()
	if err != nil {
		return GatewaySelection{}, err
	}
	if mode == "fake" {
		log.Printf("sunat=fake (creds incompletas; SUNAT_FAKE_MODE=%s)", cfg.SunatFakeMode)
		return GatewaySelection{
			Gateway: NewFake(Mode(cfg.SunatFakeMode)),
			Mode:    "fake",
		}, nil
	}

	endpoint := strings.TrimSpace(url)
	endpoint = strings.TrimSuffix(endpoint, "?wsdl")
	endpoint = strings.TrimSuffix(endpoint, "?WSDL")

	minInterval := cfg.SunatMinInterval
	if minInterval <= 0 {
		minInterval = 30 * time.Second
	}

	client := &Client{
		Endpoint:    endpoint,
		RUC:         cfg.SunatRUC,
		SOLUser:     cfg.SunatSOLUser,
		SOLPassword: cfg.SunatSOLPassword,
		MinInterval: minInterval,
		Mode:        mode,
	}
	log.Printf("sunat=%s url=%s min_interval=%s ruc=%s", mode, endpoint, minInterval, cfg.SunatRUC)
	return GatewaySelection{Gateway: client, Mode: mode, URL: endpoint}, nil
}

// MustNewFromConfig igual que NewFromConfig pero con panic controlado vía error tipado para main.
func MustExplain(sel GatewaySelection) string {
	if sel.Mode == "fake" {
		return fmt.Sprintf("sunat=%s", sel.Mode)
	}
	return fmt.Sprintf("sunat=%s url=%s", sel.Mode, sel.URL)
}
