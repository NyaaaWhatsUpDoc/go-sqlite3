package util

import (
	"context"
	"math/bits"
	"unsafe"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type Funcs struct {
	fn    [32]api.Function
	id    [32]*byte
	mask  uint32
	stack [8]uint64
}

func (f *Funcs) getfn(mod api.Module, name string) api.Function {
	p := unsafe.StringData(name)
	for i := range f.id {
		if f.id[i] == p {
			f.id[i] = nil
			f.mask &^= uint32(1) << i
			return f.fn[i]
		}
	}
	return mod.ExportedFunction(name)
}

func (f *Funcs) putfn(name string, fn api.Function) {
	p := unsafe.StringData(name)
	i := bits.TrailingZeros32(^f.mask)
	if i < 32 {
		f.id[i] = p
		f.fn[i] = fn
		f.mask |= uint32(1) << i
	} else {
		f.id[0] = p
		f.fn[0] = fn
		f.mask = uint32(1)
	}
}

func (f *Funcs) Call(ctx context.Context, mod api.Module, name string, params ...uint64) uint64 {
	copy(f.stack[:], params)
	fn := f.getfn(mod, name)
	err := fn.CallWithStack(ctx, f.stack[:])
	if err != nil {
		panic(err)
	}
	f.putfn(name, fn)
	return f.stack[0]
}

type i32 interface{ ~int32 | ~uint32 }
type i64 interface{ ~int64 | ~uint64 }

type funcVI[T0 i32] func(context.Context, api.Module, T0)

func (fn funcVI[T0]) Call(ctx context.Context, mod api.Module, stack []uint64) {
	fn(ctx, mod, T0(stack[0]))
}

func ExportFuncVI[T0 i32](mod wazero.HostModuleBuilder, name string, fn func(context.Context, api.Module, T0)) {
	mod.NewFunctionBuilder().
		WithGoModuleFunction(funcVI[T0](fn),
			[]api.ValueType{api.ValueTypeI32}, nil).
		Export(name)
}

type funcVII[T0, T1 i32] func(context.Context, api.Module, T0, T1)

func (fn funcVII[T0, T1]) Call(ctx context.Context, mod api.Module, stack []uint64) {
	fn(ctx, mod, T0(stack[0]), T1(stack[1]))
}

func ExportFuncVII[T0, T1 i32](mod wazero.HostModuleBuilder, name string, fn func(context.Context, api.Module, T0, T1)) {
	mod.NewFunctionBuilder().
		WithGoModuleFunction(funcVII[T0, T1](fn),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, nil).
		Export(name)
}

type funcVIII[T0, T1, T2 i32] func(context.Context, api.Module, T0, T1, T2)

func (fn funcVIII[T0, T1, T2]) Call(ctx context.Context, mod api.Module, stack []uint64) {
	fn(ctx, mod, T0(stack[0]), T1(stack[1]), T2(stack[2]))
}

func ExportFuncVIII[T0, T1, T2 i32](mod wazero.HostModuleBuilder, name string, fn func(context.Context, api.Module, T0, T1, T2)) {
	mod.NewFunctionBuilder().
		WithGoModuleFunction(funcVIII[T0, T1, T2](fn),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, nil).
		Export(name)
}

type funcVIIII[T0, T1, T2, T3 i32] func(context.Context, api.Module, T0, T1, T2, T3)

func (fn funcVIIII[T0, T1, T2, T3]) Call(ctx context.Context, mod api.Module, stack []uint64) {
	fn(ctx, mod, T0(stack[0]), T1(stack[1]), T2(stack[2]), T3(stack[3]))
}

func ExportFuncVIIII[T0, T1, T2, T3 i32](mod wazero.HostModuleBuilder, name string, fn func(context.Context, api.Module, T0, T1, T2, T3)) {
	mod.NewFunctionBuilder().
		WithGoModuleFunction(funcVIIII[T0, T1, T2, T3](fn),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, nil).
		Export(name)
}

type funcVIIIII[T0, T1, T2, T3, T4 i32] func(context.Context, api.Module, T0, T1, T2, T3, T4)

func (fn funcVIIIII[T0, T1, T2, T3, T4]) Call(ctx context.Context, mod api.Module, stack []uint64) {
	fn(ctx, mod, T0(stack[0]), T1(stack[1]), T2(stack[2]), T3(stack[3]), T4(stack[4]))
}

func ExportFuncVIIIII[T0, T1, T2, T3, T4 i32](mod wazero.HostModuleBuilder, name string, fn func(context.Context, api.Module, T0, T1, T2, T3, T4)) {
	mod.NewFunctionBuilder().
		WithGoModuleFunction(funcVIIIII[T0, T1, T2, T3, T4](fn),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, nil).
		Export(name)
}

type funcVIIIIJ[T0, T1, T2, T3 i32, T4 i64] func(context.Context, api.Module, T0, T1, T2, T3, T4)

func (fn funcVIIIIJ[T0, T1, T2, T3, T4]) Call(ctx context.Context, mod api.Module, stack []uint64) {
	fn(ctx, mod, T0(stack[0]), T1(stack[1]), T2(stack[2]), T3(stack[3]), T4(stack[4]))
}

func ExportFuncVIIIIJ[T0, T1, T2, T3 i32, T4 i64](mod wazero.HostModuleBuilder, name string, fn func(context.Context, api.Module, T0, T1, T2, T3, T4)) {
	mod.NewFunctionBuilder().
		WithGoModuleFunction(funcVIIIIJ[T0, T1, T2, T3, T4](fn),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI64}, nil).
		Export(name)
}

