package controller

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ConfessionController 控制器
// 需要在 router 注册时加 JWT 中间件
// 注册数据包
type ConfessionController struct {
	DB *gorm.DB
}

func CreateConfessionController(db *gorm.DB) *ConfessionController {
	return &ConfessionController{DB: db}
}

// QueryUint 从请求 query 中获取参数并转换为 uint
func QueryUint(c *gin.Context, key string) (uint, error) {
	valStr := c.Query(key)
	if valStr == "" {
		return 0, errors.New(key + " 参数为空")
	}
	valUint64, err := strconv.ParseUint(valStr, 10, 64)
	if err != nil {
		return 0, errors.New(key + " 参数无效")
	}
	return uint(valUint64), nil
}

// 发布表白
func (ctrl *ConfessionController) CreateConfession(c *gin.Context) {
	content := c.PostForm("content")
	anonymous := c.PostForm("anonymous") == "true" //发布的表白是否匿名
	private := c.PostForm("private") == "true"     //发布的表白是否私密
	//判断用户有没有登录
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "只有登录的孩子才能发布表白喵~"})
		return
	}

	// 调用封装的图片上传函数
	imagePaths, err := service.UploadImages(c, userID.(uint))
	//终于简洁的多了（狂喜
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 保存表白与发布人的ID和表白属性
	confession := model.Confession{
		UserID:    userID.(uint),
		Content:   content,
		Images:    imagePaths,
		Anonymous: anonymous,
		Private:   private,
	}
	if err := service.CreateConfession(ctrl.DB, &confession); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "发布成功了喵~"})
}

// 修改表白
func (ctrl *ConfessionController) UpdateConfession(c *gin.Context) {
	confessionIDStr := c.PostForm("confession_id")
	confessionID, err := strconv.ParseUint(confessionIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "服务器娘没有查询到这个表白，可能已经被删除了喵~"}) //错误处理
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "只有登录的孩子才能修改表白喵~"}) //拒绝没有登陆的用户修改
		return
	}
	var confession model.Confession
	if err := ctrl.DB.First(&confession, confessionID).Error; err != nil {
		c.JSON(404, gin.H{"error": "服务器娘没有查询到这个表白，可能已经被删除了喵~"}) //错误处理
		return
	}
	if confession.UserID != userID.(uint) {
		c.JSON(403, gin.H{"error": "你居然想修改别人的表白，hentai！"}) //不准修改别人的表白
		return
	}
	newContent := c.PostForm("content")
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(400, gin.H{"error": "图片因为某些原因上传失败了喵"})
		return
	}
	files := form.File["images"]
	if len(files) > 9 {
		c.JSON(400, gin.H{"error": "9张图片已经能让对方感受到你的心意了，不要再上传了喵~"})
		return
	}

	// MD，再写我要疯了，直接给原有图片文件全部删了重新上传，这样下面我就可以直接从创建表白那里复制过来了
	for _, oldPath := range confession.Images {
		_ = os.Remove(oldPath)
	}
	//对重新传回来的图片进行格式和大小检查
	// 调用封装的图片上传函数
	imagePaths, err := service.UploadImages(c, userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) //imagePath会返回更多的错误信息，这里只做最简单的错误处理
		return
	}

	confession.Images = imagePaths
	confession.Content = newContent
	if err := ctrl.DB.Save(&confession).Error; err != nil {
		c.JSON(500, gin.H{"error": "修改失败"})
		return
	}
	c.JSON(200, gin.H{"msg": "修改成功"})
}

// 查看社区表白
func (ctrl *ConfessionController) ListPublicConfessions(c *gin.Context) {
	var uid uint = 0
	// 获取当前用户ID，若未登录则为0，这样做可以让未登录用户也能看到公共表白
	if userID, exists := c.Get("user_id"); exists {
		// 调试输出
		fmt.Printf("[ListPublicConfessions] exists=true, userID=%d\n", userID)
		uid = userID.(uint)
		fmt.Printf("[ListPublicConfessions] uid=%d\n", uid)
	}
	confessions, err := service.ListPublicConfessions(ctrl.DB, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败了喵"})
		return
	}
	// 匿名处理
	for i := range confessions {
		if confessions[i].Anonymous {
			confessions[i].UserID = 0
			// 可选：清空昵称、头像等，但是我根本不想写，让前段去处理去（）
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": confessions})
}

// 发布评论
func (ctrl *ConfessionController) AddComment(c *gin.Context) {
	var req model.Comment
	//简单的错误处理
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "你向服务器娘发送了一个奇怪的请求喵~"})
		return
	}
	//拒绝没登录的用户发布评论
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "你需要登录才能发表评论哦喵~"})
		return
	}
	// 检查 confession_id 是否存在，将评论绑定到对应的表白
	if req.ConfessionID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 confession_id，服务器娘不知道你要在哪条表白下评论啊喵"})
		return
	}

	//给评论加上用户ID
	req.UserID = userID.(uint)
	//保存评论
	if err := service.AddComment(ctrl.DB, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "评论发布成功了，对方收到你的心意了喵~"})
}

// 删除评论？哇泼出去的水还想收回？做梦
func (ctrl *ConfessionController) DeleteComment(c *gin.Context) {
	// 从 query 获取 conmment_id
	commentID, err := QueryUint(c, "comment_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前登录用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "你需要登录才能删除评论喵~"})
		return
	}

	// 查询评论，确认是自己发的评论才能删除
	var comment model.Comment
	if err := ctrl.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到这个评论喵~"})
		return
	}

	if comment.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能删除别人的评论喵~，你这个大hentai！"})
		return
	}

	// 调用 service 删除
	if err := service.DeleteComment(ctrl.DB, commentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除评论失败喵~"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "评论已成功删除喵~"})
}

// 嘴上说着不要，身体还是诚实的乖乖写了删除评论的controller了呢……

// 查看某条表白的评论
func (ctrl *ConfessionController) ListComments(c *gin.Context) {
	// 从 query 获取 confession_id
	confessionID, err := QueryUint(c, "confession_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 调用 service 层获取评论列表
	comments, err := service.ListComments(ctrl.DB, confessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取评论失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": comments,
	})
}
