# GoMonkey

[![Go Reference](https://pkg.go.dev/badge/github.com/bhuisgen/gomonkey.svg)](https://pkg.go.dev/github.com/bhuisgen/gomonkey)

Go bindings to Mozilla Javascript Engine [SpiderMonkey](https://spidermonkey.dev).


## Usage

```bash
$ go get github.com/bhuisgen/gomonkey
```

### Using JS contexts

```go
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
```

### Using JS objects

```go
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
```

### Evaluate code

To evaluate some JS code, evaluate it directly:

```go
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
```

### Execute a script

If you need to execute multiple times the same JS code, compile it as a script and execute it as many times as needed:

```go
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
```

### Execute a script shared between contexts

To share a script between several contexts, compile it as a stencil:

```go
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
```

### Create a global function

```go
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
```

### Create a function object

```go
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
```

### Create an object method

```go
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
```

## Setup

The shared library `libmozjs-115.so` is required for compilation and execution.

Prebuilt libraries are available for different OS/arch in the [deps/lib](./deps/lib) directory:

| OS      | CPU   | Version    |
|---------|-------|------------|
| FreeBSD | amd64 | FreeBSD 14 |
| Linux   | amd64 | Debian 12  |
| NetBSD  | amd64 | NetBSD 9   |
| OpenBSD | amd64 | OpenBSD 7  |

If your system is not available, you have to build the library. Please follow the instructions in the [build](./docs/BUILD.md) document.

Copy the library in your system library path :

**FreeBSD:**

    $ sudo cp deps/lib/freebsd_amd64/release/lib/libmozjs-115.so /usr/lib/

**Linux:**

    $ sudo cp deps/lib/linux_amd64/release/lib/libmozjs-115.so /usr/lib/x86_64-linux-gnu/

**NetBSD:**

    $ sudo cp deps/lib/netbsd_amd64/release/lib/libmozjs-115.so /usr/lib/

**OpenBSD:**

    $ sudo cp deps/lib/openbsd_amd64/release/lib/libmozjs-115.so /usr/lib/

To troubleshoot any missing system libraries, use the `ldd` command.

## Development

### Build SpiderMonkey with debug support

The SpiderMonkey library should be built with debugging support during development. Please refer to the instructions in the [build](./docs/BUILD.md) document.

### Test the Go bindings

First step is to run the unit tests:

```shell
$ go test ./tests/...
```

Then run the smoke tests:

```shell
$ go run cmd/smoke/main.go
```
