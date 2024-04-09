// This command executes the smoke tests.
package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/bhuisgen/gomonkey"
)

func main() {
	var count int
	flag.IntVar(&count, "n", 100, "number of executions")
	flag.Usage = func() {
		fmt.Println("Usage: smoke [OPTIONS]")
		fmt.Println()
		fmt.Println("Run the smoke tests.")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
	}

	flag.Parse()

	if count <= 0 {
		flag.Usage()
		os.Exit(1)
	}

	gomonkey.Init()
	log.Print("Init OK")

	log.Printf("Version: %s\n", gomonkey.Version())

	if gomonkey.Version() != "115.9" {
		log.Print("Invalid version, aborting")
		os.Exit(2)
	}
	smoke(count)

	gomonkey.ShutDown()
	log.Print("ShutDown OK")
}

func smoke(count int) {
	log.Printf("Number of iterations per test: %d\n", count)

	{
		log.Print("Testing Context ...")
		start := time.Now()
		for i := 0; i < count; i++ {
			smokeTestContext()
		}
		log.Printf("Test OK: %d ms\n", time.Since(start).Abs().Milliseconds())
	}

	{
		log.Print("Testing Value ...")
		start := time.Now()
		for i := 0; i < count; i++ {
			smokeTestValue()
		}
		log.Printf("Test OK: %d ms\n", time.Since(start).Abs().Milliseconds())
	}

	{
		log.Print("Testing Object ...")
		start := time.Now()
		for i := 0; i < count; i++ {
			smokeTestObject()
		}
		log.Printf("Test OK: %d ms\n", time.Since(start).Abs().Milliseconds())
	}

	{
		log.Print("Testing Evaluate ...")
		start := time.Now()
		for i := 0; i < count; i++ {
			smokeTestEvaluate(10)
		}
		log.Printf("Test OK: %d ms\n", time.Since(start).Abs().Milliseconds())
	}

	{
		log.Print("Testing EvaluateFunction ...")
		start := time.Now()
		for i := 0; i < count; i++ {
			smokeTestEvaluateFunction(10)
		}
		log.Printf("Test OK: %d ms\n", time.Since(start).Abs().Milliseconds())
	}

	{
		log.Print("Testing CallFunctionName ...")
		start := time.Now()
		for i := 0; i < count; i++ {
			smokeTestCallFunctionName(10)
		}
		log.Printf("Test OK: %d ms\n", time.Since(start).Abs().Milliseconds())
	}

	{
		log.Print("Testing CallFunctionValue ...")
		start := time.Now()
		for i := 0; i < count; i++ {
			smokeTestCallFunctionValue(10)
		}
		log.Printf("Test OK: %d ms\n", time.Since(start).Abs().Milliseconds())
	}

	{
		log.Print("Testing CallMethod ...")
		start := time.Now()
		for i := 0; i < count; i++ {
			smokeTestCallMethod(10)
		}
		log.Printf("Test OK: %d ms\n", time.Since(start).Abs().Milliseconds())
	}

	{
		log.Print("Testing ExecuteScript ...")
		start := time.Now()
		for i := 0; i < count; i++ {
			smokeTestExecuteScript(10)
		}
		log.Printf("Test OK: %d ms\n", time.Since(start).Abs().Milliseconds())
	}

	{
		log.Print("Testing EvaluateScriptFromStencil ...")
		start := time.Now()
		for i := 0; i < count; i++ {
			smokeTestExecuteScriptFromStencil(10)
		}
		log.Printf("Test OK: %d ms\n", time.Since(start).Abs().Milliseconds())
	}
}

func smokeTestContext() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	ctx.Destroy()
}

