#include "gomonkey.h"

#include <js/Array.h>
#include <js/CompilationAndEvaluation.h>
#include <js/Conversions.h>
#include <js/Initialization.h>
#include <js/JSON.h>
#include <js/MapAndSet.h>
#include <js/Object.h>
#include <js/SourceText.h>

#include <cstdint>
#include <cstdlib>
#include <string>
#include <vector>

#ifdef CGO
#include "_cgo_export.h"
#endif

/*
 * Private objects.
 */

class Context {
 public:
  enum class Slots : uint8_t {
    REF,
    SLOT_COUNT,
  };

 public:
  explicit Context(unsigned ref, JSContext *cx, JS::HandleObject global)
      : ref(ref), ptr(cx), globalPtr(global) {
    if (globalPtr) JS_AddExtraGCRootsTracer(ptr, traceGlobal, &globalPtr);
  }
  ~Context() {
    if (globalPtr) JS_RemoveExtraGCRootsTracer(ptr, traceGlobal, &globalPtr);
  }

 private:
  Context(const Context &) = delete;

 public:
  unsigned getRef() const { return ref; }
  JSContext *getJSContext() const { return ptr; }
  JSObject *getGlobalJSObject() const { return globalPtr; };

 private:
  Context &operator=(const Context &) = delete;

 private:
  static void traceGlobal(JSTracer *trc, void *data) {
    JS::TraceEdge(trc, (JS::Heap<JSObject *> *)data, "global");
  }

 private:
  unsigned ref;
  JSContext *ptr;
  JS::Heap<JSObject *> globalPtr;
};

class Script {
 public:
  explicit Script(Context *ctx, JS::HandleScript script)
      : ctx(ctx), ptr(script) {
    if (ptr) JS_AddExtraGCRootsTracer(ctx->getJSContext(), traceScript, &ptr);
  };
  ~Script() {
    if (ptr)
      JS_RemoveExtraGCRootsTracer(ctx->getJSContext(), traceScript, &ptr);
  }

 private:
  Script(const Script &) = delete;

 public:
  Context *getContext() const { return ctx; };
  JSScript *getJSScript() const { return ptr.get(); };

 private:
  Script &operator=(const Script &) = delete;

 private:
  static void traceScript(JSTracer *trc, void *data) {
    JS::TraceEdge(trc, (JS::Heap<JSScript *> *)data, "script");
  }

 private:
  Context *ctx;
  JS::Heap<JSScript *> ptr;
};

class Value {
 public:
  explicit Value(Context *ctx, JS::HandleValue value) : ctx(ctx), ptr(value) {
    if (ptr) JS_AddExtraGCRootsTracer(ctx->getJSContext(), traceValue, &ptr);
  };
  ~Value() {
    if (ptr) JS_RemoveExtraGCRootsTracer(ctx->getJSContext(), traceValue, &ptr);
  }

 private:
  explicit Value(const Value &) = delete;

 public:
  Context *getContext() const { return ctx; };
  JS::Value getJSValue() const { return ptr.get(); };

 private:
  Value &operator=(const Value &) = delete;

 private:
  static void traceValue(JSTracer *trc, void *data) {
    JS::TraceEdge(trc, (JS::Heap<JS::Value> *)data, "value");
  }

 private:
  Context *ctx;
  JS::Heap<JS::Value> ptr;
};

class FrontendContext {
 public:
  explicit FrontendContext(JS::FrontendContext *cx) : ptr(cx) {}

 private:
  FrontendContext(const FrontendContext &c) = delete;

 public:
  JS::FrontendContext *getJSFrontendContext() const { return ptr; }

 private:
  FrontendContext &operator=(const FrontendContext &) = delete;

 private:
  JS::FrontendContext *ptr;
};

class Stencil {
 public:
  explicit Stencil(RefPtr<JS::Stencil> stencil) : ptr(stencil){};

 public:
  JS::Stencil *getStencil() const { return ptr.get(); }

 private:
  RefPtr<JS::Stencil> ptr;
};

/*
 * Go callbacks.
 */

extern ContextPtr goFunctionContext(unsigned contextRef);

extern ResultGoFunctionCallback goFunctionCallback(unsigned contextRef,
                                                   char *name, unsigned argc,
                                                   ValuePtr *vp);

/*
 * Private functions.
 */

static JSObject *CreateGlobalObject(JSContext *cx) {
  JS::RealmOptions options;
  static JSClass GlobalClass = {"Global",
                                JSCLASS_GLOBAL_FLAGS_WITH_SLOTS(1),
                                &JS::DefaultGlobalClassOps,
                                nullptr,
                                nullptr,
                                nullptr};

  return JS_NewGlobalObject(cx, &GlobalClass, nullptr, JS::FireOnNewGlobalHook,
                            options);
}

static Error GetError(JSContext *cx) {
  Error err = {};

  JS::ExceptionStack stack(cx);
  if (!JS::StealPendingExceptionStack(cx, &stack)) {
    return err;
  }
  JS::ErrorReportBuilder builder(cx);
  if (!builder.init(cx, stack, JS::ErrorReportBuilder::WithSideEffects)) {
    return err;
  }
  JSErrorReport *report = builder.report();

  char *message = strdup(report->message().c_str());
  if (!message) {
    return err;
  }

  err.message = message;
  err.filename = report->filename;
  err.lineno = report->lineno;
  err.number = report->errorNumber;
  return err;
}

static bool InterruptCallback(JSContext *cx) {
  JS_ResetInterruptCallback(cx, true);

  JS_ReportErrorUTF8(cx, "Execution interrupted");

  return false;
}

