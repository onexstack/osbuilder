// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/miniblog. The professional
// version of this repository is https://github.com/onexstack/onex.

package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {
	log.Printf("Listen at port: 6060")
	go func() {
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()
	for {
		_ = fmt.Sprint("test sprint")
		time.Sleep(time.Millisecond)
	}
}
