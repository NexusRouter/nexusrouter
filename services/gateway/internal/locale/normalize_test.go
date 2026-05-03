package locale

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeFromAcceptLanguage(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", TagEN},
		{"   ", TagEN}, // 不以 zh 开头
		{"en-US", TagEN},
		{"EN", TagEN},
		{"zh-CN", TagZH},
		{"zh-TW", TagZH},
		{"ZH-hans", TagZH},
		{"zh,en;q=0.9", TagZH},
		{"en;q=0.9, zh;q=0.8", TagEN},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, NormalizeFromAcceptLanguage(tc.in))
		})
	}
}

func TestWithLocale_ChildContextInherits(t *testing.T) {
	ctx := context.Background()
	ctx = WithLocale(ctx, TagZH)
	child, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	assert.Equal(t, TagZH, FromContext(child))
}
