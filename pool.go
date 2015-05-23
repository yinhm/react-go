package react

import (
	"errors"
	"log"
	"math"

	"github.com/yinhm/v8worker"
)

type Pool struct {
	opt  *Option
	size int
	ch   chan *vm
}

type Option struct {
	Source []byte
	// size for javascript vm pool.
	PoolSize int
	// name for variable includes component objects. ex. "self"
	GlobalObjectName string

	MaxRender int
}

func (opt *Option) Validate() error {
	if opt.Source == nil {
		return errors.New("react: nil []byte opt.Source")
	}
	if opt.PoolSize <= 0 {
		return errors.New("react: opt.PoolSize must be greater than or equal to 1")
	}
	if opt.GlobalObjectName == "" {
		return errors.New("react: empty string opt.GlobalObjectName")
	}
	return nil
}

func NewPool(opt *Option) (*Pool, error) {
	if opt.MaxRender == 0 {
		opt.MaxRender = math.MaxInt32
	}

	pool := &Pool{
		opt:  opt,
		size: opt.PoolSize,
	}
	pool.ch = make(chan *vm, opt.PoolSize)
	for i := 0; i < pool.size; i++ {
		vm, err := newVM(opt.Source, opt.GlobalObjectName)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		pool.ch <- vm
	}
	return pool, nil
}

func (pl *Pool) Get() *vm {
	return <-pl.ch
}

func (pl *Pool) Put(vm *vm) {
	select {
	case <-vm.ch:
	default:
	}

	vm.renderCount += 1
	if vm.renderCount > pl.opt.MaxRender {
		vm2, err := newVM(pl.opt.Source, pl.opt.GlobalObjectName)
		if err == nil {
			vm = vm2
		}
	}

	pl.ch <- vm
}

type vm struct {
	worker *v8worker.Worker
	ch     chan string

	// die after max render
	renderCount int
}

func (v *vm) callback(msg string) {
	v.ch <- msg
}

func newVM(src []byte, objName string) (*vm, error) {
	vm := &vm{
		ch: make(chan string, 1),
	}
	worker := v8worker.New(vm.callback)

	// tpl := "var %v = %v || {};\n"
	tpl := "var self = {};\n"
	// source := fmt.Sprintf(tpl, objName, objName) + string(src)
	source := tpl + string(src)
	// log.Println(source)
	err := worker.Load("reactbundle.js", source)
	if err != nil {
		return nil, err
	}

	vm.worker = worker
	return vm, nil
}
