package firma_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/firma"
	"software.sslmate.com/src/go-pkcs12"
)

func TestPKCS12_FirmarXML_Enveloped(t *testing.T) {
	path, pass := writeTestPFX(t)
	signer, err := firma.NewPKCS12FromFile(path, pass)
	if err != nil {
		t.Fatalf("NewPKCS12FromFile: %v", err)
	}
	xml := []byte(`<?xml version="1.0" encoding="UTF-8"?><Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"><cbc:ID xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">F001-1</cbc:ID></Invoice>`)
	out, err := signer.FirmarXML(context.Background(), xml)
	if err != nil {
		t.Fatalf("FirmarXML: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "Signature") {
		t.Fatalf("expected Signature element, got %s", s[:min(200, len(s))])
	}
	if strings.HasPrefix(s, "SIGNED:") {
		t.Fatal("should not keep fake prefix")
	}
}

func TestPKCS12_StripsFakePrefix(t *testing.T) {
	path, pass := writeTestPFX(t)
	signer, err := firma.NewPKCS12FromFile(path, pass)
	if err != nil {
		t.Fatal(err)
	}
	xml := []byte(`SIGNED:<?xml version="1.0"?><Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"><x/></Invoice>`)
	out, err := signer.FirmarXML(context.Background(), xml)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(string(out), "SIGNED:") {
		t.Fatal("fake prefix not stripped")
	}
}

func TestNewFromPaths_FakeWhenEmpty(t *testing.T) {
	p, mode, err := firma.NewFromPaths("", "")
	if err != nil || mode != "fake" {
		t.Fatalf("mode=%q err=%v", mode, err)
	}
	out, err := p.FirmarXML(context.Background(), []byte("<a/>"))
	if err != nil || !strings.HasPrefix(string(out), "SIGNED:") {
		t.Fatalf("fake behavior broken: %q err=%v", out, err)
	}
}

func writeTestPFX(t *testing.T) (path, password string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test-emisor"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	password = "testpass"
	pfx, err := pkcs12.Legacy.Encode(key, cert, nil, password)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path = filepath.Join(dir, "test.pfx")
	if err := os.WriteFile(path, pfx, 0o600); err != nil {
		t.Fatal(err)
	}
	return path, password
}
