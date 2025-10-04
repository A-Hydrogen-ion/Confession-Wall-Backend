package controller

import (
	"errors"
	"net/http"
	"os"
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

// ParsePublishTime 统一解析和校验定时发布时间
func ParsePublishTime(publishTimeStr string, maxDelay time.Duration) (time.Time, error) {
	now := time.Now()
	if publishTimeStr == "" {
		return now, nil
	}
	publishTime, err := time.Parse(time.RFC3339, publishTimeStr)
	if err != nil {
		return time.Time{}, errors.New("定时发布时间格式不正确喵，请使用RFC3339格式哦~")
	}
	if publishTime.After(now.Add(maxDelay)) {
		return time.Time{}, errors.New("定时发布时间不能超过一周喵~")
	}
	if publishTime.Before(now.Add(-1 * time.Minute)) {
		return time.Time{}, errors.New("过去的表白时间不被允许哦喵~")
	}
	return publishTime, nil
}
func GetUserID(c *gin.Context) (uint, error) { //获取用户ID的统一辅助函数
	uid, exists := c.Get("user_id")
	if !exists {
		respondJSON(c, http.StatusUnauthorized, "只有登录的孩子才能发布表白喵~", nil)
		return 0, errors.New("未登录")
	}
	return uid.(uint), nil
}

// CreateConfession 发布表白
func (ctrl *ConfessionController) CreateConfession(c *gin.Context) {
	var req model.CreateConfessionRequest
	if err := c.ShouldBind(&req); err != nil {
		respondJSON(c, http.StatusBadRequest, "参数格式不对喵~", nil)
		return
	}
	var publishTime time.Time
	maxDelay := 7 * 24 * time.Hour                                  // 最大允许定时发布延迟为7天
	publishTime, err := ParsePublishTime(req.PublishTime, maxDelay) //调用统一的时间解析函数
	if err != nil {
		respondJSON(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	//判断用户有没有登录
	userID, err := GetUserID(c)
	if err != nil {
		return
	}

	// 调用封装的图片上传函数
	imagePaths, err := service.UploadImages(c, userID)
	//终于简洁的多了（狂喜
	if err != nil { //对于可能的错误处理
		ReturnError400(c, err)
		return
	}
	// 保存表白与发布人的ID和表白属性
	confession := model.Confession{
		UserID:      userID,
		Content:     req.Content,
		Images:      imagePaths,
		Anonymous:   req.Anonymous,
		Private:     req.Private,
		PublishedAt: publishTime, //单独在controller里处理发布时间
	}
	if err := service.CreateConfession(ctrl.DB, &confession); err != nil {
		ReturnError400(c, err)
		return
	}
	respondJSON(c, http.StatusOK, "发布成功了喵~", nil)
}

// UpdateConfession 修改表白,返回错误统一调用authcontroller里的returnmsg
func (ctrl *ConfessionController) UpdateConfession(c *gin.Context) {
	//检查输入和权限
	confession := ctrl.CheckInput(c)
	if confession.ID == 0 {
		return //检查失败直接返回
	}
	//获取用户ID
	userID, err := GetUserID(c)
	if err != nil {
		return
	}
	//获取新的内容和图片
	newContent := c.PostForm("content")
	publishTimeStr := c.PostForm("publish_time")
	maxDelay := 7 * 24 * time.Hour // 最大允许定时发布延迟为7天
	if publishTimeStr != "" {
		publishTime, err := ParsePublishTime(publishTimeStr, maxDelay) //调用统一的时间解析函数
		if err != nil {
			ReturnMsg(c, http.StatusBadRequest, err.Error())
			return
		}
		confession.PublishedAt = publishTime //此处与发布表白不同！不传入则保持原有发布时间不变
	}
	form, err := c.MultipartForm()
	if err != nil {
		ReturnMsg(c, http.StatusBadRequest, "图片因为某些原因上传失败了喵")
		return
	}
	files := form.File["images"]
	if len(files) > 9 {
		ReturnMsg(c, http.StatusBadRequest, "9张图片已经能让对方感受到你的心意了，不要再上传了喵~")
		return
	}

	// MD，再写我要疯了，直接给原有图片文件全部删了重新上传，这样下面我就可以直接从创建表白那里复制过来了
	for _, oldPath := range confession.Images {
		_ = os.Remove(oldPath)
	}
	//对重新传回来的图片进行格式和大小检查
	// 调用封装的图片上传函数
	imagePaths, err := service.UploadImages(c, userID)
	if err != nil {
		ReturnError400(c, err) //imagePath会返回更多的错误信息，这里只做最简单的错误处理
		return
	}

	confession.Images = imagePaths
	confession.Content = newContent
	if err := ctrl.DB.Save(&confession).Error; err != nil {
		ReturnMsg(c, http.StatusInternalServerError, "服务器娘宕机了，修改失败了喵~")
		return
	}
	ReturnMsg(c, http.StatusOK, "修改成功了喵~")
}

// CheckInput 检查输入和权限的辅助函数
func (ctrl *ConfessionController) CheckInput(c *gin.Context) model.Confession {
	confessionID, err := GetUintParam(c, "confession_id")
	if err != nil {
		ReturnMsg(c, http.StatusBadRequest, "服务器娘没有查询到这个表白，可能已经被删除了喵~") //错误处理
		return model.Confession{}
	}
	userID, err := GetUserID(c)
	if err != nil {
		return model.Confession{}
	}
	var confession model.Confession //查询表白
	if err := ctrl.DB.First(&confession, confessionID).Error; err != nil {
		ReturnMsg(c, http.StatusNotFound, "服务器娘没有查询到这个表白，可能已经被删除了喵~")
		return model.Confession{}
	}
	if confession.UserID != userID {
		ReturnMsg(c, http.StatusForbidden, "你居然想修改别人的表白，hentai！") //不准修改别人的表白，对象错误处理
		return model.Confession{}
	}
	return confession

}

// ListPublicConfessions 查看社区表白
func (ctrl *ConfessionController) ListPublicConfessions(c *gin.Context) {
	var uid uint = 0
	if userID, exists := c.Get("user_id"); exists {
		uid = userID.(uint)
	}
	// 解析分页参数
	limit, offset, ok := ParsePagination(c)
	if !ok {
		return
	}
	confessions, err := service.ListPublicConfessions(ctrl.DB, uid, limit, offset)
	if err != nil {
		ReturnMsg(c, http.StatusInternalServerError, "获取失败了喵~")
		return
	}
	for i := range confessions {
		if confessions[i].Anonymous {
			confessions[i].UserID = 0
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": confessions,
		"msg":  "获取成功了喵~",
	})
}

// GetConfessionByID 根据ID获取单条表白
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
	// 浏览量+1
	_ = service.ViewCount(confessionID)
	// 匿名处理
	if confession.Anonymous {
		confession.UserID = 0
	}
	// 查询当前用户是否已点赞
	liked := false
	if userID, exists := c.Get("user_id"); exists {
		l, err := service.HasLiked(confessionID, userID.(uint))
		if err == nil {
			liked = l
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  confession,
		"liked": liked,
	})
}

// GetUserConfessions 获取某用户的所有表白（需登录，排除黑名单和私密）
func (ctrl *ConfessionController) GetUserConfessions(c *gin.Context) {
	currentUserID, exists := c.Get("user_id")
	if !exists {
		respondJSON(c, http.StatusUnauthorized, "你需要登录才能查看哦喵~", nil)
		return
	}
	targetUserID, err := QueryUint(c, "user_id")
	if err != nil {
		ReturnError400(c, err)
		return
	}
	// 解析分页参数
	limit, offset, ok := ParsePagination(c)
	if !ok {
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

// DeleteConfession 删除表白（同时删除所有评论）
func (ctrl *ConfessionController) DeleteConfession(c *gin.Context) {
	confessionID, err := QueryUint(c, "confession_id")
	if err != nil {
		ReturnError400(c, err)
		return
	}
	userID, exists := c.Get("user_id")
	if !exists {
		respondJSON(c, http.StatusUnauthorized, "你需要登录才能删除表白喵~", nil)
		return
	}
	// 查询表白，确认是自己发的才能删
	var confession model.Confession
	if err := ctrl.DB.First(&confession, confessionID).Error; err != nil {
		ReturnMsg(c, http.StatusNotFound, "服务器娘没有查询到这个表白，可能已经被删除了喵~")
		return
	}
	if confession.UserID != userID.(uint) {
		ReturnMsg(c, http.StatusForbidden, "不能删除别人的表白，你个hentai!")
		return
	}
	// 调用 service 层删除表白和评论
	if err := service.DeleteConfession(ctrl.DB, confessionID); err != nil {
		ReturnMsg(c, 500, "服务器娘宕机了，删除失败了喵~")
		return
	}
	ReturnMsg(c, http.StatusOK, "表白成功删除了喵~")
}
