package ubl

import (
	"bytes"
	"fmt"
	"html"
	"strings"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
)

const nsInvoice = "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"

func Generar(c *domain.Comprobante, emisorRUC, emisorRazon string) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("ubl: comprobante nil")
	}
	typeCode := "03"
	if c.Tipo == domain.TipoFactura {
		typeCode = "01"
	}
	id := fmt.Sprintf("%s-%d", c.Serie, c.Numero)
	issueDate := c.CreadoEn.UTC().Format("2006-01-02")
	if c.CreadoEn.IsZero() {
		issueDate = ""
	}

	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Fprintf(&b, `<Invoice xmlns="%s">`, nsInvoice)
	fmt.Fprintf(&b, `<cbc:ID xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">%s</cbc:ID>`, esc(id))
	fmt.Fprintf(&b, `<cbc:InvoiceTypeCode xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">%s</cbc:InvoiceTypeCode>`, typeCode)
	if issueDate != "" {
		fmt.Fprintf(&b, `<cbc:IssueDate xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">%s</cbc:IssueDate>`, issueDate)
	}
	b.WriteString(`<cac:AccountingSupplierParty xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2">`)
	b.WriteString(`<cac:Party><cac:PartyIdentification>`)
	fmt.Fprintf(&b, `<cbc:ID xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">%s</cbc:ID>`, esc(emisorRUC))
	b.WriteString(`</cac:PartyIdentification><cac:PartyLegalEntity>`)
	fmt.Fprintf(&b, `<cbc:RegistrationName xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">%s</cbc:RegistrationName>`, esc(emisorRazon))
	b.WriteString(`</cac:PartyLegalEntity></cac:Party></cac:AccountingSupplierParty>`)

	if c.Cliente != nil {
		b.WriteString(`<cac:AccountingCustomerParty xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2">`)
		b.WriteString(`<cac:Party><cac:PartyIdentification>`)
		fmt.Fprintf(&b, `<cbc:ID xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">%s</cbc:ID>`, esc(c.Cliente.Numero))
		b.WriteString(`</cac:PartyIdentification><cac:PartyLegalEntity>`)
		fmt.Fprintf(&b, `<cbc:RegistrationName xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">%s</cbc:RegistrationName>`, esc(c.Cliente.RazonSocial))
		b.WriteString(`</cac:PartyLegalEntity></cac:Party></cac:AccountingCustomerParty>`)
	}

	for i, linea := range c.Lineas {
		b.WriteString(`<cac:InvoiceLine xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2">`)
		fmt.Fprintf(&b, `<cbc:ID xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">%d</cbc:ID>`, i+1)
		fmt.Fprintf(&b, `<cbc:InvoicedQuantity xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">%d</cbc:InvoicedQuantity>`, linea.Cantidad)
		fmt.Fprintf(&b, `<cbc:LineExtensionAmount xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">%s</cbc:LineExtensionAmount>`, dec(linea.ValorVenta))
		fmt.Fprintf(&b, `<cbc:TaxAmount xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">%s</cbc:TaxAmount>`, dec(linea.IGV))
		fmt.Fprintf(&b, `<cbc:PayableAmount xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">%s</cbc:PayableAmount>`, dec(linea.Total))
		b.WriteString(`</cac:InvoiceLine>`)
	}

	b.WriteString(`<cac:LegalMonetaryTotal xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2">`)
	fmt.Fprintf(&b, `<cbc:LineExtensionAmount xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">%s</cbc:LineExtensionAmount>`, dec(c.ValorVenta))
	fmt.Fprintf(&b, `<cbc:TaxInclusiveAmount xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">%s</cbc:TaxInclusiveAmount>`, dec(c.Total))
	fmt.Fprintf(&b, `<cbc:PayableAmount xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">%s</cbc:PayableAmount>`, dec(c.Total))
	b.WriteString(`</cac:LegalMonetaryTotal>`)
	b.WriteString(`</Invoice>`)
	return b.Bytes(), nil
}

func dec(d interface{ StringFixed(int32) string }) string {
	return d.StringFixed(2)
}

func esc(s string) string {
	return html.EscapeString(strings.TrimSpace(s))
}
