// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/miniblog. The professional
// version of this repository is https://github.com/onexstack/onex.

package main

import (
	"fmt"
	"strings"
)

// 定义日志级别
const (
	DebugLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// Logger 定义日志接口，包含常用日志方法.
type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

// customLogger 是 Logger 接口的具体实现.
type customLogger struct {
	level  int    // 日志级别
	prefix string // 日志前缀
}

// New 创建一个 customLogger 实例.
func New(level int, prefix string) Logger {
	return &customLogger{level: level, prefix: prefix}
}

// logMessage 是内部方法，根据日志级别打印日志.
func (l *customLogger) logMessage(level int, levelStr string, msg string) {
	if level >= l.level { // 只有当日志级别大于等于设定的级别时才打印
		fmt.Printf("[%s] %s: %s\n", strings.ToUpper(levelStr), l.prefix, msg)
	}
}

// Debug 实现 Logger 接口的 Debug 方法.
func (l *customLogger) Debug(msg string) {
	l.logMessage(DebugLevel, "debug", msg)
}

// Info 实现 Logger 接口的 Info 方法.
func (l *customLogger) Info(msg string) {
	l.logMessage(InfoLevel, "info", msg)
}

// Warn 实现 Logger 接口的 Warn 方法.
func (l *customLogger) Warn(msg string) {
	l.logMessage(WarnLevel, "warn", msg)
}

// Error 实现 Logger 接口的 Error 方法.
func (l *customLogger) Error(msg string) {
	l.logMessage(ErrorLevel, "error", msg)
}

func main() {
	// 创建一个日志实例，日志级别为 Info
	logger := New(InfoLevel, "MyApp")

	// 根据不同级别打印日志
	logger.Debug("This is a debug message") // 不会打印（级别低于 Info）
	logger.Info("This is an info message")  // 会打印
}
