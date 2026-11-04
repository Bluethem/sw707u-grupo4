package sunat

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
)

const (
	soapNS    = "http://schemas.xmlsoap.org/soap/envelope/"
	serNS     = "http://service.sunat.gob.pe"
	wsseNS    = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
	defaultTO = 60 * time.Second
)

var (
	reInvoiceID   = regexp.MustCompile(`(?s)<(?:[\w-]+:)?ID[^>]*>([^<]+)</(?:[\w-]+:)?ID>`)
	reTypeCode    = regexp.MustCompile(`(?s)<(?:[\w-]+:)?InvoiceTypeCode[^>]*>([^<]+)</(?:[\w-]+:)?InvoiceTypeCode>`)
	reAppResponse = regexp.MustCompile(`(?s)<(?:[\w-]+:)?applicationResponse[^>]*>([^<]+)</(?:[\w-]+:)?applicationResponse>`)
	reFaultString = regexp.MustCompile(`(?s)<(?:[\w-]+:)?faultstring[^>]*>([^<]+)</(?:[\w-]+:)?faultstring>`)
)

// ErrRateLimited se devuelve cuando aún no pasó MinInterval (retriable vía outbox).
var ErrRateLimited = fmt.Errorf("sunat: rate limit local (espere SUNAT_MIN_INTERVAL)")

// ErrResumenUBL indica que falta UBL de resumen/baja (fuera del MVP beta sendBill).
var ErrResumenUBL = fmt.Errorf("sunat: UBL de resumen/baja no generado; solo sendBill de factura/boleta firmada está soportado en este adaptador")

// Client habla con billService (beta o producción).
type Client struct {
	Endpoint    string
	RUC         string
	SOLUser     string
	SOLPassword string
	MinInterval time.Duration
	HTTPClient  *http.Client
	Mode        string // beta|production (solo informativo)

	mu       sync.Mutex
	lastCall time.Time
}

func (c *Client) waitRateLimit(_ context.Context) error {
	if c.MinInterval <= 0 {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Since(c.lastCall) < c.MinInterval {
		return ErrRateLimited
	}
	c.lastCall = time.Now()
	return nil
}

func (c *Client) http() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: defaultTO}
}

func (c *Client) Enviar(ctx context.Context, comprobanteID string, xmlBytes []byte) (ports.SunatResult, error) {
	xmlBytes = stripFakeFirmaPrefix(xmlBytes)
	fileBase, err := buildFileBase(c.RUC, xmlBytes, comprobanteID)
	if err != nil {
		return ports.SunatResult{}, err
	}
	zipBytes, err := zipXML(fileBase+".xml", xmlBytes)
	if err != nil {
		return ports.SunatResult{}, err
	}
	return c.sendBill(ctx, fileBase+".zip", zipBytes)
}

func (c *Client) EnviarBaja(ctx context.Context, comprobanteID, motivo string) (ports.SunatResult, error) {
	_ = ctx
	_ = comprobanteID
	_ = motivo
	return ports.SunatResult{}, ErrResumenUBL
}

func (c *Client) EnviarResumen(ctx context.Context, resumenID string, payload []byte) (ports.SunatResult, error) {
	_ = ctx
	_ = resumenID
	_ = payload
	return ports.SunatResult{}, ErrResumenUBL
}

