package main

import (
	"bytes"
	"encoding/gob"
	"strconv"
	"testing"
)

type reasonString string

type numberInterface interface {
	number() int
}

type numberImpl1 struct {
	N int
}

func (n *numberImpl1) number() int {
	return n.N
}

type numberImpl2 struct {
	N string
}

func (n *numberImpl2) number() int {
	o, _ := strconv.Atoi(n.N)
	return o
}

func init() {
	gob.Register(&numberImpl1{})
	gob.Register(&numberImpl2{})
}

type subobj struct {
	Foo int
	Bar int64
	Baz struct {
		Zab reasonString
	}
	Num numberInterface
}

type obj struct {
	Str   string
	KV    map[string]interface{}
	List  []string
	FList []float64
	Sub   *subobj
}

var o = obj{
	Str: "baz",
	KV: map[string]interface{}{
		"foo": "bar",
		"num": 123,
		"pct": 98.7623,
	},
	List:  []string{"a", "bb", "ccc"},
	FList: []float64{1.1, 2.2, 3.3, 4.4, 5.5},
	Sub: &subobj{
		Num: &numberImpl1{10},
	},
}

func BenchmarkEncode(b *testing.B) {
	b.Run("WithTypes", func(b *testing.B) {
		buf := new(bytes.Buffer)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			enc := gob.NewEncoder(buf)
			err := enc.Encode(o)
			if err != nil {
				panic(err)
			}
		}
	})
	b.Run("WithoutTypes", func(b *testing.B) {
		buf := new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		err := enc.Encode(o)
		if err != nil {
			panic(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := enc.Encode(o)
			if err != nil {
				panic(err)
			}
		}
	})
}

func BenchmarkDecode(b *testing.B) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(o)
	if err != nil {
		panic(err)
	}
	tmp := buf.Bytes()
	gobTyped := make([]byte, len(tmp))
	copy(gobTyped, tmp)

	buf.Reset()
	err = enc.Encode(o)
	if err != nil {
		panic(err)
	}
	gobUntyped := make([]byte, len(tmp))
	copy(gobUntyped, tmp)

	buf.Reset()
	enc = gob.NewEncoder(buf)
	p := o
	p.KV = nil
	p.Sub.Num = &numberImpl2{"12"}
	err = enc.Encode(p)
	if err != nil {
		panic(err)
	}
	gobTyped2 := make([]byte, len(tmp))
	copy(gobTyped2, tmp)

	b.Run("WithTypes", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var j obj
			buf := bytes.NewReader(gobTyped)
			enc := gob.NewDecoder(buf)
			err := enc.Decode(&j)
			if err != nil {
				panic(err)
			}
		}
	})
	b.Run("WithoutTypes", func(b *testing.B) {
		buf := bytes.NewReader(gobTyped)
		var j obj
		enc := gob.NewDecoder(buf)
		err := enc.Decode(&j)
		if err != nil {
			panic(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset(gobUntyped)
			err := enc.Decode(&j)
			if err != nil {
				panic(err)
			}
		}
	})
	b.Run("WithReuse", func(b *testing.B) {
		buf := bytes.NewReader(gobTyped)
		var j obj
		enc := gob.NewDecoder(buf)
		err := enc.Decode(&j)
		if err != nil {
			panic(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset(gobTyped2)
			err := enc.Decode(&j)
			if err != nil {
				panic(err)
			}
		}
	})
}
