package app

import (
	"context"
	"fmt"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/pdf"
	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/ubl"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/ports"
)

type Documentos struct {
	Query   ports.ComprobanteQuery
	Storage ports.StoragePort
	Firma   ports.FirmaPort
	Pdf     ports.PdfPort
	Empresa func(ctx context.Context, empresaID string) (ruc, razon string, err error)
}

func (d *Documentos) PDF(ctx context.Context, empresaID, id string) ([]byte, error) {
	c, err := d.comprobante(ctx, empresaID, id)
	if err != nil {
		return nil, err
	}
	return d.regenerarPDF(ctx, c)
}

func (d *Documentos) XML(ctx context.Context, empresaID, id string) ([]byte, error) {
	c, err := d.comprobante(ctx, empresaID, id)
	if err != nil {
		return nil, err
	}
	if d.Storage != nil {
		if data, err := d.Storage.Leer(ctx, id+".xml"); err == nil {
			return data, nil
		}
	}
	return d.regenerarXML(ctx, c)
}

func (d *Documentos) CDR(ctx context.Context, empresaID, id string) ([]byte, error) {
	c, err := d.comprobante(ctx, empresaID, id)
	if err != nil {
		return nil, err
	}
	if d.Storage == nil {
		return nil, d.cdrNoDisponible(c)
	}
	data, err := d.Storage.Leer(ctx, id+".cdr")
	if err == nil {
		return data, nil
	}
	return nil, d.cdrNoDisponible(c)
}

func (d *Documentos) cdrNoDisponible(c *domain.Comprobante) error {
	switch c.Estado {
	case domain.EstadoPendiente, domain.EstadoEnviado:
		return domain.ErrCDRNoDisponible
	default:
		return domain.ErrCDRNoEncontrado
	}
}

func (d *Documentos) comprobante(ctx context.Context, empresaID, id string) (*domain.Comprobante, error) {
	if d.Query == nil {
		return nil, fmt.Errorf("servicio documentos no configurado")
	}
	c, err := d.Query.GetByID(ctx, empresaID, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, domain.ErrComprobanteNoEncontrado
	}
	return c, nil
}

func (d *Documentos) emisor(ctx context.Context, empresaID string) (ruc, razon string) {
	if d.Empresa == nil {
		return "", ""
	}
	ruc, razon, err := d.Empresa(ctx, empresaID)
	if err != nil {
		return "", ""
	}
	return ruc, razon
}

func (d *Documentos) regenerarXML(ctx context.Context, c *domain.Comprobante) ([]byte, error) {
	if d.Firma == nil {
		return nil, fmt.Errorf("xml no disponible")
	}
	emisorRUC, emisorRazon := d.emisor(ctx, c.EmpresaID)
	xmlData, err := ubl.Generar(c, emisorRUC, emisorRazon)
	if err != nil {
		return nil, err
	}
	firmado, err := d.Firma.FirmarXML(ctx, xmlData)
	if err != nil {
		return nil, err
	}
	if d.Storage != nil {
		_ = d.Storage.Guardar(ctx, c.ID+".xml", firmado)
	}
	return firmado, nil
}

func (d *Documentos) regenerarPDF(ctx context.Context, c *domain.Comprobante) ([]byte, error) {
	if d.Pdf == nil {
		return nil, fmt.Errorf("pdf no disponible")
	}
	emisorRUC, emisorRazon := d.emisor(ctx, c.EmpresaID)
	pdfCtx := pdf.WithEmisor(ctx, emisorRUC, emisorRazon)
	data, err := d.Pdf.Generar(pdfCtx, c)
	if err != nil {
		return nil, err
	}
	if d.Storage != nil {
		_ = d.Storage.Guardar(ctx, c.ID+".pdf", data)
	}
	return data, nil
}