func (c *Client) sendBill(ctx context.Context, fileName string, zipContent []byte) (ports.SunatResult, error) {
	if err := c.waitRateLimit(ctx); err != nil {
		return ports.SunatResult{}, err
	}

	body := buildSendBillEnvelope(c.RUC, c.SOLUser, c.SOLPassword, fileName, zipContent)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(body))
	if err != nil {
		return ports.SunatResult{}, err
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", `""`)
	req.SetBasicAuth(c.RUC+c.SOLUser, c.SOLPassword)

	res, err := c.http().Do(req)
	if err != nil {
		return ports.SunatResult{}, fmt.Errorf("sunat sendBill: %w", err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return ports.SunatResult{}, err
	}
	return parseSendBillResponse(res.StatusCode, raw)
}

func buildSendBillEnvelope(ruc, user, password, fileName string, zipContent []byte) []byte {
	b64 := base64.StdEncoding.EncodeToString(zipContent)
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString(`<soapenv:Envelope xmlns:soapenv="` + soapNS + `" xmlns:ser="` + serNS + `" xmlns:wsse="` + wsseNS + `">`)
	b.WriteString(`<soapenv:Header><wsse:Security>`)
	b.WriteString(`<wsse:UsernameToken>`)
	fmt.Fprintf(&b, `<wsse:Username>%s</wsse:Username>`, xmlEscape(ruc+user))
	fmt.Fprintf(&b, `<wsse:Password>%s</wsse:Password>`, xmlEscape(password))
	b.WriteString(`</wsse:UsernameToken></wsse:Security></soapenv:Header>`)
	b.WriteString(`<soapenv:Body><ser:sendBill>`)
	fmt.Fprintf(&b, `<fileName>%s</fileName>`, xmlEscape(fileName))
	fmt.Fprintf(&b, `<contentFile>%s</contentFile>`, b64)
	b.WriteString(`</ser:sendBill></soapenv:Body></soapenv:Envelope>`)
	return []byte(b.String())
}

func parseSendBillResponse(status int, raw []byte) (ports.SunatResult, error) {
	s := string(raw)
	if m := reFaultString.FindStringSubmatch(s); len(m) == 2 {
		msg := strings.TrimSpace(m[1])
		return ports.SunatResult{Aceptado: false, Codigo: "SOAP_FAULT", Mensaje: msg}, nil
	}
	if m := reAppResponse.FindStringSubmatch(s); len(m) == 2 {
		cdr, err := base64.StdEncoding.DecodeString(strings.TrimSpace(m[1]))
		if err != nil {
			return ports.SunatResult{}, fmt.Errorf("sunat: decode CDR: %w", err)
		}
		aceptado, codigo, mensaje := interpretCDR(cdr)
		return ports.SunatResult{Aceptado: aceptado, Codigo: codigo, Mensaje: mensaje, CDR: cdr}, nil
	}
	if status >= 400 {
		return ports.SunatResult{}, fmt.Errorf("sunat: HTTP %d: %s", status, truncate(s, 300))
	}
	return ports.SunatResult{}, fmt.Errorf("sunat: respuesta sin applicationResponse: %s", truncate(s, 300))
}

func interpretCDR(cdr []byte) (aceptado bool, codigo, mensaje string) {
	// Heurística: ResponseCode 0 / "0" = aceptado; otros = observado/rechazo.
	reCode := regexp.MustCompile(`(?s)<(?:[\w-]+:)?ResponseCode[^>]*>([^<]+)</(?:[\w-]+:)?ResponseCode>`)
	reDesc := regexp.MustCompile(`(?s)<(?:[\w-]+:)?Description[^>]*>([^<]+)</(?:[\w-]+:)?Description>`)
	codigo = "0"
	mensaje = "CDR recibido"
	if m := reCode.FindStringSubmatch(string(cdr)); len(m) == 2 {
		codigo = strings.TrimSpace(m[1])
	}
	if m := reDesc.FindStringSubmatch(string(cdr)); len(m) == 2 {
		mensaje = strings.TrimSpace(m[1])
	}
	aceptado = codigo == "0" || codigo == "00"
	return aceptado, codigo, mensaje
}

func zipXML(name string, xmlBytes []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.Create(name)
	if err != nil {
		return nil, err
	}
	if _, err := f.Write(xmlBytes); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func buildFileBase(ruc string, xmlBytes []byte, fallbackID string) (string, error) {
	tipo := "01"
	if m := reTypeCode.FindStringSubmatch(string(xmlBytes)); len(m) == 2 {
		tipo = strings.TrimSpace(m[1])
	}
	id := fallbackID
	if m := reInvoiceID.FindStringSubmatch(string(xmlBytes)); len(m) == 2 {
		id = strings.TrimSpace(m[1])
	}
	id = strings.ReplaceAll(id, " ", "")
	if ruc == "" || id == "" {
		return "", fmt.Errorf("sunat: no se pudo armar nombre de archivo (ruc/id)")
	}
	return fmt.Sprintf("%s-%s-%s", ruc, tipo, id), nil
}

func stripFakeFirmaPrefix(xml []byte) []byte {
	const prefix = "SIGNED:"
	if bytes.HasPrefix(xml, []byte(prefix)) {
		return xml[len(prefix):]
	}
	return xml
}

func xmlEscape(s string) string {
	var b strings.Builder
	xml.EscapeText(&b, []byte(s))
	return b.String()
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// Ensure Client implements the port.
var _ ports.SunatGateway = (*Client)(nil)
