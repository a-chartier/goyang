// Copyright 2015 Google Inc.
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

package yang

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	useMin = -999
	useMax = 999
)

func R(a, b int64) YRange {
	n1 := FromInt(a)
	n2 := FromInt(b)
	if a == useMin {
		n1 = minNumber
	}
	if b == useMax {
		n2 = maxNumber
	}
	return YRange{n1, n2}
}

func TestRangeEqual(t *testing.T) {
	for x, tt := range []struct {
		r1, r2 YangRange
		ok     bool
	}{
		{ok: true},                          // empty range contained in empty range
		{r1: YangRange{R(1, 2)}, ok: false}, // empty range contained in range
		{r2: YangRange{R(1, 2)}, ok: false}, // range contained in empty range
		{
			YangRange{R(1, 2)},
			YangRange{R(1, 2)},
			true,
		},
		{
			YangRange{R(1, 3)},
			YangRange{R(1, 2)},
			false,
		},
		{
			YangRange{R(1, 2), R(4, 5)},
			YangRange{R(1, 2), R(4, 5)},
			true,
		},
		{
			YangRange{R(1, 2), R(4, 6)},
			YangRange{R(1, 2), R(4, 5)},
			false,
		},
		{
			YangRange{R(1, 2)},
			YangRange{R(1, 2), R(4, 5)},
			false,
		},
		{
			YangRange{R(1, 2), R(4, 5)},
			YangRange{R(1, 2)},
			false,
		},
	} {
		if ok := tt.r1.Equal(tt.r2); ok != tt.ok {
			t.Errorf("#%d: got %v, want %v", x, ok, tt.ok)
		}
	}
}

func TestRangeContains(t *testing.T) {
	for x, tt := range []struct {
		r1, r2 YangRange
		ok     bool
	}{
		{ok: true},
		{r1: YangRange{R(1, 2)}, ok: true},
		{r2: YangRange{R(1, 2)}, ok: true},
		{
			r1: YangRange{R(1, 2)},
			r2: YangRange{R(1, 2)},
			ok: true,
		},
		{
			r1: YangRange{R(1, 5)},
			r2: YangRange{R(2, 3)},
			ok: true,
		},
		{
			r1: YangRange{R(2, 3)},
			r2: YangRange{R(1, 5)},
			ok: false,
		},
		{
			r1: YangRange{R(1, 10)},
			r2: YangRange{R(1, 2), R(4, 5), R(7, 10)},
			ok: true,
		},
		{
			r1: YangRange{R(1, 10)},
			r2: YangRange{R(1, 2), R(7, 11)},
			ok: false,
		},
		{
			r1: YangRange{R(1, 9), R(11, 19), R(21, 29)},
			r2: YangRange{R(23, 25)},
			ok: true,
		},
		{
			r1: YangRange{R(1, 9), R(11, 19), R(21, 29)},
			r2: YangRange{R(23, 23)},
			ok: true,
		},
		{
			r1: YangRange{R(1, 9), R(11, 19), R(21, 29)},
			r2: YangRange{R(20, 20)},
			ok: false,
		},
		{
			r1: YangRange{R(1, 10)},
			r2: YangRange{R(useMin, useMax)},
			ok: true,
		},
		{
			r1: YangRange{R(useMin, useMax)},
			r2: YangRange{R(1, 10)},
			ok: true,
		},
		{
			r1: YangRange{R(1024, 65535)},
			r2: YangRange{R(useMin, 4096), R(5120, useMax)},
			ok: true,
		},
		{
			r1: YangRange{R(1024, 65535)},
			r2: YangRange{R(-999999, 4096), R(5120, useMax)},
			ok: false,
		},
		{
			r1: YangRange{R(1024, 65535)},
			r2: YangRange{R(useMin, 4096), R(5120, 999999)},
			ok: false,
		},
	} {
		if ok := tt.r1.Contains(tt.r2); ok != tt.ok {
			t.Errorf("#%d: got %v, want %v", x, ok, tt.ok)
		}
	}
}

func TestCoalesce(t *testing.T) {
	for x, tt := range []struct {
		in, out YangRange
	}{
		{},
		{YangRange{R(1, 4)}, YangRange{R(1, 4)}},
		{YangRange{R(1, 2), R(3, 4)}, YangRange{R(1, 4)}},
		{YangRange{R(1, 2), R(2, 4)}, YangRange{R(1, 4)}},
		{YangRange{R(1, 2), R(4, 5)}, YangRange{R(1, 2), R(4, 5)}},
		{YangRange{R(1, 3), R(2, 5)}, YangRange{R(1, 5)}},
		{YangRange{R(1, 10), R(2, 5)}, YangRange{R(1, 10)}},
		{YangRange{R(1, 10), R(1, 2), R(4, 5), R(7, 8)}, YangRange{R(1, 10)}},
	} {
		out := coalesce(tt.in)
		if !out.Equal(tt.out) {
			t.Errorf("#%d: got %v, want %v", x, out, tt.out)
		}
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		desc  string
		inVal Number
		inAdd uint64
		want  Number
	}{{
		desc:  "add one to integer",
		inVal: FromInt(1),
		inAdd: 1,
		want:  FromInt(2),
	}, {
		desc:  "add one to decimal64",
		inVal: FromFloat(1.0),
		inAdd: 1,
		want:  FromFloat(2.0),
	}, {
		desc:  "negative int becomes positive",
		inVal: FromInt(-2),
		inAdd: 3,
		want:  FromInt(1),
	}, {
		desc:  "negative int stays negative",
		inVal: FromInt(-3),
		inAdd: 1,
		want:  FromInt(-2),
	}, {
		desc:  "negative decimal becomes positive",
		inVal: FromFloat(-2),
		inAdd: 3,
		want:  FromFloat(1.0),
	}, {
		desc:  "negative decimal stays negative",
		inVal: FromFloat(-42),
		inAdd: 41,
		want:  FromFloat(-1.0),
	}, {
		desc:  "explicitly set fraction digits",
		inVal: Number{Value: 10000, FractionDigits: 5},
		inAdd: 1,
		want:  Number{Value: 110000, FractionDigits: 5},
	}, {
		desc:  "explicitly set fraction digits - negative",
		inVal: Number{Value: 0, FractionDigits: 10},
		inAdd: 42,
		want:  FromFloat(42),
	}}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := tt.inVal.add(tt.inAdd)
			if !cmp.Equal(got, tt.want) {
				t.Fatalf("did get expected result, got: %s, want: %s", got.String(), tt.want.String())
			}
		})
	}
}
