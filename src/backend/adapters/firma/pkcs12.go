package firma

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/beevik/etree"
	dsig "github.com/russellhaering/goxmldsig"
	"software.sslmate.com/src/go-pkcs12"
)

// PKCS12 firma XML UBL con certificado .pfx/.p12 (XMLDSig enveloped).
type PKCS12 struct {
	key  *rsa.PrivateKey
	cert *x509.Certificate
}

func NewPKCS12FromFile(path, password string) (*PKCS12, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("firma: leer cert: %w", err)
	}
	key, cert, err := pkcs12.Decode(raw, password)
	if err != nil {
		return nil, fmt.Errorf("firma: decode pkcs12: %w", err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("firma: la clave del certificado no es RSA")
	}
	return &PKCS12{key: rsaKey, cert: cert}, nil
}

func (p *PKCS12) FirmarXML(_ context.Context, xml []byte) ([]byte, error) {
	if p == nil || p.key == nil || p.cert == nil {
		return nil, fmt.Errorf("firma: pkcs12 no inicializado")
	}
	xml = stripFakePrefix(xml)
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(xml); err != nil {
		return nil, fmt.Errorf("firma: parse xml: %w", err)
	}
	root := doc.Root()
	if root == nil {
		return nil, fmt.Errorf("firma: xml sin raíz")
	}

	ctx, err := dsig.NewSigningContext(p.key, [][]byte{p.cert.Raw})
	if err != nil {
		return nil, fmt.Errorf("firma: signing context: %w", err)
	}
	signed, err := ctx.SignEnveloped(root)
	if err != nil {
		return nil, fmt.Errorf("firma: sign enveloped: %w", err)
	}
	out := etree.NewDocument()
	out.SetRoot(signed)
	out.WriteSettings.CanonicalAttrVal = true
	b, err := out.WriteToBytes()
	if err != nil {
		return nil, fmt.Errorf("firma: serialize: %w", err)
	}
	return b, nil
}

func stripFakePrefix(xml []byte) []byte {
	const prefix = "SIGNED:"
	s := string(xml)
	if strings.HasPrefix(s, prefix) {
		return []byte(strings.TrimPrefix(s, prefix))
	}
	return xml
}
