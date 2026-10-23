package pdf

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
	"github.com/jung-kurt/gofpdf"
	qrcode "github.com/skip2/go-qrcode"
)

type emisorCtxKey struct{}

type emisorInfo struct {
	RUC   string
	Razon string
}

// WithEmisor adjunta datos del emisor para el PDF y el QR de consulta pública.
func WithEmisor(ctx context.Context, ruc, razon string) context.Context {
	return context.WithValue(ctx, emisorCtxKey{}, emisorInfo{
		RUC:   strings.TrimSpace(ruc),
		Razon: strings.TrimSpace(razon),
	})
}

func emisorFrom(ctx context.Context) emisorInfo {
	if v, ok := ctx.Value(emisorCtxKey{}).(emisorInfo); ok {
		return v
	}
	return emisorInfo{}
}

type Generator struct {
	baseURL string
}

func New(frontendPublicURL string) *Generator {
	return &Generator{baseURL: strings.TrimRight(strings.TrimSpace(frontendPublicURL), "/")}
}

const (
	pageWidthMM = 210.0
	marginMM    = 10.0
	leftColW    = 138.0
	qrBoxW      = 44.0
	qrSizeMM    = 32.0
	headerTopY  = 10.0
)

func (g *Generator) Generar(ctx context.Context, c *domain.Comprobante) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("pdf: comprobante nil")
	}
	emisor := emisorFrom(ctx)
	consultaURL := g.consultaURL(c, emisor.RUC)

	doc := gofpdf.New("P", "mm", "A4", "")
	doc.AddPage()

	qrBoxX := pageWidthMM - marginMM - qrBoxW
	identH := 16.0
	doc.SetDrawColor(40, 40, 40)
	doc.Rect(qrBoxX, headerTopY, qrBoxW, identH, "D")
	doc.SetXY(qrBoxX, headerTopY+1.5)
	doc.SetFont("Arial", "B", 9)
	doc.MultiCell(qrBoxW, 5, fmt.Sprintf("%s\n%s-%d", strings.ToUpper(string(c.Tipo)), c.Serie, c.Numero), "", "C", false)

	qrX := qrBoxX + (qrBoxW-qrSizeMM)/2
	qrY := headerTopY + identH + 3
	doc.Rect(qrX-1, qrY-1, qrSizeMM+2, qrSizeMM+2, "D")
	if consultaURL != "" {
		png, err := qrcode.Encode(consultaURL, qrcode.Medium, 128)
		if err != nil {
			return nil, fmt.Errorf("pdf qr: %w", err)
		}
		name := fmt.Sprintf("qr-%s.png", c.ID)
		doc.RegisterImageOptionsReader(name, gofpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(png))
		doc.ImageOptions(name, qrX, qrY, qrSizeMM, qrSizeMM, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}
	rightBottom := qrY + qrSizeMM + 2

	doc.SetXY(marginMM, headerTopY)
	doc.SetFont("Arial", "B", 12)
	if emisor.Razon != "" {
		doc.CellFormat(leftColW, 6, emisor.Razon, "", 1, "L", false, 0, "")
	}
	doc.SetFont("Arial", "", 10)
	if emisor.RUC != "" {
		doc.CellFormat(leftColW, 6, "RUC: "+emisor.RUC, "", 1, "L", false, 0, "")
	}
	doc.Ln(1)
	doc.SetFont("Arial", "B", 10)
	doc.CellFormat(leftColW, 6, "Cliente", "", 1, "L", false, 0, "")
	doc.SetFont("Arial", "", 10)
	if c.Cliente != nil {
		if c.Cliente.RazonSocial != "" {
			doc.CellFormat(leftColW, 5, c.Cliente.RazonSocial, "", 1, "L", false, 0, "")
		}
		if c.Cliente.TipoDoc != "" || c.Cliente.Numero != "" {
			doc.CellFormat(leftColW, 5, fmt.Sprintf("%s %s", c.Cliente.TipoDoc, c.Cliente.Numero), "", 1, "L", false, 0, "")
		}
	} else {
		doc.CellFormat(leftColW, 5, "—", "", 1, "L", false, 0, "")
	}
	doc.CellFormat(leftColW, 5, fmt.Sprintf("Estado: %s", c.Estado), "", 1, "L", false, 0, "")

	headerBottom := doc.GetY()
	if rightBottom > headerBottom {
		headerBottom = rightBottom
	}
	doc.SetY(headerBottom + 4)

	doc.SetFont("Arial", "B", 9)
	colW := []float64{70, 18, 28, 22, 28, 24}
	headers := []string{"Descripción", "Cant.", "P. unit.", "Afect.", "V. venta", "Total"}
	for i, h := range headers {
		doc.CellFormat(colW[i], 7, h, "1", 0, "C", false, 0, "")
	}
	doc.Ln(-1)
	doc.SetFont("Arial", "", 9)
	for _, l := range c.Lineas {
		desc := strings.TrimSpace(l.ProductoID)
		if desc == "" {
			desc = "—"
		}
		doc.CellFormat(colW[0], 7, truncateRunes(desc, 42), "1", 0, "L", false, 0, "")
		doc.CellFormat(colW[1], 7, fmt.Sprintf("%d", l.Cantidad), "1", 0, "R", false, 0, "")
		doc.CellFormat(colW[2], 7, l.PrecioUnit.StringFixed(2), "1", 0, "R", false, 0, "")
		doc.CellFormat(colW[3], 7, string(l.Afectacion), "1", 0, "C", false, 0, "")
		doc.CellFormat(colW[4], 7, l.ValorVenta.StringFixed(2), "1", 0, "R", false, 0, "")
		doc.CellFormat(colW[5], 7, l.Total.StringFixed(2), "1", 0, "R", false, 0, "")
		doc.Ln(-1)
	}
	if len(c.Lineas) == 0 {
		doc.CellFormat(190, 7, "Sin líneas", "1", 1, "C", false, 0, "")
	}

	doc.Ln(4)
	doc.SetFont("Arial", "", 10)
	doc.CellFormat(130, 6, "", "", 0, "L", false, 0, "")
	doc.CellFormat(30, 6, "Valor venta:", "", 0, "R", false, 0, "")
	doc.CellFormat(30, 6, "S/ "+c.ValorVenta.StringFixed(2), "", 1, "R", false, 0, "")
	doc.CellFormat(130, 6, "", "", 0, "L", false, 0, "")
	doc.CellFormat(30, 6, "IGV:", "", 0, "R", false, 0, "")
	doc.CellFormat(30, 6, "S/ "+c.IGV.StringFixed(2), "", 1, "R", false, 0, "")
	doc.SetFont("Arial", "B", 11)
	doc.CellFormat(130, 7, "", "", 0, "L", false, 0, "")
	doc.CellFormat(30, 7, "Total:", "", 0, "R", false, 0, "")
	doc.CellFormat(30, 7, "S/ "+c.Total.StringFixed(2), "", 1, "R", false, 0, "")

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf output: %w", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
		return nil, fmt.Errorf("pdf: salida invalida")
	}
	return buf.Bytes(), nil
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}

func (g *Generator) consultaURL(c *domain.Comprobante, ruc string) string {
	if g.baseURL == "" {
		return ""
	}
	q := url.Values{}
	if ruc != "" {
		q.Set("ruc", ruc)
	}
	q.Set("serie", c.Serie)
	q.Set("numero", fmt.Sprintf("%d", c.Numero))
	return g.baseURL + "/consulta?" + q.Encode()
}
