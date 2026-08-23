package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRequiredProjectArtifactsAndReadmeSections(t *testing.T) {
	root := ".."
	required := []string{"Makefile", ".env.example", ".gitignore", "compose.yaml", "migrations/001_initial.sql", "migrations/002_seed.sql", "api/openapi/openapi.yaml", "web/package.json", "web/src/App.vue"}
	for _, relative := range required {
		if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(relative))); err != nil || info.IsDir() {
			t.Errorf("required file %s missing: %v", relative, err)
		}
	}
	content, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	readme := string(content)
	sections := []string{"## 模块职责", "## 本地启动", "## 配置", "## 迁移与演示数据", "## API 示例", "## 主要状态规则", "## 测试与构建", "## 实际验证结果"}
	for _, section := range sections {
		if !strings.Contains(readme, section) {
			t.Errorf("README section missing: %s", section)
		}
	}
}
