package react

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

type JSX struct {
	pool *Pool
	opt  *Option
}

func NewJSX() (*JSX, error) {
	return NewJSXWithOption(DefaultJSXOption())
}

func NewJSXWithOption(opt *Option) (*JSX, error) {
	if opt == nil {
		return nil, errors.New("react: nil opt *Option")
	}
	err := opt.Validate()
	if err != nil {
		return nil, err
	}

	pool, err := NewPool(opt)
	if err != nil {
		return nil, err
	}

	return &JSX{pool: pool, opt: opt}, nil
}

func DefaultJSXOption() *Option {
	src, err := Asset("assets/JSXTransformer.js")
	if err != nil {
		panic(err)
	}
	return &Option{
		Source:           src,
		PoolSize:         10,
		GlobalObjectName: "self",
	}
}

func (jx *JSX) Transform(source []byte, opt map[string]interface{}) ([]byte, error) {
	optJSON, err := json.Marshal(opt)
	if err != nil {
		return nil, err
	}
	vm := jx.pool.Get()
	defer jx.pool.Put(vm)

	js := fmt.Sprintf(`(function(){$send(%v.JSXTransformer.transform(%#v, %v)['code']);})();`, jx.opt.GlobalObjectName, source, string(optJSON))
	err = vm.worker.Load("runscript.js", js)
	if err != nil {
		return nil, err
	}
	v := <-vm.ch
	return []byte(v), nil
}

func (jx *JSX) TransformFile(path string, opt map[string]interface{}) ([]byte, error) {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	optJSON, err := json.Marshal(opt)
	if err != nil {
		return nil, err
	}
	vm := jx.pool.Get()
	defer jx.pool.Put(vm)

	js := fmt.Sprintf(`(function(){$send(%v.JSXTransformer.transform(%#v, %v)['code']);})();`, jx.opt.GlobalObjectName, string(src), string(optJSON))
	err = vm.worker.Load("runscript.js", js)
	if err != nil {
		return nil, err
	}
	v := <-vm.ch
	return []byte(v), nil
}
