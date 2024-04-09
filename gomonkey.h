#ifndef GOMONKEY_H
#define GOMONKEY_H

#include <stdbool.h>
#include <stdint.h>

#ifdef __cplusplus

#include <js-config.h>
#include <js/experimental/CompileScript.h>
#include <js/experimental/JSStencil.h>
#include <jsapi.h>

extern "C" {
#endif

typedef struct Context Context;
typedef Context* ContextPtr;

struct ContextOptions {
  uint32_t heapMaxBytes;
  uint32_t stackSize;
  uint32_t gcMaxBytes;
  uint32_t gcIncrementalEnabled;
  uint32_t gcSliceTimeBudgetMs;
};
typedef struct ContextOptions ContextOptions;

typedef struct Script Script;
typedef Script* ScriptPtr;

typedef struct Value Value;
typedef Value* ValuePtr;

typedef struct FrontendContext FrontendContext;
typedef FrontendContext* FrontendContextPtr;

struct FrontendContextOptions {
  uint32_t stackSize;
};
typedef struct FrontendContextOptions FrontendContextOptions;

typedef struct Stencil Stencil;
typedef Stencil* StencilPtr;

struct Error {
  const char* message;
  const char* filename;
  int lineno;
  int number;
};
typedef struct Error Error;

struct Result {
  bool ok;
  Error err;
};
typedef struct Result Result;

struct ResultBool {
  bool ok;
  Error err;
  bool value;
};
typedef struct ResultBool ResultBool;

struct ResultUInt32 {
  bool ok;
  Error err;
  uint32_t value;
};
typedef struct ResultUInt32 ResultUInt32;

struct ResultValue {
  bool ok;
  Error err;
  ValuePtr ptr;
};
typedef struct ResultValue ResultValue;

struct ResultString {
  bool ok;
  Error err;
  const char* data;
  int len;
};
typedef struct ResultString ResultString;

struct ResultCompileScript {
  bool ok;
  Error err;
  ScriptPtr ptr;
};
typedef struct ResultCompileScript ResultCompileScript;

struct ResultCompileStencil {
  bool ok;
  StencilPtr ptr;
};
typedef struct ResultCompileStencil ResultCompileStencil;

struct ResultGoFunctionCallback {
  ValuePtr ptr;
  char* err;
};
typedef struct ResultGoFunctionCallback ResultGoFunctionCallback;

bool Init();
void ShutDown();
const char* Version();

ContextPtr NewContext(unsigned ref, ContextOptions options);
void DestroyContext(ContextPtr ctx);
void RequestInterruptContext(ContextPtr ctx);
ResultValue GetGlobalObject(ContextPtr ctx);
ResultValue DefineObject(ContextPtr ctx, ValuePtr recv, char* name,
                         unsigned attrs);
Result DefineProperty(ContextPtr ctx, ValuePtr recv, char* name, ValuePtr value,
                      unsigned attrs);
Result DefineElement(ContextPtr ctx, ValuePtr recv, uint32_t index,
                     ValuePtr value, unsigned attrs);
Result DefineFunction(ContextPtr ctx, ValuePtr recv, char* name, unsigned nargs,
                      unsigned attrs);
ResultValue CallFunctionName(ContextPtr ctx, char* name, ValuePtr recv,
                             int argc, ValuePtr* argv);
ResultValue CallFunctionValue(ContextPtr ctx, ValuePtr func, ValuePtr recv,
                              int argc, ValuePtr* argv);
ResultValue Evaluate(ContextPtr ctx, char* code);
ResultCompileScript CompileScript(ContextPtr ctx, char* filename, char* code);
void ReleaseScript(ScriptPtr script);
ResultValue ExecuteScript(ContextPtr ctx, ScriptPtr script);
ResultValue ExecuteScriptFromStencil(ContextPtr ctx, StencilPtr stencil);

FrontendContextPtr NewFrontendContext(FrontendContextOptions options);
void DestroyFrontendContext(FrontendContextPtr ctx);
ResultCompileStencil CompileScriptToStencil(FrontendContextPtr ctx,
                                            char* filename, char* code);
void ReleaseStencil(StencilPtr stencil);

ResultValue NewValueUndefined(ContextPtr ctx);
ResultValue NewValueNull(ContextPtr ctx);
ResultValue NewValueString(ContextPtr ctx, char* str, int len);
ResultValue NewValueBoolean(ContextPtr ctx, bool b);
ResultValue NewValueNumber(ContextPtr ctx, double d);
ResultValue NewValueInt32(ContextPtr ctx, int32_t i);
void ReleaseValue(ValuePtr value);
ResultString ToString(ValuePtr value);
bool ValueIs(ValuePtr value1, ValuePtr value2);
bool ValueIsUndefined(ValuePtr value);
bool ValueIsNull(ValuePtr value);
bool ValueIsNullOrUndefined(ValuePtr value);
bool ValueIsTrue(ValuePtr value);
bool ValueIsFalse(ValuePtr value);
bool ValueIsObject(ValuePtr value);
bool ValueIsFunction(ValuePtr value);
bool ValueIsSymbol(ValuePtr value);
bool ValueIsString(ValuePtr value);
bool ValueIsBoolean(ValuePtr value);
bool ValueIsNumber(ValuePtr value);
bool ValueIsInt32(ValuePtr value);
ResultString ValueToString(ValuePtr value);
int ValueToBoolean(ValuePtr value);
double ValueToNumber(ValuePtr value);
int32_t ValueToInt32(ValuePtr value);

ResultValue NewPlainObject(ContextPtr ctx);
ResultBool ObjectHasProperty(ValuePtr object, char* key);
ResultValue ObjectGetProperty(ValuePtr object, char* key);
Result ObjectSetProperty(ValuePtr object, char* key, ValuePtr value);
Result ObjectDeleteProperty(ValuePtr object, char* key);
ResultBool ObjectHasElement(ValuePtr object, uint32_t index);
ResultValue ObjectGetElement(ValuePtr object, uint32_t index);
Result ObjectSetElement(ValuePtr object, uint32_t index, ValuePtr value);
Result ObjectDeleteElement(ValuePtr object, uint32_t index);

ResultValue NewFunction(ContextPtr ctx, char* name);

ResultValue NewArrayObject(ContextPtr ctx, int argc, ValuePtr* argv);
ResultUInt32 GetArrayObjectLength(ValuePtr array);

ResultValue NewMapObject(ContextPtr ctx);
ResultUInt32 MapObjectSize(ValuePtr map);
ResultBool MapObjectHas(ValuePtr map, ValuePtr key);
ResultValue MapObjectGet(ValuePtr map, ValuePtr key);
Result MapObjectSet(ValuePtr map, ValuePtr key, ValuePtr val);
ResultBool MapObjectDelete(ValuePtr map, ValuePtr key);
Result MapObjectClear(ValuePtr map);
ResultValue MapObjectKeys(ValuePtr map);
ResultValue MapObjectValues(ValuePtr map);
ResultValue MapObjectEntries(ValuePtr map);

ResultValue NewSetObject(ContextPtr ctx);
ResultUInt32 SetObjectSize(ValuePtr set);
ResultBool SetObjectHas(ValuePtr set, ValuePtr key);
Result SetObjectAdd(ValuePtr set, ValuePtr val);
ResultBool SetObjectDelete(ValuePtr set, ValuePtr key);
Result SetObjectClear(ValuePtr set);
ResultValue SetObjectKeys(ValuePtr set);
ResultValue SetObjectValues(ValuePtr set);
ResultValue SetObjectEntries(ValuePtr set);

ResultValue JSONParse(ContextPtr ctx, const char* data);
ResultString JSONStringify(ContextPtr ctx, ValuePtr value);

#ifdef __cplusplus
}  // extern "C"
#endif

#endif  // GOMONKEY_H
