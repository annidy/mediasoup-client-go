package util

import (
	"encoding/json"
	"reflect"

	"github.com/imdario/mergo"
	"github.com/pion/randutil"
)

const (
	runesAlpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	runesNum   = "123456789"
)

// Use global random generator to properly seed by crypto grade random.
var globalMathRandomGenerator = randutil.NewMathRandomGenerator() // nolint:gochecknoglobals

// RandomAlpha generates a mathematical random alphabet sequence of the requested length.
func RandomAlpha(n int) string {
	return globalMathRandomGenerator.GenerateString(n, runesAlpha)
}

func RandomNumber() uint32 {
	return globalMathRandomGenerator.Uint32()
}

func Clone(from, to interface{}) (err error) {
	data, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, to)
}

func Override(dst, src interface{}) error {
	return mergo.Merge(dst, src,
		mergo.WithOverride,
		mergo.WithTypeCheck,
		mergo.WithTransformers(ptrTransformers{}),
	)
}

type ptrTransformers struct{}

// overwrites pointer type
func (ptrTransformers) Transformer(tp reflect.Type) func(dst, src reflect.Value) error {
	if tp.Kind() == reflect.Ptr {
		return func(dst, src reflect.Value) error {
			if !src.IsNil() {
				if dst.CanSet() {
					dst.Set(src)
				} else {
					dst = src
				}
			}
			return nil
		}
	}
	return nil
}
