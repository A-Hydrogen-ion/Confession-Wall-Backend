package controller

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/model"
	"github.com/A-Hydrogen-ion/Confession-Wall-Backend/app/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ConfessionController 控制器
// 需要在 router 注册时加 JWT 中间件

type ConfessionController struct {
	DB *gorm.DB
}

func CreateConfessionController(db *gorm.DB) *ConfessionController {
	return &ConfessionController{DB: db}
}

// 发布表白
func (ctrl *ConfessionController) CreateConfession(c *gin.Context) {
	content := c.PostForm("content")
	anonymous := c.PostForm("anonymous") == "true"
	private := c.PostForm("private") == "true"
	//判断用户有没有登录
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "只有登录的孩子才能发布表白喵~"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(400, gin.H{"error": "图片因为某些原因上传失败了喵"})
		return
	}
	//限制上传的图片
	files := form.File["images"]
	if len(files) > 9 {
		c.JSON(400, gin.H{"error": "9张图片已经能让对方感受到你的心意了，不要再上传了喵~"})
		return
	}

	var imagePaths []string
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	//对上传图片进行格式和大小检查
	for i, fileHeader := range files {
		if fileHeader.Size > 15*1024*1024 {
			c.JSON(400, gin.H{"error": "单张图片不能超过15MB喵~"})
			return
		}
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(400, gin.H{"error": "服务器娘不知道你上传的图片是什么东西喵~"})
			return
		}
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		filetype := http.DetectContentType(buf[:n])
		if !allowed[filetype] {
			c.JSON(400, gin.H{"error": "请不要上传除jpg/png/webp格式以外的图片，服务器娘处理不了这些图片喵~"})
			file.Close()
			return
		}
		// 保留原始后缀，并重命名为时间和用户ID+图片张数，防止重名
		ext := fileHeader.Filename[strings.LastIndex(fileHeader.Filename, "."):]
		timestamp := time.Now().UnixNano()
		saveName := fmt.Sprintf("%d_%d_%d%s", userID.(uint), timestamp, i, ext)
		savePath := "uploads/" + saveName
		out, err := os.Create(savePath)
		if err != nil {
			c.JSON(500, gin.H{"error": "服务器娘说图片保存失败惹"})
			return
		}
		file.Seek(0, 0)
		out.Close()
		file.Close()
		imagePaths = append(imagePaths, savePath)
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

// 发布评论
func (ctrl *ConfessionController) AddComment(c *gin.Context) {
	var req model.Comment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "你需要登录才能发表评论哦喵~"})
		return
	}
	req.UserID = userID.(uint)
	if err := service.AddComment(ctrl.DB, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "评论发布成功了，对方收到你的心意了喵~"})
}

// 查看社区表白
func (ctrl *ConfessionController) ListPublicConfessions(c *gin.Context) {
	var uid uint = 0
	// 获取当前用户ID，若未登录则为0，这样做可以让未登录用户也能看到公共表白
	if userID, exists := c.Get("user_id"); exists {
		uid = userID.(uint)
	}
	confessions, err := service.ListPublicConfessions(ctrl.DB, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败"})
		return
	}
	// 匿名处理
	for i := range confessions {
		if confessions[i].Anonymous {
			confessions[i].UserID = 0
			// 可选：清空昵称、头像等
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": confessions})
}

// 查看某条表白的评论
func (ctrl *ConfessionController) ListComments(c *gin.Context) {
	confessionID := c.Query("confession_id")
	if confessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "confession_id是空的？？"})
		return
	}
	var comments []model.Comment
	if err := ctrl.DB.Where("confession_id = ?", confessionID).Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取评论失败了喵~"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": comments})
}

// 修改表白
func (ctrl *ConfessionController) UpdateConfession(c *gin.Context) {
	confessionIDStr := c.PostForm("confession_id")
	confessionID, err := strconv.ParseUint(confessionIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "服务器娘没有查询到这个表白，可能已经被删除了喵~"})
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "只有登录的孩子才能修改表白喵~"})
		return
	}
	var confession model.Confession
	if err := ctrl.DB.First(&confession, confessionID).Error; err != nil {
		c.JSON(404, gin.H{"error": "服务器娘没有查询到这个表白，可能已经被删除了喵~"})
		return
	}
	if confession.UserID != userID.(uint) {
		c.JSON(403, gin.H{"error": "你居然想修改别人的表白，hentai！"})
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

	var imagePaths []string
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	// MD，再写我要疯了，直接给原有图片文件全部删了重新上传，这样下面我就可以直接从创建表白那里复制过来了
	for _, oldPath := range confession.Images {
		_ = os.Remove(oldPath)
	}
	//对重新传回来的图片进行格式和大小检查
	for i, fileHeader := range files {
		if fileHeader.Size > 15*1024*1024 {
			c.JSON(400, gin.H{"error": "单张图片不能超过15MB喵~"})
			return
		}
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(400, gin.H{"error": "服务器娘不知道你上传的图片是什么东西喵~"})
			return
		}
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		filetype := http.DetectContentType(buf[:n])
		if !allowed[filetype] {
			c.JSON(400, gin.H{"error": "请不要上传除jpg/png/webp格式以外的图片，服务器娘处理不了这些图片喵~"})
			file.Close()
			return
		}
		// 保留原始后缀，并重命名为时间和用户ID+图片张数，防止重名
		ext := fileHeader.Filename[strings.LastIndex(fileHeader.Filename, "."):]
		timestamp := time.Now().UnixNano()
		saveName := fmt.Sprintf("%d_%d_%d%s", userID.(uint), timestamp, i, ext)
		savePath := "uploads/" + saveName
		out, err := os.Create(savePath)
		if err != nil {
			c.JSON(500, gin.H{"error": "服务器娘说图片保存失败惹"})
			return
		}
		file.Seek(0, 0)
		out.Close()
		file.Close()
		imagePaths = append(imagePaths, savePath)
	}
	confession.Images = imagePaths
	confession.Content = newContent
	if err := ctrl.DB.Save(&confession).Error; err != nil {
		c.JSON(500, gin.H{"error": "修改失败"})
		return
	}
	c.JSON(200, gin.H{"msg": "修改成功"})
}
