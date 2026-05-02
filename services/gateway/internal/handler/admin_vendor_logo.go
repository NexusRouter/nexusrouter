package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const maxVendorLogoBytes = 512 << 10

var allowedVendorLogoExt = map[string]struct{}{
	".svg": {}, ".png": {}, ".jpg": {}, ".jpeg": {}, ".gif": {}, ".webp": {},
}

// adminUploadVendorLogo 接收 multipart 文件，保存到上传目录并返回可写入 model_vendor.logo 的 URL 路径。
func adminUploadVendorLogo(cfg *config.Config, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg == nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "配置为空")
			return
		}
		fh, err := c.FormFile("file")
		if err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_REQUEST", "缺少 multipart 字段 file")
			return
		}
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		if _, ok := allowedVendorLogoExt[ext]; !ok {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_TYPE", "仅支持 svg、png、jpg、jpeg、gif、webp")
			return
		}
		src, err := fh.Open()
		if err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_FILE", "无法读取上传文件")
			return
		}
		defer func() { _ = src.Close() }()

		limited := io.LimitReader(src, maxVendorLogoBytes+1)
		data, err := io.ReadAll(limited)
		if err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_FILE", "读取失败")
			return
		}
		if len(data) > maxVendorLogoBytes {
			WriteGatewayError(c, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "文件超过 512KB")
			return
		}
		if err := validateVendorLogoContent(ext, data); err != nil {
			WriteGatewayError(c, http.StatusBadRequest, "INVALID_CONTENT", err.Error())
			return
		}

		var id [16]byte
		if _, err := rand.Read(id[:]); err != nil {
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "生成文件名失败")
			return
		}
		name := hex.EncodeToString(id[:]) + ext
		root := cfg.EffectiveUploadsDir()
		dir := filepath.Join(root, "vendor-logos")
		if err := os.MkdirAll(dir, 0755); err != nil {
			if log != nil {
				log.Error("vendor logo mkdir", zap.Error(err))
			}
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "无法创建上传目录")
			return
		}
		dstPath := filepath.Join(dir, name)
		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			if log != nil {
				log.Error("vendor logo write", zap.Error(err))
			}
			WriteGatewayError(c, http.StatusInternalServerError, "INTERNAL", "写入失败")
			return
		}
		publicPath := "/uploads/vendor-logos/" + name
		c.JSON(http.StatusOK, gin.H{"path": publicPath})
	}
}

func validateVendorLogoContent(ext string, data []byte) error {
	if len(data) < 8 {
		return errors.New("文件过短")
	}
	switch ext {
	case ".svg":
		s := strings.TrimSpace(string(data[:min(256, len(data))]))
		if strings.HasPrefix(s, "<svg") || strings.HasPrefix(s, "<?xml") {
			return nil
		}
		return errors.New("SVG 格式无效")
	case ".png":
		if string(data[0:8]) != "\x89PNG\r\n\x1a\n" {
			return errors.New("PNG 头无效")
		}
	case ".jpg", ".jpeg":
		if data[0] != 0xff || data[1] != 0xd8 {
			return errors.New("JPEG 头无效")
		}
	case ".gif":
		if string(data[0:6]) != "GIF87a" && string(data[0:6]) != "GIF89a" {
			return errors.New("GIF 头无效")
		}
	case ".webp":
		if string(data[0:4]) != "RIFF" || len(data) < 12 || string(data[8:12]) != "WEBP" {
			return errors.New("WEBP 头无效")
		}
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
