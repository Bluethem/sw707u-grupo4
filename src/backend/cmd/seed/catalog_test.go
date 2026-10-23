package main

import (
	"testing"

	factdomain "github.com/Bluethem/sw707u-grupo4/src/backend/internal/facturacion/domain"
)

func TestRucDesdeBase_Valido(t *testing.T) {
	for _, base := range []string{"2012345678", "2051234567", "2010007097"} {
		ruc := rucDesdeBase(base)
		if err := factdomain.ValidarRUC(ruc); err != nil {
			t.Fatalf("RUC generado invalido %s: %v", ruc, err)
		}
	}
	if got := rucDesdeBase("2010007097"); got != "20100070970" {
		t.Fatalf("check digit empresa demo: got %s", got)
	}
}

func TestConstruirDemoComprobantes_CubreReportesYResumenes(t *testing.T) {
	docs, resumenDias := construirDemoComprobantes(1, 1)
	if len(docs) < 40 {
		t.Fatalf("esperaba al menos 40 comprobantes, got %d", len(docs))
	}

	hoyBoletas, hoyFacturas, hoyAnulados := 0, 0, 0
	boletasPorDia := map[int]int{}
	for _, d := range docs {
		if d.diasAtras == 0 {
			if d.estado == "ANULADA" || d.estado == "PENDIENTE_BAJA" {
				hoyAnulados++
			}
			if d.tipo == "boleta" {
				hoyBoletas++
			} else {
				hoyFacturas++
			}
		}
		if d.tipo == "boleta" && d.estado != "ANULADA" {
			boletasPorDia[d.diasAtras]++
		}
	}
	if hoyBoletas < 4 || hoyFacturas < 2 || hoyAnulados < 1 {
		t.Fatalf("hoy insuficiente para reportes: boletas=%d facturas=%d anulados=%d", hoyBoletas, hoyFacturas, hoyAnulados)
	}
	if boletasPorDia[0] < 2 || boletasPorDia[1] < 2 {
		t.Fatalf("hoy/ayer deben tener boletas libres para generar resumen")
	}
	for _, dia := range resumenDias {
		if boletasPorDia[dia] == 0 {
			t.Fatalf("dia -%d no tiene boletas para resumen aceptado", dia)
		}
	}
}
