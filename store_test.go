package floki_test

import (
	"testing"

	"github.com/rodolfo-mora/floki" // Adjust the import path as needed
)

func TestNewMemoryStore(t *testing.T) {
	store := floki.NewMemoryStore()
	if store == nil {
		t.Error("Expected non-nil MemoryStore")
	}
	if len(store.Users) != 0 {
		t.Errorf("Expected empty Users map, got %d", len(store.Users))
	}
}

func TestUserExists(t *testing.T) {
	store := floki.NewMemoryStore()
	store.Users["testuser"] = floki.User{Email: "testuser"}

	exists := store.UserExists("testuser")
	if !exists {
		t.Error("Expected user to exist")
	}

	exists = store.UserExists("nonexistentuser")
	if exists {
		t.Error("Expected user not to exist")
	}
}

func TestSave(t *testing.T) {
	store := floki.NewMemoryStore()

	err := store.Save("testuser", []string{"group1", "group2"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	user, exists := store.Users["testuser"]
	if !exists {
		t.Error("Expected user to be saved")
	}

	if user.Email != "testuser" || len(user.SSOGroups) != 2 {
		t.Error("User data was not saved correctly")
	}
}

func TestGetSSOGroups(t *testing.T) {
	store := floki.NewMemoryStore()
	store.Save("testuser", []string{"group1", "group2"})

	groups := store.GetSSOGroups("testuser")
	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}

	groups = store.GetSSOGroups("nonexistentuser")
	if len(groups) != 0 {
		t.Error("Expected no groups for non-existent user")
	}
}