static bool FunctionCallback(JSContext *cx, unsigned argc, JS::Value *vp) {
  JS::CallArgs args = JS::CallArgsFromVp(argc, vp);

  JS::RootedValue funcVal(cx, args.calleev());
  JSFunction *func = JS_ValueToFunction(cx, funcVal);
  if (!func) {
    JS_ReportOutOfMemory(cx);
    return false;
  }

  JS::PersistentRootedObject global(cx,
                                    JS::GetNonCCWObjectGlobal(&args.callee()));
  if (!global) {
    JS_ReportOutOfMemory(cx);
    return false;
  }
  JS::RootedValue contextRefVal(
      cx,
      JS::GetReservedSlot(global, static_cast<size_t>(Context::Slots::REF)));
  if (!contextRefVal.isInt32()) {
    JS_ReportOutOfMemory(cx);
    return false;
  }
  unsigned contextRef = contextRefVal.toInt32();

  ContextPtr ctx = goFunctionContext(contextRef);
  if (!ctx) {
    JS_ReportOutOfMemory(cx);
    return false;
  }

  JSString *nameStr = JS_GetFunctionId(func);
  if (!nameStr) {
    JS_ReportOutOfMemory(cx);
    return false;
  }
  size_t len = JS_GetStringEncodingLength(cx, nameStr);
  char *name = static_cast<char *>(JS_malloc(cx, len + 1));
  if (!name) {
    JS_ReportOutOfMemory(cx);
    return false;
  }
  if (!JS_EncodeStringToBuffer(cx, nameStr, name, len)) {
    JS_ReportOutOfMemory(cx);
    JS_free(cx, name);
    return false;
  }
  name[len] = 0;

  std::vector<Value *> vals = {};
  for (unsigned i = 0; i < args.length(); i++) {
    Value *v = new Value(ctx, args.get(i));
    if (!v) {
      JS_ReportOutOfMemory(cx);
      return false;
    }
    vals.push_back(v);
  }
  JS::RootedValue rval(cx);

  ResultGoFunctionCallback result =
      goFunctionCallback(contextRef, name, vals.size(), vals.data());

  for (const auto &val : vals) {
    delete val;
  }
  JS_free(cx, name);

  if (result.err) {
    JS_ReportErrorUTF8(cx, "%s", result.err);
    JS_free(cx, result.err);
    return false;
  }
  if (result.ptr) {
    rval.set(result.ptr->getJSValue());

    delete result.ptr;
  } else {
    rval.setUndefined();
  }

  args.rval().set(rval);
  return true;
}

static bool StringifyCallback(const char16_t *buf, uint32_t len, void *data) {
  std::u16string *str = static_cast<std::u16string *>(data);
  str->append(buf, len);
  return true;
}

/*
 * Public functions.
 */

bool Init() { return JS_Init(); }

void ShutDown() { JS_ShutDown(); }

const char *Version() {
  const std::string version = std::to_string(MOZJS_MAJOR_VERSION) + "." +
                              std::to_string(MOZJS_MINOR_VERSION);

  char *str = strdup(version.c_str());

  return str;
}

ContextPtr NewContext(unsigned ref, ContextOptions options) {
  uint32_t heapMaxBytes = JS::DefaultHeapMaxBytes;
  if (options.heapMaxBytes) {
    heapMaxBytes = options.heapMaxBytes;
  }

  JSContext *cx = JS_NewContext(heapMaxBytes);
  if (!cx) {
    return nullptr;
  }
  if (options.stackSize) {
    JS_SetNativeStackQuota(cx, options.stackSize);
  }
#ifdef DEBUG
  JS_SetGCZeal(cx, 14, 1);
#endif
  if (options.gcMaxBytes) {
    JS_SetGCParameter(cx, JSGC_MAX_BYTES, options.gcMaxBytes);
  }
  if (options.gcIncrementalEnabled) {
    JS_SetGCParameter(cx, JSGC_INCREMENTAL_GC_ENABLED,
                      options.gcIncrementalEnabled);
  }
  if (options.gcSliceTimeBudgetMs) {
    JS_SetGCParameter(cx, JSGC_SLICE_TIME_BUDGET_MS,
                      options.gcSliceTimeBudgetMs);
  }

  if (!JS_AddInterruptCallback(cx, &InterruptCallback)) {
    return nullptr;
  }

  if (!JS::InitSelfHostedCode(cx)) {
    return nullptr;
  }

  JS::RootedObject global(cx, CreateGlobalObject(cx));
  if (!global) {
    return nullptr;
  }
  JSAutoRealm ar(cx, global);

  JS::RootedValue contextRefVal(cx, JS::Int32Value(ref));
  JS_SetReservedSlot(global, static_cast<uint32_t>(Context::Slots::REF),
                     contextRefVal);

  Context *ctx = new Context(ref, cx, global);
  if (!ctx) {
    return nullptr;
  }
  return ctx;
}

void DestroyContext(ContextPtr ctx) {
  JS_DestroyContext(ctx->getJSContext());
  delete ctx;
}

void RequestInterruptContext(ContextPtr ctx) {
  JS_RequestInterruptCallback(ctx->getJSContext());
}

