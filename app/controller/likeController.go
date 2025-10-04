package controller

import (
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/service"
	"github.com/gin-gonic/gin"
)

// 全局使用common中的respondJSON函数
// 点赞/取消点赞接口（POST请求自动切换状态）
func LikeConfession(c *gin.Context) {
	confessionID, err := QueryUint(c, "confession_id")
	if err != nil {
		respondJSON(c, 400, "不知道这是谁的表白喵~", nil)
		return
	}
	userID := c.GetUint("user_id") // 阻止没有登录的用户操作点赞
	if userID == 0 {
		respondJSON(c, 401, "你需要登陆才能操作点赞哦喵~", nil)
		return
	}
	liked, err := service.HasLiked(confessionID, userID)
	if err != nil {
		respondJSON(c, 500, "查询点赞状态失败了喵~", nil)
		return
	}
	if liked {
		// 已点赞则取消点赞
		err = service.UnlikeConfession(confessionID, userID)
		if err != nil {
			respondJSON(c, 500, "取消点赞失败了喵~", nil)
			return
		}
		respondJSON(c, 200, "已取消点赞喵~", gin.H{"liked": false})
		return
	} else {
		// 未点赞则点赞
		err = service.LikeConfession(confessionID, userID)
		if err != nil {
			respondJSON(c, 500, "点赞失败了喵~", nil)
			return
		}
		respondJSON(c, 200, "点赞成功了喵~", gin.H{"liked": true})
		return
	}
}

// 已移除单独取消点赞接口
