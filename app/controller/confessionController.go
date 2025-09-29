package controller

import (
	"errors"
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
		ReturnMsg(c, 401, "只有登录的孩子才能发布表白喵~")
		return
	}

	// 调用封装的图片上传函数
	imagePaths, err := service.UploadImages(c, userID.(uint))
	//终于简洁的多了（狂喜
	if err != nil {
		ReturnError400(c, err)
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
		ReturnError400(c, err)
		return
	}
	ReturnMsg(c, http.StatusOK, "发布成功了喵~")
}

// 修改表白,返回错误统一调用authcontroller里的returnmsg
func (ctrl *ConfessionController) UpdateConfession(c *gin.Context) {
	confessionIDStr := c.PostForm("confession_id")
	confessionID, err := strconv.ParseUint(confessionIDStr, 10, 64)
	if err != nil {
		ReturnMsg(c, 400, "服务器娘没有查询到这个表白，可能已经被删除了喵~")
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		ReturnMsg(c, 401, "只有登录的孩子才能修改表白喵~")
		return
	}
	var confession model.Confession
	if err := ctrl.DB.First(&confession, confessionID).Error; err != nil {
		ReturnMsg(c, 404, "服务器娘没有查询到这个表白，可能已经被删除了喵~")
		return
	}
	if confession.UserID != userID.(uint) {
		ReturnMsg(c, 403, "你居然想修改别人的表白，hentai！") //不准修改别人的表白
		return
	}
	newContent := c.PostForm("content")
	form, err := c.MultipartForm()
	if err != nil {
		ReturnMsg(c, 400, "图片因为某些原因上传失败了喵")
		return
	}
	files := form.File["images"]
	if len(files) > 9 {
		ReturnMsg(c, 400, "9张图片已经能让对方感受到你的心意了，不要再上传了喵~")
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
		ReturnError400(c, err) //imagePath会返回更多的错误信息，这里只做最简单的错误处理
		return
	}

	confession.Images = imagePaths
	confession.Content = newContent
	if err := ctrl.DB.Save(&confession).Error; err != nil {
		ReturnMsg(c, 500, "服务器娘宕机了，修改失败了喵~")
		return
	}
	ReturnMsg(c, 200, "修改成功了喵~")
}

// 查看社区表白
func (ctrl *ConfessionController) ListPublicConfessions(c *gin.Context) {
	var uid uint = 0
	if userID, exists := c.Get("user_id"); exists {
		uid = userID.(uint)
	}
	// 解析分页参数
	limitStr := c.DefaultQuery("PageLimit", "10")
	offsetStr := c.DefaultQuery("Page", "0")
	limit, err1 := strconv.Atoi(limitStr)
	offset, err2 := strconv.Atoi(offsetStr)
	if err1 != nil || err2 != nil || limit < 1 || offset < 0 {
		ReturnMsg(c, 400, "分页参数不合法喵，你看看你都传入了些什么分页，服务器娘愤怒的告诉你她找不到负数的页码")
		return
	}
	confessions, err := service.ListPublicConfessions(ctrl.DB, uid, limit, offset)
	if err != nil {
		ReturnMsg(c, 500, "获取失败了喵~")
		return
	}
	for i := range confessions {
		if confessions[i].Anonymous {
			confessions[i].UserID = 0
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": confessions,
		"msg":  "获取成功了喵~",
	})
}

// 根据ID获取单条表白
func (ctrl *ConfessionController) GetConfessionByID(c *gin.Context) {
	confessionID, err := QueryUint(c, "id")
	if err != nil {
		ReturnError400(c, err)
		return
	}
	confession, err := service.GetConfessionByID(ctrl.DB, confessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "没有找到这条表白喵~"})
		return
	}
	// 匿名处理
	if confession.Anonymous {
		confession.UserID = 0
	}
	c.JSON(http.StatusOK, gin.H{"data": confession})
}

// 获取某用户的所有表白（需登录，排除黑名单和私密）
func (ctrl *ConfessionController) GetUserConfessions(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		ReturnMsg(c, 401, "你需要登录才能查看哦喵~")
		return
	}
	targetUserID, err := QueryUint(c, "user_id")
	if err != nil {
		ReturnError400(c, err)
		return
	}
	// 解析分页参数
	limitStr := c.DefaultQuery("PageLimit", "10")
	offsetStr := c.DefaultQuery("Page", "0")
	limit, err1 := strconv.Atoi(limitStr)
	offset, err2 := strconv.Atoi(offsetStr)
	if err1 != nil || err2 != nil || limit < 1 || offset < 0 {
		ReturnMsg(c, 400, "分页参数不合法喵，你看看你都传入了些什么分页，服务器娘愤怒的告诉你她找不到负数的页码")
		return
	}
	confessions, err := service.GetUserConfessions(ctrl.DB, targetUserID, currentUserID.(uint), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器娘宕机了,获取TA的表白失败了喵"})
		return
	}
	for i := range confessions {
		if confessions[i].Anonymous {
			confessions[i].UserID = 0
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": confessions})
}

// 删除表白（同时删除所有评论）
func (ctrl *ConfessionController) DeleteConfession(c *gin.Context) {
	confessionID, err := QueryUint(c, "confession_id")
	if err != nil {
		ReturnError400(c, err)
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		ReturnMsg(c, 401, "你需要登录才能删除表白喵~")
		return
	}
	// 查询表白，确认是自己发的才能删
	var confession model.Confession
	if err := ctrl.DB.First(&confession, confessionID).Error; err != nil {
		ReturnMsg(c, 404, "服务器娘没有查询到这个表白，可能已经被删除了喵~")
		return
	}
	if confession.UserID != userID.(uint) {
		ReturnMsg(c, 403, "不能删除别人的表白，你个hentai!")
		return
	}
	// 调用 service 层删除表白和评论
	if err := service.DeleteConfession(ctrl.DB, confessionID); err != nil {
		ReturnMsg(c, 500, "服务器娘宕机了，删除失败了喵~")
		return
	}
	ReturnMsg(c, http.StatusOK, "表白成功删除了喵~")
}

// 发布评论
func (ctrl *ConfessionController) AddComment(c *gin.Context) {
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
func (ctrl *ConfessionController) DeleteComment(c *gin.Context) {
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
func (ctrl *ConfessionController) ListComments(c *gin.Context) {
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
