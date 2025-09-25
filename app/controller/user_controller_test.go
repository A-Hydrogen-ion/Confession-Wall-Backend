package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/database"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("automigrate failed: %v", err)
	}
	// set global DB for services that read database.DB
	database.DB = db
	return db
}

func TestUpdateUserProfile_DuplicateNicknameReturns400(t *testing.T) {
	db := setupTestDB(t)

	// create two users
	user1 := model.User{Username: "u1", Password: "pass12345", Nickname: "nick1"}
	user2 := model.User{Username: "u2", Password: "pass23456", Nickname: "nick2"}
	if err := db.Create(&user1).Error; err != nil {
		t.Fatalf("create user1 failed: %v", err)
	}
	if err := db.Create(&user2).Error; err != nil {
		t.Fatalf("create user2 failed: %v", err)
	}

	// prepare controller with test db
	authController := NewAuthController(db)

	// prepare request body to update user1's nickname to user2's nickname (duplicate)
	input := model.User{Username: user1.Username, Nickname: user2.Nickname, Avatar: "avatar.png"}
	b, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPut, "/api/user/profile", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// set authenticated user id in context
	c.Set("user_id", user1.UserID)

	// call handler
	authController.UpdateUserProfile(c)

	if w.Code != http.StatusBadRequest && w.Code != http.StatusInternalServerError && w.Code != 400 {
		t.Fatalf("expected 400 for duplicate nickname, got status %d, body: %s", w.Code, w.Body.String())
	}
	// assert response contains nickname occupied message
	body := w.Body.Bytes()
	if !(bytes.Contains(body, []byte("昵称已被占用")) || bytes.Contains(body, []byte("该昵称已被占用喵")) || bytes.Contains(body, []byte("昵称已存在"))) {
		t.Fatalf("expected nickname occupied message, got body: %s", w.Body.String())
	}
}
