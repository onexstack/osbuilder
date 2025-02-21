package validation

import (
	"testing"
)

// 性能测试用例
func BenchmarkIsValidUsername(b *testing.B) {
	// 定义测试的用户名集合，包括合法和非法的输入
	testUsernames := []string{
		"valid_user123",         // 合法，常规输入
		"user_too_long_example", // 长度超过 20
		"sh",                    // 长度不足 3
		"in*valid",              // 包含非法字符
		"12345678901234567890",  // 合法，刚好 20 个字符
	}

	// 重置计时器
	b.ResetTimer()

	// 性能基准测试
	for i := 0; i < b.N; i++ {
		// 模拟不同的测试用例
		for _, username := range testUsernames {
			isValidUsername(username)
		}
	}
}
