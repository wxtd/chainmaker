/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package convert

import (
	"errors"
	"unicode"
)

const BASE = 48

func Int32ToString(number int32) string {
	var negativeFlag bool
	var convertNumber int32
	convertNumber = number
	if number < 0 {
		negativeFlag = true
		convertNumber = -1 * number
	}

	var res string
	for convertNumber != 0 {
		res = getChar(convertNumber%10) + res
		convertNumber = convertNumber / 10
	}

	if negativeFlag {
		res = "-" + res
	}
	return res
}

func Int64ToString(number int64) string {
	var negativeFlag bool
	var convertNumber int64
	convertNumber = number
	if number < 0 {
		negativeFlag = true
		convertNumber = -1 * number
	}

	var res string
	for convertNumber != 0 {
		res = getChar(int32(convertNumber%10)) + res
		convertNumber = convertNumber / 10
	}

	if negativeFlag {
		res = "-" + res
	}
	return res
}

func StringToInt32(str string) (int32, error) { //no verify isNumber
	var negativeFlag bool
	if str[0] == '-' {
		negativeFlag = true
		str = str[1:]
	}

	var num int32 = 0
	for _, v := range str {
		if !unicode.IsNumber(v) {
			return 0, errors.New("check str is number ")
		}

		addResult := num*10 + getNumber(v)
		if addResult < num {
			return 0, errors.New("too big")
		}
		num = addResult
	}

	if negativeFlag {
		num *= -1
	}

	return num, nil
}

func StringToInt64(str string) (int64, error) { //no verify isNumber
	var negativeFlag bool
	if str[0] == '-' {
		negativeFlag = true
		str = str[1:]
	}

	var num int64 = 0
	for _, v := range str {
		if !unicode.IsNumber(v) {
			return 0, errors.New("check str is number ")
		}

		addResult := num*10 + int64(getNumber(v))
		if addResult < num {
			return 0, errors.New("too big")
		}
		num = addResult
	}

	if negativeFlag {
		num *= -1
	}

	return num, nil
}


func getChar(number int32) string {
	return string(number + BASE)
}

func getNumber(char int32) int32 {
	return char - BASE
}
