package sunat_test

import (
	"context"
	"testing"

	"github.com/Bluethem/sw707u-grupo4/src/backend/adapters/sunat"
)

func TestFakeSUNAT_Accept(t *testing.T) {
	res, err := sunat.NewFake(sunat.ModeAccept).Enviar(context.Background(), "c-1", []byte("<xml/>"))
	if err != nil {
		t.Fatalf("Enviar: %v", err)
	}
	if !res.Aceptado {
		t.Fatal("Aceptado = false; want true")
	}
	if string(res.CDR) != "CDR-OK:c-1" {
		t.Fatalf("CDR = %q", res.CDR)
	}
}

func TestFakeSUNAT_Reject(t *testing.T) {
	res, err := sunat.NewFake(sunat.ModeReject).Enviar(context.Background(), "c-1", nil)
	if err != nil {
		t.Fatalf("Enviar: %v", err)
	}
	if res.Aceptado {
		t.Fatal("Aceptado = true; want false")
	}
}

func TestFakeSUNAT_Error(t *testing.T) {
	_, err := sunat.NewFake(sunat.ModeError).Enviar(context.Background(), "c-1", nil)
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestFakeSUNAT_BajaAccept(t *testing.T) {
	res, err := sunat.NewFake(sunat.ModeAccept).EnviarBaja(context.Background(), "c-1", "error de emision")
	if err != nil {
		t.Fatalf("EnviarBaja: %v", err)
	}
	if !res.Aceptado || string(res.CDR) != "CDR-BAJA:c-1" {
		t.Fatalf("resultado baja: %+v", res)
	}
}

func TestFakeSUNAT_BajaReject(t *testing.T) {
	res, err := sunat.NewFake(sunat.ModeReject).EnviarBaja(context.Background(), "c-1", "motivo")
	if err != nil || res.Aceptado {
		t.Fatalf("EnviarBaja reject: err=%v res=%+v", err, res)
	}
}

func TestFakeSUNAT_ResumenAccept(t *testing.T) {
	res, err := sunat.NewFake(sunat.ModeAccept).EnviarResumen(context.Background(), "r-1", []byte("{}"))
	if err != nil {
		t.Fatalf("EnviarResumen: %v", err)
	}
	if !res.Aceptado || string(res.CDR) != "CDR-RESUMEN:r-1" {
		t.Fatalf("resultado resumen: %+v", res)
	}
}

func TestFakeSUNAT_ResumenError(t *testing.T) {
	_, err := sunat.NewFake(sunat.ModeError).EnviarResumen(context.Background(), "r-1", nil)
	if err == nil {
		t.Fatal("expected transport error on resumen")
	}
}
