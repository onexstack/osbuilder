// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

// this file contains factories with no other dependencies

package util

import (
	"k8s.io/klog/v2"

	clioptions "github.com/onexstack/osbuilder/internal/osbuilder/util/options"
)

type factoryImpl struct {
	opts *clioptions.Options
}

var _ Factory = (*factoryImpl)(nil)

func NewFactory(opts *clioptions.Options) Factory {
	if opts == nil {
		klog.Fatal("attempt to instantiate client_access_factory with nil clientGetter")
	}

	return &factoryImpl{opts: opts}
}

func (f *factoryImpl) GetOptions() *clioptions.Options {
	return f.opts
}
