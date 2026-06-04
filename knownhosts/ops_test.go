package knownhosts

import (
	"reflect"
	"testing"
)

func TestSearch(t *testing.T) {
	type args struct {
		input   []string
		pattern string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{"first", args{[]string{"1", "2", "3", "4", "5"}, "1"}, []string{"1"}},
		{"last", args{[]string{"1", "2", "3", "4", "5"}, "5"}, []string{"5"}},
		{"middle", args{[]string{"1", "2", "3", "4", "5"}, "3"}, []string{"3"}},
		{"multi", args{[]string{"12", "21", "33", "44", "55"}, "1"}, []string{"12", "21"}},
		{"empty input", args{nil, "1"}, nil},
		{"empty pattern", args{[]string{"1", "2", "3"}, ""}, []string{"1", "2", "3"}},
		{"not found", args{[]string{"1", "2", "3"}, "99"}, nil},
		{"full host line", args{[]string{"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"}, "github"}, []string{"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"}},
		{"host with comma", args{[]string{"myserver,192.168.1.1 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"}, "myserver"}, []string{"myserver,192.168.1.1 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"}},
		{"ip search", args{[]string{"myserver,192.168.1.1 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"}, "192.168.1.1"}, []string{"myserver,192.168.1.1 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC"}},
		{"partial ip match", args{[]string{"myserver,192.168.1.1 ssh-rsa key"}, "192.168"}, []string{"myserver,192.168.1.1 ssh-rsa key"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Search(test.args.input, test.args.pattern)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("Not equal, want: %v, got: %v", test.want, got)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		input   []string
		pattern string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{"first", args{[]string{"1", "2", "3", "4", "5"}, "1"}, []string{"2", "3", "4", "5"}},
		{"last", args{[]string{"1", "2", "3", "4", "5"}, "5"}, []string{"1", "2", "3", "4"}},
		{"middle", args{[]string{"1", "2", "3", "4", "5"}, "3"}, []string{"1", "2", "4", "5"}},
		{"multi-1 exact", args{[]string{"11", "11", "33", "44", "55"}, "11"}, []string{"33", "44", "55"}},
		{"multi-2 exact", args{[]string{"11", "22", "11", "44", "55"}, "11"}, []string{"22", "44", "55"}},
		{"empty input", args{nil, "1"}, nil},
		{"not found", args{[]string{"1", "2", "3"}, "99"}, []string{"1", "2", "3"}},
		{"delete all", args{[]string{"1", "1", "1"}, "1"}, nil},
		{"exact host match", args{[]string{"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC", "gitlab.com ssh-rsa key"}, "github.com"}, []string{"gitlab.com ssh-rsa key"}},
		{"empty string skip", args{[]string{"1", "", "2"}, "1"}, []string{"2"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Delete(test.args.input, test.args.pattern)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("Not equal, want: %v, got: %v", test.want, got)
			}
		})
	}
}

func TestDeleteMatches(t *testing.T) {
	tests := []struct {
		name          string
		input         []string
		pattern       string
		wantRemaining []string
		wantRemoved   []string
	}{
		{
			name: "remove exact host part",
			input: []string{
				"github.com ssh-rsa key1",
				"github.com ssh-ed25519 key2",
				"gitlab.com ssh-rsa key3",
			},
			pattern: "github.com",
			wantRemaining: []string{
				"gitlab.com ssh-rsa key3",
			},
			wantRemoved: []string{
				"github.com ssh-rsa key1",
				"github.com ssh-ed25519 key2",
			},
		},
		{
			name: "remove exact full line",
			input: []string{
				"github.com ssh-rsa key1",
				"github.com ssh-ed25519 key2",
			},
			pattern: "github.com ssh-rsa key1",
			wantRemaining: []string{
				"github.com ssh-ed25519 key2",
			},
			wantRemoved: []string{
				"github.com ssh-rsa key1",
			},
		},
		{
			name: "no match",
			input: []string{
				"github.com ssh-rsa key1",
			},
			pattern: "bitbucket.org",
			wantRemaining: []string{
				"github.com ssh-rsa key1",
			},
			wantRemoved: []string(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRemaining, gotRemoved := DeleteMatches(tt.input, tt.pattern)
			if !reflect.DeepEqual(gotRemaining, tt.wantRemaining) {
				t.Errorf("DeleteMatches() remaining = %v, want %v", gotRemaining, tt.wantRemaining)
			}
			if !reflect.DeepEqual(gotRemoved, tt.wantRemoved) {
				t.Errorf("DeleteMatches() removed = %v, want %v", gotRemoved, tt.wantRemoved)
			}
		})
	}
}

// Security tests: exact-match delete contract

func TestDelete_ExactMatchForCLI(t *testing.T) {
	t.Run("CLI精确主机名匹配应仅删除该主机条目", func(t *testing.T) {
		input := []string{
			"github.com ssh-rsa key1",
			"gitlab.com ssh-rsa key2",
			"gitea.example.com ssh-rsa key3",
		}

		got := Delete(input, "github.com")

		want := []string{"gitlab.com ssh-rsa key2", "gitea.example.com ssh-rsa key3"}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("精确匹配删除失败: got %v, want %v", got, want)
		}
	})

	t.Run("CLI精确IP匹配应仅删除该IP条目", func(t *testing.T) {
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

	t.Run("CLI精确主机加IP格式匹配应仅删除该完整条目", func(t *testing.T) {
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

	t.Run("CLI使用短字符串前缀应拒绝删除以防意外批量删除", func(t *testing.T) {
		input := []string{
			"github.com ssh-rsa key1",
			"gitlab.com ssh-rsa key2",
			"bitbucket.org ssh-rsa key3",
		}

		got := Delete(input, "git")

		want := []string{
			"github.com ssh-rsa key1",
			"gitlab.com ssh-rsa key2",
			"bitbucket.org ssh-rsa key3",
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("模糊匹配应该被拒绝: got %v, want %v", got, want)
		}
	})

	t.Run("CLI使用不存在的精确主机应返回原列表不变", func(t *testing.T) {
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

func TestDelete_FullLineMatchForTUI(t *testing.T) {
	t.Run("TUI完整主机行匹配应删除该条目", func(t *testing.T) {
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

	t.Run("TUI完整行不匹配时应回退到主机部分精确匹配", func(t *testing.T) {
		input := []string{
			"github.com ssh-rsa key1",
			"gitlab.com ssh-rsa key2",
		}

		got := Delete(input, "github.com")
		want := []string{"gitlab.com ssh-rsa key2"}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("主机部分精确匹配失败: got %v, want %v", got, want)
		}
	})
}

func TestDelete_SecurityEdgeCases(t *testing.T) {
	t.Run("删除空字符串应保留所有主机", func(t *testing.T) {
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

	t.Run("删除包含特殊字符的主机名应正确处理", func(t *testing.T) {
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

	t.Run("删除空行输入应返回空切片", func(t *testing.T) {
		input := []string{"", ""}
		got := Delete(input, "anything")
		var want []string

		if !reflect.DeepEqual(got, want) {
			t.Errorf("空行输入应返回空切片: got %v, want %v", got, want)
		}
	})
}

func TestSearch_NoFuzzyMatchingOnDeletionPath(t *testing.T) {
	t.Run("搜索功能可使用模糊匹配但不应影响删除的精确性", func(t *testing.T) {
		input := []string{
			"github.com ssh-rsa key1",
			"gitlab.com ssh-rsa key2",
			"bitbucket.org ssh-rsa key3",
		}

		searchResults := Search(input, "git")
		if len(searchResults) != 2 {
			t.Errorf("搜索 'git' 应返回2个结果, got %d", len(searchResults))
		}

		afterDelete := Delete(input, "git")
		if len(afterDelete) != 3 {
			t.Errorf("删除 'git' 应保留所有3个主机, got %d", len(afterDelete))
		}
	})
}