ResultValue GetGlobalObject(ContextPtr ctx) {
  ResultValue result = {};

  JS::PersistentRootedObject global(ctx->getJSContext(),
                                    ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  JS::RootedValue globalVal(ctx->getJSContext());
  globalVal.setObject(*global);

  Value *v = new Value(ctx, globalVal);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultValue DefineObject(ContextPtr ctx, ValuePtr recv, char *name,
                         unsigned attrs) {
  ResultValue result = {};

  JS::PersistentRootedObject global(ctx->getJSContext(),
                                    ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), global);

  JS::RootedValue recvValue(ctx->getJSContext(), recv->getJSValue());
  JS::RootedObject recvObject(ctx->getJSContext(),
                              JS::ToObject(ctx->getJSContext(), recvValue));
  if (!recvObject) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  JS::RootedObject object(
      ctx->getJSContext(),
      JS_DefineObject(ctx->getJSContext(), recvObject, name, nullptr, attrs));
  if (!object) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  JS::RootedValue objectVal(ctx->getJSContext());
  objectVal.setObject(*object);

  Value *v = new Value(ctx, objectVal);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

Result DefineProperty(ContextPtr ctx, ValuePtr recv, char *name, ValuePtr value,
                      unsigned attrs) {
  Result result = {};

  JS::PersistentRootedObject global(ctx->getJSContext(),
                                    ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), global);

  JS::RootedValue recvValue(ctx->getJSContext(), recv->getJSValue());
  JS::RootedObject recvObject(ctx->getJSContext(),
                              JS::ToObject(ctx->getJSContext(), recvValue));
  if (!recvObject) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  JS::RootedValue propValue(ctx->getJSContext(), value->getJSValue());

  if (!JS_DefineProperty(ctx->getJSContext(), recvObject, name, propValue,
                         attrs)) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  result.ok = true;
  return result;
}

Result DefineElement(ContextPtr ctx, ValuePtr recv, uint32_t index,
                     ValuePtr value, unsigned attrs) {
  Result result = {};

  JS::PersistentRootedObject global(ctx->getJSContext(),
                                    ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), global);

  JS::RootedValue recvValue(ctx->getJSContext(), recv->getJSValue());
  JS::RootedObject recvObject(ctx->getJSContext(),
                              JS::ToObject(ctx->getJSContext(), recvValue));
  if (!recvObject) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  JS::RootedValue propValue(ctx->getJSContext(), value->getJSValue());

  if (!JS_DefineElement(ctx->getJSContext(), recvObject, index, propValue,
                        attrs)) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  result.ok = true;

  return result;
}

Result DefineFunction(ContextPtr ctx, ValuePtr recv, char *name, unsigned nargs,
                      unsigned attrs) {
  Result result = {};

  JS::PersistentRootedObject global(ctx->getJSContext(),
                                    ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), global);

  JS::RootedValue recvValue(ctx->getJSContext(), recv->getJSValue());
  JS::RootedObject recvObject(ctx->getJSContext(),
                              JS::ToObject(ctx->getJSContext(), recvValue));
  if (!recvObject) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JS::RootedFunction func(
      ctx->getJSContext(),
      JS_DefineFunction(ctx->getJSContext(), recvObject, name,
                        &FunctionCallback, nargs, attrs));
  if (!func) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JS::RootedObject funcObj(ctx->getJSContext(), JS_GetFunctionObject(func));
  if (!funcObj) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  result.ok = true;
  return result;
}

ResultValue CallFunctionName(ContextPtr ctx, char *name, ValuePtr recv,
                             int argc, ValuePtr *argv) {
  ResultValue result = {};

  JS::PersistentRootedObject global(ctx->getJSContext(),
                                    ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), global);

  JS::RootedValue recvValue(ctx->getJSContext(), recv->getJSValue());
  JS::RootedObject recvObject(ctx->getJSContext(),
                              JS::ToObject(ctx->getJSContext(), recvValue));
  if (!recvObject) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JS::RootedValueVector values(ctx->getJSContext());
  if (!values.resize(argc)) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  for (int i = 0; i < argc; i++) {
    values[i].set(argv[i]->getJSValue());
  }
  JS::HandleValueArray args(values);
  JS::RootedValue rval(ctx->getJSContext());
  if (!JS_CallFunctionName(ctx->getJSContext(), recvObject, name, args,
                           &rval)) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  Value *v = new Value(ctx, rval);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultValue CallFunctionValue(ContextPtr ctx, ValuePtr func, ValuePtr recv,
                              int argc, ValuePtr *argv) {
  ResultValue result = {};

  JS::PersistentRootedObject global(ctx->getJSContext(),
                                    ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), global);

  JS::RootedValue recvValue(ctx->getJSContext(), recv->getJSValue());
  JS::RootedObject recvObject(ctx->getJSContext(),
                              JS::ToObject(ctx->getJSContext(), recvValue));
  if (!recvObject) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JS::RootedValue funcVal(ctx->getJSContext(), func->getJSValue());
  JS::RootedValueVector values(ctx->getJSContext());
  if (!values.resize(argc)) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  for (int i = 0; i < argc; i++) {
    values[i].set(argv[i]->getJSValue());
  }
  JS::HandleValueArray args(values);
  JS::RootedValue rval(ctx->getJSContext());
  if (!JS_CallFunctionValue(ctx->getJSContext(), recvObject, funcVal, args,
                            &rval)) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  Value *v = new Value(ctx, rval);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultValue Evaluate(ContextPtr ctx, char *code) {
  ResultValue result = {};

  JS::PersistentRootedObject global(ctx->getJSContext(),
                                    ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), global);

  JS::CompileOptions options(ctx->getJSContext());
  options.setFileAndLine("", 1);

  JS::SourceText<mozilla::Utf8Unit> source;
  if (!source.init(ctx->getJSContext(), code, strlen(code),
                   JS::SourceOwnership::Borrowed)) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  JS::RootedValue rval(ctx->getJSContext());
  if (!JS::Evaluate(ctx->getJSContext(), options, source, &rval)) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  Value *v = new Value(ctx, rval);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultCompileScript CompileScript(ContextPtr ctx, char *filename, char *code) {
  ResultCompileScript result = {};

  JS::PersistentRootedObject global(ctx->getJSContext(),
                                    ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), global);

  JS::CompileOptions options(ctx->getJSContext());
  options.setFileAndLine(filename, 1);

  JS::SourceText<mozilla::Utf8Unit> source;
  if (!source.init(ctx->getJSContext(), code, strlen(code),
                   JS::SourceOwnership::Borrowed)) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  JS::RootedScript rscript(ctx->getJSContext(),
                           JS::Compile(ctx->getJSContext(), options, source));
  if (!rscript) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  Script *script = new Script(ctx, rscript);
  if (!script) {
    return result;
  }

  result.ok = true;
  result.ptr = script;
  return result;
}

