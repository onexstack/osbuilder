// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package util

import clioptions "github.com/onexstack/osbuilder/internal/osbuilder/util/options"

type Factory interface {
	GetOptions() *clioptions.Options
}
