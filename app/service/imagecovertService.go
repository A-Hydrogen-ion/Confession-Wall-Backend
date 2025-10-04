package service

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
)

// 太棒了，这个处理头像的service成终极大史山了

// UploadImages 上传图片并返回文件路径
func UploadImages(c *gin.Context, userID uint) ([]string, error) {
	const ( //常量集中定义
		maxCount  = 9                     // 最大上传数量
		maxSize   = 15 * 1024 * 1024      // 单个图片最大 15MB
		uploadDir = "uploads/confessions" // 统一目录常量
	) //确保目录存在
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("服务器娘说存储目录创建失败喵: %v", err)
	}
	form, err := c.MultipartForm()
	if err != nil {
		return nil, fmt.Errorf("图片上传失败: %v", err)
	}
	// 获取上传的文件
	files := form.File["images"]
	// 限制上传图片的数量和大小
	if len(files) == 0 {
		return nil, fmt.Errorf("你还没有上传图片喵~")
	}
	if len(files) > maxCount {
		return nil, fmt.Errorf("一次最多上传%d张图片喵~", maxCount)
	}
	var imagePaths []string
	for i, fileHeader := range files {
		if err := validateImage(fileHeader, maxSize); err != nil {
			return nil, err
		} // 校验文件
		if err := saveUploadedFile(fileHeader, userID, i, &imagePaths); err != nil {
			return nil, err
		} // 调用辅助函数保存单个文件
	}

	return imagePaths, nil
}

// saveUploadedFile 保存单个上传的文件
func saveUploadedFile(fileHeader *multipart.FileHeader, userID uint, index int, imagePaths *[]string) error {
	// 打开上传的源文件
	src, err := fileHeader.Open() //src为客户端上传的图片源文件
	if err != nil {
		return fmt.Errorf("服务器娘看不懂你上传的图片喵: %v", err)
	}
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {
			log.Printf("关闭 out 文件失败: %v", err)
		}
	}(src) // 此处 defer 在函数退出时执行，不会影响循环

	// 生成保存路径
	ext := filepath.Ext(fileHeader.Filename)
	timestamp := time.Now().UnixNano()
	saveName := fmt.Sprintf("%d_%d_%d%s", userID, timestamp, index, ext)
	savePath := "uploads/confessions/" + saveName

	// 创建目标文件
	dst, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("服务器娘说图片保存失败惹: %v", err)
	}
	defer func(dst *os.File) {
		err := dst.Close()
		if err != nil {
			log.Printf("关闭 out 文件失败: %v", err)
		}
	}(dst)

	// 复制文件内容
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("服务器娘说图片写入失败惹: %v", err)
	}

	// 添加到结果列表
	*imagePaths = append(*imagePaths, savePath)
	return nil
}

// UploadAvatar 上传用户头像，返回头像路径
func UploadAvatar(c *gin.Context, userID uint) (string, error) {
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		return "", fmt.Errorf("获取文件失败: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("关闭文件失败: %v", err) // 这里只能打印日志，不能 return，因为 defer 里 return 没意义
		}
	}()
	//限制图片最大大小
	maxSize := int64(5 * 1024 * 1024)
	//调用校验图片类型和大小的函数
	if err := validateImage(header, maxSize); err != nil {
		return "", err
	}
	// 读取文件内容
	//校验通过后重新打开文件
	file, err = header.Open()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	//对图片进行解码
	var img image.Image
	switch {
	case ext == ".png":
		img, err = png.Decode(file)
		if err != nil {
			return "", fmt.Errorf("PNG图片解码失败: %v", err)
		}
	case ext == ".jpg" || ext == ".jpeg":
		img, err = jpeg.Decode(file)
		if err != nil {
			return "", fmt.Errorf("JPG图片解码失败: %v", err)
		}
	case ext == ".webp": //image 包默认不支持 webp 格式的解码
		return "", fmt.Errorf("webp格式的图片暂时不支持作为头像喵，请转换成jpg或png格式后再上传")
	default:
		return "", fmt.Errorf("服务器娘看不懂你上传的图片喵: 不支持的图片格式")
	}
	resizedImg := cutImage(img)
	//设置图片存储路径
	saveDir := "uploads/avatars/"
	savePath := fmt.Sprintf("%savatar_%d%s", saveDir, userID, ext)
	out, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("保存头像失败，服务器娘不小心把你的头像弄丢了: %v", err)
	}
	defer func() {
		if err := out.Close(); err != nil {
			log.Printf("关闭 out 文件失败: %v", err)
		}
	}()
	if ext == ".png" {
		err = png.Encode(out, resizedImg)
	} else {
		//写入时将图片压缩
		err = jpeg.Encode(out, resizedImg, &jpeg.Options{Quality: 85})
	}
	if err != nil {
		return "", fmt.Errorf("写入头像失败，服务器娘不小心把你的头像弄丢了: %v", err)
	}
	return savePath, nil
}

// 裁剪与压缩传入的图片
func cutImage(img image.Image) image.Image {
	// 裁剪成正方形（以中心为基准）
	var cropLength int
	if img.Bounds().Dx() < img.Bounds().Dy() {
		cropLength = img.Bounds().Dx()
	} else {
		cropLength = img.Bounds().Dy()
	}
	cropRect := image.Rect(
		(img.Bounds().Dx()-cropLength)/2,
		(img.Bounds().Dy()-cropLength)/2,
		(img.Bounds().Dx()+cropLength)/2,
		(img.Bounds().Dy()+cropLength)/2,
	)
	croppedImg := image.NewRGBA(image.Rect(0, 0, cropLength, cropLength))
	draw.Draw(croppedImg, croppedImg.Bounds(), img, cropRect.Min, draw.Src)

	// 压缩到固定尺寸 512x512
	//我知道这个库万年没更新现在还read-only了，但是我真的找不到替代而且有完整教程的了
	resizedImg := resize.Resize(512, 512, croppedImg, resize.Lanczos3)
	return resizedImg
}

// 校验图片类型和大小
func validateImage(fileHeader *multipart.FileHeader, maxSize int64) error {
	if fileHeader.Size > maxSize {
		return fmt.Errorf("图片大小不能超过 %d MB", maxSize/1024/1024)
	}
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("图片打开失败，服务器娘理解不了你上传了什么: %v", err)//
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("关闭文件失败: %v", err) // 这里只能打印日志，不能 return，因为 defer 里 return 没意义
		}
	}()
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	filetype := http.DetectContentType(buf[:n])
	if !(filetype == "image/jpeg" || filetype == "image/png" || filetype == "image/webp") {
		return fmt.Errorf("服务器娘看不懂你上传的图片喵，只允许jpg/png/webp类型的图片哦: %s", filetype) // 只允许jpg/png/webp类型
	}
	return nil
}
