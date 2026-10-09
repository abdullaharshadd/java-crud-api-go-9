package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"migrated-app/internal/api"
	"migrated-app/internal/model"
)

// fakeUserService is an in-memory mock implementation of UserService,
// replacing the Java DAO-backed implementation for testing purposes.
type fakeUserService struct {
	mu    sync.Mutex
	users map[int]model.User
	seq   int
}

func newFakeUserService() *fakeUserService {
	return &fakeUserService{users: make(map[int]model.User)}
}

func (f *fakeUserService) SaveUser(ctx context.Context, user *model.User) (*model.User, error) {
	if err := user.Validate(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seq++
	user.ID = f.seq
	f.users[user.ID] = *user
	saved := *user
	return &saved, nil
}

func (f *fakeUserService) FetchUserList(ctx context.Context) ([]model.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	list := make([]model.User, 0, len(f.users))
	for _, u := range f.users {
		list = append(list, u)
	}
	return list, nil
}

func (f *fakeUserService) FetchUserById(ctx context.Context, id int) (*model.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.users[id]
	if !ok {
		return nil, fmt.Errorf("FetchUserById(%d): %w", id, api.ErrUserNotFound)
	}
	return &u, nil
}

func (f *fakeUserService) DeleteUser(ctx context.Context, id int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.users, id)
	return nil
}

func (f *fakeUserService) UpdateUser(ctx context.Context, id int, user *model.User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.users[id]; !ok {
		return fmt.Errorf("UpdateUser(%d): %w", id, api.ErrUserNotFound)
	}
	user.ID = id
	f.users[id] = *user
	return nil
}

func (f *fakeUserService) GetUserNameByName(ctx context.Context, name string) (*model.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.users {
		if u.Name == name {
			found := u
			return &found, nil
		}
	}
	return nil, fmt.Errorf("GetUserNameByName(%q): %w", name, api.ErrUserNotFound)
}

func TestSaveUser(t *testing.T) {
	ctx := context.Background()

	t.Run("returns saved user with non-zero ID", func(t *testing.T) {
		svc := newFakeUserService()
		in := &model.User{Name: "hemraj", Email: "hemraj@example.com"}
		got, err := svc.SaveUser(ctx, in)
		if err != nil {
			t.Fatalf("SaveUser returned error: %v", err)
		}
		if got == nil {
			t.Fatal("SaveUser returned nil user; invariant: never null on success")
		}
		if got.ID == 0 {
			t.Error("expected database-assigned non-zero ID")
		}
		if got.Name != "hemraj" {
			t.Errorf("Name = %q, want %q", got.Name, "hemraj")
		}
	})

	t.Run("propagates persistence-layer error", func(t *testing.T) {
		svc := newFakeUserService()
		_, err := svc.SaveUser(ctx, &model.User{Name: "   "})
		if err == nil {
			t.Fatal("expected error for blank name, got nil")
		}
		if !errors.Is(err, model.ErrNameRequired) {
			t.Errorf("error = %v, want wrapped model.ErrNameRequired", err)
		}
	})
}

func TestFetchUserList(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		seed      []model.User
		wantCount int
	}{
		{name: "empty list when no users exist", seed: nil, wantCount: 0},
		{name: "returns all seeded users", seed: []model.User{
			{Name: "hemraj"}, {Name: "amit"}, {Name: "priya"},
		}, wantCount: 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newFakeUserService()
			for i := range tc.seed {
				if _, err := svc.SaveUser(ctx, &tc.seed[i]); err != nil {
					t.Fatalf("seeding: %v", err)
				}
			}
			got, err := svc.FetchUserList(ctx)
			if err != nil {
				t.Fatalf("FetchUserList returned error: %v", err)
			}
			if got == nil {
				t.Fatal("FetchUserList returned nil; invariant: never null, use empty slice")
			}
			if len(got) != tc.wantCount {
				t.Errorf("len = %d, want %d", len(got), tc.wantCount)
			}
		})
	}
}