void ReleaseScript(ScriptPtr script) { delete script; }

ResultValue ExecuteScript(ContextPtr ctx, ScriptPtr script) {
  ResultValue result = {};

  JS::PersistentRootedObject global(ctx->getJSContext(),
                                    ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), ctx->getGlobalJSObject());

  JS::RootedValue rval(script->getContext()->getJSContext());
  JS::RootedScript rscript(script->getContext()->getJSContext(),
                           script->getJSScript());
  if (!JS_ExecuteScript(ctx->getJSContext(), rscript, &rval)) {
    result.err = GetError(script->getContext()->getJSContext());
    return result;
  }

  Value *v = new Value(script->getContext(), rval);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultValue ExecuteScriptFromStencil(ContextPtr ctx, StencilPtr stencil) {
  ResultValue result = {};

  JSAutoRealm ar(ctx->getJSContext(), ctx->getGlobalJSObject());

  JS::CompileOptions options(ctx->getJSContext());
  JS::InstantiateOptions instantiateOptions(options);
  JS::RootedScript script(
      ctx->getJSContext(),
      JS::InstantiateGlobalStencil(ctx->getJSContext(), instantiateOptions,
                                   stencil->getStencil()));
  if (!script) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  JS::RootedValue rval(ctx->getJSContext());
  if (!JS_ExecuteScript(ctx->getJSContext(), script, &rval)) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  Value *v = new Value(ctx, rval);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

FrontendContextPtr NewFrontendContext(FrontendContextOptions options) {
  JS::FrontendContext *fc = JS::NewFrontendContext();
  if (!fc) {
    return nullptr;
  }
  if (options.stackSize) {
    JS::SetNativeStackQuota(fc, 0);
  }

  FrontendContext *ctx = new FrontendContext(fc);
  if (!ctx) {
    return nullptr;
  }
  return ctx;
}

void DestroyFrontendContext(FrontendContextPtr ctx) {
  JS::DestroyFrontendContext(ctx->getJSFrontendContext());
  delete ctx;
}

ResultCompileStencil CompileScriptToStencil(FrontendContextPtr ctx,
                                            char *filename, char *code) {
  ResultCompileStencil result = {};

  JS::CompileOptions options(JS::CompileOptions::ForFrontendContext{});
  options.setFileAndLine(filename, 1);

  JS::SourceText<mozilla::Utf8Unit> source;
  if (!source.init(ctx->getJSFrontendContext(), code, strlen(code),
                   JS::SourceOwnership::Borrowed)) {
    return result;
  }

  JS::CompilationStorage compileStorage;
  RefPtr<JS::Stencil> st = JS::CompileGlobalScriptToStencil(
      ctx->getJSFrontendContext(), options, source, compileStorage);
  if (!st) {
    return result;
  }

  Stencil *stencil = new Stencil(st);
  if (!stencil) {
    return result;
  }

  result.ok = true;
  result.ptr = stencil;
  return result;
}

void ReleaseStencil(StencilPtr stencil) { delete stencil; }

ResultValue NewValueUndefined(ContextPtr ctx) {
  ResultValue result = {};

  JS::RootedValue val(ctx->getJSContext());
  val.setUndefined();

  Value *v = new Value(ctx, val);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultValue NewValueNull(ContextPtr ctx) {
  ResultValue result = {};

  JS::RootedValue val(ctx->getJSContext());
  val.setNull();

  Value *v = new Value(ctx, val);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultValue NewValueString(ContextPtr ctx, char *str, int len) {
  ResultValue result = {};

  JS::RootedObject global(ctx->getJSContext(), ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), global);

  JS::RootedValue val(ctx->getJSContext());
  JSString *s = JS_NewStringCopyN(ctx->getJSContext(), str, len);
  if (!s) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  val.setString(s);

  Value *v = new Value(ctx, val);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultValue NewValueBoolean(ContextPtr ctx, bool b) {
  ResultValue result = {};

  JS::RootedValue val(ctx->getJSContext());
  val.setBoolean(b);

  Value *v = new Value(ctx, val);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultValue NewValueNumber(ContextPtr ctx, double d) {
  ResultValue result = {};

  JS::RootedValue val(ctx->getJSContext());
  val.setNumber(d);

  Value *v = new Value(ctx, val);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultValue NewValueInt32(ContextPtr ctx, int32_t i) {
  ResultValue result = {};

  JS::RootedValue val(ctx->getJSContext());
  val.setInt32(i);

  Value *v = new Value(ctx, val);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

void ReleaseValue(ValuePtr value) { delete value; }

ResultString ToString(ValuePtr value) {
  ResultString result = {};

  JS::RootedObject global(value->getContext()->getJSContext(),
                          value->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(value->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(value->getContext()->getJSContext(), global);

  JS::RootedValue val(value->getContext()->getJSContext(), value->getJSValue());

  JS::RootedString str(value->getContext()->getJSContext(),
                       JS::ToString(value->getContext()->getJSContext(), val));
  if (!str) {
    result.err = GetError(value->getContext()->getJSContext());
    return result;
  }

  size_t len =
      JS_GetStringEncodingLength(value->getContext()->getJSContext(), str);
  char *data =
      static_cast<char *>(JS_malloc(value->getContext()->getJSContext(), len));
  if (!data) {
    result.err = GetError(value->getContext()->getJSContext());
    return result;
  }
  if (!JS_EncodeStringToBuffer(value->getContext()->getJSContext(), str, data,
                               len)) {
    result.err = GetError(value->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  result.data = data;
  result.len = len;
  return result;
}

bool ValueIs(ValuePtr value, ValuePtr other) {
  return value->getJSValue() == other->getJSValue();
}

bool ValueIsUndefined(ValuePtr value) {
  return value->getJSValue().isUndefined();
}

bool ValueIsNull(ValuePtr value) { return value->getJSValue().isNull(); }

bool ValueIsNullOrUndefined(ValuePtr value) {
  return value->getJSValue().isNullOrUndefined();
}

bool ValueIsTrue(ValuePtr value) { return value->getJSValue().isTrue(); }

bool ValueIsFalse(ValuePtr value) { return value->getJSValue().isFalse(); }

bool ValueIsObject(ValuePtr value) { return value->getJSValue().isObject(); }

bool ValueIsFunction(ValuePtr value) {
  return value->getJSValue().isObject() &&
         JS::IsCallable(&value->getJSValue().toObject());
}

bool ValueIsSymbol(ValuePtr value) { return value->getJSValue().isSymbol(); }

bool ValueIsString(ValuePtr value) { return value->getJSValue().isString(); }

bool ValueIsBoolean(ValuePtr value) { return value->getJSValue().isBoolean(); }

bool ValueIsNumber(ValuePtr value) { return value->getJSValue().isNumber(); }

bool ValueIsInt32(ValuePtr value) { return value->getJSValue().isInt32(); }

bool ValueIsBigInt(ValuePtr value) { return value->getJSValue().isBigInt(); }

ResultString ValueToString(ValuePtr value) {
  ResultString result = {};

  JS::RootedObject global(value->getContext()->getJSContext(),
                          value->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(value->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(value->getContext()->getJSContext(), global);

  JS::RootedValue val(value->getContext()->getJSContext());
  if (!value->getJSValue().isString()) {
    return result;
  }

  JSString *str = value->getJSValue().toString();
  if (!str) {
    result.err = GetError(value->getContext()->getJSContext());
    return result;
  }
  size_t len =
      JS_GetStringEncodingLength(value->getContext()->getJSContext(), str);
  char *data =
      static_cast<char *>(JS_malloc(value->getContext()->getJSContext(), len));
  if (!data) {
    result.err = GetError(value->getContext()->getJSContext());
    return result;
  }
  if (!JS_EncodeStringToBuffer(value->getContext()->getJSContext(), str, data,
                               len)) {
    result.err = GetError(value->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  result.data = data;
  result.len = len;
  return result;
}

int ValueToBoolean(ValuePtr value) { return value->getJSValue().toBoolean(); }

double ValueToNumber(ValuePtr value) { return value->getJSValue().toNumber(); }

int32_t ValueToInt32(ValuePtr value) { return value->getJSValue().toInt32(); }

ResultValue NewPlainObject(ContextPtr ctx) {
  ResultValue result = {};

  JS::PersistentRootedObject global(ctx->getJSContext(),
                                    ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), global);

  JS::RootedObject obj(ctx->getJSContext(),
                       JS_NewPlainObject(ctx->getJSContext()));
  JS::RootedValue objectVal(ctx->getJSContext());
  objectVal.setObject(*obj);

  Value *v = new Value(ctx, objectVal);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultBool ObjectHasProperty(ValuePtr object, char *key) {
  ResultBool result = {};

  JS::RootedObject global(object->getContext()->getJSContext(),
                          object->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(object->getContext()->getJSContext(), global);

  JS::RootedObject obj(object->getContext()->getJSContext(),
                       &object->getJSValue().toObject());
  if (!obj) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }

  bool found;
  if (!JS_HasProperty(object->getContext()->getJSContext(), obj, key, &found)) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  result.value = found;
  return result;
}

ResultValue ObjectGetProperty(ValuePtr object, char *key) {
  ResultValue result = {};

  JS::RootedObject global(object->getContext()->getJSContext(),
                          object->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(object->getContext()->getJSContext(), global);

  JS::RootedObject obj(object->getContext()->getJSContext(),
                       &object->getJSValue().toObject());
  if (!obj) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }

  JS::RootedValue propValue(object->getContext()->getJSContext());
  if (!JS_GetProperty(object->getContext()->getJSContext(), obj, key,
                      &propValue)) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }

  Value *v = new Value(object->getContext(), propValue);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

Result ObjectSetProperty(ValuePtr object, char *key, ValuePtr value) {
  Result result = {};

  JS::RootedObject global(object->getContext()->getJSContext(),
                          object->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(object->getContext()->getJSContext(), global);

  JS::RootedObject obj(object->getContext()->getJSContext(),
                       &object->getJSValue().toObject());
  if (!obj) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }

  JS::RootedValue propValue(object->getContext()->getJSContext(),
                            value->getJSValue());
  if (!JS_SetProperty(object->getContext()->getJSContext(), obj, key,
                      propValue)) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  return result;
}

Result ObjectDeleteProperty(ValuePtr object, char *key) {
  Result result = {};

  JS::RootedObject global(object->getContext()->getJSContext(),
                          object->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(object->getContext()->getJSContext(), global);

  JS::RootedObject obj(object->getContext()->getJSContext(),
                       &object->getJSValue().toObject());
  if (!obj) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }

  if (!JS_DeleteProperty(object->getContext()->getJSContext(), obj, key)) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  return result;
}

ResultBool ObjectHasElement(ValuePtr object, uint32_t index) {
  ResultBool result = {};

  JS::RootedObject global(object->getContext()->getJSContext(),
                          object->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(object->getContext()->getJSContext(), global);

  JS::RootedObject obj(object->getContext()->getJSContext(),
                       &object->getJSValue().toObject());
  if (!obj) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }

  bool found;
  if (!JS_HasElement(object->getContext()->getJSContext(), obj, index,
                     &found)) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  result.value = found;
  return result;
}

ResultValue ObjectGetElement(ValuePtr object, uint32_t index) {
  ResultValue result = {};

  JS::RootedObject global(object->getContext()->getJSContext(),
                          object->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(object->getContext()->getJSContext(), global);

  JS::RootedObject obj(object->getContext()->getJSContext(),
                       &object->getJSValue().toObject());
  if (!obj) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }

  JS::RootedValue propValue(object->getContext()->getJSContext());
  if (!JS_GetElement(object->getContext()->getJSContext(), obj, index,
                     &propValue)) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }

  Value *v = new Value(object->getContext(), propValue);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

Result ObjectSetElement(ValuePtr object, uint32_t index, ValuePtr value) {
  Result result = {};

  JS::RootedObject global(object->getContext()->getJSContext(),
                          object->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(object->getContext()->getJSContext(), global);

  JS::RootedObject obj(object->getContext()->getJSContext(),
                       &object->getJSValue().toObject());
  if (!obj) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }

  JS::RootedValue propValue(object->getContext()->getJSContext(),
                            value->getJSValue());
  if (!JS_SetElement(object->getContext()->getJSContext(), obj, index,
                     propValue)) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  return result;
}

Result ObjectDeleteElement(ValuePtr object, uint32_t index) {
  Result result = {};

  JS::RootedObject global(object->getContext()->getJSContext(),
                          object->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(object->getContext()->getJSContext(), global);

  JS::RootedObject obj(object->getContext()->getJSContext(),
                       &object->getJSValue().toObject());
  if (!obj) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }

  if (!JS_DeleteElement(object->getContext()->getJSContext(), obj, index)) {
    result.err = GetError(object->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  return result;
}

ResultValue NewFunction(ContextPtr ctx, char *name) {
  ResultValue result = {};

  JS::PersistentRootedObject global(ctx->getJSContext(),
                                    ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), global);

  JS::RootedFunction func(
      ctx->getJSContext(),
      JS_NewFunction(ctx->getJSContext(), &FunctionCallback, 0, 0, name));
  if (!func) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  JS::RootedObject funcObj(ctx->getJSContext(), JS_GetFunctionObject(func));
  if (!funcObj) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JS::RootedValue funcValue(ctx->getJSContext());
  funcValue.setObject(*funcObj);

  Value *v = new Value(ctx, funcValue);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultValue NewArrayObject(ContextPtr ctx, int argc, ValuePtr *argv) {
  ResultValue result = {};

  JS::RootedObject global(ctx->getJSContext(), ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), global);

  JS::RootedValueVector values(ctx->getJSContext());
  if (!values.resize(argc)) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  for (int i = 0; i < argc; i++) {
    values[i].set(argv[i]->getJSValue());
  }
  JS::HandleValueArray args(values);

  JS::RootedObject obj(ctx->getJSContext(),
                       JS::NewArrayObject(ctx->getJSContext(), values));
  JS::RootedValue objectVal(ctx->getJSContext());
  objectVal.setObject(*obj);

  Value *v = new Value(ctx, objectVal);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultUInt32 GetArrayObjectLength(ValuePtr array) {
  ResultUInt32 result = {};

  JS::RootedObject global(array->getContext()->getJSContext(),
                          array->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(array->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(array->getContext()->getJSContext(), global);

  JS::RootedObject obj(array->getContext()->getJSContext(),
                       &array->getJSValue().toObject());
  if (!obj) {
    result.err = GetError(array->getContext()->getJSContext());
    return result;
  }

  uint32_t length;
  if (!JS::GetArrayLength(array->getContext()->getJSContext(), obj, &length)) {
    result.err = GetError(array->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  result.value = length;
  return result;
}

ResultValue NewMapObject(ContextPtr ctx) {
  ResultValue result = {};

  JS::RootedObject global(ctx->getJSContext(), ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), global);

  JS::RootedObject obj(ctx->getJSContext(),
                       JS::NewMapObject(ctx->getJSContext()));
  JS::RootedValue val(ctx->getJSContext());
  val.setObject(*obj);

  Value *v = new Value(ctx, val);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultUInt32 MapObjectSize(ValuePtr map) {
  ResultUInt32 result = {};

  JS::RootedObject global(map->getContext()->getJSContext(),
                          map->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(map->getContext()->getJSContext(), global);

  JS::RootedObject mapObj(map->getContext()->getJSContext(),
                          &map->getJSValue().toObject());
  if (!mapObj) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }

  uint32_t size = JS::MapSize(map->getContext()->getJSContext(), mapObj);

  result.ok = true;
  result.value = size;
  return result;
}

ResultBool MapObjectHas(ValuePtr map, ValuePtr key) {
  ResultBool result = {};

  JS::RootedObject global(map->getContext()->getJSContext(),
                          map->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(map->getContext()->getJSContext(), global);

  JS::RootedObject mapObj(map->getContext()->getJSContext(),
                          &map->getJSValue().toObject());
  if (!mapObj) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }
  JS::RootedValue keyVal(map->getContext()->getJSContext(), key->getJSValue());

  bool found;
  if (!JS::MapHas(map->getContext()->getJSContext(), mapObj, keyVal, &found)) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  result.value = found;
  return result;
}

ResultValue MapObjectGet(ValuePtr map, ValuePtr key) {
  ResultValue result = {};

  JS::RootedObject global(map->getContext()->getJSContext(),
                          map->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(map->getContext()->getJSContext(), global);

  JS::RootedObject mapObj(map->getContext()->getJSContext(),
                          &map->getJSValue().toObject());
  if (!mapObj) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }
  JS::RootedValue keyVal(map->getContext()->getJSContext(), key->getJSValue());

  JS::RootedValue val(map->getContext()->getJSContext());
  if (!JS::MapGet(map->getContext()->getJSContext(), mapObj, keyVal, &val)) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }

  Value *v = new Value(map->getContext(), val);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

Result MapObjectSet(ValuePtr map, ValuePtr key, ValuePtr val) {
  Result result = {};

  JS::RootedObject global(map->getContext()->getJSContext(),
                          map->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(map->getContext()->getJSContext(), global);

  JS::RootedObject mapObj(map->getContext()->getJSContext(),
                          &map->getJSValue().toObject());
  if (!mapObj) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }
  JS::RootedValue keyVal(map->getContext()->getJSContext(), key->getJSValue());
  JS::RootedValue valVal(map->getContext()->getJSContext(), val->getJSValue());

  if (!JS::MapSet(map->getContext()->getJSContext(), mapObj, keyVal, valVal)) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  return result;
}

ResultBool MapObjectDelete(ValuePtr map, ValuePtr key) {
  ResultBool result = {};

  JS::RootedObject global(map->getContext()->getJSContext(),
                          map->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(map->getContext()->getJSContext(), global);

  JS::RootedObject mapObj(map->getContext()->getJSContext(),
                          &map->getJSValue().toObject());
  if (!mapObj) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }
  JS::RootedValue keyVal(map->getContext()->getJSContext(), key->getJSValue());

  bool found;
  if (!JS::MapDelete(map->getContext()->getJSContext(), mapObj, keyVal,
                     &found)) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  result.value = found;
  return result;
}

Result MapObjectClear(ValuePtr map) {
  Result result = {};

  JS::RootedObject global(map->getContext()->getJSContext(),
                          map->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(map->getContext()->getJSContext(), global);

  JS::RootedObject mapObj(map->getContext()->getJSContext(),
                          &map->getJSValue().toObject());
  if (!mapObj) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }

  if (!JS::MapClear(map->getContext()->getJSContext(), mapObj)) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  return result;
}

ResultValue MapObjectKeys(ValuePtr map) {
  ResultValue result = {};

  JS::RootedObject global(map->getContext()->getJSContext(),
                          map->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(map->getContext()->getJSContext(), global);

  JS::RootedObject mapObj(map->getContext()->getJSContext(),
                          &map->getJSValue().toObject());
  if (!mapObj) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }

  JS::RootedValue val(map->getContext()->getJSContext());
  if (!JS::MapKeys(map->getContext()->getJSContext(), mapObj, &val)) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }

  Value *v = new Value(map->getContext(), val);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultValue MapObjectValues(ValuePtr map) {
  ResultValue result = {};

  JS::RootedObject global(map->getContext()->getJSContext(),
                          map->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(map->getContext()->getJSContext(), global);

  JS::RootedObject mapObj(map->getContext()->getJSContext(),
                          &map->getJSValue().toObject());
  if (!mapObj) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }

  JS::RootedValue val(map->getContext()->getJSContext());
  if (!JS::MapValues(map->getContext()->getJSContext(), mapObj, &val)) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }

  Value *v = new Value(map->getContext(), val);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultValue MapObjectEntries(ValuePtr map) {
  ResultValue result = {};

  JS::RootedObject global(map->getContext()->getJSContext(),
                          map->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(map->getContext()->getJSContext(), global);

  JS::RootedObject mapObj(map->getContext()->getJSContext(),
                          &map->getJSValue().toObject());
  if (!mapObj) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }

  JS::RootedValue val(map->getContext()->getJSContext());
  if (!JS::MapEntries(map->getContext()->getJSContext(), mapObj, &val)) {
    result.err = GetError(map->getContext()->getJSContext());
    return result;
  }

  Value *v = new Value(map->getContext(), val);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultValue NewSetObject(ContextPtr ctx) {
  ResultValue result = {};

  JS::RootedObject global(ctx->getJSContext(), ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), global);

  JS::RootedObject obj(ctx->getJSContext(),
                       JS::NewSetObject(ctx->getJSContext()));
  JS::RootedValue val(ctx->getJSContext());
  val.setObject(*obj);

  Value *v = new Value(ctx, val);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultUInt32 SetObjectSize(ValuePtr set) {
  ResultUInt32 result = {};

  JS::RootedObject global(set->getContext()->getJSContext(),
                          set->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(set->getContext()->getJSContext(), global);

  JS::RootedObject setObj(set->getContext()->getJSContext(),
                          &set->getJSValue().toObject());
  if (!setObj) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }

  uint32_t size = JS::SetSize(set->getContext()->getJSContext(), setObj);

  result.ok = true;
  result.value = size;
  return result;
}

ResultBool SetObjectHas(ValuePtr set, ValuePtr key) {
  ResultBool result = {};

  JS::RootedObject global(set->getContext()->getJSContext(),
                          set->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(set->getContext()->getJSContext(), global);

  JS::RootedObject setObj(set->getContext()->getJSContext(),
                          &set->getJSValue().toObject());
  if (!setObj) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }
  JS::RootedValue keyVal(set->getContext()->getJSContext(), key->getJSValue());

  bool found;
  if (!JS::SetHas(set->getContext()->getJSContext(), setObj, keyVal, &found)) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  result.value = found;
  return result;
}

Result SetObjectAdd(ValuePtr set, ValuePtr val) {
  Result result = {};

  JS::RootedObject global(set->getContext()->getJSContext(),
                          set->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(set->getContext()->getJSContext(), global);

  JS::RootedObject mapObj(set->getContext()->getJSContext(),
                          &set->getJSValue().toObject());
  if (!mapObj) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }
  JS::RootedValue valVal(set->getContext()->getJSContext(), val->getJSValue());

  if (!JS::SetAdd(set->getContext()->getJSContext(), mapObj, valVal)) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  return result;
}

ResultBool SetObjectDelete(ValuePtr set, ValuePtr key) {
  ResultBool result = {};

  JS::RootedObject global(set->getContext()->getJSContext(),
                          set->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(set->getContext()->getJSContext(), global);

  JS::RootedObject setObj(set->getContext()->getJSContext(),
                          &set->getJSValue().toObject());
  if (!setObj) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }
  JS::RootedValue keyVal(set->getContext()->getJSContext(), key->getJSValue());

  bool found;
  if (!JS::SetDelete(set->getContext()->getJSContext(), setObj, keyVal,
                     &found)) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  result.value = found;
  return result;
}

Result SetObjectClear(ValuePtr set) {
  Result result = {};

  JS::RootedObject global(set->getContext()->getJSContext(),
                          set->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(set->getContext()->getJSContext(), global);

  JS::RootedObject setObj(set->getContext()->getJSContext(),
                          &set->getJSValue().toObject());
  if (!setObj) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }

  if (!JS::SetClear(set->getContext()->getJSContext(), setObj)) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  return result;
}

ResultValue SetObjectKeys(ValuePtr set) {
  ResultValue result = {};

  JS::RootedObject global(set->getContext()->getJSContext(),
                          set->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(set->getContext()->getJSContext(), global);

  JS::RootedObject setObj(set->getContext()->getJSContext(),
                          &set->getJSValue().toObject());
  if (!setObj) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }

  JS::RootedValue val(set->getContext()->getJSContext());
  if (!JS::SetKeys(set->getContext()->getJSContext(), setObj, &val)) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }

  Value *v = new Value(set->getContext(), val);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultValue SetObjectValues(ValuePtr set) {
  ResultValue result = {};

  JS::RootedObject global(set->getContext()->getJSContext(),
                          set->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(set->getContext()->getJSContext(), global);

  JS::RootedObject setObj(set->getContext()->getJSContext(),
                          &set->getJSValue().toObject());
  if (!setObj) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }

  JS::RootedValue val(set->getContext()->getJSContext());
  if (!JS::SetValues(set->getContext()->getJSContext(), setObj, &val)) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }

  Value *v = new Value(set->getContext(), val);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultValue SetObjectEntries(ValuePtr set) {
  ResultValue result = {};

  JS::RootedObject global(set->getContext()->getJSContext(),
                          set->getContext()->getGlobalJSObject());
  if (!global) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }
  JSAutoRealm ar(set->getContext()->getJSContext(), global);

  JS::RootedObject setObj(set->getContext()->getJSContext(),
                          &set->getJSValue().toObject());
  if (!setObj) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }

  JS::RootedValue val(set->getContext()->getJSContext());
  if (!JS::SetEntries(set->getContext()->getJSContext(), setObj, &val)) {
    result.err = GetError(set->getContext()->getJSContext());
    return result;
  }

  Value *v = new Value(set->getContext(), val);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultValue JSONParse(ContextPtr ctx, const char *data) {
  ResultValue result = {};

  JS::RootedObject global(ctx->getJSContext(), ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), global);

  JS::ConstUTF8CharsZ utf8Data(data, strlen(data));
  JS::RootedString str(ctx->getJSContext(),
                       JS_NewStringCopyUTF8Z(ctx->getJSContext(), utf8Data));
  if (!str) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  JS::RootedValue val(ctx->getJSContext());
  if (!JS_ParseJSON(ctx->getJSContext(), str, &val)) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  Value *v = new Value(ctx, val);
  if (!v) {
    return result;
  }

  result.ok = true;
  result.ptr = v;
  return result;
}

ResultString JSONStringify(ContextPtr ctx, ValuePtr value) {
  ResultString result = {};

  JS::RootedObject global(ctx->getJSContext(), ctx->getGlobalJSObject());
  if (!global) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }
  JSAutoRealm ar(ctx->getJSContext(), global);

  JS::RootedValue val(ctx->getJSContext(), value->getJSValue());

  std::u16string u16str;
  if (!JS_Stringify(ctx->getJSContext(), &val, nullptr, JS::NullHandleValue,
                    &StringifyCallback, &u16str)) {
    result.err = GetError(ctx->getJSContext());
    return result;
  }

  JS::RootedString str(
      ctx->getJSContext(),
      JS_NewUCStringCopyZ(ctx->getJSContext(), u16str.c_str()));

  size_t len = JS_GetStringEncodingLength(ctx->getJSContext(), str);
  char *data =
      static_cast<char *>(JS_malloc(value->getContext()->getJSContext(), len));
  if (!data) {
    result.err = GetError(value->getContext()->getJSContext());
    return result;
  }
  if (!JS_EncodeStringToBuffer(ctx->getJSContext(), str, data, len)) {
    result.err = GetError(value->getContext()->getJSContext());
    return result;
  }

  result.ok = true;
  result.data = data;
  result.len = len;
  return result;
}
