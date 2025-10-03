package controller

import (
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/service"
	"github.com/gin-gonic/gin"
)

// 全局使用common中的respondJSON函数
// 点赞接口
func LikeConfession(c *gin.Context) {
	confessionID, err := QueryUint(c, "confession_id") // 获取表白ID
	if err != nil {
		respondJSON(c, 400, "不知道这是谁的表白喵~", nil)
		return
	}
	userID := c.GetUint("user_id") // 从中间件获取当前用户ID
	if userID == 0 {
		respondJSON(c, 401, "你需要登陆才能点赞哦喵~", nil)
		return
	}
	// 调用service的点赞函数
	err = service.LikeConfession(confessionID, userID)
	if err != nil {
		respondJSON(c, 500, "点赞失败了喵~", nil)
		return
	}
	respondJSON(c, 200, "点赞成功了喵~", nil)
}

// 取消点赞接口
func UnlikeConfession(c *gin.Context) {
	confessionID, err := QueryUint(c, "confession_id")
	if err != nil {
		respondJSON(c, 400, "不知道这是谁的表白喵~", nil)
		return
	}
	userID := c.GetUint("user_id")
	// 调用service的取消点赞函数
	err = service.UnlikeConfession(confessionID, userID)
	if err != nil {
		respondJSON(c, 500, "取消点赞失败了喵~", nil)
		return
	}
	respondJSON(c, 200, "取消点赞成功了喵~", nil)
}
