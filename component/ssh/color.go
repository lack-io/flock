// Copyright 2018 xingyys, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ssh

import "fmt"

// 在终端上输出带颜色的文本
//  echo -e "\033[背景;字体颜色;特殊效果m 字符串 \033[0m"

// 文本颜色
const (
	Blank   = 30
	Red     = 31
	Green   = 32
	Yellow  = 33
	Bblue   = 34
	Purple  = 35
	SkyBlue = 36
	White   = 37
)

// 文本背景颜色
const (
	BBlank   = 40
	BRed     = 41
	BGreen   = 42
	BYellow  = 43
	BBlue    = 44
	BPurple  = 45
	BSkyBlue = 46
	BWhite   = 47
)

// 特殊效果
const (
	// 关闭所有属性
	FDefault = 0

	// 高亮
	FHighLight = 1

	// 下划线
	FUnderline = 4

	// 闪烁
	FFlash = 4

	// 反显
	FReverse = 7

	// 消隐
	FBlanking = 8
)

// 使文本在终端带颜色输出，只在终端有效
// @msg : 输出文本
// @fcolor : 文本颜色
// @args: 只有1，2参数有效，分别是背景色和特殊效果
func Color(msg interface{}, fcolor int, params ... int) string {
	if len(params) == 0 {
		return fmt.Sprintf("\033[%dm%s\033[0m", fcolor, msg)
	} else if len(params) == 1 {
		return fmt.Sprintf("\033[%d;%dm%s\033[0m", fcolor, params[0], msg)
	}
	return fmt.Sprintf("\033[%d;%d;%dm%s\033[0m", fcolor, params[0], params[1], msg)
}

func SuccessColor(msg interface{}) string {
	return Color(msg, Green)
}

func ChangeColor(msg interface{}) string {
	return Color(msg, Yellow)
}

func WarnColor(msg interface{}) string {
	return Color(msg, Purple)
}

func FailColor(msg interface{}) string {
	return Color(msg, Red)
}