func smokeTestValue() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Destroy()

	stringValue, err := gomonkey.NewValueString(ctx, "string")
	if err != nil {
		log.Fatal(err)
	}
	defer stringValue.Release()
	if !stringValue.IsString() && stringValue.ToString() != "string" {
		log.Fatal(errors.New("invalid string"))
	}
	if stringValue.String() != "string" {
		log.Fatal(errors.New("invalid string conversion"))
	}

	booleanValue, err := gomonkey.NewValueBoolean(ctx, true)
	if err != nil {
		log.Fatal(err)
	}
	defer booleanValue.Release()
	if !booleanValue.IsBoolean() && !booleanValue.ToBoolean() {
		log.Fatal(errors.New("invalid boolean"))
	}
	if booleanValue.String() != "true" {
		log.Fatal(errors.New("invalid string conversion"))
	}

	numberValue, err := gomonkey.NewValueNumber(ctx, 123.456)
	if err != nil {
		log.Fatal(err)
	}
	defer numberValue.Release()
	if !numberValue.IsNumber() && numberValue.ToNumber() != 123.456 {
		log.Fatal(errors.New("invalid number"))
	}
	if numberValue.String() != "123.456" {
		log.Fatal(errors.New("invalid string conversion"))
	}

	int32Value, err := gomonkey.NewValueInt32(ctx, 123)
	if err != nil {
		log.Fatal(err)
	}
	defer int32Value.Release()
	if !int32Value.IsInt32() && int32Value.ToInt32() != 123 {
		log.Fatal(errors.New("invalid int32"))
	}
	if int32Value.String() != "123" {
		log.Fatal(errors.New("invalid string conversion"))
	}
}

func smokeTestObject() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Destroy()

	object, err := gomonkey.NewObject(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer object.Release()

	if object.Has("test") {
		log.Fatal(errors.New("invalid plain object"))
	}

	{
		value, err := gomonkey.NewValueString(ctx, "string")
		if err != nil {
			log.Fatal(err)
		}
		defer value.Release()
		if err := object.Set("test", value); err != nil {
			log.Fatal(err)
		}
		if !object.Has("test") {
			log.Fatal(errors.New("missing property"))
		}
		propValue, err := object.Get("test")
		if err != nil {
			log.Fatal(err)
		}
		defer propValue.Release()
		if !propValue.IsString() || propValue.ToString() != "string" {
			log.Fatal(errors.New("invalid property value"))
		}
		if err := object.Delete("test"); err != nil {
			log.Fatal(err)
		}
		if object.Has("test") {
			log.Fatal(errors.New("existing property"))
		}
	}

	{
		value, err := gomonkey.NewValueBoolean(ctx, true)
		if err != nil {
			log.Fatal(err)
		}
		defer value.Release()
		if err := object.Set("test", value); err != nil {
			log.Fatal(err)
		}
		if !object.Has("test") {
			log.Fatal(errors.New("missing property"))
		}
		propValue, err := object.Get("test")
		if err != nil {
			log.Fatal(err)
		}
		defer propValue.Release()
		if !propValue.IsBoolean() || !propValue.ToBoolean() {
			log.Fatal(errors.New("invalid property value"))
		}
		if err := object.Delete("test"); err != nil {
			log.Fatal(err)
		}
		if object.Has("test") {
			log.Fatal(errors.New("existing property"))
		}
	}

	{
		value, err := gomonkey.NewValueNumber(ctx, 123.456)
		if err != nil {
			log.Fatal(err)
		}
		defer value.Release()
		if err := object.Set("test", value); err != nil {
			log.Fatal(err)
		}
		if !object.Has("test") {
			log.Fatal(errors.New("missing property"))
		}
		propValue, err := object.Get("test")
		if err != nil {
			log.Fatal(err)
		}
		defer propValue.Release()
		if !propValue.IsNumber() || propValue.ToNumber() != 123.456 {
			log.Fatal(errors.New("invalid property value"))
		}
		if err := object.Delete("test"); err != nil {
			log.Fatal(err)
		}
		if object.Has("test") {
			log.Fatal(errors.New("existing property"))
		}
	}

	{
		value, err := gomonkey.NewValueInt32(ctx, 123)
		if err != nil {
			log.Fatal(err)
		}
		defer value.Release()
		if err := object.Set("test", value); err != nil {
			log.Fatal(err)
		}
		if !object.Has("test") {
			log.Fatal(errors.New("missing property"))
		}
		propValue, err := object.Get("test")
		if err != nil {
			log.Fatal(err)
		}
		defer propValue.Release()
		if !propValue.IsInt32() || propValue.ToInt32() != 123 {
			log.Fatal(errors.New("invalid property value"))
		}
		if err := object.Delete("test"); err != nil {
			log.Fatal(err)
		}
		if object.Has("test") {
			log.Fatal(errors.New("existing property"))
		}
	}

}

