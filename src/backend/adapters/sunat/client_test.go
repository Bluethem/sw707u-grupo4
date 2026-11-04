package sunat_test

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/sunat"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/platform/config"
)

func TestClient_Enviar_SendBillOK(t *testing.T) {
	cdr := []byte(`<ApplicationResponse><cbc:ResponseCode xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">0</cbc:ResponseCode><cbc:Description xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">Aceptado</cbc:Description></ApplicationResponse>`)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "sendBill") {
			t.Errorf("expected sendBill in body")
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "20123456789MODDATOS" || pass != "secret" {
			t.Errorf("basic auth=%q %q ok=%v", user, pass, ok)
		}
		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <sendBillResponse>
      <applicationResponse>` + base64.StdEncoding.EncodeToString(cdr) + `</applicationResponse>
    </sendBillResponse>
  </soap:Body>
</soap:Envelope>`))
	}))
	defer srv.Close()

	c := &sunat.Client{
		Endpoint:    srv.URL,
		RUC:         "20123456789",
		SOLUser:     "MODDATOS",
		SOLPassword: "secret",
		MinInterval: 0,
		HTTPClient:  srv.Client(),
	}
	xml := []byte(`<?xml version="1.0"?><Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"><cbc:ID xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">F001-1</cbc:ID><cbc:InvoiceTypeCode xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">01</cbc:InvoiceTypeCode></Invoice>`)
	res, err := c.Enviar(context.Background(), "c-1", xml)
	if err != nil {
		t.Fatalf("Enviar: %v", err)
	}
	if !res.Aceptado || res.Codigo != "0" {
		t.Fatalf("res=%+v", res)
	}
	if string(res.CDR) != string(cdr) {
		t.Fatalf("CDR mismatch")
	}
}

func TestClient_Enviar_RateLimit(t *testing.T) {
	n := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		cdr := base64.StdEncoding.EncodeToString([]byte(`<ApplicationResponse><ResponseCode>0</ResponseCode></ApplicationResponse>`))
		_, _ = w.Write([]byte(`<Envelope><Body><applicationResponse>` + cdr + `</applicationResponse></Body></Envelope>`))
	}))
	defer srv.Close()

	c := &sunat.Client{
		Endpoint: srv.URL, RUC: "20123456789", SOLUser: "U", SOLPassword: "p",
		MinInterval: 1 * time.Hour, HTTPClient: srv.Client(),
	}
	xml := []byte(`<Invoice><ID>F001-1</ID><InvoiceTypeCode>01</InvoiceTypeCode></Invoice>`)
	if _, err := c.Enviar(context.Background(), "c1", xml); err != nil {
		t.Fatalf("first: %v", err)
	}
	_, err := c.Enviar(context.Background(), "c2", xml)
	if err == nil || !strings.Contains(err.Error(), "rate limit") {
		t.Fatalf("expected rate limit, got %v", err)
	}
	if n != 1 {
		t.Fatalf("server hits=%d want 1", n)
	}
}

func TestClient_EnviarResumen_NotSupported(t *testing.T) {
	c := &sunat.Client{}
	_, err := c.EnviarResumen(context.Background(), "r1", []byte(`{}`))
	if err == nil || err != sunat.ErrResumenUBL {
		t.Fatalf("err=%v", err)
	}
}

func TestNewFromConfig_FakeAndBeta(t *testing.T) {
	sel, err := sunat.NewFromConfig(config.Config{SunatFakeMode: "accept"})
	if err != nil || sel.Mode != "fake" {
		t.Fatalf("sel=%+v err=%v", sel, err)
	}

	sel, err = sunat.NewFromConfig(config.Config{
		SunatRUC: "20123456789", SunatSOLUser: "U", SunatSOLPassword: "p",
		SunatCertPath: "/c.pfx", SunatCertPassword: "x",
		SunatEnv: "beta", SunatMinInterval: time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if sel.Mode != "beta" || sel.URL != config.DefaultSunatBetaURL {
		t.Fatalf("sel=%+v", sel)
	}
	if _, ok := sel.Gateway.(*sunat.Client); !ok {
		t.Fatalf("gateway type %T", sel.Gateway)
	}
}

func TestNewFromConfig_ProductionBlocked(t *testing.T) {
	_, err := sunat.NewFromConfig(config.Config{
		SunatRUC: "20123456789", SunatSOLUser: "U", SunatSOLPassword: "p",
		SunatCertPath: "/c.pfx", SunatCertPassword: "x",
		SunatEnv: "production", SunatAllowProduction: false,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
