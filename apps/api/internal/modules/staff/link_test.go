package staff_test

import (
	"context"
	"errors"
	"testing"

	"system-barbershop/internal/modules/staff"
	"system-barbershop/internal/platform/apperr"
)

const (
	linkShop   = "11111111-1111-1111-1111-111111111111"
	linkUserID = "aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2"
	linkBarber = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"
)

type recordingObserver struct {
	released []string
}

func (o *recordingObserver) BarberLinkReleased(_ context.Context, _ string, barberID string) {
	o.released = append(o.released, barberID)
}

func requireKind(t *testing.T, err error, want apperr.Kind) {
	t.Helper()
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != want {
		t.Fatalf("expected apperr kind %q, got %v", want, err)
	}
}

func TestLinkedBarber_WithoutLink_IsNotFound(t *testing.T) {
	repo := &fakeRepository{getLinkedFn: func(context.Context, string, string) (staff.Barber, bool, error) {
		return staff.Barber{}, false, nil
	}}
	_, err := staff.NewService(repo).LinkedBarber(context.Background(), linkShop, linkUserID)
	requireKind(t, err, apperr.KindNotFound)
}

func TestLinkedBarber_ReturnsTheUsersBarber(t *testing.T) {
	repo := &fakeRepository{getLinkedFn: func(_ context.Context, shop, user string) (staff.Barber, bool, error) {
		if shop != linkShop || user != linkUserID {
			t.Fatalf("la consulta debe usar el principal, recibió %q/%q", shop, user)
		}
		return staff.Barber{ID: linkBarber, FullName: "Carlos"}, true, nil
	}}
	barber, err := staff.NewService(repo).LinkedBarber(context.Background(), linkShop, linkUserID)
	if err != nil || barber.ID != linkBarber {
		t.Fatalf("expected the linked barber, got %+v, %v", barber, err)
	}
}

func TestLinkMyBarber_InvalidID_NeverReachesRepository(t *testing.T) {
	repo := &fakeRepository{linkFn: func(context.Context, string, string, string) (staff.LinkResult, error) {
		t.Fatal("un identificador inválido no debe llegar al repositorio")
		return staff.LinkResult{}, nil
	}}
	_, err := staff.NewService(repo).LinkMyBarber(context.Background(), linkShop, linkUserID, "no-es-uuid")
	requireKind(t, err, apperr.KindNotFound)
}

func TestLinkMyBarber_UnknownOrForeignBarber_IsNotFound(t *testing.T) {
	repo := &fakeRepository{linkFn: func(context.Context, string, string, string) (staff.LinkResult, error) {
		return staff.LinkResult{Found: false}, nil
	}}
	_, err := staff.NewService(repo).LinkMyBarber(context.Background(), linkShop, linkUserID, linkBarber)
	requireKind(t, err, apperr.KindNotFound)
}

func TestLinkMyBarber_TakenByAnotherUser_IsConflictAndNotifiesNobody(t *testing.T) {
	repo := &fakeRepository{linkFn: func(context.Context, string, string, string) (staff.LinkResult, error) {
		return staff.LinkResult{Found: true, Taken: true, ReleasedBarberID: "no-debe-avisar"}, nil
	}}
	service := staff.NewService(repo)
	observer := &recordingObserver{}
	service.ObserveLinkReleased(observer)

	_, err := service.LinkMyBarber(context.Background(), linkShop, linkUserID, linkBarber)
	requireKind(t, err, apperr.KindConflict)
	if len(observer.released) != 0 {
		t.Fatalf("un conflicto no libera nada, pero avisó %v", observer.released)
	}
}

