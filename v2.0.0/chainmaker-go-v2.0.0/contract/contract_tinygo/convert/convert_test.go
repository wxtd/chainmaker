/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package convert

import "testing"

func Test_getChar(t *testing.T) {
	type args struct {
		number int32
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "t1",
			args: args{
				number: 1,
			},
			want: "1",
		},
		{
			name: "t2",
			args: args{
				number: 2,
			},
			want: "2",
		},
		{
			name: "t3",
			args: args{
				number: 3,
			},
			want: "3",
		},
		{
			name: "t4",
			args: args{
				number: 4,
			},
			want: "4",
		},
		{
			name: "t5",
			args: args{
				number: 5,
			},
			want: "5",
		},
		{
			name: "t0",
			args: args{
				number: 0,
			},
			want: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getChar(tt.args.number); got != tt.want {
				t.Errorf("getChar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getNumber(t *testing.T) {
	type args struct {
		char int32
	}
	tests := []struct {
		name string
		args args
		want int32
	}{

		{
			name: "t0",
			args: args{
				char: 48,
			},
			want: 0,
		},
		{
			name: "t1",
			args: args{
				char: 49,
			},
			want: 1,
		},
		{
			name: "t2",
			args: args{
				char: 50,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNumber(tt.args.char); got != tt.want {
				t.Errorf("getNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertInt64ToString(t *testing.T) {
	type args struct {
		number int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "t1",
			args: args{
				number: 89328938293892,
			},
			want: "89328938293892",
		},
		{
			name: "t2",
			args: args{
				number: -89328938293892,
			},
			want: "-89328938293892",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Int64ToString(tt.args.number); got != tt.want {
				t.Errorf("Int64ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertIntToString(t *testing.T) {
	type args struct {
		number int32
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "t1",
			args: args{
				number: 1223332323,
			},
			want: "1223332323",
		},
		{
			name: "t2",
			args: args{
				number: -1223332323,
			},
			want: "-1223332323",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Int32ToString(tt.args.number); got != tt.want {
				t.Errorf("Int32ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertStringToInt(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want int32
	}{
		{
			name: "t1",
			args: args{
				str: "1630662366",
			},
			want: 1630662366,
		},
		{
			name: "t2",
			args: args{
				str: "1223124314",
			},
			want: 1223124314,
		},
		{
			name: "t3",
			args: args{
				str: "-1223124314",
			},
			want: -1223124314,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := StringToInt32(tt.args.str); got != tt.want {
				t.Errorf("StringToInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertStringToInt64(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "t1",
			args: args{
				str: "172834342343",
			},
			want: 172834342343,
		},
		{
			name: "t2",
			args: args{
				str: "-172834342343",
			},
			want: -172834342343,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := StringToInt64(tt.args.str); got != tt.want {
				if err != nil {
					t.Error(err)
				}
				t.Errorf("StringToInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

