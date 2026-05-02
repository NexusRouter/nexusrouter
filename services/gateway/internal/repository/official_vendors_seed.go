package repository

import (
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// vendorLogoPublicPath 管理台静态资源路径（Vite public/vendor-logos/<vendor_code>.svg）。
func vendorLogoPublicPath(vendorCode string) string {
	return "/vendor-logos/" + strings.TrimSpace(vendorCode) + ".svg"
}

// officialVendorSeeds 首版官方原厂（vendor_type=1），vendor_code 稳定唯一。
// Logo 指向 web/dashboard/public/vendor-logos/ 下同名 SVG（Simple Icons 与自绘占位，见该目录 README）。
// 命名与 PangaeaHub relay/channeltype 渠道概念对照见 openspec 变更 design.md；不含聚合商（如 OpenRouter）。
var officialVendorSeeds = []ModelVendor{
	{VendorName: "OpenAI", VendorType: 1, VendorCode: "openai", Logo: vendorLogoPublicPath("openai"), Status: 1},
	{VendorName: "Anthropic", VendorType: 1, VendorCode: "anthropic", Logo: vendorLogoPublicPath("anthropic"), Status: 1},
	{VendorName: "Google Gemini", VendorType: 1, VendorCode: "google_gemini", Logo: vendorLogoPublicPath("google_gemini"), Status: 1},
	{VendorName: "Azure OpenAI", VendorType: 1, VendorCode: "azure_openai", Logo: vendorLogoPublicPath("azure_openai"), Status: 1},
	{VendorName: "Baidu Qianfan", VendorType: 1, VendorCode: "baidu", Logo: vendorLogoPublicPath("baidu"), Status: 1},
	{VendorName: "Zhipu AI", VendorType: 1, VendorCode: "zhipu", Logo: vendorLogoPublicPath("zhipu"), Status: 1},
	{VendorName: "Alibaba DashScope", VendorType: 1, VendorCode: "aliyun_dashscope", Logo: vendorLogoPublicPath("aliyun_dashscope"), Status: 1},
	{VendorName: "Moonshot", VendorType: 1, VendorCode: "moonshot", Logo: vendorLogoPublicPath("moonshot"), Status: 1},
	{VendorName: "Baichuan", VendorType: 1, VendorCode: "baichuan", Logo: vendorLogoPublicPath("baichuan"), Status: 1},
	{VendorName: "MiniMax", VendorType: 1, VendorCode: "minimax", Logo: vendorLogoPublicPath("minimax"), Status: 1},
	{VendorName: "Mistral AI", VendorType: 1, VendorCode: "mistral", Logo: vendorLogoPublicPath("mistral"), Status: 1},
	{VendorName: "Groq", VendorType: 1, VendorCode: "groq", Logo: vendorLogoPublicPath("groq"), Status: 1},
	{VendorName: "DeepSeek", VendorType: 1, VendorCode: "deepseek", Logo: vendorLogoPublicPath("deepseek"), Status: 1},
	{VendorName: "Cohere", VendorType: 1, VendorCode: "cohere", Logo: vendorLogoPublicPath("cohere"), Status: 1},
	{VendorName: "xAI", VendorType: 1, VendorCode: "xai", Logo: vendorLogoPublicPath("xai"), Status: 1},
	{VendorName: "Together AI", VendorType: 1, VendorCode: "together", Logo: vendorLogoPublicPath("together"), Status: 1},
	{VendorName: "Cloudflare Workers AI", VendorType: 1, VendorCode: "cloudflare", Logo: vendorLogoPublicPath("cloudflare"), Status: 1},
	{VendorName: "Volcengine Doubao", VendorType: 1, VendorCode: "doubao", Logo: vendorLogoPublicPath("doubao"), Status: 1},
	{VendorName: "Novita", VendorType: 1, VendorCode: "novita", Logo: vendorLogoPublicPath("novita"), Status: 1},
	{VendorName: "Replicate", VendorType: 1, VendorCode: "replicate", Logo: vendorLogoPublicPath("replicate"), Status: 1},
	{VendorName: "Tencent Hunyuan", VendorType: 1, VendorCode: "hunyuan", Logo: vendorLogoPublicPath("hunyuan"), Status: 1},
}

// OfficialVendorSeedCount 预置官方厂商条数（测试断言用）。
func OfficialVendorSeedCount() int { return len(officialVendorSeeds) }

// SeedOfficialVendors 按 vendor_code 幂等插入官方厂商；已存在则跳过且不更新。
func SeedOfficialVendors(db *gorm.DB, log *zap.Logger) error {
	if db == nil {
		return nil
	}
	if log == nil {
		log = zap.NewNop()
	}
	for _, seed := range officialVendorSeeds {
		code := strings.TrimSpace(seed.VendorCode)
		if code == "" {
			continue
		}
		var existing ModelVendor
		err := db.Where("vendor_code = ?", code).First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("official vendor seed lookup %q: %w", code, err)
		}
		row := ModelVendor{
			VendorName: strings.TrimSpace(seed.VendorName),
			VendorType: seed.VendorType,
			VendorCode: code,
			Logo:       strings.TrimSpace(seed.Logo),
			Status:     seed.Status,
		}
		if row.Status == 0 {
			row.Status = 1
		}
		if err := db.Create(&row).Error; err != nil {
			if isUniqueConstraintVendorCode(err) {
				continue
			}
			return fmt.Errorf("official vendor seed insert %q: %w", code, err)
		}
		log.Info("预置官方厂商", zap.String("vendor_code", code), zap.String("vendor_name", row.VendorName))
	}
	return nil
}

func isUniqueConstraintVendorCode(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "vendor_code") &&
		(strings.Contains(s, "unique") || strings.Contains(s, "duplicate"))
}