type funcII[TR, T0 i32] func(context.Context, api.Module, T0) TR

func (fn funcII[TR, T0]) Call(ctx context.Context, mod api.Module, stack []uint64) {
	stack[0] = uint64(fn(ctx, mod, T0(stack[0])))
}

func ExportFuncII[TR, T0 i32](mod wazero.HostModuleBuilder, name string, fn func(context.Context, api.Module, T0) TR) {
	mod.NewFunctionBuilder().
		WithGoModuleFunction(funcII[TR, T0](fn),
			[]api.ValueType{api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export(name)
}

type funcIII[TR, T0, T1 i32] func(context.Context, api.Module, T0, T1) TR

func (fn funcIII[TR, T0, T1]) Call(ctx context.Context, mod api.Module, stack []uint64) {
	stack[0] = uint64(fn(ctx, mod, T0(stack[0]), T1(stack[1])))
}

func ExportFuncIII[TR, T0, T1 i32](mod wazero.HostModuleBuilder, name string, fn func(context.Context, api.Module, T0, T1) TR) {
	mod.NewFunctionBuilder().
		WithGoModuleFunction(funcIII[TR, T0, T1](fn),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export(name)
}

type funcIIII[TR, T0, T1, T2 i32] func(context.Context, api.Module, T0, T1, T2) TR

func (fn funcIIII[TR, T0, T1, T2]) Call(ctx context.Context, mod api.Module, stack []uint64) {
	stack[0] = uint64(fn(ctx, mod, T0(stack[0]), T1(stack[1]), T2(stack[2])))
}

func ExportFuncIIII[TR, T0, T1, T2 i32](mod wazero.HostModuleBuilder, name string, fn func(context.Context, api.Module, T0, T1, T2) TR) {
	mod.NewFunctionBuilder().
		WithGoModuleFunction(funcIIII[TR, T0, T1, T2](fn),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export(name)
}

type funcIIIII[TR, T0, T1, T2, T3 i32] func(context.Context, api.Module, T0, T1, T2, T3) TR

func (fn funcIIIII[TR, T0, T1, T2, T3]) Call(ctx context.Context, mod api.Module, stack []uint64) {
	stack[0] = uint64(fn(ctx, mod, T0(stack[0]), T1(stack[1]), T2(stack[2]), T3(stack[3])))
}

func ExportFuncIIIII[TR, T0, T1, T2, T3 i32](mod wazero.HostModuleBuilder, name string, fn func(context.Context, api.Module, T0, T1, T2, T3) TR) {
	mod.NewFunctionBuilder().
		WithGoModuleFunction(funcIIIII[TR, T0, T1, T2, T3](fn),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export(name)
}

type funcIIIIII[TR, T0, T1, T2, T3, T4 i32] func(context.Context, api.Module, T0, T1, T2, T3, T4) TR

func (fn funcIIIIII[TR, T0, T1, T2, T3, T4]) Call(ctx context.Context, mod api.Module, stack []uint64) {
	stack[0] = uint64(fn(ctx, mod, T0(stack[0]), T1(stack[1]), T2(stack[2]), T3(stack[3]), T4(stack[4])))
}

func ExportFuncIIIIII[TR, T0, T1, T2, T3, T4 i32](mod wazero.HostModuleBuilder, name string, fn func(context.Context, api.Module, T0, T1, T2, T3, T4) TR) {
	mod.NewFunctionBuilder().
		WithGoModuleFunction(funcIIIIII[TR, T0, T1, T2, T3, T4](fn),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export(name)
}

type funcIIIIIII[TR, T0, T1, T2, T3, T4, T5 i32] func(context.Context, api.Module, T0, T1, T2, T3, T4, T5) TR

func (fn funcIIIIIII[TR, T0, T1, T2, T3, T4, T5]) Call(ctx context.Context, mod api.Module, stack []uint64) {
	stack[0] = uint64(fn(ctx, mod, T0(stack[0]), T1(stack[1]), T2(stack[2]), T3(stack[3]), T4(stack[4]), T5(stack[5])))
}

func ExportFuncIIIIIII[TR, T0, T1, T2, T3, T4, T5 i32](mod wazero.HostModuleBuilder, name string, fn func(context.Context, api.Module, T0, T1, T2, T3, T4, T5) TR) {
	mod.NewFunctionBuilder().
		WithGoModuleFunction(funcIIIIIII[TR, T0, T1, T2, T3, T4, T5](fn),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32}).
		Export(name)
}

type funcIIIIJ[TR, T0, T1, T2 i32, T3 i64] func(context.Context, api.Module, T0, T1, T2, T3) TR

func (fn funcIIIIJ[TR, T0, T1, T2, T3]) Call(ctx context.Context, mod api.Module, stack []uint64) {
	stack[0] = uint64(fn(ctx, mod, T0(stack[0]), T1(stack[1]), T2(stack[2]), T3(stack[3])))
}

func ExportFuncIIIIJ[TR, T0, T1, T2 i32, T3 i64](mod wazero.HostModuleBuilder, name string, fn func(context.Context, api.Module, T0, T1, T2, T3) TR) {
	mod.NewFunctionBuilder().
		WithGoModuleFunction(funcIIIIJ[TR, T0, T1, T2, T3](fn),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI64}, []api.ValueType{api.ValueTypeI32}).
		Export(name)
}

type funcIIJ[TR, T0 i32, T1 i64] func(context.Context, api.Module, T0, T1) TR

func (fn funcIIJ[TR, T0, T1]) Call(ctx context.Context, mod api.Module, stack []uint64) {
	stack[0] = uint64(fn(ctx, mod, T0(stack[0]), T1(stack[1])))
}

func ExportFuncIIJ[TR, T0 i32, T1 i64](mod wazero.HostModuleBuilder, name string, fn func(context.Context, api.Module, T0, T1) TR) {
	mod.NewFunctionBuilder().
		WithGoModuleFunction(funcIIJ[TR, T0, T1](fn),
			[]api.ValueType{api.ValueTypeI32, api.ValueTypeI64}, []api.ValueType{api.ValueTypeI32}).
		Export(name)
}
