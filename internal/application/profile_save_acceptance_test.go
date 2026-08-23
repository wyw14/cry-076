package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/wyw14/cry-076/internal/application"
	"github.com/wyw14/cry-076/internal/domain"
)

type versionedProfiles struct{ current domain.Profile }

func (r *versionedProfiles) Get(context.Context, string) (domain.Profile, error) {
	return r.current, nil
}

func (r *versionedProfiles) Save(_ context.Context, item domain.Profile, expected int64) (domain.Profile, error) {
	if r.current.Version != expected {
		return domain.Profile{}, domain.ErrConflict
	}
	item.Version = expected + 1
	r.current = item
	return item, nil
}

func TestProfileSaveReturnsCommittedVersionForNextWrite(t *testing.T) {
	repository := &versionedProfiles{current: domain.Profile{OwnerID: "u1", Version: 7}}
	service := application.NewProfileService(repository, &fakeAudits{}, fixedClock{}, &ids{})
	actor := domain.Actor{ID: "u1", Role: domain.RoleOwner}

	first, err := service.Save(context.Background(), actor, domain.Profile{OwnerID: "u1", Contact: domain.ContactProfile{FullName: "林未"}}, 7, "request-1")
	if err != nil {
		t.Fatal(err)
	}
	if first.Version != 8 {
		t.Fatalf("first response version=%d, want committed version 8", first.Version)
	}
	first.Contact.Summary = "第二次保存"
	second, err := service.Save(context.Background(), actor, first, first.Version, "request-2")
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			t.Fatalf("response supplied a stale version for the next save: %v", err)
		}
		t.Fatal(err)
	}
	if second.Version != 9 {
		t.Fatalf("second response version=%d, want 9", second.Version)
	}
}

