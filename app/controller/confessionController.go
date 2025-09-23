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
// 注册数据包
type ConfessionController struct {
	DB *gorm.DB
}

func CreateConfessionController(db *gorm.DB) *ConfessionController {
	return &ConfessionController{DB: db}
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
	imagePaths, err := ctrl.ctrlImageUpload(c, userID.(uint))
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
	//给评论加上用户ID
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
	imagePaths, err := ctrl.ctrlImageUpload(c, userID.(uint))
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

// 上传图片并返回文件路径
func (ctrl *ConfessionController) ctrlImageUpload(c *gin.Context, userID uint) ([]string, error) {
	// 处理上传图片
	form, err := c.MultipartForm()
	if err != nil {
		return nil, fmt.Errorf("图片上传失败: %v", err)
	}
	//判断图片数量
	files := form.File["images"]
	if len(files) > 9 {
		return nil, fmt.Errorf("一次最多上传9张图片喵~")
	}
	//允许的文件类型和保存路径
	var imagePaths []string
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}

	// 对上传图片进行格式和大小检查
	for i, fileHeader := range files {
		if fileHeader.Size > 15*1024*1024 {
			return nil, fmt.Errorf("服务器娘不接受超过15MB的图片喵~")
		}

		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("服务器娘看不懂你上传的图片喵: %v", err)
		}
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		filetype := http.DetectContentType(buf[:n])

		if !allowed[filetype] {
			file.Close()
			return nil, fmt.Errorf("请不要上传除jpg/png/webp格式以外的图片，服务器娘处理不了这些图片喵~")
		}

		// 保存图片文件
		ext := fileHeader.Filename[strings.LastIndex(fileHeader.Filename, "."):]
		timestamp := time.Now().UnixNano()                               //获取当前系统时间
		saveName := fmt.Sprintf("%d_%d_%d%s", userID, timestamp, i, ext) //获取时间、用户ID，将图片重命名为这样的格式
		savePath := "uploads/" + saveName                                //将图片存储到本地的路径
		out, err := os.Create(savePath)                                  //保存图片
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("服务器娘说图片保存失败惹: %v", err)
		}
		file.Seek(0, 0)
		out.Close()
		file.Close()
		imagePaths = append(imagePaths, savePath)
	}

	return imagePaths, nil
}
