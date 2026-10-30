package shops_test

import (
	"context"
	"errors"
	"testing"

	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/platform/apperr"
)

type fakeLoginOTPRepository struct {
	channel shops.LoginOTPChannel
	found   bool
	err     error
	updates []shops.LoginOTPChannel
}

func (f *fakeLoginOTPRepository) Get(context.Context, string) (shops.LoginOTPChannel, bool, error) {
	return f.channel, f.found, f.err
}

func (f *fakeLoginOTPRepository) Update(_ context.Context, _ string, channel shops.LoginOTPChannel) (shops.LoginOTPChannel, bool, error) {
	f.updates = append(f.updates, channel)
	return channel, f.found, f.err
}

func TestLoginOTPChannel_Valid_OnlyTheThreeChannels(t *testing.T) {
	for _, ok := range []shops.LoginOTPChannel{"email", "whatsapp", "both"} {
		if !ok.Valid() {
			t.Errorf("%q debe ser válido", ok)
		}
	}
	for _, bad := range []shops.LoginOTPChannel{"", "sms", "Email", "WHATSAPP", "email,whatsapp", " both"} {
		if bad.Valid() {
			t.Errorf("%q no debe ser válido", bad)
		}
	}
}

func TestLoginOTPService_Update_InvalidChannel_IsValidationAndNeverWrites(t *testing.T) {
	repo := &fakeLoginOTPRepository{found: true}
	_, err := shops.NewLoginOTPService(repo).Update(context.Background(), "shop", "sms")

	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindValidation {
		t.Fatalf("esperaba un error de validación, obtuvo %v", err)
	}
	if len(repo.updates) != 0 {
		t.Fatal("un canal inválido no debe llegar al repositorio")
	}
}

func TestLoginOTPService_Update_Stores(t *testing.T) {
	repo := &fakeLoginOTPRepository{found: true}
	got, err := shops.NewLoginOTPService(repo).Update(context.Background(), "shop", shops.LoginOTPChannelBoth)
	if err != nil || got != shops.LoginOTPChannelBoth || len(repo.updates) != 1 {
		t.Fatalf("got=%q err=%v updates=%v", got, err, repo.updates)
	}
}

func TestLoginOTPService_NotFoundAndRepositoryErrors(t *testing.T) {
	svc := shops.NewLoginOTPService(&fakeLoginOTPRepository{found: false})
	if _, err := svc.Get(context.Background(), "shop"); err == nil {
		t.Fatal("una barbería no visible debe responder not found")
	} else if e, _ := apperr.As(err); e == nil || e.Kind != apperr.KindNotFound {
		t.Fatalf("esperaba not found, obtuvo %v", err)
	}
	if _, err := svc.Update(context.Background(), "shop", shops.LoginOTPChannelEmail); err == nil {
		t.Fatal("actualizar una barbería no visible debe responder not found")
	}

	boom := shops.NewLoginOTPService(&fakeLoginOTPRepository{err: errors.New("db caída")})
	if _, err := boom.Get(context.Background(), "shop"); err == nil {
		t.Fatal("un fallo de base de datos debe propagarse")
	} else if e, _ := apperr.As(err); e == nil || e.Kind != apperr.KindInternal {
		t.Fatalf("esperaba error interno, obtuvo %v", err)
	}
}

func TestLoginOTPService_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	svc := shops.NewLoginOTPService(&fakeLoginOTPRepository{found: true})
	if _, err := svc.Get(ctx, "shop"); err == nil {
		t.Fatal("un contexto cancelado debe fallar antes de leer")
	}
	if _, err := svc.Update(ctx, "shop", shops.LoginOTPChannelEmail); err == nil {
		t.Fatal("un contexto cancelado debe fallar antes de escribir")
	}
}
