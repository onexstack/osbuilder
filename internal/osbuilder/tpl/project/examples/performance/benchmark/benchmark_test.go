// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/miniblog. The professional
// version of this repository is https://github.com/onexstack/onex.

package benchmark

import (
	"math"
	"testing"
)

func BenchmarkExp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = math.Exp(3.5)
	}
}
