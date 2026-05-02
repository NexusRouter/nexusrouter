package deps

import "testing"

// TestDepsPackageBuilds 空白导入包在 -cover 下需存在测试文件，否则工具链报错。
func TestDepsPackageBuilds(t *testing.T) {
	t.Parallel()
}
