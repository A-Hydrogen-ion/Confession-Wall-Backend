package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

// respondJSON 是内部统一响应函数，保持和现有接口返回格式一致
func respondJSON(c *gin.Context, status int, msg string, data interface{}) {
	payload := gin.H{"code": status, "msg": msg}
	if data != nil { // 仅当 data 不为 nil 时添加 data 字段
		payload["data"] = data
	} else {
		// 保留兼容性字段
		payload["token"] = nil
	}
	c.JSON(status, payload) // 使用 c.JSON 直接返回 JSON 响应
}

// QueryUint 从 query 参数中解析 uint，返回 (value, error)
// 与现有控制器兼容：当参数缺失或解析失败时返回 error
func QueryUint(c *gin.Context, key string) (uint, error) {
	valStr := c.Query(key) // 只从 Query 参数中获取
	if valStr == "" {
		return 0, errors.New(key + " 参数为空")
	}
	v, err := strconv.ParseUint(valStr, 10, 64)
	if err != nil {
		return 0, errors.New(key + " 参数无效")
	}
	return uint(v), nil // 转换为 uint 并返回
}

// ParsePagination 解析 PageLimit / Page 查询参数并返回 (limit, offset, ok)
// 兼容之前 controller 中的 ParsePagination 实现
func ParsePagination(c *gin.Context) (limit int, offset int, ok bool) {
	limitStr := c.DefaultQuery("PageLimit", "10") // 默认每页10条
	offsetStr := c.DefaultQuery("Page", "0")      // 默认第0页
	l, err1 := strconv.Atoi(limitStr)             // 转换为整数
	o, err2 := strconv.Atoi(offsetStr)
	if err1 != nil || err2 != nil || l < 1 || o < 0 { //合法性检查
		respondJSON(c, 400, "分页参数不合法喵，你看看你都传入了些什么分页，服务器娘愤怒的告诉你她找不到负数的页码", nil)
		return 0, 0, false
	}
	return l, o, true
}

// GetUintParam 尝试从 Query, PostForm, Path Param 中解析 uint 值，顺序为 Query -> PostForm -> Param
func GetUintParam(c *gin.Context, key string) (uint, error) {
	if v := c.Query(key); v != "" { // 优先从 Query 参数中获取
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			return uint(n), nil
		}
		return 0, errors.New(key + " 参数无效")
	}
	if v := c.PostForm(key); v != "" { //其次从 PostForm 参数中获取
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			return uint(n), nil
		}
		return 0, errors.New(key + " 参数无效")
	}
	if v := c.Param(key); v != "" { // 最后从 Path 参数中获取
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			return uint(n), nil
		}
		return 0, errors.New(key + " 参数无效")
	}
	return 0, errors.New(key + " 参数为空")
}
