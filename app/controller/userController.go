package controller

import (
	"net/http"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/config/database"
	"github.com/gin-gonic/gin"
)

func (authController *AuthController) GetMyProfile(c *gin.Context) {
	// 获取从中间件设置的user_id
	userID, _ := c.Get("user_id")
	var profile model.User
	result := database.DB.First(&profile, userID)
	c.JSON(http.StatusInternalServerError, gin.H{"result": result})
}

func (authController *AuthController) UpdateUserProfile(c *gin.Context) {

}
