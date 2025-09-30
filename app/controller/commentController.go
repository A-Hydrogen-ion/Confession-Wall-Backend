package controller

import (
	"net/http"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CommentController 控制器
// 需要在 router 注册时加 JWT 中间件
// 注册数据包

type CommentController struct {
	DB *gorm.DB
}

func NewCommentController(db *gorm.DB) *CommentController {
	return &CommentController{DB: db}
}

// 发布评论
func (ctrl *CommentController) AddComment(c *gin.Context) {
	var req model.Comment
	//简单的错误处理
	if err := c.ShouldBindJSON(&req); err != nil {
		ReturnMsg(c, 400, "你向服务器娘发送了一个奇怪的请求喵~")
		return
	}
	//拒绝没登录的用户发布评论
	userID, exists := c.Get("user_id")
	if !exists {
		ReturnMsg(c, 401, "你需要登录才能发表评论哦喵~")
		return
	}
	// 检查 confession_id 是否存在，将评论绑定到对应的表白
	if req.ConfessionID == 0 {
		ReturnMsg(c, 400, "缺少 confession_id，服务器娘不知道你要在哪条表白下评论啊喵~")
		return
	}

	//给评论加上用户ID
	req.UserID = userID.(uint)
	//保存评论
	if err := service.AddComment(ctrl.DB, &req); err != nil {
		ReturnError400(c, err)
		return
	}
	ReturnMsg(c, http.StatusOK, "评论发布成功了，对方收到你的心意了喵~")
}

// 删除评论？哇泼出去的水还想收回？做梦
func (ctrl *CommentController) DeleteComment(c *gin.Context) {
	// 从 query 获取 conmment_id
	commentID, err := QueryUint(c, "comment_id")
	if err != nil {
		ReturnError400(c, err)
		return
	}

	// 获取当前登录用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ReturnMsg(c, 401, "你需要登录才能删除评论喵~")
		return
	}

	// 查询评论，确认是自己发的评论才能删除
	var comment model.Comment
	if err := ctrl.DB.First(&comment, commentID).Error; err != nil {
		ReturnMsg(c, 404, "服务器娘没有查询到这个评论，可能已经被删除了喵~")
		return
	}

	if comment.UserID != userID.(uint) {
		ReturnMsg(c, 403, "不能删除别人的评论喵~，你这个大hentai！")
		return
	}

	// 调用 service 删除
	if err := service.DeleteComment(ctrl.DB, commentID); err != nil {
		ReturnMsg(c, 500, "服务器娘宕机了，删除评论失败了喵~")
		return
	}

	ReturnMsg(c, http.StatusOK, "评论已成功删除喵~")
}

// 嘴上说着不要，身体还是诚实的乖乖写了删除评论的controller了呢……

// 查看某条表白的评论
func (ctrl *CommentController) ListComments(c *gin.Context) {
	// 从 query 获取 confession_id
	confessionID, err := QueryUint(c, "confession_id")
	if err != nil {
		ReturnError400(c, err)
		return
	}
	// 调用 service 层获取评论列表
	comments, err := service.ListComments(ctrl.DB, confessionID)
	if err != nil {
		ReturnMsg(c, 500, "服务器娘宕机了，获取评论失败了喵~")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": comments,
		"msg":  "获取评论成功了喵~",
	})
}