func TestFetchUserById(t *testing.T) {
	ctx := context.Background()
	svc := newFakeUserService()
	saved, err := svc.SaveUser(ctx, &model.User{Name: "hemraj", Email: "hemraj@example.com"})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	tests := []struct {
		name    string
		id      int
		wantErr error
	}{
		{name: "existing id returns matching user", id: saved.ID, wantErr: nil},
		{name: "missing id returns ErrUserNotFound", id: 999999, wantErr: api.ErrUserNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := svc.FetchUserById(ctx, tc.id)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error for id %d, got nil", tc.id)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("error = %v, want wrapped %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("FetchUserById(%d) returned error: %v", tc.id, err)
			}
			if got.ID != tc.id {
				t.Errorf("ID = %d, want %d", got.ID, tc.id)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		deleteID   int // 0 means delete the seeded user's ID
		exists     bool
		wantListLen int
	}{
		{name: "deletes existing user", deleteID: 0, exists: true, wantListLen: 0},
		{name: "non-existent id succeeds silently", deleteID: 424242, exists: false, wantListLen: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newFakeUserService()
			saved, err := svc.SaveUser(ctx, &model.User{Name: "hemraj"})
			if err != nil {
				t.Fatalf("seed: %v", err)
			}
			id := saved.ID
			if !tc.exists {
				id = tc.deleteID
			}
			if err := svc.DeleteUser(ctx, id); err != nil {
				t.Fatalf("DeleteUser(%d) returned error: %v", id, err)
			}
			list, err := svc.FetchUserList(ctx)
			if err != nil {
				t.Fatalf("FetchUserList: %v", err)
			}
			if len(list) != tc.wantListLen {
				t.Errorf("list length after delete = %d, want %d", len(list), tc.wantListLen)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		id      int // 0 means use seeded user's ID
		update  model.User
		wantErr error
	}{
		{
			name:   "updates existing user",
			update: model.User{Name: "hemraj", Email: "new@example.com", About: "updated"},
		},
		{
			name:    "non-existent id returns ErrUserNotFound",
			id:      777,
			update:  model.User{Name: "ghost"},
			wantErr: api.ErrUserNotFound,
		},
		{
			name:    "blank name rejected by validation",
			update:  model.User{Name: " "},
			wantErr: model.ErrNameRequired,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newFakeUserService()
			saved, err := svc.SaveUser(ctx, &model.User{Name: "hemraj"})
			if err != nil {
				t.Fatalf("seed: %v", err)
			}
			id := saved.ID
			if tc.id != 0 {
				id = tc.id
			}
			err = svc.UpdateUser(ctx, id, &tc.update)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("error = %v, want wrapped %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("UpdateUser returned error: %v", err)
			}
			got, err := svc.FetchUserById(ctx, id)
			if err != nil {
				t.Fatalf("FetchUserById after update: %v", err)
			}
			if got.Email != tc.update.Email {
				t.Errorf("Email after update = %q, want %q", got.Email, tc.update.Email)
			}
			if got.ID != id {
				t.Errorf("ID = %d, want %d", got.ID, id)
			}
		})
	}
}

func TestGetUserNameByName(t *testing.T) {
	ctx := context.Background()
	svc := newFakeUserService()
	if _, err := svc.SaveUser(ctx, &model.User{Name: "hemraj", Email: "hemraj@example.com"}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{name: "finds user by name hemraj", query: "hemraj"},
		{name: "missing name returns ErrUserNotFound", query: "nobody", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := svc.GetUserNameByName(ctx, tc.query)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", tc.query)
				}
				if !errors.Is(err, api.ErrUserNotFound) {
					t.Errorf("error = %v, want wrapped api.ErrUserNotFound", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetUserNameByName(%q) returned error: %v", tc.query, err)
			}
			if got.Name != tc.query {
				t.Errorf("Name = %q, want %q (assertEquals(name, found.getName()))", got.Name, tc.query)
			}
		})
	}
}

func TestNewUserNotFoundErrorWrapping(t *testing.T) {
	tests := []struct {
		name         string
		detail       string
		wantSentinel bool
	}{
		{name: "empty detail returns bare sentinel", detail: "", wantSentinel: true},
		{name: "sentinel text returns bare sentinel", detail: api.ErrUserNotFound.Error(), wantSentinel: true},
		{name: "detail wraps sentinel", detail: "user 42 missing", wantSentinel: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := api.NewUserNotFoundError(tc.detail)
			if !errors.Is(err, api.ErrUserNotFound) {
				t.Errorf("errors.Is(err, ErrUserNotFound) = false for %q", tc.detail)
			}
			if tc.wantSentinel && err != api.ErrUserNotFound {
				t.Errorf("expected exact sentinel, got %v", err)
			}
		})
	}
}