func smokeTestEvaluate(executions int) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Destroy()

	execute := func() {
		value, err := ctx.Evaluate([]byte("(() => { return 1 + 2; })();"))
		if err != nil {
			log.Fatal(err)
		}
		defer value.Release()
		if !value.IsInt32() || value.ToInt32() != 3 {
			log.Fatal(errors.New("invalid result"))
		}
	}

	for i := 0; i < executions; i++ {
		execute()
	}
}

func smokeTestEvaluateFunction(executions int) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Destroy()
	global, err := ctx.Global()
	if err != nil {
		log.Fatal(err)
	}
	defer global.Release()

	add := func(args []*gomonkey.Value) (*gomonkey.Value, error) {
		if len(args) != 2 {
			log.Fatal(errors.New("invalid args"))
		}
		retValue, err := gomonkey.NewValueInt32(ctx, args[0].ToInt32()+args[1].ToInt32())
		if err != nil {
			log.Fatal(err)
		}
		return retValue, nil
	}
	if err := ctx.DefineFunction(global, "add", add, 2, 0); err != nil {
		log.Fatal(err)
	}

	execute := func() {
		result, err := ctx.Evaluate([]byte("(() => { return add(1, 2); })()"))
		if err != nil {
			log.Fatal(err)
		}
		defer result.Release()
		if !result.IsInt32() || result.ToInt32() != 3 {
			log.Fatal(errors.New("invalid result"))
		}
	}

	for i := 0; i < executions; i++ {
		execute()
	}
}

func smokeTestCallFunctionName(executions int) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Destroy()
	global, err := ctx.Global()
	if err != nil {
		log.Fatal(err)
	}
	defer global.Release()

	add := func(args []*gomonkey.Value) (*gomonkey.Value, error) {
		if len(args) != 2 {
			log.Fatal(errors.New("invalid args"))
		}
		retValue, err := gomonkey.NewValueInt32(ctx, args[0].ToInt32()+args[1].ToInt32())
		if err != nil {
			log.Fatal(err)
		}
		return retValue, nil
	}
	if err := ctx.DefineFunction(global, "add", add, 2, 0); err != nil {
		log.Fatal(err)
	}

	execute := func() {
		arg0, err := gomonkey.NewValueInt32(ctx, 1)
		if err != nil {
			log.Fatal(err)
		}
		defer arg0.Release()
		arg1, err := gomonkey.NewValueInt32(ctx, 2)
		if err != nil {
			log.Fatal(err)
		}
		defer arg1.Release()
		result, err := ctx.CallFunctionName("add", global.AsValue(), arg0, arg1)
		if err != nil {
			log.Fatal(err)
		}
		defer result.Release()
		if !result.IsInt32() || result.ToInt32() != 3 {
			log.Fatal(errors.New("invalid result"))
		}
	}

	for i := 0; i < executions; i++ {
		execute()
	}
}

func smokeTestCallFunctionValue(executions int) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Destroy()
	global, err := ctx.Global()
	if err != nil {
		log.Fatal(err)
	}
	defer global.Release()

	fn, err := gomonkey.NewFunction(ctx, "add", func(args []*gomonkey.Value) (*gomonkey.Value, error) {
		if len(args) != 2 {
			log.Fatal(errors.New("invalid args"))
		}
		retValue, err := gomonkey.NewValueInt32(ctx, args[0].ToInt32()+args[1].ToInt32())
		if err != nil {
			log.Fatal(err)
		}
		return retValue, nil
	})
	if err != nil {
		log.Fatal(err)
	}

	execute := func() {
		arg0, err := gomonkey.NewValueInt32(ctx, 1)
		if err != nil {
			log.Fatal(err)
		}
		defer arg0.Release()
		arg1, err := gomonkey.NewValueInt32(ctx, 2)
		if err != nil {
			log.Fatal(err)
		}
		defer arg1.Release()
		result, err := ctx.CallFunctionValue(fn.AsValue(), global.AsValue(), arg0, arg1)
		if err != nil {
			log.Fatal(err)
		}
		defer result.Release()
		if !result.IsInt32() || result.ToInt32() != 3 {
			log.Fatal(errors.New("invalid result"))
		}
	}

	for i := 0; i < executions; i++ {
		execute()
	}
}

