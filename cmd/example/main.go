// This command executes the code examples in the README file.
package main

import (
	"errors"
	"os"
	"runtime"
	"sync"

	"github.com/bhuisgen/gomonkey"
)

func main() {
	gomonkey.Init()

	_ = contexts()
	_ = objects()
	_ = evaluate()
	_ = executeScript()
	_ = executeScriptFromStencil()
	_ = createGlobalFunction()
	_ = createFunctionObject()
	_ = createObjectMethod()
}

func contexts() error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		// lock the goroutine to its own OS thread
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		// create the context
		ctx, err := gomonkey.NewContext()
		if err != nil {
			return
		}
		defer ctx.Destroy() // destroy after usage

		// use the context only inside this goroutine
	}()

	wg.Wait()

	return nil
}

func objects() error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		runtime.LockOSThread()
		defer func() {
			runtime.UnlockOSThread()
			wg.Done()
		}()

		ctx, err := gomonkey.NewContext()
		if err != nil {
			return
		}
		defer ctx.Destroy()

		// get the global object ...

		global, err := ctx.Global()
		if err != nil {
			return
		}
		defer global.Release() // release after usage

		// ... or create a new object ...

		object, err := gomonkey.NewObject(ctx)
		if err != nil {
			return
		}
		defer object.Release() // release after usage

		// ... add a property ...

		propValue1, err := gomonkey.NewValueString(ctx, "string")
		if err != nil {
			return
		}
		defer propValue1.Release() // release after usage
		if err := object.Set("key1", propValue1); err != nil {
			return
		}

		// ... check if a property exists ...

		b := object.Has("key2")
		_ = b // use result

		// ... get a property value ...

		propValue3, err := object.Get("key3")
		if err != nil {
			return
		}
		defer propValue3.Release() // release after usage

		// ... delete a property ...

		if err := object.Delete("key4"); err != nil {
			return
		}
	}()

	wg.Wait()

	return nil
}

func evaluate() error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		runtime.LockOSThread()
		defer func() {
			runtime.UnlockOSThread()
			wg.Done()
		}()

		ctx, err := gomonkey.NewContext()
		if err != nil {
			return
		}
		defer ctx.Destroy()

		result, err := ctx.Evaluate([]byte("(() => { return 'result'; })()"))
		if err != nil {
			return
		}
		defer result.Release() // release after usage

		if !result.IsString() { // check value type
			return
		}
		_ = result.String() // use the value
	}()

	wg.Wait()

	return nil
}

func executeScript() error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		runtime.LockOSThread()
		defer func() {
			runtime.UnlockOSThread()
			wg.Done()
		}()

		ctx, err := gomonkey.NewContext()
		if err != nil {
			return
		}
		defer ctx.Destroy()

		// load your script ...

		code, err := os.ReadFile("script.js")
		if err != nil {
			return
		}

		// compile the script ...

		script, err := ctx.CompileScript("script.js", code)
		if err != nil {
			return
		}
		defer script.Release() // release after usage

		// .. and execute it ...

		result1, err := ctx.ExecuteScript(script)
		if err != nil {
			return
		}
		defer result1.Release() // release after usage

		if !result1.IsString() { // check value type
			return
		}
		_ = result1.String() // use the value

		// ... one more time please ...

		result2, err := ctx.ExecuteScript(script)
		if err != nil {
			return
		}
		defer result2.Release() // release after usage

		if !result2.IsString() { // check value type
			return
		}
		_ = result2.String() // use the value
	}()

	wg.Wait()

	return nil
}

