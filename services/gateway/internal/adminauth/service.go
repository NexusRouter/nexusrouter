// Package adminauth 提供管理控制台 JWT 签发与校验。
package adminauth

import (
	"errors"
	"strings"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/NexusRouter/nexusrouter/services/gateway/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CtxClaims Gin 上下文中 Claims 的键名。
const CtxClaims = "admin_jwt_claims"

// Claims 管理端 JWT。
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// Service 管理端认证。
type Service struct {
	cfg *config.Config
	db  *gorm.DB
}

// New 若控制台未完整配置则返回 nil。
func New(cfg *config.Config, db *gorm.DB) *Service {
	if !IsConsoleConfigured(cfg, db) {
		return nil
	}
	return &Service{cfg: cfg, db: db}
}

// Login 校验用户名密码并返回 JWT、过期时刻与角色（admin | operator）。
func (s *Service) Login(username, password string, remember bool) (token string, exp time.Time, role string, err error) {
	if s == nil {
		return "", time.Time{}, "", errors.New("adminauth: 未启用")
	}
	u := strings.TrimSpace(username)
	p := password
	expDur := s.cfg.AdminJWTExpire
	if remember && s.cfg.AdminRefreshExpire > expDur {
		expDur = s.cfg.AdminRefreshExpire
	}
	exp = time.Now().UTC().Add(expDur)

	if s.db != nil {
		var row repository.AdminUserModel
		if err := s.db.Where("username = ?", u).First(&row).Error; err == nil {
			if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordBcrypt), []byte(p)); err != nil {
				return "", time.Time{}, "", errors.New("凭据无效")
			}
			r := strings.TrimSpace(row.Role)
			if r != "admin" && r != "operator" {
				return "", time.Time{}, "", errors.New("凭据无效")
			}
			tok, err := s.signJWT(r, row.Username, exp)
			if err != nil {
				return "", time.Time{}, "", err
			}
			return tok, exp, r, nil
		}
	}

	opUser := strings.TrimSpace(s.cfg.AdminOperatorUsername)
	opHash := strings.TrimSpace(s.cfg.AdminOperatorPasswordBcrypt)
	if opUser != "" && opHash != "" && u == opUser {
		if err := bcrypt.CompareHashAndPassword([]byte(opHash), []byte(p)); err != nil {
			return "", time.Time{}, "", errors.New("凭据无效")
		}
		tok, err := s.signJWT("operator", opUser, exp)
		if err != nil {
			return "", time.Time{}, "", err
		}
		return tok, exp, "operator", nil
	}

	if u != s.cfg.AdminUsername {
		return "", time.Time{}, "", errors.New("凭据无效")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(s.cfg.AdminPasswordBcrypt), []byte(p)); err != nil {
		return "", time.Time{}, "", errors.New("凭据无效")
	}
	tok, err := s.signJWT("admin", s.cfg.AdminUsername, exp)
	if err != nil {
		return "", time.Time{}, "", err
	}
	return tok, exp, "admin", nil
}

func (s *Service) signJWT(role, subject string, exp time.Time) (string, error) {
	claims := Claims{
		Role: strings.TrimSpace(role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    "nexusrouter-admin",
			Subject:   subject,
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.cfg.AdminJWTSecret))
}

// Parse 解析并校验 Bearer JWT。
func (s *Service) Parse(authz string) (*Claims, error) {
	if s == nil {
		return nil, errors.New("adminauth: 未启用")
	}
	h := strings.TrimSpace(authz)
	if len(h) < 8 || !strings.EqualFold(h[:7], "bearer ") {
		return nil, errors.New("缺少 Bearer")
	}
	raw := strings.TrimSpace(h[7:])
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.AdminJWTSecret), nil
	})
	if err != nil || !tok.Valid {
		return nil, err
	}
	c, ok := tok.Claims.(*Claims)
	if !ok || c == nil {
		return nil, errors.New("invalid claims")
	}
	return c, nil
}
