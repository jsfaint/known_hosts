package main

import (
	"reflect"
	"testing"
)

// ============================================================================
// Zero-Doc Spec Coding: Security-First Test Contracts
// ============================================================================
// 本文件定义了安全相关的测试契约，涵盖主机删除的安全行为。
// 业务规则通过测试断言体现，无需额外文档。

// TestDelete_ExactMatchForCLI 测试 CLI 模式下的精确匹配删除行为
// @UseCase DeletingHostByExactMatch: 用户通过 CLI 删除主机时必须使用精确匹配以防止意外删除
func TestDelete_ExactMatchForCLI(t *testing.T) {
	t.Run("@MSS-1: CLI精确主机名匹配应仅删除该主机条目", func(t *testing.T) {
		// Arrange: 准备包含多个相似主机的测试数据
		input := []string{
			"github.com ssh-rsa key1",
			"gitlab.com ssh-rsa key2",
			"gitea.example.com ssh-rsa key3",
		}

		// Act: 执行精确匹配删除 github.com
		got := Delete(input, "github.com")

		// Assert: 只删除了 github.com，其他包含 "git" 的主机保留
		want := []string{"gitlab.com ssh-rsa key2", "gitea.example.com ssh-rsa key3"}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("精确匹配删除失败: got %v, want %v", got, want)
		}
	})

	t.Run("@MSS-2: CLI精确IP匹配应仅删除该IP条目", func(t *testing.T) {
		input := []string{
			"192.168.1.1 ssh-rsa key1",
			"192.168.1.2 ssh-rsa key2",
			"192.168.2.1 ssh-rsa key3",
		}

		got := Delete(input, "192.168.1.1")
		want := []string{"192.168.1.2 ssh-rsa key2", "192.168.2.1 ssh-rsa key3"}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("精确IP匹配删除失败: got %v, want %v", got, want)
		}
	})

	t.Run("@MSS-3: CLI精确主机加IP格式匹配应仅删除该完整条目", func(t *testing.T) {
		input := []string{
			"myserver,192.168.1.1 ssh-rsa key1",
			"myserver,192.168.1.2 ssh-rsa key2",
		}

		got := Delete(input, "myserver,192.168.1.1")
		want := []string{"myserver,192.168.1.2 ssh-rsa key2"}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("精确主机加IP匹配删除失败: got %v, want %v", got, want)
		}
	})

	t.Run("@Ext-4a: CLI使用短字符串前缀应拒绝删除以防意外批量删除", func(t *testing.T) {
		input := []string{
			"github.com ssh-rsa key1",
			"gitlab.com ssh-rsa key2",
			"bitbucket.org ssh-rsa key3",
		}

		// 尝试使用 "git" 删除（模糊匹配）
		got := Delete(input, "git")

		// 应该不删除任何内容，因为 "git" 不精确匹配任何主机部分
		want := []string{
			"github.com ssh-rsa key1",
			"gitlab.com ssh-rsa key2",
			"bitbucket.org ssh-rsa key3",
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("模糊匹配应该被拒绝: got %v, want %v", got, want)
		}
	})

	t.Run("@Ext-4b: CLI使用不存在的精确主机应返回原列表不变", func(t *testing.T) {
		input := []string{
			"github.com ssh-rsa key1",
			"gitlab.com ssh-rsa key2",
		}

		got := Delete(input, "nonexistent.com")
		want := input

		if !reflect.DeepEqual(got, want) {
			t.Errorf("删除不存在主机应返回原列表: got %v, want %v", got, want)
		}
	})
}

// TestDelete_FullLineMatchForTUI 测试 TUI 模式下的完整行匹配删除行为
// @UseCase DeletingHostByFullLine: TUI模式下使用完整行匹配确保删除操作的精确性
func TestDelete_FullLineMatchForTUI(t *testing.T) {
	t.Run("@MSS-1: TUI完整主机行匹配应删除该条目", func(t *testing.T) {
		input := []string{
			"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC1",
			"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC2",
			"gitlab.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC3",
		}

		fullLine := "github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC1"
		got := Delete(input, fullLine)
		want := []string{
			"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC2",
			"gitlab.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC3",
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("完整行匹配删除失败: got %v, want %v", got, want)
		}
	})

	t.Run("@Ext-2a: TUI完整行不匹配时应回退到主机部分精确匹配", func(t *testing.T) {
		input := []string{
			"github.com ssh-rsa key1",
			"gitlab.com ssh-rsa key2",
		}

		// 输入只有主机名（CLI 回退场景）
		got := Delete(input, "github.com")
		want := []string{"gitlab.com ssh-rsa key2"}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("主机部分精确匹配失败: got %v, want %v", got, want)
		}
	})
}

// TestDelete_SecurityEdgeCases 测试删除操作的安全边界情况
// @UseCase HandlingDeleteEdgeCases: 正确处理边界情况以防止安全漏洞
func TestDelete_SecurityEdgeCases(t *testing.T) {
	t.Run("@Ext-1a: 删除空字符串应保留所有主机", func(t *testing.T) {
		input := []string{
			"github.com ssh-rsa key1",
			"gitlab.com ssh-rsa key2",
		}

		got := Delete(input, "")
		want := input

		if !reflect.DeepEqual(got, want) {
			t.Errorf("空字符串不应删除任何内容: got %v, want %v", got, want)
		}
	})

	t.Run("@Ext-1b: 删除包含特殊字符的主机名应正确处理", func(t *testing.T) {
		input := []string{
			"my-server_01.example.com ssh-rsa key1",
			"my-server_02.example.com ssh-rsa key2",
		}

		got := Delete(input, "my-server_01.example.com")
		want := []string{"my-server_02.example.com ssh-rsa key2"}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("特殊字符主机名删除失败: got %v, want %v", got, want)
		}
	})

	t.Run("@Ext-1c: 删除空行输入应返回空切片", func(t *testing.T) {
		input := []string{"", ""}
		got := Delete(input, "anything")
		var want []string

		if !reflect.DeepEqual(got, want) {
			t.Errorf("空行输入应返回空切片: got %v, want %v", got, want)
		}
	})
}

// TestSearch_NoFuzzyMatchingOnDeletionPath 测试搜索结果与删除路径的安全性
// @UseCase SearchingHostsSafely: 搜索结果应用于删除时需防止误操作
func TestSearch_NoFuzzyMatchingOnDeletionPath(t *testing.T) {
	t.Run("@MSS-1: 搜索功能可使用模糊匹配但不应影响删除的精确性", func(t *testing.T) {
		input := []string{
			"github.com ssh-rsa key1",
			"gitlab.com ssh-rsa key2",
			"bitbucket.org ssh-rsa key3",
		}

		// 搜索 "git" 应返回两个结果
		searchResults := Search(input, "git")
		if len(searchResults) != 2 {
			t.Errorf("搜索 'git' 应返回2个结果, got %d", len(searchResults))
		}

		// 但删除 "git" 应该不删除任何内容（精确匹配）
		afterDelete := Delete(input, "git")
		if len(afterDelete) != 3 {
			t.Errorf("删除 'git' 应保留所有3个主机, got %d", len(afterDelete))
		}
	})
}