func TestLinkMyBarber_Success_NotifiesPreviousBarberRelease(t *testing.T) {
	repo := &fakeRepository{linkFn: func(_ context.Context, shop, user, barber string) (staff.LinkResult, error) {
		if shop != linkShop || user != linkUserID || barber != linkBarber {
			t.Fatalf("argumentos inesperados %q/%q/%q", shop, user, barber)
		}
		return staff.LinkResult{
			Found:            true,
			Barber:           staff.Barber{ID: linkBarber, FullName: "Carlos"},
			ReleasedBarberID: "previous-barber",
		}, nil
	}}
	service := staff.NewService(repo)
	observer := &recordingObserver{}
	service.ObserveLinkReleased(observer)

	barber, err := service.LinkMyBarber(context.Background(), linkShop, linkUserID, linkBarber)
	if err != nil || barber.ID != linkBarber {
		t.Fatalf("expected the new barber, got %+v, %v", barber, err)
	}
	if len(observer.released) != 1 || observer.released[0] != "previous-barber" {
		t.Fatalf("expected one release notice for the previous barber, got %v", observer.released)
	}
}

func TestLinkMyBarber_SameBarberAgain_IsIdempotentWithoutNotice(t *testing.T) {
	repo := &fakeRepository{linkFn: func(context.Context, string, string, string) (staff.LinkResult, error) {
		return staff.LinkResult{Found: true, Barber: staff.Barber{ID: linkBarber}}, nil
	}}
	service := staff.NewService(repo)
	observer := &recordingObserver{}
	service.ObserveLinkReleased(observer)

	if _, err := service.LinkMyBarber(context.Background(), linkShop, linkUserID, linkBarber); err != nil {
		t.Fatalf("repetir la selección debe ser un éxito: %v", err)
	}
	if len(observer.released) != 0 {
		t.Fatalf("repetir la selección no libera nada, avisó %v", observer.released)
	}
}

func TestUnlinkMyBarber_NotifiesOnlyWhenSomethingWasReleased(t *testing.T) {
	for name, released := range map[string]string{"con vínculo": linkBarber, "sin vínculo": ""} {
		t.Run(name, func(t *testing.T) {
			repo := &fakeRepository{unlinkFn: func(context.Context, string, string) (string, error) { return released, nil }}
			service := staff.NewService(repo)
			observer := &recordingObserver{}
			service.ObserveLinkReleased(observer)

			if err := service.UnlinkMyBarber(context.Background(), linkShop, linkUserID); err != nil {
				t.Fatalf("quitar el vínculo es idempotente: %v", err)
			}
			wantNotices := 0
			if released != "" {
				wantNotices = 1
			}
			if len(observer.released) != wantNotices {
				t.Fatalf("expected %d notices, got %v", wantNotices, observer.released)
			}
		})
	}
}

func TestLinkOperations_RepositoryFailure_IsInternal(t *testing.T) {
	boom := errors.New("boom")
	repo := &fakeRepository{
		getLinkedFn: func(context.Context, string, string) (staff.Barber, bool, error) { return staff.Barber{}, false, boom },
		linkFn: func(context.Context, string, string, string) (staff.LinkResult, error) {
			return staff.LinkResult{}, boom
		},
		unlinkFn: func(context.Context, string, string) (string, error) { return "", boom },
	}
	service := staff.NewService(repo)

	_, err := service.LinkedBarber(context.Background(), linkShop, linkUserID)
	requireKind(t, err, apperr.KindInternal)
	_, err = service.LinkMyBarber(context.Background(), linkShop, linkUserID, linkBarber)
	requireKind(t, err, apperr.KindInternal)
	requireKind(t, service.UnlinkMyBarber(context.Background(), linkShop, linkUserID), apperr.KindInternal)
}

func TestUserBarberLookup_MapsNoLinkToNotFoundFlag(t *testing.T) {
	repo := &fakeRepository{getLinkedFn: func(_ context.Context, _, user string) (staff.Barber, bool, error) {
		if user == linkUserID {
			return staff.Barber{ID: linkBarber}, true, nil
		}
		return staff.Barber{}, false, nil
	}}
	lookup := staff.NewUserBarberLookup(staff.NewService(repo))

	id, found, err := lookup.BarberIDOfUser(context.Background(), linkShop, linkUserID)
	if err != nil || !found || id != linkBarber {
		t.Fatalf("expected the linked barber, got %q %v %v", id, found, err)
	}
	id, found, err = lookup.BarberIDOfUser(context.Background(), linkShop, "otro-usuario")
	if err != nil || found || id != "" {
		t.Fatalf("un usuario sin barbero no es un error, got %q %v %v", id, found, err)
	}
}