func smokeTestCallMethod(executions int) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewObject(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer object.Release()

	fn, err := gomonkey.NewFunction(ctx, "add", func(args []*gomonkey.Value) (*gomonkey.Value, error) {
		if len(args) != 2 {
			log.Fatal(errors.New("invalid args"))
		}
		retValue, err := gomonkey.NewValueInt32(ctx, args[0].ToInt32()+args[1].ToInt32())
		if err != nil {
			log.Fatal(err)
		}
		return retValue, nil
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := object.Set("method", fn.AsValue()); err != nil {
		log.Fatal(err)
	}

	execute := func() {
		arg0, err := gomonkey.NewValueInt32(ctx, 1)
		if err != nil {
			log.Fatal(err)
		}
		defer arg0.Release()
		arg1, err := gomonkey.NewValueInt32(ctx, 2)
		if err != nil {
			log.Fatal(err)
		}
		defer arg1.Release()
		result, err := object.Call("method", arg0, arg1)
		if err != nil {
			log.Fatal(err)
		}
		defer result.Release()
		if !result.IsInt32() || result.ToInt32() != 3 {
			log.Fatal(errors.New("invalid result"))
		}
	}

	for i := 0; i < executions; i++ {
		execute()
	}
}

func smokeTestExecuteScript(executions int) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Destroy()

	script, err := ctx.CompileScript("script.js", []byte("(() => { return 1 + 2; })();"))
	if err != nil {
		log.Fatal(err)
	}
	defer script.Release()

	execute := func() {
		result, err := ctx.ExecuteScript(script)
		if err != nil {
			log.Fatal(err)
		}
		defer result.Release()
		if !result.IsInt32() || result.ToInt32() != 3 {
			log.Fatal(errors.New("invalid result"))
		}
	}

	for i := 0; i < executions; i++ {
		execute()
	}
}

func smokeTestExecuteScriptFromStencil(executions int) {
	var wg sync.WaitGroup
	ch := make(chan *gomonkey.Stencil, executions+1)

	wg.Add(1)
	go func(ch chan<- *gomonkey.Stencil) {
		runtime.LockOSThread()
		defer func() {
			runtime.UnlockOSThread()
			close(ch)
			wg.Done()
		}()

		fc, err := gomonkey.NewFrontendContext()
		if err != nil {
			log.Fatal(err)
		}
		defer fc.Destroy()

		stencil, err := fc.CompileScriptToStencil("script.js", []byte("(() => { return 1 + 2; })();"))
		if err != nil {
			log.Fatal(err)
		}
		for i := 0; i < executions+1; i++ {
			ch <- stencil
		}
	}(ch)

	wg.Add(1)
	go func(ch <-chan *gomonkey.Stencil) {
		runtime.LockOSThread()
		defer func() {
			runtime.UnlockOSThread()
			wg.Done()
		}()

		ctx, err := gomonkey.NewContext()
		if err != nil {
			log.Fatal(err)
		}
		defer ctx.Destroy()

		execute := func() {
			stencil := <-ch

			result, err := ctx.ExecuteScriptFromStencil(stencil)
			if err != nil {
				return
			}
			defer result.Release()
			if !result.IsInt32() || result.ToInt32() != 3 {
				log.Fatal(errors.New("invalid result"))
			}
		}

		for i := 0; i < executions; i++ {
			execute()
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
}