func executeScriptFromStencil() error {
	var wg sync.WaitGroup

	ch := make(chan *gomonkey.Stencil)

	// compile script as a stencil in a frontend context ...

	wg.Add(1)
	go func(ch chan<- *gomonkey.Stencil) {
		runtime.LockOSThread()
		defer func() {
			runtime.UnlockOSThread()
			wg.Done()
		}()

		ctx, err := gomonkey.NewFrontendContext()
		if err != nil {
			return
		}
		defer ctx.Destroy()

		code, err := os.ReadFile("script.js")
		if err != nil {
			return
		}

		stencil, err := ctx.CompileScriptToStencil("script.js", code)
		if err != nil {
			return
		}

		ch <- stencil
		close(ch)
	}(ch)

	// ... and execute it from another context ...

	wg.Add(1)
	go func(ch <-chan *gomonkey.Stencil) {
		runtime.LockOSThread()
		defer func() {
			runtime.UnlockOSThread()
			wg.Done()
		}()

		ctx, err := gomonkey.NewContext()
		if err != nil {
			return
		}
		defer ctx.Destroy()

		stencil := <-ch         // share the stencil
		defer stencil.Release() // release the stencil from the last goroutine

		for i := 0; i < 10; i++ {
			value, err := ctx.ExecuteScriptFromStencil(stencil)
			if err != nil {
				return
			}
			defer value.Release() // release after usage
		}
	}(ch)

	wg.Wait()

	return nil
}

func createGlobalFunction() error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		runtime.LockOSThread()
		defer func() {
			runtime.UnlockOSThread()
			wg.Done()
		}()

		ctx, err := gomonkey.NewContext()
		if err != nil {
			return
		}
		defer ctx.Destroy()

		// get the global object ...

		global, err := ctx.Global()
		if err != nil {
			return
		}
		defer global.Release() // release after usage

		// ... implement your Go function ...

		hello := func(args []*gomonkey.Value) (*gomonkey.Value, error) {
			// parse any arguments but do not release them
			for _, arg := range args {
				_ = arg
			}
			// create the returned value
			retValue, err := gomonkey.NewValueString(ctx, "hello world!")
			if err != nil {
				return nil, errors.New("create value")
			}
			return retValue, nil // do not release it
		}

		// ... and register it as a JS function

		if err := ctx.DefineFunction(global, "hello", hello, 0, 0); err != nil {
			return
		}
	}()

	wg.Wait()

	return nil
}

func createFunctionObject() error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		runtime.LockOSThread()
		defer func() {
			runtime.UnlockOSThread()
			wg.Done()
		}()

		ctx, err := gomonkey.NewContext()
		if err != nil {
			return
		}
		defer ctx.Destroy()

		// create a new function ...

		fn, err := gomonkey.NewFunction(ctx, "name", func(args []*gomonkey.Value) (*gomonkey.Value, error) {
			// parse any arguments but do not release them
			for _, arg := range args {
				_ = arg
			}
			// create the returned value
			retValue, err := gomonkey.NewValueString(ctx, "hello world!")
			if err != nil {
				return nil, errors.New("create value")
			}
			return retValue, nil // do not release it
		})
		if err != nil {
			return
		}
		defer fn.Release() // release after usage

		// ... you can call it with a receiver (object) ...

		receiver, err := gomonkey.NewObject(ctx)
		if err != nil {
			return
		}
		defer receiver.Release() // release after usage

		arg0, err := gomonkey.NewValueString(ctx, "value")
		if err != nil {
			return
		}
		defer arg0.Release() // release after usage

		result, err := fn.Call(receiver, arg0)
		if err != nil {
			return
		}
		defer result.Release() // release after usage
	}()

	wg.Wait()

	return nil
}

func createObjectMethod() error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		runtime.LockOSThread()
		defer func() {
			runtime.UnlockOSThread()
			wg.Done()
		}()

		ctx, err := gomonkey.NewContext()
		if err != nil {
			return
		}
		defer ctx.Destroy()

		// create a new object ...

		object, err := gomonkey.NewObject(ctx)
		if err != nil {
			return
		}
		defer object.Release() // release after usage

		// ... create a function ...

		fn, err := gomonkey.NewFunction(ctx, "hello", func(args []*gomonkey.Value) (*gomonkey.Value, error) {
			// parse any arguments but do not release them
			for _, arg := range args {
				_ = arg
			}
			// create the returned value
			retValue, err := gomonkey.NewValueString(ctx, "hello world!")
			if err != nil {
				return nil, errors.New("create value")
			}
			return retValue, nil // do not release it
		})
		if err != nil {
			return
		}

		// ... set it as a property on the object ...

		if err := object.Set("showHello", fn.AsValue()); err != nil {
			return
		}

		// ... and call this new method ...

		result, err := object.Call("showHello")
		if err != nil {
			return
		}
		defer result.Release() // release after usage
	}()

	wg.Wait()

	return nil
}
