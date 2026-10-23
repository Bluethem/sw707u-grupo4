package pdf

import (
	"bytes"
	"compress/zlib"
	"context"
	"regexp"
	"strconv"
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/shopspring/decimal"
)

func TestGenerar_QREnHeaderDerechaFueraDeTabla(t *testing.T) {
	gen := New("http://localhost:3000")
	ctx := WithEmisor(context.Background(), "20100070970", "FlowFacture Demo SAC")
	c := &domain.Comprobante{
		ID:     "cmp-qr-1",
		Tipo:   domain.TipoBoleta,
		Serie:  "B001",
		Numero: 12,
		Estado: domain.EstadoAceptado,
		Cliente: &domain.DatosCliente{
			TipoDoc: "DNI", Numero: "12345678", RazonSocial: "Ana Perez",
		},
		ValorVenta: decimal.RequireFromString("50.85"),
		IGV:        decimal.RequireFromString("9.15"),
		Total:      decimal.RequireFromString("60.00"),
		Lineas: []domain.LineaComprobante{
			lineaPDF("p1", 1, "10.00"),
			lineaPDF("p2", 2, "12.00"),
			lineaPDF("p3", 1, "14.00"),
			lineaPDF("p4", 3, "8.00"),
			lineaPDF("p5", 1, "6.00"),
		},
	}

	data, err := gen.Generar(ctx, c)
	if err != nil {
		t.Fatalf("Generar: %v", err)
	}
	boxes := pdfImageBoxesMM(t, data)
	if len(boxes) != 1 {
		t.Fatalf("esperaba 1 imagen QR, got %d (%+v)", len(boxes), boxes)
	}
	qr := boxes[0]
	if qr.X < 145 {
		t.Errorf("QR debe ir a la derecha del header, x=%.1f", qr.X)
	}
	if qr.Y >= 40 {
		t.Errorf("QR solapa la tabla de productos: y=%.1f", qr.Y)
	}
	if qr.Y+qr.H > 70 {
		t.Errorf("QR encuadrado debe terminar antes de la tabla: bottom=%.1f", qr.Y+qr.H)
	}
	if qr.W < 28 || qr.H < 28 {
		t.Errorf("QR demasiado pequeño: %+v", qr)
	}
}

func lineaPDF(id string, cant int64, precio string) domain.LineaComprobante {
	p := decimal.RequireFromString(precio)
	total := p.Mul(decimal.NewFromInt(cant))
	return domain.LineaComprobante{
		ProductoID: id, Cantidad: cant, PrecioUnit: p,
		Afectacion: domain.AfectacionGravado, ValorVenta: total, Total: total,
	}
}

type imageBox struct {
	X, Y, W, H float64
}

func pdfImageBoxesMM(t *testing.T, pdf []byte) []imageBox {
	t.Helper()
	const (
		ptToMM = 25.4 / 72.0
		pageH  = 297.0
	)
	content := inflatePDFStreams(pdf)
	re := regexp.MustCompile(`([0-9.]+) 0 0 ([0-9.]+) ([0-9.]+) ([0-9.]+) cm`)
	matches := re.FindAllSubmatch(content, -1)
	var out []imageBox
	for _, m := range matches {
		wPt, _ := strconv.ParseFloat(string(m[1]), 64)
		hPt, _ := strconv.ParseFloat(string(m[2]), 64)
		xPt, _ := strconv.ParseFloat(string(m[3]), 64)
		pdfY, _ := strconv.ParseFloat(string(m[4]), 64)
		w, h := wPt*ptToMM, hPt*ptToMM
		if w < 10 || h < 10 {
			continue
		}
		out = append(out, imageBox{
			X: xPt * ptToMM,
			Y: pageH - (pdfY * ptToMM) - h,
			W: w,
			H: h,
		})
	}
	if len(out) == 0 {
		t.Fatalf("no se encontró imagen en PDF (len=%d inflated=%d)", len(pdf), len(content))
	}
	return out
}

func inflatePDFStreams(pdf []byte) []byte {
	re := regexp.MustCompile(`(?s)stream\r?\n(.*?)\r?\nendstream`)
	var buf bytes.Buffer
	buf.Write(pdf)
	for _, m := range re.FindAllSubmatch(pdf, -1) {
		raw := m[1]
		r, err := zlib.NewReader(bytes.NewReader(raw))
		if err != nil {
			continue
		}
		_, _ = buf.ReadFrom(r)
		_ = r.Close()
	}
	return buf.Bytes()
}
