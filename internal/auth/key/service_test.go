package key

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ulibaysya/caseit/internal/user"
)

type createServiceTest struct {
	name    string
	err     error
	key     string
	storage mockStorage
	keygen  mockKeygen
}

func (test createServiceTest) run(t *testing.T) {
	key, err := NewService(discardLogger, test.storage, test.keygen).Create(t.Context(), 71289)

	if !errors.Is(err, test.err) {
		t.Fatalf("unexpected error %v; expected %v", err, test.err)
	}

	if key != test.key {
		t.Fatalf("unexpected key %v; expected %v", key, test.key)
	}
}

func TestCreate(t *testing.T) {
	someErr := fmt.Errorf("some error")

	tests := []createServiceTest{
		{
			name: "SavingError",
			err:  someErr,
			storage: mockStorage{
				err: someErr,
			},
		},
		{
			name: "Success",
			key:  "someRandomKey",
			keygen: mockKeygen{
				key: "someRandomKey",
			},
		},
	}

	for _, i := range tests {
		if success := t.Run(i.name, i.run); !success {
			t.FailNow()
		}
	}
}

type mockStorage struct {
	err error
}

func (m mockStorage) Save(ctx context.Context, authKey string, userID user.ID) error {
	return m.err
}

type mockKeygen struct {
	key string
}

func (m mockKeygen) Generate() string {
	return m.key
}
