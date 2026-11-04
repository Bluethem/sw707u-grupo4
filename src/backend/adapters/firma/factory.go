package firma

import (
	"fmt"
	"log"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
)

// NewFromPaths usa PKCS#12 si path/password no están vacíos; si no, Fake.
func NewFromPaths(certPath, certPassword string) (ports.FirmaPort, string, error) {
	if certPath == "" || certPassword == "" {
		return Fake{}, "fake", nil
	}
	p, err := NewPKCS12FromFile(certPath, certPassword)
	if err != nil {
		return nil, "", fmt.Errorf("firma real: %w", err)
	}
	log.Printf("firma: pkcs12 loaded from %s", certPath)
	return p, "pkcs12", nil
}
