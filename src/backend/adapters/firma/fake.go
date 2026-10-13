package firma

import (
	"context"
)

type Fake struct{}

func (Fake) FirmarXML(_ context.Context, xml []byte) ([]byte, error) {
	return append([]byte("SIGNED:"), xml...), nil
}
