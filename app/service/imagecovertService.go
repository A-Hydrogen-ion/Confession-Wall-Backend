package service

//太棒了，这个处理头像的service成终极大史山了
import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
)

// 上传图片并返回文件路径
func UploadImages(c *gin.Context, userID uint) ([]string, error) {
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
		savePath := "uploads/confessions/" + saveName                    //将图片存储到本地的路径
		out, err := os.Create(savePath)                                  //保存图片
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("服务器娘说图片保存失败惹: %v", err)
		}
		file.Seek(0, 0)
		// 写入内容
		if _, err := io.Copy(out, file); err != nil {
			out.Close()
			file.Close()
			return nil, fmt.Errorf("图片写入失败: %v", err)
		}
		out.Close()
		file.Close()
		imagePaths = append(imagePaths, savePath)
	}

	return imagePaths, nil
}

func UploadAvatar(c *gin.Context, userID uint) (string, error) {
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		return "", fmt.Errorf("获取文件失败: %v", err)
	}
	defer file.Close()
	// 限制文件大小5MB
	const maxSize = 5 * 1024 * 1024
	if header.Size > maxSize {
		return "", fmt.Errorf("服务器娘不接受超过5MB的图片喵")
	}

	// 检查文件类型，这个就没必要分类保存了
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", fmt.Errorf("请不要上传除jpg/png格式以外的图片，服务器娘处理不了这些图片喵~")
	}

	// 解码图片
	var img image.Image
	if ext == ".png" {
		img, err = png.Decode(file)
	} else {
		img, err = jpeg.Decode(file)
	}
	if err != nil {
		return "", fmt.Errorf("服务器娘看不懂你上传的图片喵: %v", err)
	}

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

	// 压缩到固定尺寸 256x256
	//我知道这个库万年没更新现在还read-only了，但是我真的找不到替代而且有完整教程的了
	resizedImg := resize.Resize(256, 256, croppedImg, resize.Lanczos3)

	// 保存文件
	saveDir := "uploads/avatars/"
	//由于用户头像唯一，所以只采用用户ID作为文件名，这样用户上传的时候旧的头像会被自动覆写掉
	savePath := fmt.Sprintf("%savatar_%d%s", saveDir, userID, ext)
	out, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("保存头像失败: %v", err)
	}
	defer out.Close()

	if ext == ".png" {
		err = png.Encode(out, resizedImg)
	} else {
		err = jpeg.Encode(out, resizedImg, &jpeg.Options{Quality: 85})
	}
	if err != nil {
		return "", fmt.Errorf("写入头像失败: %v", err)
	}

	return savePath, nil
}
