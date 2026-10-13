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
