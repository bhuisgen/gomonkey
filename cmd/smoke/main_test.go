package main

import (
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/bhuisgen/gomonkey"
)

func TestMain(m *testing.M) {
	gomonkey.Init()
	code := m.Run()
	gomonkey.ShutDown()
	os.Exit(code)
}

func TestSmoke(t *testing.T) {
	smoke(1)
}

func TestSmokeFull(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	smoke(1000)
}

func BenchmarkExecute1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkExecute(b, 1)
	}
}

func BenchmarkExecute100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkExecute(b, 100)
	}
}

func benchmarkExecute(b *testing.B, executions int) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		b.Fatal(err)
	}
	defer ctx.Destroy()

	script, err := ctx.CompileScript("script.js", []byte("(() => { return 1 + 2; })()"))
	if err != nil {
		b.Fatal(err)
	}
	defer script.Release()

	for i := 0; i < executions; i++ {
		result, err := ctx.ExecuteScript(script)
		if err != nil {
			b.Fatal(err)
		}
		defer result.Release()
		if !result.IsInt32() || result.ToInt32() != 3 {
			b.Fatal(errors.New("invalid result"))
		}
	}
}

func BenchmarkExecuteWithStencil1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkExecuteWithStencil(b, 1)
	}
}
func BenchmarkExecuteWithStencil100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkExecuteWithStencil(b, 100)
	}
}

func benchmarkExecuteWithStencil(b *testing.B, executions int) {
	var wg sync.WaitGroup
	ch := make(chan *gomonkey.Stencil, executions+1)

	wg.Add(1)
	go func(ch chan<- *gomonkey.Stencil) {
		fc, err := gomonkey.NewFrontendContext()
		if err != nil {
			panic(err)
		}
		defer func() {
			wg.Done()
		}()
		stencil, err := fc.CompileScriptToStencil("script.js", []byte("(() => { return 1 + 2; })()"))
		if err != nil {
			panic(err)
		}
		for i := 0; i < executions+1; i++ {
			ch <- stencil
		}
		close(ch)
	}(ch)

	wg.Add(1)
	go func(ch <-chan *gomonkey.Stencil) {
		ctx, err := gomonkey.NewContext()
		if err != nil {
			panic(err)
		}
		defer func() {
			ctx.Destroy()
			wg.Done()
		}()

		for i := 0; i < executions; i++ {
			stencil := <-ch

			result, err := ctx.ExecuteScriptFromStencil(stencil)
			if err != nil {
				return
			}
			defer result.Release()
			if !result.IsInt32() || result.ToInt32() != 3 {
				panic(errors.New("invalid result"))
			}
		}
	}(ch)

	wg.Wait()

	wg.Add(1)
	go func(ch <-chan *gomonkey.Stencil) {
		stencil := <-ch
		stencil.Release()
		wg.Done()
	}(ch)

	wg.Wait()

	_ = b
}
