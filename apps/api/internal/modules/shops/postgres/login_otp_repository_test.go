// Pruebas de integración de shops.LoginOTPRepository (DEC-139) contra PostgreSQL
// REAL con el par shopA/shopB: el canal es por barbería, el CHECK vive en la base
// y RLS impide que una sesión toque la fila de la otra.
package postgres_test

import (
	"context"
	"testing"

	"system-barbershop/internal/modules/shops"
	shopspostgres "system-barbershop/internal/modules/shops/postgres"
	"system-barbershop/internal/platform/database"
)

// restoreLoginChannel devuelve el canal a WhatsApp, que es como
// testdata/hu007_reto_telefonico.sql deja las barberías de las pruebas.
func restoreLoginChannel(t *testing.T, shopID string) {
	t.Helper()
	t.Cleanup(func() {
		db := setupTestDB(t)
		defer db.Close()
		_, _, err := shopspostgres.NewLoginOTPRepository(db).Update(context.Background(), shopID, shops.LoginOTPChannelWhatsApp)
		if err != nil {
			t.Fatalf("restaurar canal: %v", err)
		}
	})
}

func TestLoginOTPRepository_UpdateAndGet_RoundTripPerChannel(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	restoreLoginChannel(t, shopA)
	repo := shopspostgres.NewLoginOTPRepository(db)

	for _, channel := range []shops.LoginOTPChannel{shops.LoginOTPChannelEmail, shops.LoginOTPChannelBoth, shops.LoginOTPChannelWhatsApp} {
		stored, found, err := repo.Update(context.Background(), shopA, channel)
		if err != nil || !found || stored != channel {
			t.Fatalf("Update(%s): stored=%q found=%v err=%v", channel, stored, found, err)
		}
		got, found, err := repo.Get(context.Background(), shopA)
		if err != nil || !found || got != channel {
			t.Fatalf("Get tras Update(%s): got=%q found=%v err=%v", channel, got, found, err)
		}
	}
}

// RN-TEN-01: cambiar el canal de A no cambia el de B, y una sesión de A no puede
// escribir la fila de B aunque conozca su id.
func TestLoginOTPRepository_IsolatedBetweenTenants(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	restoreLoginChannel(t, shopA)
	restoreLoginChannel(t, shopB)
	repo := shopspostgres.NewLoginOTPRepository(db)

	if _, _, err := repo.Update(context.Background(), shopB, shops.LoginOTPChannelWhatsApp); err != nil {
		t.Fatalf("preparar B: %v", err)
	}
	if _, _, err := repo.Update(context.Background(), shopA, shops.LoginOTPChannelEmail); err != nil {
		t.Fatalf("Update A: %v", err)
	}
	gotB, _, err := repo.Get(context.Background(), shopB)
	if err != nil || gotB != shops.LoginOTPChannelWhatsApp {
		t.Fatalf("B no debe cambiar al cambiar A: got=%q err=%v", gotB, err)
	}

	// Contexto de A escribiendo el id de B: RLS no deja ver ni actualizar la fila.
	var affected int64
	err = db.InTenantTx(context.Background(), database.BarbershopID(shopA), func(ctx context.Context, q database.Queries) error {
		tag, err := q.Exec(ctx, `UPDATE barbershop SET login_otp_channel = 'both' WHERE id = $1`, shopB)
		affected = tag.RowsAffected()
		return err
	})
	if err != nil {
		t.Fatalf("UPDATE cruzado: %v", err)
	}
	if affected != 0 {
		t.Fatalf("una sesión de A modificó %d fila(s) de B", affected)
	}
	gotB, _, _ = repo.Get(context.Background(), shopB)
	if gotB != shops.LoginOTPChannelWhatsApp {
		t.Fatalf("B cambió por una escritura cruzada: %q", gotB)
	}
}

func TestLoginOTPRepository_DatabaseRejectsAnUnknownChannel(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	restoreLoginChannel(t, shopA)

	_, _, err := shopspostgres.NewLoginOTPRepository(db).Update(context.Background(), shopA, "sms")
	if err == nil {
		t.Fatal("el CHECK de la base debe rechazar un canal fuera de la lista aunque el servicio falle en validar")
	}
}

func TestLoginOTPRepository_NotVisibleBarbershop_IsNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := shopspostgres.NewLoginOTPRepository(db)

	const unknown = "99999999-9999-4999-8999-999999999998"
	if _, found, err := repo.Get(context.Background(), unknown); err != nil || found {
		t.Fatalf("Get: found=%v err=%v", found, err)
	}
	if _, found, err := repo.Update(context.Background(), unknown, shops.LoginOTPChannelEmail); err != nil || found {
		t.Fatalf("Update: found=%v err=%v", found, err)
	}
}
