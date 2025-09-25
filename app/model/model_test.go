package model

import (
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestIsBcryptHash(t *testing.T) {
	// random string
	if isBcryptHash("plainpassword") {
		t.Fatal("plain string should not be recognized as bcrypt hash")
	}

	// generate bcrypt hash
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt generate failed: %v", err)
	}
	if !isBcryptHash(string(hash)) {
		t.Fatal("generated bcrypt hash should be recognized")
	}
}

func TestBeforeSaveAndCheckPassword(t *testing.T) {
	u := &User{
		Username: "u1",
		Password: "myPassword123",
	}
	// call BeforeSave to hash password
	if err := u.BeforeSave(nil); err != nil {
		t.Fatalf("BeforeSave failed: %v", err)
	}
	// password should be hashed and not equal to plain
	if u.Password == "myPassword123" {
		t.Fatal("password should be hashed after BeforeSave")
	}
	// check password
	if err := u.CheckPassword("myPassword123"); err != nil {
		t.Fatalf("CheckPassword failed: %v", err)
	}
}

func TestBeforeCreateSetsDefaults(t *testing.T) {
	u := &User{
		Username: "user2",
	}
	// Nickname empty should be set to username
	if err := u.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate failed: %v", err)
	}
	if u.Nickname != "user2" {
		t.Fatalf("expected nickname to be set to username, got %s", u.Nickname)
	}
	// createdAt and UpdateAt should be set
	if u.CreatedAt.IsZero() || u.UpdateAt.IsZero() {
		t.Fatalf("timestamps should be set")
	}
	// When createdAt is already set, BeforeCreate should not overwrite
	past := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	u2 := &User{Username: "user3", CreatedAt: past}
	if err := u2.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate failed: %v", err)
	}
	if !u2.CreatedAt.Equal(past) {
		t.Fatalf("CreatedAt should not be overwritten")
	}
}
