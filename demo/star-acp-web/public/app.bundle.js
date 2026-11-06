var __defProp = Object.defineProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};

// node_modules/@agentclientprotocol/sdk/dist/v1/schema/index.js
var AGENT_METHODS = {
  initialize: "initialize",
  authenticate: "authenticate",
  session_new: "session/new",
  session_load: "session/load",
  session_set_mode: "session/set_mode",
  session_set_config_option: "session/set_config_option",
  session_prompt: "session/prompt",
  session_cancel: "session/cancel",
  session_list: "session/list",
  session_delete: "session/delete",
  session_resume: "session/resume",
  session_close: "session/close",
  logout: "logout"
};
var CLIENT_METHODS = {
  session_request_permission: "session/request_permission",
  session_update: "session/update",
  fs_write_text_file: "fs/write_text_file",
  fs_read_text_file: "fs/read_text_file",
  terminal_create: "terminal/create",
  terminal_output: "terminal/output",
  terminal_release: "terminal/release",
  terminal_wait_for_exit: "terminal/wait_for_exit",
  terminal_kill: "terminal/kill",
  elicitation_create: "elicitation/create",
  elicitation_complete: "elicitation/complete"
};
var PROTOCOL_METHODS = {
  cancel_request: "$/cancel_request"
};
var PROTOCOL_VERSION = 1;

// node_modules/zod/v4/core/util.js
var util_exports = {};
__export(util_exports, {
  BIGINT_FORMAT_RANGES: () => BIGINT_FORMAT_RANGES,
  CONSTANT_CATCH: () => CONSTANT_CATCH,
  Class: () => Class,
  NUMBER_FORMAT_RANGES: () => NUMBER_FORMAT_RANGES,
  aborted: () => aborted,
  allowsEval: () => allowsEval,
  assert: () => assert,
  assertEqual: () => assertEqual,
  assertIs: () => assertIs,
  assertNever: () => assertNever,
  assertNotEqual: () => assertNotEqual,
  assignProp: () => assignProp,
  attachSchema: () => attachSchema,
  base64ToUint8Array: () => base64ToUint8Array,
  base64urlToUint8Array: () => base64urlToUint8Array,
  cached: () => cached,
  captureStackTrace: () => captureStackTrace,
  cleanEnum: () => cleanEnum,
  cleanRegex: () => cleanRegex,
  clone: () => clone,
  cloneDef: () => cloneDef,
  codePointLength: () => codePointLength,
  constantCatch: () => constantCatch,
  createTransparentProxy: () => createTransparentProxy,
  defineLazy: () => defineLazy,
  defineLazyInternal: () => defineLazyInternal,
  derived: () => derived,
  esc: () => esc,
  escapeRegex: () => escapeRegex,
  explicitlyAborted: () => explicitlyAborted,
  extend: () => extend,
  finalizeIssue: () => finalizeIssue,
  floatSafeRemainder: () => floatSafeRemainder,
  getElementAtPath: () => getElementAtPath,
  getEnumValues: () => getEnumValues,
  getLengthableOrigin: () => getLengthableOrigin,
  getParsedType: () => getParsedType,
  getSizableOrigin: () => getSizableOrigin,
  hexToUint8Array: () => hexToUint8Array,
  hide: () => hide,
  installLazyProp: () => installLazyProp,
  isObject: () => isObject,
  isPlainObject: () => isPlainObject,
  issue: () => issue,
  joinValues: () => joinValues,
  jsonStringifyReplacer: () => jsonStringifyReplacer,
  members: () => members,
  merge: () => merge,
  mergeDefs: () => mergeDefs,
  normalizeParams: () => normalizeParams,
  nullish: () => nullish,
  numKeys: () => numKeys,
  objectClone: () => objectClone,
  omit: () => omit,
  optionalKeys: () => optionalKeys,
  own: () => own,
  parsedType: () => parsedType,
  partial: () => partial,
  pick: () => pick,
  prefixIssues: () => prefixIssues,
  primitiveTypes: () => primitiveTypes,
  promiseAllObject: () => promiseAllObject,
  propertyKeyTypes: () => propertyKeyTypes,
  randomString: () => randomString,
  rawShape: () => rawShape,
  required: () => required,
  safeExtend: () => safeExtend,
  shallowClone: () => shallowClone,
  sizeCaps: () => sizeCaps,
  slugify: () => slugify,
  stringifyPrimitive: () => stringifyPrimitive,
  toZod: () => toZod,
  uint8ArrayToBase64: () => uint8ArrayToBase64,
  uint8ArrayToBase64url: () => uint8ArrayToBase64url,
  uint8ArrayToHex: () => uint8ArrayToHex,
  unwrapMessage: () => unwrapMessage
});
function assertEqual(val) {
  return val;
}
function assertNotEqual(val) {
  return val;
}
function toZod() {
  return (schema) => schema;
}
function assertIs(_arg) {
}
function assertNever(_x) {
  throw new Error("Unexpected value in exhaustive check");
}
function assert(_) {
}
function getEnumValues(entries) {
  const numericValues = Object.values(entries).filter((v) => typeof v === "number");
  const values = Object.entries(entries).filter(([k, _]) => numericValues.indexOf(+k) === -1).map(([_, v]) => v);
  return values;
}
function joinValues(array2, separator = "|") {
  return array2.map((val) => stringifyPrimitive(val)).join(separator);
}
function jsonStringifyReplacer(_, value) {
  if (typeof value === "bigint")
    return value.toString();
  return value;
}
var Cached = class {
  constructor(getter) {
    this._getter = getter;
    this._value = void 0;
  }
  get value() {
    const getter = this._getter;
    if (getter !== void 0) {
      this._value = getter();
      this._getter = void 0;
    }
    return this._value;
  }
};
function cached(getter) {
  return new Cached(getter);
}
function nullish(input) {
  return input === null || input === void 0;
}
function cleanRegex(source) {
  const start = source.startsWith("^") ? 1 : 0;
  const end = source.endsWith("$") ? source.length - 1 : source.length;
  return source.slice(start, end);
}
function floatSafeRemainder(val, step) {
  const ratio = val / step;
  const roundedRatio = Math.round(ratio);
  const tolerance = 4 * Number.EPSILON * Math.max(Math.abs(ratio), 1);
  if (Math.abs(ratio - roundedRatio) < tolerance)
    return 0;
  return ratio - roundedRatio;
}
var EVALUATING = /* @__PURE__ */ Symbol("evaluating");
function defineLazy(object2, key, getter) {
  let value = void 0;
  Object.defineProperty(object2, key, {
    get() {
      if (value === EVALUATING) {
        return void 0;
      }
      if (value === void 0) {
        value = EVALUATING;
        value = getter();
      }
      return value;
    },
    set(v) {
      Object.defineProperty(object2, key, {
        value: v
        // configurable: true,
      });
    },
    configurable: true
  });
}
function objectClone(obj) {
  return Object.create(Object.getPrototypeOf(obj), Object.getOwnPropertyDescriptors(obj));
}
function assignProp(target, prop, value) {
  Object.defineProperty(target, prop, {
    value,
    writable: true,
    enumerable: true,
    configurable: true
  });
}
function rawShape(def2) {
  const desc = Object.getOwnPropertyDescriptor(def2, "shape");
  return desc?.get ? desc.get.raw : desc?.value;
}
function sourceShape(schema) {
  return rawShape(schema._zod.def) ?? schema._zod.def.shape;
}
function deferProp(target, key, getter) {
  Object.defineProperty(target, key, {
    get() {
      const value = getter();
      assignProp(this, key, value);
      return value;
    },
    enumerable: true,
    configurable: true
  });
}
function putProp(target, key, value) {
  if (key in target)
    assignProp(target, key, value);
  else
    target[key] = value;
}
function mirrorShape(target, source, keys, wrap) {
  const raw = sourceShape(source);
  for (const key of keys) {
    const desc = Object.getOwnPropertyDescriptor(raw, key);
    if (!desc.enumerable)
      continue;
    if (desc.get) {
      deferProp(target, key, () => {
        const value = source._zod.def.shape[key];
        return wrap ? wrap(value, key) : value;
      });
    } else
      putProp(target, key, wrap ? wrap(desc.value, key) : desc.value);
  }
}
function mirrorProps(target, source) {
  for (const key of Reflect.ownKeys(source)) {
    const desc = Object.getOwnPropertyDescriptor(source, key);
    if (!desc.enumerable)
      continue;
    if (desc.get)
      deferProp(target, key, () => source[key]);
    else
      putProp(target, key, desc.value);
  }
}
function mergeDefs(...defs) {
  const mergedDescriptors = {};
  for (const def2 of defs) {
    const descriptors = Object.getOwnPropertyDescriptors(def2);
    Object.assign(mergedDescriptors, descriptors);
  }
  return Object.defineProperties({}, mergedDescriptors);
}
function cloneDef(schema) {
  return mergeDefs(schema._zod.def);
}
function getElementAtPath(obj, path) {
  if (!path)
    return obj;
  return path.reduce((acc, key) => acc?.[key], obj);
}
function promiseAllObject(promisesObj) {
  const keys = Object.keys(promisesObj);
  const promises = keys.map((key) => promisesObj[key]);
  return Promise.all(promises).then((results) => {
    const resolvedObj = {};
    for (let i = 0; i < keys.length; i++) {
      resolvedObj[keys[i]] = results[i];
    }
    return resolvedObj;
  });
}
function randomString(length = 10) {
  const chars = "abcdefghijklmnopqrstuvwxyz";
  let str = "";
  for (let i = 0; i < length; i++) {
    str += chars[Math.floor(Math.random() * chars.length)];
  }
  return str;
}
function esc(str) {
  return JSON.stringify(str);
}
function slugify(input) {
  return input.toLowerCase().trim().replace(/[^\w\s-]/g, "").replace(/[\s_-]+/g, "-").replace(/^-+|-+$/g, "");
}
var captureStackTrace = "captureStackTrace" in Error ? Error.captureStackTrace : (..._args) => {
};
function isObject(data) {
  return typeof data === "object" && data !== null && !Array.isArray(data);
}
var allowsEval = /* @__PURE__ */ cached(() => {
  if (globalConfig.jitless) {
    return false;
  }
  if (typeof navigator !== "undefined" && navigator?.userAgent?.includes("Cloudflare")) {
    return false;
  }
  try {
    const F = Function;
    new F("");
    return true;
  } catch (_) {
    return false;
  }
});
function isPlainObject(o) {
  if (isObject(o) === false)
    return false;
  const ctor = o.constructor;
  if (ctor === void 0)
    return true;
  if (typeof ctor !== "function")
    return true;
  const prot = ctor.prototype;
  if (isObject(prot) === false)
    return false;
  if (Object.prototype.hasOwnProperty.call(prot, "isPrototypeOf") === false) {
    return false;
  }
  return true;
}
function shallowClone(o) {
  if (isPlainObject(o))
    return { ...o };
  if (Array.isArray(o))
    return [...o];
  if (o instanceof Map)
    return new Map(o);
  if (o instanceof Set)
    return new Set(o);
  return o;
}
function numKeys(data) {
  let keyCount = 0;
  for (const key in data) {
    if (Object.prototype.hasOwnProperty.call(data, key)) {
      keyCount++;
    }
  }
  return keyCount;
}
var getParsedType = (data) => {
  const t = typeof data;
  switch (t) {
    case "undefined":
      return "undefined";
    case "string":
      return "string";
    case "number":
      return Number.isNaN(data) ? "nan" : "number";
    case "boolean":
      return "boolean";
    case "function":
      return "function";
    case "bigint":
      return "bigint";
    case "symbol":
      return "symbol";
    case "object":
      if (Array.isArray(data)) {
        return "array";
      }
      if (data === null) {
        return "null";
      }
      if (data.then && typeof data.then === "function" && data.catch && typeof data.catch === "function") {
        return "promise";
      }
      if (typeof Map !== "undefined" && data instanceof Map) {
        return "map";
      }
      if (typeof Set !== "undefined" && data instanceof Set) {
        return "set";
      }
      if (typeof Date !== "undefined" && data instanceof Date) {
        return "date";
      }
      if (typeof File !== "undefined" && data instanceof File) {
        return "file";
      }
      return "object";
    default:
      throw new Error(`Unknown data type: ${t}`);
  }
};
var propertyKeyTypes = /* @__PURE__ */ new Set(["string", "number", "symbol"]);
var primitiveTypes = /* @__PURE__ */ new Set([
  "string",
  "number",
  "bigint",
  "boolean",
  "symbol",
  "undefined"
]);
function escapeRegex(str) {
  return str.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
function clone(inst, def2, params) {
  const cl = new inst._zod.constr(def2 ?? inst._zod.def);
  if (!def2 || params?.parent)
    cl._zod.parent = inst;
  return cl;
}
function normalizeParams(_params) {
  const params = _params;
  if (!params)
    return {};
  if (typeof params === "string")
    return { error: () => params };
  if (params?.message !== void 0) {
    if (params?.error !== void 0)
      throw new Error("Cannot specify both `message` and `error` params");
    params.error = params.message;
  }
  delete params.message;
  if (typeof params.error === "string")
    return { ...params, error: () => params.error };
  return params;
}
function createTransparentProxy(getter) {
  let target;
  return new Proxy({}, {
    get(_, prop, receiver) {
      target ?? (target = getter());
      return Reflect.get(target, prop, receiver);
    },
    set(_, prop, value, receiver) {
      target ?? (target = getter());
      return Reflect.set(target, prop, value, receiver);
    },
    has(_, prop) {
      target ?? (target = getter());
      return Reflect.has(target, prop);
    },
    deleteProperty(_, prop) {
      target ?? (target = getter());
      return Reflect.deleteProperty(target, prop);
    },
    ownKeys(_) {
      target ?? (target = getter());
      return Reflect.ownKeys(target);
    },
    getOwnPropertyDescriptor(_, prop) {
      target ?? (target = getter());
      return Reflect.getOwnPropertyDescriptor(target, prop);
    },
    defineProperty(_, prop, descriptor) {
      target ?? (target = getter());
      return Reflect.defineProperty(target, prop, descriptor);
    }
  });
}
function stringifyPrimitive(value) {
  if (typeof value === "bigint")
    return value.toString() + "n";
  if (typeof value === "string")
    return `"${value}"`;
  return `${value}`;
}
function optionalKeys(shape) {
  return Object.keys(shape).filter((k) => {
    return shape[k]._zod.optin !== void 0 && shape[k]._zod.optout === "optional";
  });
}
var NUMBER_FORMAT_RANGES = /* @__PURE__ */ (() => ({
  safeint: [Number.MIN_SAFE_INTEGER, Number.MAX_SAFE_INTEGER],
  int32: [-2147483648, 2147483647],
  uint32: [0, 4294967295],
  float32: [-34028234663852886e22, 34028234663852886e22],
  float64: [-Number.MAX_VALUE, Number.MAX_VALUE]
}))();
var BIGINT_FORMAT_RANGES = {
  int64: [/* @__PURE__ */ BigInt("-9223372036854775808"), /* @__PURE__ */ BigInt("9223372036854775807")],
  uint64: [/* @__PURE__ */ BigInt(0), /* @__PURE__ */ BigInt("18446744073709551615")]
};
function pick(schema, mask) {
  const currDef = schema._zod.def;
  const checks = currDef.checks;
  const hasChecks = checks && checks.length > 0;
  if (hasChecks) {
    throw new Error(".pick() cannot be used on object schemas containing refinements");
  }
  const newShape = {};
  mirrorShape(newShape, schema, maskedKeys(schema, mask));
  return clone(schema, mergeDefs(currDef, { shape: newShape, checks: [] }));
}
function maskedKeys(schema, mask) {
  const raw = sourceShape(schema);
  const keys = [];
  for (const key of Reflect.ownKeys(mask)) {
    if (!Object.getOwnPropertyDescriptor(raw, key)?.enumerable) {
      throw new Error(`Unrecognized key: "${String(key)}"`);
    }
    if (mask[key])
      keys.push(key);
  }
  return keys;
}
function omit(schema, mask) {
  const currDef = schema._zod.def;
  const checks = currDef.checks;
  const hasChecks = checks && checks.length > 0;
  if (hasChecks) {
    throw new Error(".omit() cannot be used on object schemas containing refinements");
  }
  const omitted = new Set(maskedKeys(schema, mask));
  const newShape = {};
  mirrorShape(newShape, schema, Reflect.ownKeys(sourceShape(schema)).filter((key) => !omitted.has(key)));
  return clone(schema, mergeDefs(currDef, { shape: newShape, checks: [] }));
}
function extend(schema, shape) {
  if (!isPlainObject(shape)) {
    throw new Error("Invalid input to extend: expected a plain object");
  }
  const checks = schema._zod.def.checks;
  const hasChecks = checks && checks.length > 0;
  if (hasChecks) {
    const existingShape = sourceShape(schema);
    for (const key of Reflect.ownKeys(shape)) {
      if (Object.getOwnPropertyDescriptor(existingShape, key) !== void 0) {
        throw new Error("Cannot overwrite keys on object schemas containing refinements. Use `.safeExtend()` instead.");
      }
    }
  }
  return clone(schema, mergeDefs(schema._zod.def, { shape: extended(schema, shape) }));
}
function extended(schema, shape) {
  const newShape = {};
  mirrorShape(newShape, schema, Reflect.ownKeys(sourceShape(schema)));
  mirrorProps(newShape, shape);
  return newShape;
}
function safeExtend(schema, shape) {
  if (!isPlainObject(shape)) {
    throw new Error("Invalid input to safeExtend: expected a plain object");
  }
  return clone(schema, mergeDefs(schema._zod.def, { shape: extended(schema, shape) }));
}
function merge(a, b) {
  if (!b?._zod?.def) {
    throw new Error("Invalid input to merge: expected an object schema. To merge a plain shape, use `.extend()`.");
  }
  if (a._zod.def.checks?.length) {
    throw new Error(".merge() cannot be used on object schemas containing refinements. Use .safeExtend() instead.");
  }
  const newShape = {};
  mirrorShape(newShape, a, Reflect.ownKeys(sourceShape(a)));
  mirrorShape(newShape, b, Reflect.ownKeys(sourceShape(b)));
  const def2 = mergeDefs(a._zod.def, {
    shape: newShape,
    get catchall() {
      return b._zod.def.catchall;
    },
    checks: b._zod.def.checks ?? []
  });
  return clone(a, def2);
}
function partial(Class2, schema, mask, name = "partial") {
  const currDef = schema._zod.def;
  const checks = currDef.checks;
  const hasChecks = checks && checks.length > 0;
  if (hasChecks) {
    throw new Error(`.${name}() cannot be used on object schemas containing refinements`);
  }
  const selected = mask ? new Set(maskedKeys(schema, mask)) : void 0;
  const newShape = {};
  mirrorShape(newShape, schema, Reflect.ownKeys(sourceShape(schema)), Class2 && ((value, key) => selected && !selected.has(key) ? value : new Class2({ type: "optional", innerType: value })));
  return clone(schema, mergeDefs(schema._zod.def, { shape: newShape, checks: [] }));
}
function required(Class2, schema, mask) {
  const selected = mask ? new Set(maskedKeys(schema, mask)) : void 0;
  const newShape = {};
  mirrorShape(newShape, schema, Reflect.ownKeys(sourceShape(schema)), (value, key) => (
    // overwrite with non-optional
    selected && !selected.has(key) ? value : new Class2({ type: "nonoptional", innerType: value })
  ));
  return clone(schema, mergeDefs(schema._zod.def, { shape: newShape }));
}
function aborted(x, startIndex = 0) {
  if (x.aborted === true)
    return true;
  for (let i = startIndex; i < x.issues.length; i++) {
    if (x.issues[i]?.continue !== true) {
      return true;
    }
  }
  return false;
}
function explicitlyAborted(x, startIndex = 0) {
  if (x.aborted === true)
    return true;
  for (let i = startIndex; i < x.issues.length; i++) {
    if (x.issues[i]?.continue === false) {
      return true;
    }
  }
  return false;
}
function prefixIssues(path, issues) {
  return issues.map((iss) => {
    var _a3;
    (_a3 = iss).path ?? (_a3.path = []);
    iss.path.unshift(path);
    return iss;
  });
}
function unwrapMessage(message) {
  return typeof message === "string" ? message : message?.message;
}
function sizeCaps(defChecks) {
  const caps = [];
  for (const ch of defChecks ?? []) {
    const d = ch._zod.def;
    if (!/_(length|size)$|_equals$/.test(d.check))
      break;
    const max = d.maximum ?? d.length ?? d.size;
    if (max >= 0)
      caps.push([max, ch]);
  }
  return caps;
}
function attachSchema(issues, start, inst) {
  var _a3;
  for (let i = start; i < issues.length; i++) {
    (_a3 = issues[i]).schema ?? (_a3.schema = inst);
  }
}
function finalizeIssue(iss, ctx, config2) {
  var _a3;
  const traits = iss.inst?._zod?.traits;
  if (traits?.has("$ZodType")) {
    if (traits.has("$ZodCheck"))
      (_a3 = iss).schema ?? (_a3.schema = iss.inst);
    else
      iss.schema = iss.inst;
  }
  const schemaError = iss.schema !== iss.inst ? iss.schema?._zod.def?.error : void 0;
  const message = iss.message ? iss.message : unwrapMessage(iss.inst?._zod.def?.error?.(iss)) ?? unwrapMessage(schemaError?.(iss)) ?? unwrapMessage(ctx?.error?.(iss)) ?? unwrapMessage(config2.customError?.(iss)) ?? unwrapMessage(config2.localeError?.(iss)) ?? "Invalid input";
  const full = {};
  for (const k of Object.keys(iss)) {
    if (k === "inst" || k === "schema" || k === "continue" || k === "input" || k === "__proto__")
      continue;
    full[k] = iss[k];
  }
  full.path ?? (full.path = []);
  full.message = message;
  if (ctx?.reportInput) {
    full.input = iss.input;
  }
  return full;
}
function getSizableOrigin(input) {
  if (input instanceof Set)
    return "set";
  if (input instanceof Map)
    return "map";
  if (input instanceof File)
    return "file";
  return "unknown";
}
var highSurrogate = /[\uD800-\uDBFF]/;
function codePointLength(str) {
  const units = str.length;
  if (!highSurrogate.test(str))
    return units;
  let count = units;
  for (let i = 0; i < units - 1; i++) {
    if ((str.charCodeAt(i) & 64512) === 55296 && (str.charCodeAt(i + 1) & 64512) === 56320) {
      count--;
      i++;
    }
  }
  return count;
}
function getLengthableOrigin(input) {
  if (Array.isArray(input))
    return "array";
  if (typeof input === "string")
    return "string";
  return "unknown";
}
function parsedType(data) {
  const t = typeof data;
  switch (t) {
    case "number": {
      return Number.isNaN(data) ? "nan" : "number";
    }
    case "object": {
      if (data === null) {
        return "null";
      }
      if (Array.isArray(data)) {
        return "array";
      }
      const obj = data;
      if (obj && Object.getPrototypeOf(obj) !== Object.prototype && "constructor" in obj && obj.constructor) {
        return obj.constructor.name;
      }
    }
  }
  return t;
}
function issue(...args) {
  const [iss, input, inst] = args;
  if (typeof iss === "string") {
    return {
      message: iss,
      code: "custom",
      input,
      inst
    };
  }
  return { ...iss };
}
function cleanEnum(obj) {
  return Object.entries(obj).filter(([k, _]) => {
    return Number.isNaN(Number.parseInt(k, 10));
  }).map((el) => el[1]);
}
function base64ToUint8Array(base642) {
  const binaryString = atob(base642);
  const bytes = new Uint8Array(binaryString.length);
  for (let i = 0; i < binaryString.length; i++) {
    bytes[i] = binaryString.charCodeAt(i);
  }
  return bytes;
}
function uint8ArrayToBase64(bytes) {
  let binaryString = "";
  for (let i = 0; i < bytes.length; i++) {
    binaryString += String.fromCharCode(bytes[i]);
  }
  return btoa(binaryString);
}
function base64urlToUint8Array(base64url2) {
  const base642 = base64url2.replace(/-/g, "+").replace(/_/g, "/");
  const padding = "=".repeat((4 - base642.length % 4) % 4);
  return base64ToUint8Array(base642 + padding);
}
function uint8ArrayToBase64url(bytes) {
  return uint8ArrayToBase64(bytes).replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
}
function hexToUint8Array(hex) {
  const cleanHex = hex.replace(/^0x/, "");
  if (cleanHex.length % 2 !== 0) {
    throw new Error("Invalid hex string length");
  }
  const bytes = new Uint8Array(cleanHex.length / 2);
  for (let i = 0; i < cleanHex.length; i += 2) {
    bytes[i / 2] = Number.parseInt(cleanHex.slice(i, i + 2), 16);
  }
  return bytes;
}
function uint8ArrayToHex(bytes) {
  return Array.from(bytes).map((b) => b.toString(16).padStart(2, "0")).join("");
}
var Class = class {
  constructor(..._args) {
  }
};
function members(proto, table) {
  for (const key in table) {
    const desc = Object.getOwnPropertyDescriptor(table, key);
    if (desc.get)
      Object.defineProperty(proto, key, { ...desc, enumerable: false });
    else
      defineBound(proto, key, desc.value);
  }
}
function own(inst, key, value, enumerable = true) {
  Object.defineProperty(inst, key, { configurable: true, writable: true, enumerable, value });
  return value;
}
function hide(inst, key, value) {
  return own(inst, key, value, false);
}
// @__NO_SIDE_EFFECTS__
function derived(computes, table) {
  for (const key in computes) {
    const compute = computes[key];
    Object.defineProperty(table, key, {
      configurable: true,
      enumerable: true,
      get() {
        return own(this, key, compute(this));
      },
      set(value) {
        own(this, key, value);
      }
    });
  }
  return table;
}
function defineBound(proto, key, fn) {
  Object.defineProperty(proto, key, {
    configurable: true,
    get() {
      return this == null ? fn : own(this, key, fn.bind(this));
    },
    set(value) {
      own(this, key, value);
    }
  });
}
function claim(inst, sentinel) {
  const proto = Object.getPrototypeOf(inst);
  return sentinel in proto ? void 0 : proto;
}
var installing;
var broke = false;
var breaker = {
  configurable: true,
  get() {
    broke = true;
    return void 0;
  }
};
function defineLazyInternal(inst, key, compute) {
  const proto = Object.getPrototypeOf(inst._zod);
  if (key in proto && installing !== inst._zod) {
    installing = void 0;
    return;
  }
  installing = inst._zod;
  Object.defineProperty(proto, key, {
    configurable: true,
    get() {
      Object.defineProperty(this, key, breaker);
      const outer = broke;
      broke = false;
      try {
        const value = compute(this);
        if (broke)
          delete this[key];
        else
          Object.defineProperty(this, key, { configurable: true, writable: true, value });
        broke = broke || outer;
        return value;
      } catch (err) {
        delete this[key];
        broke = broke || outer;
        throw err;
      }
    },
    set(value) {
      Object.defineProperty(this, key, { configurable: true, writable: true, value });
    }
  });
}
function installLazyProp(inst, key, make, enumerable) {
  const proto = claim(inst, key);
  if (!proto)
    return;
  Object.defineProperty(proto, key, {
    configurable: true,
    get() {
      const desc = { configurable: true, writable: true, enumerable, value: void 0 };
      Object.defineProperty(this, key, desc);
      desc.value = make(this);
      Object.defineProperty(this, key, desc);
      return desc.value;
    },
    set(value) {
      Object.defineProperty(this, key, { configurable: true, writable: true, enumerable, value });
    }
  });
}
var CONSTANT_CATCH = "~constantCatch";
function constantCatch(value) {
  const fn = () => value;
  fn[CONSTANT_CATCH] = true;
  return fn;
}

// node_modules/zod/v4/core/core.js
var _a;
var NEVER = /* @__PURE__ */ Object.freeze({
  status: "aborted"
});
var _zodDesc = { value: void 0, enumerable: false };
var _E = "captureStackTrace" in Error ? Error : null;
function newError(Definition) {
  const E = _E;
  if (E) {
    const saved = E.stackTraceLimit;
    if (typeof saved === "number") {
      try {
        E.stackTraceLimit = 0;
      } catch {
        _E = null;
        return new Definition();
      }
      try {
        return new Definition();
      } finally {
        E.stackTraceLimit = saved;
      }
    }
  }
  return new Definition();
}
// @__NO_SIDE_EFFECTS__
function $constructor(name, initializer3, proto, params) {
  const zodProto = {};
  function Internals(def2) {
    this.def = def2;
    this.constr = _;
    this.traits = /* @__PURE__ */ new Set();
  }
  Internals.prototype = zodProto;
  const protoMembers = proto;
  const initialized = protoMembers && /* @__PURE__ */ new WeakSet();
  function init(inst, def2) {
    if (!inst._zod) {
      _zodDesc.value = new Internals(def2);
      try {
        Object.defineProperty(inst, "_zod", _zodDesc);
      } finally {
        _zodDesc.value = void 0;
      }
    } else if (inst._zod.traits.has(name)) {
      return;
    }
    inst._zod.traits.add(name);
    initializer3(inst, def2);
    if (initialized) {
      const own2 = Object.getPrototypeOf(inst);
      const ctorProto = inst._zod.constr.prototype;
      let up = own2;
      while (up && up !== ctorProto)
        up = Object.getPrototypeOf(up);
      const target = up ?? own2;
      if (!initialized.has(target)) {
        initialized.add(target);
        members(target, protoMembers);
      }
    }
    const proto2 = _.prototype;
    for (const k in proto2) {
      if (!Object.prototype.hasOwnProperty.call(proto2, k))
        continue;
      if (!(k in inst)) {
        inst[k] = proto2[k].bind(inst);
      }
    }
  }
  const Parent = params?.Parent ?? Object;
  class Definition extends Parent {
  }
  Object.defineProperty(Definition, "name", { value: name });
  function _(def2) {
    const inst = params?.Parent ? newError(Definition) : this;
    init(inst, def2);
    const deferred = inst._zod.deferred;
    if (deferred) {
      for (const fn of deferred) {
        fn();
      }
      inst._zod.deferred = void 0;
    }
    const pp = globalThis.__zod_globalConfig?.postProcessor;
    if (pp)
      pp(inst);
    return inst;
  }
  Object.defineProperty(_, "init", { value: init });
  Object.defineProperty(_, Symbol.hasInstance, {
    value: (inst) => {
      if (params?.Parent && inst instanceof params.Parent)
        return true;
      return inst?._zod?.traits?.has(name);
    }
  });
  Object.defineProperty(_, "name", { value: name });
  return _;
}
var $ZodAsyncError = class extends Error {
  constructor() {
    super(`Encountered Promise during synchronous parse. Use .parseAsync() instead.`);
  }
};
var $ZodEncodeError = class extends Error {
  constructor(name) {
    super(`Encountered unidirectional transform during encode: ${name}`);
    this.name = "ZodEncodeError";
  }
};
(_a = globalThis).__zod_globalConfig ?? (_a.__zod_globalConfig = {});
var globalConfig = globalThis.__zod_globalConfig;
function config(newConfig) {
  if (newConfig)
    Object.assign(globalConfig, newConfig);
  return globalConfig;
}

// node_modules/zod/v4/core/errors.js
function _getMessage() {
  const internals = this._zod;
  internals.message ?? (internals.message = JSON.stringify(internals.def, jsonStringifyReplacer, 2));
  return internals.message;
}
function _setMessage(value) {
  this._zod.message = value;
}
var _messageDesc = {
  get: _getMessage,
  set: _setMessage,
  enumerable: true,
  configurable: true
};
var _issuesDesc = { value: void 0, enumerable: false };
var _installedToString = /* @__PURE__ */ new WeakSet([Object.prototype, Error.prototype]);
var initializer = (inst, def2) => {
  inst.name = "$ZodError";
  _issuesDesc.value = def2;
  Object.defineProperty(inst, "issues", _issuesDesc);
  _issuesDesc.value = void 0;
  Object.defineProperty(inst, "message", _messageDesc);
  const proto = Object.getPrototypeOf(inst);
  if (!_installedToString.has(proto)) {
    _installedToString.add(proto);
    Object.defineProperty(proto, "toString", {
      configurable: true,
      enumerable: false,
      get() {
        const value = () => this.message;
        Object.defineProperty(this, "toString", { value, configurable: true, writable: true });
        return value;
      },
      set(value) {
        Object.defineProperty(this, "toString", { value, configurable: true, writable: true });
      }
    });
  }
};
var $ZodError = $constructor("$ZodError", initializer);
var $ZodRealError = $constructor("$ZodError", initializer, void 0, {
  Parent: Error
});
function node(obj, key, make) {
  if (!Object.prototype.hasOwnProperty.call(obj, key)) {
    if (key === "__proto__") {
      Object.defineProperty(obj, key, { value: make(), writable: true, enumerable: true, configurable: true });
    } else {
      obj[key] = make();
    }
  }
  return obj[key];
}
function flattenError(error2, mapper = (issue2) => issue2.message) {
  const fieldErrors = {};
  const formErrors = [];
  for (const sub of error2.issues) {
    if (sub.path.length > 0) {
      node(fieldErrors, sub.path[0], () => []).push(mapper(sub));
    } else {
      formErrors.push(mapper(sub));
    }
  }
  return { formErrors, fieldErrors };
}
function formatError(error2, mapper = (issue2) => issue2.message) {
  const fieldErrors = { _errors: [] };
  const processError = (error3, path = []) => {
    for (const issue2 of error3.issues) {
      if (issue2.code === "invalid_union" && issue2.errors.length) {
        issue2.errors.map((issues) => processError({ issues }, [...path, ...issue2.path]));
      } else if (issue2.code === "invalid_key") {
        processError({ issues: issue2.issues }, [...path, ...issue2.path]);
      } else if (issue2.code === "invalid_element") {
        processError({ issues: issue2.issues }, [...path, ...issue2.path]);
      } else {
        const fullpath = [...path, ...issue2.path];
        if (fullpath.length === 0) {
          fieldErrors._errors.push(mapper(issue2));
        } else {
          let curr = fieldErrors;
          let i = 0;
          while (i < fullpath.length) {
            const el = fullpath[i];
            const terminal = i === fullpath.length - 1;
            if (el === "_errors") {
              if (terminal)
                curr._errors.push(mapper(issue2));
              i++;
              continue;
            }
            if (!Object.prototype.hasOwnProperty.call(curr, el)) {
              Object.defineProperty(curr, el, {
                value: { _errors: [] },
                enumerable: true,
                writable: true,
                configurable: true
              });
            }
            const node2 = curr[el];
            if (terminal) {
              node2._errors.push(mapper(issue2));
            }
            curr = node2;
            i++;
          }
        }
      }
    }
  };
  processError(error2);
  return fieldErrors;
}

// node_modules/zod/v4/core/parse.js
function finalizeParams(callee, params) {
  return { callee: params?.callee ?? callee, Err: params?.Err };
}
var _parse = (_Err) => {
  const fn = (schema, value, _ctx, _params) => {
    const ctx = _ctx ? { ..._ctx, async: false } : { async: false };
    const result = schema._zod.run({ value, issues: [] }, ctx);
    if (result instanceof Promise) {
      throw new $ZodAsyncError();
    }
    if (result.issues.length) {
      const e = new (_params?.Err ?? _Err)(result.issues.map((iss) => finalizeIssue(iss, ctx, config())));
      captureStackTrace(e, _params?.callee ?? fn);
      throw e;
    }
    return result.value;
  };
  return fn;
};
var _parseAsync = (_Err) => {
  const fn = async (schema, value, _ctx, params) => {
    const ctx = _ctx ? { ..._ctx, async: true } : { async: true };
    let result = schema._zod.run({ value, issues: [] }, ctx);
    if (result instanceof Promise)
      result = await result;
    if (result.issues.length) {
      const e = new (params?.Err ?? _Err)(result.issues.map((iss) => finalizeIssue(iss, ctx, config())));
      captureStackTrace(e, params?.callee ?? fn);
      throw e;
    }
    return result.value;
  };
  return fn;
};
var _safeParse = (_Err) => (schema, value, _ctx) => {
  const ctx = _ctx ? { ..._ctx, async: false } : { async: false };
  const result = schema._zod.run({ value, issues: [] }, ctx);
  if (result instanceof Promise) {
    throw new $ZodAsyncError();
  }
  return result.issues.length ? failure(_Err, result.issues, ctx) : { success: true, data: result.value };
};
function failure(Err, issues, ctx) {
  let error2;
  return {
    success: false,
    get error() {
      if (!error2) {
        error2 = new Err(issues.map((iss) => finalizeIssue(iss, ctx, config())));
        issues = void 0;
        ctx = void 0;
      }
      return error2;
    },
    set error(e) {
      error2 = e;
      issues = void 0;
      ctx = void 0;
    }
  };
}
var _safeParseAsync = (_Err) => async (schema, value, _ctx) => {
  const ctx = _ctx ? { ..._ctx, async: true } : { async: true };
  let result = schema._zod.run({ value, issues: [] }, ctx);
  if (result instanceof Promise)
    result = await result;
  return result.issues.length ? failure(_Err, result.issues, ctx) : { success: true, data: result.value };
};
var COMPILE_INVALID = /* @__PURE__ */ Symbol.for("zod.compile.invalid");
var COMPILE_FALLBACK = /* @__PURE__ */ Symbol.for("zod.compile.fallback");
var validate = (schema, value, _ctx) => {
  const validator = schema._zod.bag.validator;
  if (validator !== void 0) {
    if (validator(value) !== COMPILE_INVALID)
      return true;
    if (validator.definite === true && _ctx === void 0)
      return false;
  }
  return validateFallback(schema, value, _ctx);
};
function validateFallback(schema, value, _ctx) {
  const ctx = _ctx ? { ..._ctx, async: false, abortEarly: true } : { async: false, abortEarly: true };
  const fallbackRun = schema._zod.bag.fallbackRun;
  let result;
  if (fallbackRun) {
    ctx[COMPILE_FALLBACK] = true;
    result = fallbackRun({ value, issues: [] }, ctx);
  } else {
    result = schema._zod.run({ value, issues: [] }, ctx);
  }
  if (result instanceof Promise) {
    throw new $ZodAsyncError();
  }
  return result.issues.length === 0;
}
var validateAsync = async (schema, value, _ctx) => {
  const ctx = _ctx ? { ..._ctx, async: true, abortEarly: true } : { async: true, abortEarly: true };
  let result = schema._zod.run({ value, issues: [] }, ctx);
  if (result instanceof Promise)
    result = await result;
  return result.issues.length === 0;
};
var _encode = (_Err) => {
  const parse2 = _parse(_Err);
  const fn = (schema, value, _ctx, _params) => {
    const ctx = _ctx ? { ..._ctx, direction: "backward" } : { direction: "backward" };
    return parse2(schema, value, ctx, finalizeParams(fn, _params));
  };
  return fn;
};
var _decode = (_Err) => {
  const parse2 = _parse(_Err);
  const fn = (schema, value, _ctx, _params) => {
    return parse2(schema, value, _ctx, finalizeParams(fn, _params));
  };
  return fn;
};
var _encodeAsync = (_Err) => {
  const parseAsync2 = _parseAsync(_Err);
  const fn = async (schema, value, _ctx, _params) => {
    const ctx = _ctx ? { ..._ctx, direction: "backward" } : { direction: "backward" };
    return await parseAsync2(schema, value, ctx, finalizeParams(fn, _params));
  };
  return fn;
};
var _decodeAsync = (_Err) => {
  const parseAsync2 = _parseAsync(_Err);
  const fn = async (schema, value, _ctx, _params) => {
    return await parseAsync2(schema, value, _ctx, finalizeParams(fn, _params));
  };
  return fn;
};
var _safeEncode = (_Err) => (schema, value, _ctx) => {
  const ctx = _ctx ? { ..._ctx, direction: "backward" } : { direction: "backward" };
  return _safeParse(_Err)(schema, value, ctx);
};
var _safeDecode = (_Err) => (schema, value, _ctx) => {
  return _safeParse(_Err)(schema, value, _ctx);
};
var _safeEncodeAsync = (_Err) => async (schema, value, _ctx) => {
  const ctx = _ctx ? { ..._ctx, direction: "backward" } : { direction: "backward" };
  return _safeParseAsync(_Err)(schema, value, ctx);
};
var _safeDecodeAsync = (_Err) => async (schema, value, _ctx) => {
  return _safeParseAsync(_Err)(schema, value, _ctx);
};

// node_modules/zod/v4/core/regexes.js
var cuid = /^[cC][0-9a-z]{6,}$/;
var cuid2 = /^[0-9a-z]+$/;
var ulid = /^[0-7][0-9A-HJKMNP-TV-Za-hjkmnp-tv-z]{25}$/;
var xid = /^[0-9a-vA-V]{20}$/;
var ksuid = /^[A-Za-z0-9]{27}$/;
var nanoid = /^[a-zA-Z0-9_-]{21}$/;
function nanoidOfLength(length) {
  return new RegExp(`^[a-zA-Z0-9_-]{${length}}$`);
}
var duration = /^P(?:(\d+W)|(?!.*W)(?=\d|T\d)(\d+Y)?(\d+M)?(\d+D)?(T(?=\d)(\d+H)?(\d+M)?(\d+([.,]\d+)?S)?)?)$/;
var guid = /^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})$/;
var uuid = (version2) => {
  if (!version2)
    return /^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-8][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}|00000000-0000-0000-0000-000000000000|ffffffff-ffff-ffff-ffff-ffffffffffff)$/;
  return new RegExp(`^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-${version2}[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12})$`);
};
var email = /^(?:[A-Za-z0-9_'+\-]+\.)*[A-Za-z0-9_'+\-]*[A-Za-z0-9_+-]@(?:[A-Za-z0-9][A-Za-z0-9\-]*\.)+[A-Za-z]{2,}$/;
var _emoji = `^(?=[\\s\\S]*[\\p{Extended_Pictographic}\\p{Regional_Indicator}\\u20E3])[\\p{Extended_Pictographic}\\p{Emoji_Component}]+$`;
function emoji() {
  return new RegExp(_emoji, "u");
}
var ipv4 = /^(?:(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\.){3}(?:25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])$/;
var ipv6 = /^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:))$/;
var cidrv4 = /^((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\/([0-9]|[1-2][0-9]|3[0-2])$/;
var cidrv6 = /^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:))\/(12[0-8]|1[01][0-9]|[1-9]?[0-9])$/;
var base64 = /^$|^(?:[0-9a-zA-Z+/]{4})*(?:(?:[0-9a-zA-Z+/]{2}==)|(?:[0-9a-zA-Z+/]{3}=))?$/;
var base64url = /^(?:[A-Za-z0-9_-]{4})*(?:[A-Za-z0-9_-]{2,3})?$/;
var httpProtocol = /^https?$/;
var e164 = /^\+[1-9]\d{6,14}$/;
var dateSource = `(?:(?:\\d\\d[2468][048]|\\d\\d[13579][26]|\\d\\d0[48]|[02468][048]00|[13579][26]00)-02-29|\\d{4}-(?:(?:0[13578]|1[02])-(?:0[1-9]|[12]\\d|3[01])|(?:0[469]|11)-(?:0[1-9]|[12]\\d|30)|(?:02)-(?:0[1-9]|1\\d|2[0-8])))`;
function anchor(source) {
  return new RegExp(`^${source}$`);
}
var date = /* @__PURE__ */ anchor(dateSource);
function timeSource(args) {
  const hhmm = `(?:[01]\\d|2[0-3]):[0-5]\\d`;
  const regex = typeof args.precision === "number" ? args.precision === -1 ? `${hhmm}` : args.precision === 0 ? `${hhmm}:[0-5]\\d` : `${hhmm}:[0-5]\\d\\.\\d{${args.precision}}` : args.seconds ? `${hhmm}:[0-5]\\d(?:\\.\\d+)?` : `${hhmm}(?::[0-5]\\d(?:\\.\\d+)?)?`;
  return regex;
}
function time(args) {
  return new RegExp(`^${timeSource(args)}$`);
}
function datetime(args) {
  const opts = ["Z"];
  if (args.offset)
    opts.push(`([+-](?:[01]\\d|2[0-3]):[0-5]\\d)`);
  const qualified = `${timeSource({ precision: args.precision, seconds: true })}(?:${opts.join("|")})`;
  const timeRegex = args.local ? `${qualified}|${timeSource({ precision: args.precision })}` : qualified;
  return new RegExp(`^${dateSource}T(?:${timeRegex})$`);
}
var anyString = /^[\s\S]{0,}$/;
var integer = /^-?\d+$/;
var number = /^-?\d+(?:\.\d+)?$/;
var boolean = /^(?:true|false)$/i;
var lowercase = /^[^A-Z]*$/;
var uppercase = /^[^a-z]*$/;

// node_modules/zod/v4/core/checks.js
var $ZodCheck = /* @__PURE__ */ $constructor("$ZodCheck", (inst, def2) => {
  var _a3;
  inst._zod ?? (inst._zod = {});
  inst._zod.def = def2;
  (_a3 = inst._zod).onattach ?? (_a3.onattach = []);
});
var _whenHasLength = (payload) => {
  const val = payload.value;
  return !nullish(val) && val.length !== void 0;
};
var numericOriginMap = {
  number: "number",
  bigint: "bigint",
  object: "date"
};
var $ZodCheckLessThan = /* @__PURE__ */ $constructor("$ZodCheckLessThan", (inst, def2) => {
  $ZodCheck.init(inst, def2);
  const origin = numericOriginMap[typeof def2.value];
  inst._zod.check = (payload) => {
    if (def2.inclusive ? payload.value <= def2.value : payload.value < def2.value) {
      return;
    }
    payload.issues.push({
      origin: numericOriginMap[typeof payload.value] ?? origin,
      code: "too_big",
      maximum: typeof def2.value === "object" ? def2.value.getTime() : def2.value,
      input: payload.value,
      inclusive: def2.inclusive,
      inst,
      continue: !def2.abort
    });
  };
});
var $ZodCheckGreaterThan = /* @__PURE__ */ $constructor("$ZodCheckGreaterThan", (inst, def2) => {
  $ZodCheck.init(inst, def2);
  const origin = numericOriginMap[typeof def2.value];
  inst._zod.check = (payload) => {
    if (def2.inclusive ? payload.value >= def2.value : payload.value > def2.value) {
      return;
    }
    payload.issues.push({
      origin: numericOriginMap[typeof payload.value] ?? origin,
      code: "too_small",
      minimum: typeof def2.value === "object" ? def2.value.getTime() : def2.value,
      input: payload.value,
      inclusive: def2.inclusive,
      inst,
      continue: !def2.abort
    });
  };
});
var $ZodCheckMultipleOf = /* @__PURE__ */ $constructor("$ZodCheckMultipleOf", (inst, def2) => {
  $ZodCheck.init(inst, def2);
  inst._zod.check = (payload) => {
    if (typeof payload.value !== typeof def2.value)
      throw new Error("Cannot mix number and bigint in multiple_of check.");
    const isMultiple = typeof payload.value === "bigint" ? (
      // `value % 0n` throws, and nothing is a multiple of zero — the number branch already fails this way via NaN
      def2.value !== BigInt(0) && payload.value % def2.value === BigInt(0)
    ) : floatSafeRemainder(payload.value, def2.value) === 0;
    if (isMultiple)
      return;
    payload.issues.push({
      origin: typeof payload.value,
      code: "not_multiple_of",
      divisor: def2.value,
      input: payload.value,
      inst,
      continue: !def2.abort
    });
  };
});
var $ZodCheckNumberFormat = /* @__PURE__ */ $constructor("$ZodCheckNumberFormat", (inst, def2) => {
  $ZodCheck.init(inst, def2);
  def2.format = def2.format || "float64";
  const isInt = def2.format?.includes("int");
  const origin = isInt ? "int" : "number";
  const [minimum, maximum] = NUMBER_FORMAT_RANGES[def2.format];
  inst._zod.check = (payload) => {
    const input = payload.value;
    if (isInt) {
      if (!Number.isInteger(input)) {
        payload.issues.push({
          expected: origin,
          format: def2.format,
          code: "invalid_type",
          continue: false,
          input,
          inst
        });
        return;
      }
      if (!Number.isSafeInteger(input)) {
        if (input > 0) {
          payload.issues.push({
            input,
            code: "too_big",
            maximum: Number.MAX_SAFE_INTEGER,
            note: "Integers must be within the safe integer range.",
            inst,
            origin,
            inclusive: true,
            continue: !def2.abort
          });
        } else {
          payload.issues.push({
            input,
            code: "too_small",
            minimum: Number.MIN_SAFE_INTEGER,
            note: "Integers must be within the safe integer range.",
            inst,
            origin,
            inclusive: true,
            continue: !def2.abort
          });
        }
        return;
      }
    }
    if (input < minimum) {
      payload.issues.push({
        origin: "number",
        input,
        code: "too_small",
        minimum,
        inclusive: true,
        inst,
        continue: !def2.abort
      });
    }
    if (input > maximum) {
      payload.issues.push({
        origin: "number",
        input,
        code: "too_big",
        maximum,
        inclusive: true,
        inst,
        continue: !def2.abort
      });
    }
  };
});
var $ZodCheckMaxLength = /* @__PURE__ */ $constructor("$ZodCheckMaxLength", (inst, def2) => {
  var _a3;
  $ZodCheck.init(inst, def2);
  (_a3 = inst._zod.def).when ?? (_a3.when = _whenHasLength);
  inst._zod.check = (payload) => {
    const input = payload.value;
    const units = input.length;
    const length = typeof input === "string" && units > def2.maximum ? codePointLength(input) : units;
    if (length <= def2.maximum)
      return;
    const origin = getLengthableOrigin(input);
    payload.issues.push({
      origin,
      code: "too_big",
      maximum: def2.maximum,
      inclusive: true,
      input,
      inst,
      continue: !def2.abort
    });
  };
});
var $ZodCheckMinLength = /* @__PURE__ */ $constructor("$ZodCheckMinLength", (inst, def2) => {
  var _a3;
  $ZodCheck.init(inst, def2);
  (_a3 = inst._zod.def).when ?? (_a3.when = _whenHasLength);
  inst._zod.check = (payload) => {
    const input = payload.value;
    const units = input.length;
    const length = typeof input === "string" && units >= def2.minimum && units < def2.minimum * 2 ? codePointLength(input) : units;
    if (length >= def2.minimum)
      return;
    const origin = getLengthableOrigin(input);
    payload.issues.push({
      origin,
      code: "too_small",
      minimum: def2.minimum,
      inclusive: true,
      input,
      inst,
      continue: !def2.abort
    });
  };
});
var $ZodCheckLengthEquals = /* @__PURE__ */ $constructor("$ZodCheckLengthEquals", (inst, def2) => {
  var _a3;
  $ZodCheck.init(inst, def2);
  (_a3 = inst._zod.def).when ?? (_a3.when = _whenHasLength);
  inst._zod.check = (payload) => {
    const input = payload.value;
    const units = input.length;
    const length = typeof input === "string" && units >= def2.length && units <= def2.length * 2 ? codePointLength(input) : units;
    if (length === def2.length)
      return;
    const origin = getLengthableOrigin(input);
    const tooBig = length > def2.length;
    payload.issues.push({
      origin,
      ...tooBig ? { code: "too_big", maximum: def2.length } : { code: "too_small", minimum: def2.length },
      inclusive: true,
      exact: true,
      input: payload.value,
      inst,
      continue: !def2.abort
    });
  };
});
var $ZodCheckStringFormat = /* @__PURE__ */ $constructor("$ZodCheckStringFormat", (inst, def2) => {
  var _a3, _b;
  $ZodCheck.init(inst, def2);
  if (def2.pattern)
    (_a3 = inst._zod).check ?? (_a3.check = (payload) => {
      def2.pattern.lastIndex = 0;
      if (def2.pattern.test(payload.value))
        return;
      payload.issues.push({
        origin: "string",
        code: "invalid_format",
        format: def2.format,
        input: payload.value,
        ...def2.pattern ? { pattern: def2.pattern.toString() } : {},
        inst,
        continue: !def2.abort
      });
    });
  else
    (_b = inst._zod).check ?? (_b.check = () => {
    });
});
var $ZodCheckRegex = /* @__PURE__ */ $constructor("$ZodCheckRegex", (inst, def2) => {
  $ZodCheckStringFormat.init(inst, def2);
  inst._zod.check = (payload) => {
    def2.pattern.lastIndex = 0;
    if (def2.pattern.test(payload.value))
      return;
    payload.issues.push({
      origin: "string",
      code: "invalid_format",
      format: "regex",
      input: payload.value,
      pattern: def2.pattern.toString(),
      inst,
      continue: !def2.abort
    });
  };
});
var $ZodCheckLowerCase = /* @__PURE__ */ $constructor("$ZodCheckLowerCase", (inst, def2) => {
  def2.pattern ?? (def2.pattern = lowercase);
  $ZodCheckStringFormat.init(inst, def2);
});
var $ZodCheckUpperCase = /* @__PURE__ */ $constructor("$ZodCheckUpperCase", (inst, def2) => {
  def2.pattern ?? (def2.pattern = uppercase);
  $ZodCheckStringFormat.init(inst, def2);
});
var $ZodCheckIncludes = /* @__PURE__ */ $constructor("$ZodCheckIncludes", (inst, def2) => {
  $ZodCheck.init(inst, def2);
  const escapedRegex = escapeRegex(def2.includes);
  const pattern = new RegExp(typeof def2.position === "number" ? `^.{${def2.position},}${escapedRegex}` : escapedRegex);
  def2.pattern = pattern;
  inst._zod.check = (payload) => {
    if (payload.value.includes(def2.includes, def2.position))
      return;
    payload.issues.push({
      origin: "string",
      code: "invalid_format",
      format: "includes",
      includes: def2.includes,
      input: payload.value,
      inst,
      continue: !def2.abort
    });
  };
});
var $ZodCheckStartsWith = /* @__PURE__ */ $constructor("$ZodCheckStartsWith", (inst, def2) => {
  $ZodCheck.init(inst, def2);
  const pattern = new RegExp(`^${escapeRegex(def2.prefix)}.*`);
  def2.pattern ?? (def2.pattern = pattern);
  inst._zod.check = (payload) => {
    if (payload.value.startsWith(def2.prefix))
      return;
    payload.issues.push({
      origin: "string",
      code: "invalid_format",
      format: "starts_with",
      prefix: def2.prefix,
      input: payload.value,
      inst,
      continue: !def2.abort
    });
  };
});
var $ZodCheckEndsWith = /* @__PURE__ */ $constructor("$ZodCheckEndsWith", (inst, def2) => {
  $ZodCheck.init(inst, def2);
  const pattern = new RegExp(`.*${escapeRegex(def2.suffix)}$`);
  def2.pattern ?? (def2.pattern = pattern);
  inst._zod.check = (payload) => {
    if (payload.value.endsWith(def2.suffix))
      return;
    payload.issues.push({
      origin: "string",
      code: "invalid_format",
      format: "ends_with",
      suffix: def2.suffix,
      input: payload.value,
      inst,
      continue: !def2.abort
    });
  };
});
var $ZodCheckOverwrite = /* @__PURE__ */ $constructor("$ZodCheckOverwrite", (inst, def2) => {
  $ZodCheck.init(inst, def2);
  inst._zod.check = (payload) => {
    payload.value = def2.tx(payload.value);
  };
});

// node_modules/zod/v4/core/doc.js
var Doc = class {
  constructor(args = [], closed = {}) {
    this.content = [];
    this.indent = 0;
    this.args = args;
    this.closed = closed;
  }
  // the compiler catches a child's throw and keeps writing into this doc, so the indent has to unwind with it
  indented(fn) {
    this.indent += 1;
    try {
      fn(this);
    } finally {
      this.indent -= 1;
    }
  }
  write(arg) {
    if (typeof arg === "function") {
      arg(this, { execution: "sync" });
      arg(this, { execution: "async" });
      return;
    }
    const content = arg;
    const lines = content.split("\n").filter((x) => x);
    const minIndent = Math.min(...lines.map((x) => x.length - x.trimStart().length));
    const dedented = lines.map((x) => x.slice(minIndent)).map((x) => " ".repeat(this.indent * 2) + x);
    for (const line of dedented) {
      this.content.push(line);
    }
  }
  compile() {
    const F = Function;
    const content = this?.content ?? [``];
    const factory = new F(...Object.keys(this.closed), `return function (${this.args.join(", ")}) {
${content.join("\n")}
};`);
    return factory(...Object.values(this.closed));
  }
};

// node_modules/zod/v4/core/versions.js
var version = {
  major: 4,
  minor: 6,
  patch: 7
};

// node_modules/zod/v4/core/schemas.js
var $ZodType = /* @__PURE__ */ $constructor("$ZodType", (inst, def2) => {
  var _a3;
  inst ?? (inst = {});
  inst._zod.def = def2;
  inst._zod.bag = inst._zod.bag || {};
  inst._zod.version = version;
  const defChecks = inst._zod.def.checks;
  const checks = inst._zod.traits.has("$ZodCheck") ? [inst, ...defChecks ?? []] : defChecks?.length ? [...defChecks] : [];
  for (const ch of checks) {
    for (const fn of ch._zod.onattach) {
      fn(inst);
    }
  }
  if (checks.length === 0) {
    (_a3 = inst._zod).deferred ?? (_a3.deferred = []);
    inst._zod.deferred?.push(() => {
      inst._zod.run = inst._zod.parse;
    });
  } else {
    const runChecks = (payload, checks2, ctx) => {
      if (payload.memo)
        return payload;
      let isAborted = aborted(payload);
      let asyncResult;
      for (const ch of checks2) {
        if (ch._zod.def.when) {
          if (explicitlyAborted(payload))
            continue;
          const shouldRun = ch._zod.def.when(payload);
          if (!shouldRun)
            continue;
        } else if (isAborted) {
          continue;
        }
        const currLen = payload.issues.length;
        const _ = ch._zod.check(payload);
        if (_ instanceof Promise && ctx?.async === false) {
          throw new $ZodAsyncError();
        }
        if (asyncResult || _ instanceof Promise) {
          asyncResult = (asyncResult ?? Promise.resolve()).then(async () => {
            await _;
            const nextLen = payload.issues.length;
            if (nextLen === currLen)
              return;
            attachSchema(payload.issues, currLen, inst);
            if (!isAborted)
              isAborted = aborted(payload, currLen);
          });
        } else {
          const nextLen = payload.issues.length;
          if (nextLen === currLen)
            continue;
          attachSchema(payload.issues, currLen, inst);
          if (!isAborted)
            isAborted = aborted(payload, currLen);
        }
      }
      if (asyncResult) {
        return asyncResult.then(() => {
          return payload;
        });
      }
      return payload;
    };
    const handleCanaryResult = (canary, payload, ctx) => {
      if (aborted(canary)) {
        canary.aborted = true;
        return canary;
      }
      const checkResult = runChecks(payload, checks, ctx);
      if (checkResult instanceof Promise) {
        if (ctx.async === false)
          throw new $ZodAsyncError();
        return checkResult.then((checkResult2) => inst._zod.parse(checkResult2, ctx));
      }
      return inst._zod.parse(checkResult, ctx);
    };
    inst._zod.run = (payload, ctx) => {
      if (ctx.skipChecks) {
        return inst._zod.parse(payload, ctx);
      }
      if (ctx.direction === "backward") {
        const canary = inst._zod.parse({ value: payload.value, issues: [] }, { ...ctx, skipChecks: true });
        if (canary instanceof Promise) {
          return canary.then((canary2) => {
            return handleCanaryResult(canary2, payload, ctx);
          });
        }
        return handleCanaryResult(canary, payload, ctx);
      }
      const result = inst._zod.parse(payload, ctx);
      if (result instanceof Promise) {
        if (ctx.async === false)
          throw new $ZodAsyncError();
        return result.then((result2) => runChecks(result2, checks, ctx));
      }
      return runChecks(result, checks, ctx);
    };
  }
}, {
  // Wrappers extend this by installing a richer factory over it; reading it eagerly would defeat the laziness.
  get "~standard"() {
    return hide(this, "~standard", standardProps(this));
  },
  set "~standard"(value) {
    own(this, "~standard", value);
  }
});
var toStandardResult = (r, ctx) => r.issues.length ? { issues: r.issues.map((iss) => finalizeIssue(iss, ctx, config())) } : { value: r.value };
async function validateAsync2(inst, value) {
  const ctx = { async: true };
  return toStandardResult(await inst._zod.run({ value, issues: [] }, ctx), ctx);
}
function standardProps(inst) {
  return {
    validate: (value) => {
      const ctx = { async: false };
      try {
        const r = inst._zod.run({ value, issues: [] }, ctx);
        if (!(r instanceof Promise))
          return toStandardResult(r, ctx);
      } catch (_) {
      }
      return validateAsync2(inst, value);
    },
    vendor: "zod",
    version: 1
  };
}
var $ZodString = /* @__PURE__ */ $constructor("$ZodString", (inst, def2) => {
  $ZodType.init(inst, def2);
  inst._zod.pattern = def2.pattern ?? anyString;
  inst._zod.parse = (payload, _) => {
    if (def2.coerce)
      try {
        payload.value = String(payload.value);
      } catch (_2) {
      }
    if (typeof payload.value === "string")
      return payload;
    payload.issues.push({
      expected: "string",
      code: "invalid_type",
      input: payload.value,
      inst
    });
    return payload;
  };
});
var $ZodStringFormat = /* @__PURE__ */ $constructor("$ZodStringFormat", (inst, def2) => {
  $ZodCheckStringFormat.init(inst, def2);
  $ZodString.init(inst, def2);
});
var $ZodGUID = /* @__PURE__ */ $constructor("$ZodGUID", (inst, def2) => {
  def2.pattern ?? (def2.pattern = guid);
  $ZodStringFormat.init(inst, def2);
});
var $ZodUUID = /* @__PURE__ */ $constructor("$ZodUUID", (inst, def2) => {
  if (def2.version) {
    const versionMap = {
      v1: 1,
      v2: 2,
      v3: 3,
      v4: 4,
      v5: 5,
      v6: 6,
      v7: 7,
      v8: 8
    };
    const v = versionMap[def2.version];
    if (v === void 0)
      throw new Error(`Invalid UUID version: "${def2.version}"`);
    def2.pattern ?? (def2.pattern = uuid(v));
  } else
    def2.pattern ?? (def2.pattern = uuid());
  $ZodStringFormat.init(inst, def2);
});
var $ZodEmail = /* @__PURE__ */ $constructor("$ZodEmail", (inst, def2) => {
  def2.pattern ?? (def2.pattern = email);
  $ZodStringFormat.init(inst, def2);
});
var URL_BAD_FORMAT = 1;
var URL_UNPARSEABLE = 2;
function canParseURL(input) {
  try {
    if (typeof URL !== "undefined" && typeof URL.canParse === "function")
      return URL.canParse(input);
    new URL(input);
    return true;
  } catch {
    return false;
  }
}
function validateURL(trimmed, def2) {
  if (!("normalize" in def2) && !("hostname" in def2) && !("protocol" in def2)) {
    return canParseURL(trimmed) || URL_UNPARSEABLE;
  }
  return parseURLObject(trimmed, def2);
}
function parseURLObject(trimmed, def2) {
  if (!def2.normalize && def2.protocol?.source === httpProtocol.source && !/^https?:\/\//i.test(trimmed)) {
    return URL_BAD_FORMAT;
  }
  try {
    if (typeof URL !== "undefined") {
      const URLStatic = URL;
      if (typeof URLStatic.parse === "function")
        return URLStatic.parse(trimmed) ?? URL_UNPARSEABLE;
    }
    return new URL(trimmed);
  } catch {
    return URL_UNPARSEABLE;
  }
}
var asciiTabOrNewline = /[\t\n\r]/g;
function stripTabAndNewline(value) {
  return value.replace(asciiTabOrNewline, "");
}
function urlHostnameOk(url2, hostname) {
  hostname.lastIndex = 0;
  return hostname.test(url2.hostname);
}
function urlProtocolOk(url2, protocol) {
  protocol.lastIndex = 0;
  return protocol.test(url2.protocol.endsWith(":") ? url2.protocol.slice(0, -1) : url2.protocol);
}
var $ZodURL = /* @__PURE__ */ $constructor("$ZodURL", (inst, def2) => {
  $ZodStringFormat.init(inst, def2);
  inst._zod.check = (payload) => {
    try {
      const trimmed = payload.value.trim();
      const url2 = validateURL(trimmed, def2);
      if (url2 === URL_BAD_FORMAT) {
        payload.issues.push({
          code: "invalid_format",
          format: "url",
          note: "Invalid URL format",
          input: payload.value,
          inst,
          continue: !def2.abort
        });
        return;
      }
      if (url2 === URL_UNPARSEABLE) {
        payload.issues.push({
          code: "invalid_format",
          format: "url",
          input: payload.value,
          inst,
          continue: !def2.abort
        });
        return;
      }
      if (url2 === true) {
        payload.value = stripTabAndNewline(trimmed);
        return;
      }
      if (def2.hostname && !urlHostnameOk(url2, def2.hostname)) {
        payload.issues.push({
          code: "invalid_format",
          format: "url",
          note: "Invalid hostname",
          pattern: def2.hostname.source,
          input: payload.value,
          inst,
          continue: !def2.abort
        });
      }
      if (def2.protocol && !urlProtocolOk(url2, def2.protocol)) {
        payload.issues.push({
          code: "invalid_format",
          format: "url",
          note: "Invalid protocol",
          pattern: def2.protocol.source,
          input: payload.value,
          inst,
          continue: !def2.abort
        });
      }
      payload.value = def2.normalize ? url2.href : stripTabAndNewline(trimmed);
      return;
    } catch (_) {
      payload.issues.push({
        code: "invalid_format",
        format: "url",
        input: payload.value,
        inst,
        continue: !def2.abort
      });
    }
  };
});
var $ZodEmoji = /* @__PURE__ */ $constructor("$ZodEmoji", (inst, def2) => {
  def2.pattern ?? (def2.pattern = emoji());
  $ZodStringFormat.init(inst, def2);
});
var $ZodNanoID = /* @__PURE__ */ $constructor("$ZodNanoID", (inst, def2) => {
  if (def2.length !== void 0 && (!Number.isInteger(def2.length) || def2.length < 1))
    throw new Error(`Invalid nanoid length: ${def2.length}`);
  def2.pattern ?? (def2.pattern = def2.length === void 0 ? nanoid : nanoidOfLength(def2.length));
  $ZodStringFormat.init(inst, def2);
});
var $ZodCUID = /* @__PURE__ */ $constructor("$ZodCUID", (inst, def2) => {
  def2.pattern ?? (def2.pattern = cuid);
  $ZodStringFormat.init(inst, def2);
});
var $ZodCUID2 = /* @__PURE__ */ $constructor("$ZodCUID2", (inst, def2) => {
  def2.pattern ?? (def2.pattern = cuid2);
  $ZodStringFormat.init(inst, def2);
});
var $ZodULID = /* @__PURE__ */ $constructor("$ZodULID", (inst, def2) => {
  def2.pattern ?? (def2.pattern = ulid);
  $ZodStringFormat.init(inst, def2);
});
var $ZodXID = /* @__PURE__ */ $constructor("$ZodXID", (inst, def2) => {
  def2.pattern ?? (def2.pattern = xid);
  $ZodStringFormat.init(inst, def2);
});
var $ZodKSUID = /* @__PURE__ */ $constructor("$ZodKSUID", (inst, def2) => {
  def2.pattern ?? (def2.pattern = ksuid);
  $ZodStringFormat.init(inst, def2);
});
var $ZodISODateTime = /* @__PURE__ */ $constructor("$ZodISODateTime", (inst, def2) => {
  def2.pattern ?? (def2.pattern = datetime(def2));
  $ZodStringFormat.init(inst, def2);
});
var $ZodISODate = /* @__PURE__ */ $constructor("$ZodISODate", (inst, def2) => {
  def2.pattern ?? (def2.pattern = date);
  $ZodStringFormat.init(inst, def2);
});
var $ZodISOTime = /* @__PURE__ */ $constructor("$ZodISOTime", (inst, def2) => {
  def2.pattern ?? (def2.pattern = time(def2));
  $ZodStringFormat.init(inst, def2);
});
var $ZodISODuration = /* @__PURE__ */ $constructor("$ZodISODuration", (inst, def2) => {
  def2.pattern ?? (def2.pattern = duration);
  $ZodStringFormat.init(inst, def2);
});
var $ZodIPv4 = /* @__PURE__ */ $constructor("$ZodIPv4", (inst, def2) => {
  def2.pattern ?? (def2.pattern = ipv4);
  $ZodStringFormat.init(inst, def2);
});
var ipv6Alphabet = /^[0-9a-fA-F:.]+$/;
function isValidIPv6(value) {
  if (!ipv6Alphabet.test(value))
    return false;
  return canParseURL(`http://[${value}]`);
}
var $ZodIPv6 = /* @__PURE__ */ $constructor("$ZodIPv6", (inst, def2) => {
  def2.pattern ?? (def2.pattern = ipv6);
  $ZodStringFormat.init(inst, def2);
  inst._zod.check = (payload) => {
    if (!isValidIPv6(payload.value)) {
      payload.issues.push({
        code: "invalid_format",
        format: "ipv6",
        input: payload.value,
        inst,
        continue: !def2.abort
      });
    }
  };
});
var $ZodCIDRv4 = /* @__PURE__ */ $constructor("$ZodCIDRv4", (inst, def2) => {
  def2.pattern ?? (def2.pattern = cidrv4);
  $ZodStringFormat.init(inst, def2);
});
function isValidCIDRv6(value) {
  const parts = value.split("/");
  if (parts.length !== 2)
    return false;
  const [address, prefix] = parts;
  if (!prefix)
    return false;
  const prefixNum = Number(prefix);
  if (`${prefixNum}` !== prefix)
    return false;
  if (prefixNum < 0 || prefixNum > 128)
    return false;
  return isValidIPv6(address);
}
var $ZodCIDRv6 = /* @__PURE__ */ $constructor("$ZodCIDRv6", (inst, def2) => {
  def2.pattern ?? (def2.pattern = cidrv6);
  $ZodStringFormat.init(inst, def2);
  inst._zod.check = (payload) => {
    if (!isValidCIDRv6(payload.value)) {
      payload.issues.push({
        code: "invalid_format",
        format: "cidrv6",
        input: payload.value,
        inst,
        continue: !def2.abort
      });
    }
  };
});
function isValidBase64(data) {
  if (data === "")
    return true;
  if (/\s/.test(data))
    return false;
  if (data.length % 4 !== 0)
    return false;
  try {
    atob(data);
    return true;
  } catch {
    return false;
  }
}
var base64Charset = /^[0-9a-zA-Z+/]*={0,2}$/;
var $ZodBase64 = /* @__PURE__ */ $constructor("$ZodBase64", (inst, def2) => {
  def2.pattern ?? (def2.pattern = base64Charset);
  $ZodStringFormat.init(inst, def2);
  inst._zod.check = (payload) => {
    if (isValidBase64(payload.value))
      return;
    payload.issues.push({
      code: "invalid_format",
      format: "base64",
      input: payload.value,
      inst,
      continue: !def2.abort
    });
  };
});
var base64urlCharset = /^[A-Za-z0-9_-]*$/;
function isValidBase64URL(data) {
  if (!base64urlCharset.test(data))
    return false;
  const base642 = data.replace(/[-_]/g, (c) => c === "-" ? "+" : "/");
  const padded = base642.padEnd(Math.ceil(base642.length / 4) * 4, "=");
  return isValidBase64(padded);
}
var $ZodBase64URL = /* @__PURE__ */ $constructor("$ZodBase64URL", (inst, def2) => {
  def2.pattern ?? (def2.pattern = base64urlCharset);
  $ZodStringFormat.init(inst, def2);
  inst._zod.check = (payload) => {
    if (isValidBase64URL(payload.value))
      return;
    payload.issues.push({
      code: "invalid_format",
      format: "base64url",
      input: payload.value,
      inst,
      continue: !def2.abort
    });
  };
});
var $ZodE164 = /* @__PURE__ */ $constructor("$ZodE164", (inst, def2) => {
  def2.pattern ?? (def2.pattern = e164);
  $ZodStringFormat.init(inst, def2);
});
function isValidJWT(token, algorithm = null) {
  try {
    const tokensParts = token.split(".");
    if (tokensParts.length !== 3)
      return false;
    const [header] = tokensParts;
    if (!header)
      return false;
    const parsedHeader = JSON.parse(atob(header));
    if ("typ" in parsedHeader && parsedHeader?.typ !== "JWT")
      return false;
    if (!parsedHeader.alg)
      return false;
    if (algorithm && (!("alg" in parsedHeader) || parsedHeader.alg !== algorithm))
      return false;
    return true;
  } catch {
    return false;
  }
}
var $ZodJWT = /* @__PURE__ */ $constructor("$ZodJWT", (inst, def2) => {
  $ZodStringFormat.init(inst, def2);
  inst._zod.check = (payload) => {
    if (isValidJWT(payload.value, def2.alg))
      return;
    payload.issues.push({
      code: "invalid_format",
      format: "jwt",
      input: payload.value,
      inst,
      continue: !def2.abort
    });
  };
});
var $ZodNumber = /* @__PURE__ */ $constructor("$ZodNumber", (inst, def2) => {
  $ZodType.init(inst, def2);
  inst._zod.pattern = number;
  inst._zod.parse = (payload, _ctx) => {
    if (def2.coerce)
      try {
        payload.value = Number(payload.value);
      } catch (_) {
      }
    const input = payload.value;
    if (typeof input === "number" && !Number.isNaN(input) && Number.isFinite(input)) {
      return payload;
    }
    const received = typeof input === "number" ? Number.isNaN(input) ? "NaN" : !Number.isFinite(input) ? String(input) : void 0 : void 0;
    payload.issues.push({
      expected: "number",
      code: "invalid_type",
      input,
      inst,
      ...received ? { received } : {}
    });
    return payload;
  };
});
var $ZodNumberFormat = /* @__PURE__ */ $constructor("$ZodNumberFormat", (inst, def2) => {
  $ZodCheckNumberFormat.init(inst, def2);
  $ZodNumber.init(inst, def2);
});
var $ZodBoolean = /* @__PURE__ */ $constructor("$ZodBoolean", (inst, def2) => {
  $ZodType.init(inst, def2);
  inst._zod.pattern = boolean;
  inst._zod.parse = (payload, _ctx) => {
    if (def2.coerce)
      try {
        payload.value = Boolean(payload.value);
      } catch (_) {
      }
    const input = payload.value;
    if (typeof input === "boolean")
      return payload;
    payload.issues.push({
      expected: "boolean",
      code: "invalid_type",
      input,
      inst
    });
    return payload;
  };
});
var $ZodUnknown = /* @__PURE__ */ $constructor("$ZodUnknown", (inst, def2) => {
  $ZodType.init(inst, def2);
  inst._zod.parse = (payload) => payload;
});
var $ZodNever = /* @__PURE__ */ $constructor("$ZodNever", (inst, def2) => {
  $ZodType.init(inst, def2);
  inst._zod.parse = (payload, _ctx) => {
    payload.issues.push({
      expected: "never",
      code: "invalid_type",
      input: payload.value,
      inst
    });
    return payload;
  };
});
function capSize(inst, size) {
  const caps = sizeCaps(inst._zod.def.checks);
  if (!caps.length)
    return;
  const cap = Math.min(...caps.map((c) => c[0]));
  const parse2 = inst._zod.parse;
  inst._zod.parse = (payload, ctx) => {
    const n = size(payload.value);
    if (n > cap && !ctx?.skipChecks && ctx?.direction !== "backward") {
      for (const [max, ch] of sizeCaps(inst._zod.def.checks)) {
        if (n > max && ch._zod.def.when(payload)) {
          const start = payload.issues.length;
          ch._zod.check(payload);
          payload.issues[start].continue = false;
          attachSchema(payload.issues, start, inst);
          return payload;
        }
      }
    }
    return parse2(payload, ctx);
  };
}
var arraySize = (input) => Array.isArray(input) ? input.length : -1;
function handleArrayResult(result, final, index) {
  if (result.issues.length) {
    final.issues.push(...prefixIssues(index, result.issues));
  }
  final.value[index] = result.value;
}
var $ZodArray = /* @__PURE__ */ $constructor("$ZodArray", (inst, def2) => {
  $ZodType.init(inst, def2);
  const memo2 = globalConfig.memoizer;
  memo2?.attach(inst);
  inst._zod.parse = (payload, ctx) => {
    const input = payload.value;
    if (!Array.isArray(input)) {
      payload.issues.push({
        expected: "array",
        code: "invalid_type",
        input,
        inst
      });
      return payload;
    }
    payload.value = memo2 ? memo2.alloc(inst, payload, Array(input.length), ctx) : Array(input.length);
    const proms = [];
    const abortEarly = ctx?.abortEarly;
    for (let i = 0; i < input.length; i++) {
      const item = input[i];
      const result = def2.element._zod.run({
        value: item,
        issues: []
      }, ctx);
      if (result instanceof Promise) {
        proms.push(result.then((result2) => handleArrayResult(result2, payload, i)));
      } else {
        handleArrayResult(result, payload, i);
        if (abortEarly && result.issues.length !== 0 && aborted(result))
          break;
      }
    }
    if (proms.length) {
      return Promise.all(proms).then(() => payload);
    }
    return payload;
  };
  capSize(inst, arraySize);
});
function handlePropertyResult(result, final, key, input, optin, optout) {
  const isPresent = key in input;
  const isOptionalOut = optout === "optional";
  if (!isPresent && isOptionalOut && optin === "optional") {
    return;
  }
  if (result.issues.length) {
    if (optin !== void 0 && isOptionalOut && !isPresent) {
      return;
    }
    final.issues.push(...prefixIssues(key, result.issues));
  }
  if (!isPresent && optin === void 0) {
    if (!result.issues.length) {
      final.issues.push({
        code: "invalid_type",
        expected: "nonoptional",
        input: void 0,
        path: [key]
      });
    }
    return;
  }
  if (result.value === void 0) {
    if (isPresent || optin === "defaulted" && !isOptionalOut) {
      final.value[key] = void 0;
    }
  } else {
    final.value[key] = result.value;
  }
}
var NO_SYMBOL_KEYS = [];
function normalizeDef(def2) {
  const keys = Object.keys(def2.shape);
  const ownSymbols = Object.getOwnPropertySymbols(def2.shape);
  const symbolKeys = ownSymbols.length ? ownSymbols : NO_SYMBOL_KEYS;
  const allKeys = symbolKeys.length ? [...keys, ...symbolKeys] : keys;
  for (const k of allKeys) {
    if (!def2.shape?.[k]?._zod?.traits?.has("$ZodType")) {
      throw new Error(`Invalid element at key "${String(k)}": expected a Zod schema`);
    }
  }
  const okeys = optionalKeys(def2.shape);
  return {
    ...def2,
    allKeys,
    symbolKeys,
    // string-only: handleCatchall matches it against `for...in`, which never yields a symbol
    keySet: new Set(keys),
    numKeys: keys.length,
    optionalKeys: new Set(okeys)
  };
}
function handleCatchall(proms, input, payload, ctx, def2, inst, abortEarly) {
  const unrecognized = [];
  const keySet = def2.keySet;
  const _catchall = def2.catchall._zod;
  const t = _catchall.def.type;
  const optin = _catchall.optin;
  const optout = _catchall.optout;
  let seen = 0;
  for (const key in input) {
    if (abortEarly && payload.issues.length !== seen) {
      if (aborted(payload, seen))
        break;
      seen = payload.issues.length;
    }
    if (keySet.has(key))
      continue;
    if (key === "__proto__") {
      if (t === "never")
        unrecognized.push(key);
      continue;
    }
    if (t === "never") {
      unrecognized.push(key);
      continue;
    }
    const r = _catchall.run({ value: input[key], issues: [] }, ctx);
    if (r instanceof Promise) {
      proms.push(r.then((r2) => handlePropertyResult(r2, payload, key, input, optin, optout)));
    } else {
      handlePropertyResult(r, payload, key, input, optin, optout);
    }
  }
  if (unrecognized.length) {
    payload.issues.push({
      code: "unrecognized_keys",
      keys: unrecognized,
      input,
      inst,
      // Describes the shape of the input, not the validity of the parsed value, so it never aborts. The parse still fails; the schema's own checks just get to run first, and an enclosing intersection can reconcile the key against a sibling operand.
      continue: true
    });
  }
  if (!proms.length)
    return payload;
  return Promise.all(proms).then(() => {
    return payload;
  });
}
var $ZodObject = /* @__PURE__ */ $constructor("$ZodObject", (inst, def2) => {
  $ZodType.init(inst, def2);
  const desc = Object.getOwnPropertyDescriptor(def2, "shape");
  const sh = desc?.get ? desc.get.raw : def2.shape ?? {};
  if (sh) {
    const get = () => {
      const newSh = { ...sh };
      Object.defineProperty(def2, "shape", { value: newSh });
      get.raw = newSh;
      return newSh;
    };
    get.raw = sh;
    Object.defineProperty(def2, "shape", { get });
  }
  const _normalized = cached(() => normalizeDef(def2));
  defineLazyInternal(inst, "propValues", (zod) => {
    const shape = zod.def.shape;
    const propValues = {};
    for (const key in shape) {
      const field = shape[key]._zod;
      if (field.values) {
        if (!Object.prototype.hasOwnProperty.call(propValues, key)) {
          assignProp(propValues, key, /* @__PURE__ */ new Set());
        }
        for (const v of field.values)
          propValues[key].add(v);
        if (field.optin !== void 0)
          propValues[key].add(void 0);
      }
    }
    return propValues;
  });
  const isObject2 = isObject;
  const catchall = def2.catchall;
  let value;
  const memo2 = globalConfig.memoizer;
  memo2?.attach(inst);
  inst._zod.parse = (payload, ctx) => {
    value ?? (value = _normalized.value);
    const input = payload.value;
    if (!isObject2(input)) {
      payload.issues.push({
        expected: "object",
        code: "invalid_type",
        input,
        inst
      });
      return payload;
    }
    payload.value = memo2 ? memo2.alloc(inst, payload, {}, ctx) : {};
    const proms = [];
    const shape = value.shape;
    const abortEarly = ctx?.abortEarly;
    let seen = payload.issues.length;
    for (const key of value.allKeys) {
      if (abortEarly && payload.issues.length !== seen) {
        if (aborted(payload, seen))
          break;
        seen = payload.issues.length;
      }
      if (key === "__proto__")
        continue;
      const el = shape[key];
      const optin = el._zod.optin;
      const optout = el._zod.optout;
      const r = el._zod.run({ value: input[key], issues: [] }, ctx);
      if (r instanceof Promise) {
        proms.push(r.then((r2) => handlePropertyResult(r2, payload, key, input, optin, optout)));
      } else {
        handlePropertyResult(r, payload, key, input, optin, optout);
      }
    }
    if (!catchall) {
      return proms.length ? Promise.all(proms).then(() => payload) : payload;
    }
    return handleCatchall(proms, input, payload, ctx, _normalized.value, inst, abortEarly === true);
  };
});
var $ZodObjectJIT = /* @__PURE__ */ $constructor("$ZodObjectJIT", (inst, def2) => {
  $ZodObject.init(inst, def2);
  const superParse = inst._zod.parse;
  const _normalized = cached(() => normalizeDef(def2));
  const memo2 = globalConfig.memoizer;
  const generateFastpass = (shape) => {
    const normalized = _normalized.value;
    const syms = normalized.symbolKeys;
    const doc = new Doc(["payload", "ctx"], { shape, inst, memo: memo2, syms });
    const parseStr = (k) => `shape[${k}]._zod.run({ value: input[${k}], issues: [] }, ctx)`;
    const prefixStr = (id, k) => `
          let ${id}_ab = false;
          for (let i = 0; i < ${id}.issues.length; i++) {
            const iss = ${id}.issues[i];
            iss.path = iss.path ? [${k}, ...iss.path] : [${k}];
            payload.issues.push(iss);
            if (iss.continue !== true) ${id}_ab = true;
          }
          if (${id}_ab && ctx && ctx.abortEarly) {
            payload.value = newResult;
            return payload;
          }`;
    doc.write(`const input = payload.value;`);
    const ids = /* @__PURE__ */ Object.create(null);
    let counter = 0;
    for (const key of normalized.allKeys) {
      ids[key] = `key_${counter++}`;
    }
    doc.write(memo2 ? `const newResult = memo.alloc(inst, payload, {}, ctx);` : `const newResult = {};`);
    for (const key of normalized.allKeys) {
      if (key === "__proto__")
        continue;
      const id = ids[key];
      const k = typeof key === "symbol" ? `syms[${syms.indexOf(key)}]` : esc(key);
      const isPresent = `${k} in input`;
      const schema = shape[key];
      const optin = schema?._zod?.optin;
      const isOptionalIn = optin !== void 0;
      const isOptionalOut = schema?._zod?.optout === "optional";
      doc.write(`const ${id} = ${parseStr(k)};`);
      if (isOptionalIn && isOptionalOut) {
        const assign = optin === "optional" ? `${id}_present` : `${id}.value !== undefined || ${id}_present`;
        doc.write(`
        const ${id}_present = ${isPresent};
        if (!${id}.issues.length || ${id}_present) {
          if (${id}.issues.length) {${prefixStr(id, k)}
          }

          if (${assign}) {
            newResult[${k}] = ${id}.value;
          }
        }

      `);
      } else if (!isOptionalIn) {
        doc.write(`
        const ${id}_present = ${isPresent};
        if (${id}.issues.length) {${prefixStr(id, k)}
        }
        if (!${id}_present && !${id}.issues.length) {
          payload.issues.push({
            code: "invalid_type",
            expected: "nonoptional",
            input: undefined,
            path: [${k}]
          });
          if (ctx && ctx.abortEarly) {
            payload.value = newResult;
            return payload;
          }
        }

        if (${id}_present) {
          newResult[${k}] = ${id}.value;
        }

      `);
      } else {
        doc.write(`
        if (${id}.issues.length) {${prefixStr(id, k)}
        }
      `);
        if (optin === "defaulted") {
          doc.write(`newResult[${k}] = ${id}.value;`);
        } else {
          doc.write(`
        if (${id}.value !== undefined || ${isPresent}) {
          newResult[${k}] = ${id}.value;
        }
      `);
        }
      }
    }
    doc.write(`payload.value = newResult;`);
    doc.write(`return payload;`);
    return doc.compile();
  };
  let fastpass;
  const isObject2 = isObject;
  const jit = !globalConfig.jitless;
  const allowsEval2 = allowsEval;
  const fastEnabled = jit && allowsEval2.value;
  const catchall = def2.catchall;
  let value;
  inst._zod.parse = (payload, ctx) => {
    value ?? (value = _normalized.value);
    const input = payload.value;
    if (!isObject2(input)) {
      payload.issues.push({
        expected: "object",
        code: "invalid_type",
        input,
        inst
      });
      return payload;
    }
    if (jit && fastEnabled && ctx?.async === false && ctx.jitless !== true) {
      if (!fastpass)
        fastpass = generateFastpass(def2.shape);
      payload = fastpass(payload, ctx);
      if (!catchall)
        return payload;
      return handleCatchall([], input, payload, ctx, value, inst, ctx?.abortEarly === true);
    }
    return superParse(payload, ctx);
  };
});
function handleUnionResults(results, final, inst, ctx) {
  for (const result of results) {
    if (result.issues.length === 0) {
      final.value = result.value;
      return final;
    }
  }
  const nonaborted = results.filter((r) => !aborted(r));
  if (nonaborted.length === 1) {
    final.value = nonaborted[0].value;
    return nonaborted[0];
  }
  final.issues.push({
    code: "invalid_union",
    input: final.value,
    inst,
    errors: results.map((result) => result.issues.map((iss) => finalizeIssue(iss, ctx, config())))
  });
  return final;
}
var $ZodUnion = /* @__PURE__ */ $constructor("$ZodUnion", (inst, def2) => {
  $ZodType.init(inst, def2);
  defineLazyInternal(inst, "optin", (zod) => zod.def.options.some((o) => o._zod.optin === "defaulted") ? "defaulted" : zod.def.options.some((o) => o._zod.optin !== void 0) ? "optional" : void 0);
  defineLazyInternal(inst, "optout", (zod) => zod.def.options.some((o) => o._zod.optout === "optional") ? "optional" : void 0);
  defineLazyInternal(inst, "values", (zod) => {
    if (zod.def.options.every((o) => o._zod.values)) {
      return new Set(zod.def.options.flatMap((option) => Array.from(option._zod.values)));
    }
    return void 0;
  });
  defineLazyInternal(inst, "pattern", (zod) => {
    if (zod.def.options.every((o) => o._zod.pattern)) {
      const patterns = zod.def.options.map((o) => o._zod.pattern);
      return new RegExp(`^(${patterns.map((p) => cleanRegex(p.source)).join("|")})$`);
    }
    return void 0;
  });
  const first = def2.options.length === 1 ? def2.options[0]._zod.run : null;
  inst._zod.parse = (payload, ctx) => {
    if (first) {
      return first(payload, ctx);
    }
    let async = false;
    const results = [];
    for (const option of def2.options) {
      const result = option._zod.run({
        value: payload.value,
        issues: []
      }, ctx);
      if (result instanceof Promise) {
        results.push(result);
        async = true;
      } else {
        if (result.issues.length === 0)
          return result;
        results.push(result);
      }
    }
    if (!async)
      return handleUnionResults(results, payload, inst, ctx);
    return Promise.all(results).then((results2) => {
      return handleUnionResults(results2, payload, inst, ctx);
    });
  };
});
var $ZodIntersection = /* @__PURE__ */ $constructor("$ZodIntersection", (inst, def2) => {
  $ZodType.init(inst, def2);
  inst._zod.parse = (payload, ctx) => {
    const input = payload.value;
    const left = def2.left._zod.run({ value: input, issues: [] }, ctx);
    const right = def2.right._zod.run({ value: input, issues: [] }, ctx);
    const async = left instanceof Promise || right instanceof Promise;
    if (async) {
      return Promise.all([left, right]).then(([left2, right2]) => {
        return handleIntersectionResults(payload, left2, right2);
      });
    }
    return handleIntersectionResults(payload, left, right);
  };
});
function mergeValues(a, b) {
  if (a === b) {
    return { valid: true, data: a };
  }
  if (a instanceof Date && b instanceof Date && +a === +b) {
    return { valid: true, data: a };
  }
  if (isPlainObject(a) && isPlainObject(b)) {
    const bKeys = Object.keys(b);
    const sharedKeys = Object.keys(a).filter((key) => bKeys.indexOf(key) !== -1);
    const newObj = { ...a, ...b };
    if (Object.prototype.hasOwnProperty.call(newObj, "__proto__"))
      delete newObj.__proto__;
    for (const key of sharedKeys) {
      if (key === "__proto__")
        continue;
      const sharedValue = mergeValues(a[key], b[key]);
      if (!sharedValue.valid) {
        return {
          valid: false,
          mergeErrorPath: [key, ...sharedValue.mergeErrorPath]
        };
      }
      newObj[key] = sharedValue.data;
    }
    return { valid: true, data: newObj };
  }
  if (Array.isArray(a) && Array.isArray(b)) {
    if (a.length !== b.length) {
      return { valid: false, mergeErrorPath: [] };
    }
    const newArray = [];
    for (let index = 0; index < a.length; index++) {
      const itemA = a[index];
      const itemB = b[index];
      const sharedValue = mergeValues(itemA, itemB);
      if (!sharedValue.valid) {
        return {
          valid: false,
          mergeErrorPath: [index, ...sharedValue.mergeErrorPath]
        };
      }
      newArray.push(sharedValue.data);
    }
    return { valid: true, data: newArray };
  }
  return { valid: false, mergeErrorPath: [] };
}
function handleIntersectionResults(result, left, right) {
  const unrecKeys = /* @__PURE__ */ new Map();
  let unrecIssue;
  const keyIssues = /* @__PURE__ */ new Map();
  const collect = (iss, side) => {
    let keys;
    if (iss.code === "unrecognized_keys" && !iss.path?.length) {
      unrecIssue ?? (unrecIssue = iss);
      keys = iss.keys;
    } else if (iss.code === "invalid_key" && iss.origin === "record" && iss.path?.length === 1) {
      const k = String(iss.path[0]);
      if (!keyIssues.has(k))
        keyIssues.set(k, iss);
      keys = [k];
    } else {
      return false;
    }
    for (const k of keys) {
      if (!unrecKeys.has(k))
        unrecKeys.set(k, {});
      unrecKeys.get(k)[side] = true;
    }
    return true;
  };
  for (const iss of left.issues) {
    if (!collect(iss, "l"))
      result.issues.push(iss);
  }
  for (const iss of right.issues) {
    if (!collect(iss, "r"))
      result.issues.push(iss);
  }
  const bothKeys = [...unrecKeys].filter(([, f]) => f.l && f.r).map(([k]) => k);
  if (bothKeys.length) {
    const aggregated = unrecIssue ? bothKeys.filter((k) => unrecIssue.keys.includes(k)) : [];
    if (aggregated.length)
      result.issues.push({ ...unrecIssue, keys: aggregated });
    for (const k of bothKeys) {
      if (!aggregated.includes(k) && keyIssues.has(k))
        result.issues.push(keyIssues.get(k));
    }
  }
  const merged = mergeValues(left.value, right.value);
  if (!merged.valid) {
    if (aborted(result))
      return result;
    throw new Error(`Unmergable intersection. Error path: ${JSON.stringify(merged.mergeErrorPath)}`);
  }
  result.value = merged.data;
  return result;
}
var $ZodRecord = /* @__PURE__ */ $constructor("$ZodRecord", (inst, def2) => {
  $ZodType.init(inst, def2);
  const memo2 = globalConfig.memoizer;
  memo2?.attach(inst);
  inst._zod.parse = (payload, ctx) => {
    const input = payload.value;
    if (!isPlainObject(input)) {
      payload.issues.push({
        expected: "record",
        code: "invalid_type",
        input,
        inst
      });
      return payload;
    }
    const proms = [];
    const values = def2.keyType._zod.values;
    if (values && !def2.partial) {
      payload.value = memo2 ? memo2.alloc(inst, payload, {}, ctx) : {};
      const recordKeys = /* @__PURE__ */ new Set();
      for (const key of values) {
        if (typeof key === "string" || typeof key === "number" || typeof key === "symbol") {
          recordKeys.add(typeof key === "number" ? key.toString() : key);
          if (key === "__proto__")
            continue;
          const keyResult = def2.keyType._zod.run({ value: key, issues: [] }, ctx);
          if (keyResult instanceof Promise) {
            throw new Error("Async schemas not supported in object keys currently");
          }
          if (keyResult.issues.length) {
            payload.issues.push({
              code: "invalid_key",
              origin: "record",
              issues: keyResult.issues.map((iss) => finalizeIssue(iss, ctx, config())),
              input: key,
              path: [key],
              inst
            });
            continue;
          }
          const outKey = keyResult.value;
          if (outKey === "__proto__")
            continue;
          const result = def2.valueType._zod.run({ value: input[key], issues: [] }, ctx);
          if (result instanceof Promise) {
            proms.push(result.then((result2) => {
              if (result2.issues.length) {
                payload.issues.push(...prefixIssues(key, result2.issues));
              }
              payload.value[outKey] = result2.value;
            }));
          } else {
            if (result.issues.length) {
              payload.issues.push(...prefixIssues(key, result.issues));
            }
            payload.value[outKey] = result.value;
          }
        }
      }
      let unrecognized;
      for (const key in input) {
        if (!recordKeys.has(key)) {
          if (def2.mode === "loose") {
            if (key === "__proto__")
              continue;
            payload.value[key] = input[key];
          } else {
            unrecognized = unrecognized ?? [];
            unrecognized.push(key);
          }
        }
      }
      if (unrecognized && unrecognized.length > 0) {
        payload.issues.push({
          code: "unrecognized_keys",
          input,
          inst,
          keys: unrecognized,
          continue: true
        });
      }
    } else {
      payload.value = memo2 ? memo2.alloc(inst, payload, {}, ctx) : {};
      let unrecognized;
      for (const key of Reflect.ownKeys(input)) {
        if (key === "__proto__")
          continue;
        if (!Object.prototype.propertyIsEnumerable.call(input, key))
          continue;
        let keyResult = def2.keyType._zod.run({ value: key, issues: [] }, ctx);
        if (keyResult instanceof Promise) {
          throw new Error("Async schemas not supported in object keys currently");
        }
        const checkNumericKey = typeof key === "string" && number.test(key) && keyResult.issues.length;
        if (checkNumericKey) {
          const retryResult = def2.keyType._zod.run({ value: Number(key), issues: [] }, ctx);
          if (retryResult instanceof Promise) {
            throw new Error("Async schemas not supported in object keys currently");
          }
          if (retryResult.issues.length === 0) {
            keyResult = retryResult;
          }
        }
        if (keyResult.issues.length) {
          if (def2.mode === "loose") {
            payload.value[key] = input[key];
          } else if (values) {
            unrecognized = unrecognized ?? [];
            unrecognized.push(key);
          } else {
            payload.issues.push({
              code: "invalid_key",
              origin: "record",
              issues: keyResult.issues.map((iss) => finalizeIssue(iss, ctx, config())),
              input: key,
              path: [key],
              inst
            });
          }
          continue;
        }
        const outKey = keyResult.value;
        if (outKey === "__proto__")
          continue;
        const result = def2.valueType._zod.run({ value: input[key], issues: [] }, ctx);
        if (result instanceof Promise) {
          proms.push(result.then((result2) => {
            if (result2.issues.length) {
              payload.issues.push(...prefixIssues(key, result2.issues));
            }
            payload.value[outKey] = result2.value;
          }));
        } else {
          if (result.issues.length) {
            payload.issues.push(...prefixIssues(key, result.issues));
          }
          payload.value[outKey] = result.value;
        }
      }
      if (unrecognized && unrecognized.length > 0) {
        payload.issues.push({
          code: "unrecognized_keys",
          input,
          inst,
          keys: unrecognized,
          continue: true
        });
      }
    }
    if (proms.length) {
      return Promise.all(proms).then(() => payload);
    }
    return payload;
  };
});
var $ZodEnum = /* @__PURE__ */ $constructor("$ZodEnum", (inst, def2) => {
  $ZodType.init(inst, def2);
  const values = getEnumValues(def2.entries);
  const valuesSet = new Set(values);
  inst._zod.values = valuesSet;
  defineLazyInternal(inst, "pattern", (zod) => {
    const patternValues = getEnumValues(zod.def.entries).filter((k) => propertyKeyTypes.has(typeof k));
    return new RegExp(patternValues.length ? `^(${patternValues.map((o) => escapeRegex(o.toString())).join("|")})$` : "^[^\\s\\S]$");
  });
  inst._zod.parse = (payload, _ctx) => {
    const input = payload.value;
    if (valuesSet.has(input)) {
      return payload;
    }
    payload.issues.push({
      code: "invalid_value",
      values,
      input,
      inst
    });
    return payload;
  };
});
var $ZodLiteral = /* @__PURE__ */ $constructor("$ZodLiteral", (inst, def2) => {
  $ZodType.init(inst, def2);
  const values = new Set(def2.values);
  inst._zod.values = values;
  defineLazyInternal(inst, "pattern", (zod) => {
    const vals = zod.def.values;
    return new RegExp(vals.length ? `^(${vals.map((o) => typeof o === "string" ? escapeRegex(o) : o ? escapeRegex(o.toString()) : String(o)).join("|")})$` : "^[^\\s\\S]$");
  });
  inst._zod.parse = (payload, _ctx) => {
    const input = payload.value;
    if (values.has(input)) {
      return payload;
    }
    payload.issues.push({
      code: "invalid_value",
      values: def2.values,
      input,
      inst
    });
    return payload;
  };
});
var $ZodTransform = /* @__PURE__ */ $constructor("$ZodTransform", (inst, def2) => {
  $ZodType.init(inst, def2);
  inst._zod.optin = "optional";
  globalConfig.memoizer?.guard(inst);
  inst._zod.parse = (payload, ctx) => {
    if (ctx.direction === "backward") {
      throw new $ZodEncodeError(inst.constructor.name);
    }
    const _out = def2.transform(payload.value, payload);
    if (ctx.async) {
      const output = _out instanceof Promise ? _out : Promise.resolve(_out);
      return output.then((output2) => {
        payload.value = output2;
        return payload;
      });
    }
    if (_out instanceof Promise) {
      throw new $ZodAsyncError();
    }
    payload.value = _out;
    return payload;
  };
});
function handleOptionalResult(payload, result) {
  payload.value = result.issues.length ? void 0 : result.value;
  return payload;
}
var $ZodOptional = /* @__PURE__ */ $constructor("$ZodOptional", (inst, def2) => {
  $ZodType.init(inst, def2);
  defineLazyInternal(inst, "optin", (zod) => zod.def.innerType._zod.optin === "defaulted" ? "defaulted" : "optional");
  inst._zod.optout = "optional";
  defineLazyInternal(inst, "values", (zod) => {
    const values = zod.def.innerType._zod.values;
    return values ? /* @__PURE__ */ new Set([...values, void 0]) : void 0;
  });
  defineLazyInternal(inst, "pattern", (zod) => {
    const pattern = zod.def.innerType._zod.pattern;
    return pattern ? new RegExp(`^(${cleanRegex(pattern.source)})?$`) : void 0;
  });
  inst._zod.parse = (payload, ctx) => {
    if (payload.value === void 0) {
      if (def2.innerType._zod.optin !== "defaulted")
        return payload;
      const result = def2.innerType._zod.run({ value: payload.value, issues: [] }, ctx);
      if (result instanceof Promise)
        return result.then((result2) => handleOptionalResult(payload, result2));
      return handleOptionalResult(payload, result);
    }
    return def2.innerType._zod.run(payload, ctx);
  };
});
var $ZodExactOptional = /* @__PURE__ */ $constructor("$ZodExactOptional", (inst, def2) => {
  $ZodOptional.init(inst, def2);
  defineLazyInternal(inst, "values", (zod) => zod.def.innerType._zod.values);
  defineLazyInternal(inst, "pattern", (zod) => zod.def.innerType._zod.pattern);
  inst._zod.parse = (payload, ctx) => {
    return def2.innerType._zod.run(payload, ctx);
  };
});
var $ZodNullable = /* @__PURE__ */ $constructor("$ZodNullable", (inst, def2) => {
  $ZodType.init(inst, def2);
  defineLazyInternal(inst, "optin", (zod) => zod.def.innerType._zod.optin);
  defineLazyInternal(inst, "optout", (zod) => zod.def.innerType._zod.optout);
  defineLazyInternal(inst, "pattern", (zod) => {
    const pattern = zod.def.innerType._zod.pattern;
    return pattern ? new RegExp(`^(${cleanRegex(pattern.source)}|null)$`) : void 0;
  });
  defineLazyInternal(inst, "values", (zod) => {
    return zod.def.innerType._zod.values ? /* @__PURE__ */ new Set([...zod.def.innerType._zod.values, null]) : void 0;
  });
  inst._zod.parse = (payload, ctx) => {
    if (payload.value === null)
      return payload;
    return def2.innerType._zod.run(payload, ctx);
  };
});
var $ZodDefault = /* @__PURE__ */ $constructor("$ZodDefault", (inst, def2) => {
  $ZodType.init(inst, def2);
  inst._zod.optin = "defaulted";
  defineLazyInternal(inst, "values", (zod) => zod.def.innerType._zod.values);
  inst._zod.parse = (payload, ctx) => {
    if (ctx.direction === "backward") {
      return def2.innerType._zod.run(payload, ctx);
    }
    if (payload.value === void 0) {
      payload.value = def2.defaultValue;
      return payload;
    }
    const result = def2.innerType._zod.run(payload, ctx);
    if (result instanceof Promise) {
      return result.then((result2) => handleDefaultResult(result2, def2));
    }
    return handleDefaultResult(result, def2);
  };
});
function handleDefaultResult(payload, def2) {
  if (payload.value === void 0) {
    payload.value = def2.defaultValue;
  }
  return payload;
}
var $ZodPrefault = /* @__PURE__ */ $constructor("$ZodPrefault", (inst, def2) => {
  $ZodType.init(inst, def2);
  inst._zod.optin = "defaulted";
  defineLazyInternal(inst, "values", (zod) => zod.def.innerType._zod.values);
  inst._zod.parse = (payload, ctx) => {
    if (ctx.direction === "backward") {
      return def2.innerType._zod.run(payload, ctx);
    }
    if (payload.value === void 0) {
      payload.value = def2.defaultValue;
    }
    return def2.innerType._zod.run(payload, ctx);
  };
});
var $ZodNonOptional = /* @__PURE__ */ $constructor("$ZodNonOptional", (inst, def2) => {
  $ZodType.init(inst, def2);
  defineLazyInternal(inst, "values", (zod) => {
    const v = zod.def.innerType._zod.values;
    return v ? new Set([...v].filter((x) => x !== void 0)) : void 0;
  });
  inst._zod.parse = (payload, ctx) => {
    const result = def2.innerType._zod.run(payload, ctx);
    if (result instanceof Promise) {
      return result.then((result2) => handleNonOptionalResult(result2, inst));
    }
    return handleNonOptionalResult(result, inst);
  };
});
function handleNonOptionalResult(payload, inst) {
  if (!payload.issues.length && payload.value === void 0) {
    payload.issues.push({
      code: "invalid_type",
      expected: "nonoptional",
      input: payload.value,
      inst
    });
  }
  return payload;
}
function handleCatchResult(payload, result, def2, ctx) {
  if (!result.issues.length) {
    payload.value = result.value;
    if (result.memo)
      payload.memo = true;
    return payload;
  }
  payload.value = def2.catchValue({
    ...result,
    value: payload.value,
    error: {
      issues: result.issues.map((iss) => finalizeIssue(iss, ctx, config()))
    },
    input: payload.value
  });
  return payload;
}
var $ZodCatch = /* @__PURE__ */ $constructor("$ZodCatch", (inst, def2) => {
  $ZodType.init(inst, def2);
  defineLazyInternal(inst, "optin", (zod) => zod.def.innerType._zod.optin === "defaulted" ? "defaulted" : "optional");
  defineLazyInternal(inst, "optout", (zod) => zod.def.innerType._zod.optout);
  defineLazyInternal(inst, "values", (zod) => zod.def.innerType._zod.values);
  inst._zod.parse = (payload, ctx) => {
    if (ctx.direction === "backward") {
      return def2.innerType._zod.run(payload, ctx);
    }
    const result = def2.innerType._zod.run({ value: payload.value, issues: [] }, ctx);
    if (result instanceof Promise) {
      return result.then((result2) => handleCatchResult(payload, result2, def2, ctx));
    }
    return handleCatchResult(payload, result, def2, ctx);
  };
});
var $ZodPipe = /* @__PURE__ */ $constructor("$ZodPipe", (inst, def2) => {
  $ZodType.init(inst, def2);
  defineLazyInternal(inst, "values", (zod) => zod.def.in._zod.values);
  defineLazyInternal(inst, "optin", (zod) => zod.def.in._zod.optin);
  defineLazyInternal(inst, "optout", (zod) => zod.def.out._zod.optout);
  defineLazyInternal(inst, "propValues", (zod) => zod.def.in._zod.propValues);
  inst._zod.parse = (payload, ctx) => {
    if (ctx.direction === "backward") {
      const right = def2.out._zod.run(payload, ctx);
      if (right instanceof Promise) {
        return right.then((right2) => handlePipeResult(right2, def2.in, ctx));
      }
      return handlePipeResult(right, def2.in, ctx);
    }
    const left = def2.in._zod.run(payload, ctx);
    if (left instanceof Promise) {
      return left.then((left2) => handlePipeResult(left2, def2.out, ctx));
    }
    return handlePipeResult(left, def2.out, ctx);
  };
});
function handlePipeResult(left, next, ctx) {
  if (left.issues.some((iss) => iss.code !== "unrecognized_keys")) {
    left.aborted = true;
    return left;
  }
  return next._zod.run({ value: left.value, issues: left.issues }, ctx);
}
var $ZodPreprocess = /* @__PURE__ */ $constructor("$ZodPreprocess", (inst, def2) => {
  $ZodPipe.init(inst, def2);
});
var $ZodReadonly = /* @__PURE__ */ $constructor("$ZodReadonly", (inst, def2) => {
  $ZodType.init(inst, def2);
  defineLazyInternal(inst, "propValues", (zod) => zod.def.innerType._zod.propValues);
  defineLazyInternal(inst, "values", (zod) => zod.def.innerType._zod.values);
  defineLazyInternal(inst, "optin", (zod) => zod.def.innerType?._zod?.optin);
  defineLazyInternal(inst, "optout", (zod) => zod.def.innerType?._zod?.optout);
  inst._zod.parse = (payload, ctx) => {
    if (ctx.direction === "backward") {
      return def2.innerType._zod.run(payload, ctx);
    }
    const result = def2.innerType._zod.run(payload, ctx);
    if (result instanceof Promise) {
      return result.then(handleReadonlyResult);
    }
    return handleReadonlyResult(result);
  };
});
function handleReadonlyResult(payload) {
  if (!payload.memo)
    payload.value = Object.freeze(payload.value);
  return payload;
}
var $ZodCustom = /* @__PURE__ */ $constructor("$ZodCustom", (inst, def2) => {
  $ZodCheck.init(inst, def2);
  $ZodType.init(inst, def2);
  inst._zod.parse = (payload, _) => {
    return payload;
  };
  inst._zod.check = (payload) => {
    const input = payload.value;
    const r = def2.fn(input);
    if (r instanceof Promise) {
      return r.then((r2) => handleRefineResult(r2, payload, input, inst));
    }
    handleRefineResult(r, payload, input, inst);
    return;
  };
});
function handleRefineResult(result, payload, input, inst) {
  if (!result) {
    const _iss = {
      code: "custom",
      input,
      inst,
      // incorporates params.error into issue reporting
      path: [...inst._zod.def.path ?? []],
      // incorporates params.error into issue reporting
      continue: !inst._zod.def.abort
      // params: inst._zod.def.params,
    };
    if (inst._zod.def.params)
      _iss.params = inst._zod.def.params;
    payload.issues.push(issue(_iss));
  }
}

// node_modules/zod/v4/core/memoizer.js
var $ZodCyclicError = class extends Error {
  constructor() {
    super(`Cannot parse a reference cycle that closes through a transform`);
    this.name = "ZodCyclicError";
  }
};
var STATE = "~memo";
var NO_ISSUES = [];
function isRef(value) {
  return value !== null && typeof value === "object";
}
function cloneIssues(issues) {
  return issues.map((iss) => iss.path ? { ...iss, path: iss.path.slice() } : { ...iss });
}
var recursive = /* @__PURE__ */ new WeakMap();
var NONE = 0;
var ASSUMED = 1;
var PROVEN = 2;
function isRecursive(inst, stack, resolve) {
  const cached2 = recursive.get(inst);
  if (cached2 !== void 0)
    return cached2 ? PROVEN : NONE;
  if (stack.has(inst))
    return PROVEN;
  stack.add(inst);
  let result = NONE;
  const check = (child) => {
    if (result !== PROVEN && child?._zod) {
      const answer = isRecursive(child, stack, resolve);
      if (answer > result)
        result = answer;
    }
  };
  const shape = (sh, spread) => {
    let answer = NONE;
    for (const key of Reflect.ownKeys(sh)) {
      const desc = Object.getOwnPropertyDescriptor(sh, key);
      if (spread && !desc.enumerable)
        continue;
      const child = desc.get ? ASSUMED : desc.value?._zod ? isRecursive(desc.value, stack, resolve) : NONE;
      if (child > answer)
        answer = child;
    }
    return answer;
  };
  const merge2 = (answer) => {
    if (answer > result)
      result = answer;
  };
  const def2 = inst._zod.def;
  const kind = def2.type;
  switch (kind) {
    case "object": {
      const raw = rawShape(def2);
      merge2(raw ? shape(raw, true) : ASSUMED);
      check(def2.catchall);
      break;
    }
    case "array":
      check(def2.element);
      break;
    case "tuple":
      for (const el of def2.items)
        check(el);
      check(def2.rest);
      break;
    case "record":
    case "map":
      check(def2.keyType);
      check(def2.valueType);
      break;
    case "set":
      check(def2.valueType);
      break;
    case "union":
      for (const el of def2.options)
        check(el);
      break;
    case "intersection":
      check(def2.left);
      check(def2.right);
      break;
    case "optional":
    case "nullable":
    case "default":
    case "prefault":
    case "catch":
    case "readonly":
    case "nonoptional":
    case "promise":
    case "success":
      check(def2.innerType);
      break;
    case "pipe":
      check(def2.in);
      check(def2.out);
      break;
    case "function":
      check(def2.input);
      check(def2.output);
      break;
    // `$ZodLazy` caches its inner on the def, so a resolved edge is followed exactly
    case "lazy": {
      const inner = def2._cachedInner ?? (resolve ? inst._zod.innerType : void 0);
      merge2(inner ? isRecursive(inner, stack, false) : ASSUMED);
      break;
    }
    // a leaf by choice: `parts` are regex fragments, not data positions
    case "template_literal":
    // leaves
    case "string":
    case "number":
    case "int":
    case "boolean":
    case "bigint":
    case "symbol":
    case "undefined":
    case "null":
    case "void":
    case "never":
    case "any":
    case "unknown":
    case "date":
    case "nan":
    case "enum":
    case "literal":
    case "file":
    case "transform":
    case "custom":
      break;
    default: {
      kind;
      for (const key in def2) {
        const desc = Object.getOwnPropertyDescriptor(def2, key);
        if (!desc || desc.get)
          continue;
        const value = desc.value;
        if (!value || typeof value !== "object")
          continue;
        if (value._zod)
          check(value);
        else if (Array.isArray(value))
          for (const el of value)
            check(el);
      }
    }
  }
  stack.delete(inst);
  return settle(inst, result);
}
function settle(inst, answer) {
  if (answer !== ASSUMED)
    recursive.set(inst, answer === PROVEN);
  return answer;
}
function bucketFor(state2, inst) {
  let bucket = state2.buckets.get(inst);
  if (!bucket) {
    bucket = /* @__PURE__ */ new WeakMap();
    state2.buckets.set(inst, bucket);
  }
  return bucket;
}
var handoff;
var open = [];
var memo = {
  alloc(_inst, payload, empty) {
    const bucket = handoff;
    if (!bucket)
      return empty;
    handoff = void 0;
    const entry = { value: empty, issues: null };
    bucket.set(payload.value, entry);
    open.push(entry);
    return empty;
  },
  guard(inst) {
    var _a3;
    (_a3 = inst._zod).deferred ?? (_a3.deferred = []);
    inst._zod.deferred.push(() => {
      const base = inst._zod.parse;
      const wrapped = (payload, ctx) => {
        if (ctx.direction !== "backward" && isBackEdge(ctx, payload.value))
          throw new $ZodCyclicError();
        return base(payload, ctx);
      };
      inst._zod.parse = wrapped;
      if (inst._zod.run === base)
        inst._zod.run = wrapped;
    });
  },
  attach(inst) {
    var _a3;
    let isRecursiveInst;
    let rechecked = false;
    let lastCtx;
    let lastBucket;
    (_a3 = inst._zod).deferred ?? (_a3.deferred = []);
    inst._zod.deferred.push(() => {
      const base = inst._zod.parse;
      const wrapped = (payload, ctx) => {
        if (isRecursiveInst === void 0) {
          const walked = isRecursive(inst, /* @__PURE__ */ new Set(), false);
          if (walked === NONE) {
            inst._zod.parse = base;
            if (inst._zod.run === wrapped)
              inst._zod.run = base;
            return base(payload, ctx);
          }
          if (walked === PROVEN || rechecked)
            isRecursiveInst = true;
          else
            rechecked = true;
        }
        const input = payload.value;
        if (!isRef(input))
          return base(payload, ctx);
        let state2 = ctx[STATE];
        if (!state2) {
          state2 = { buckets: /* @__PURE__ */ new WeakMap(), backEdges: void 0 };
          ctx[STATE] = state2;
        }
        let bucket;
        if (lastCtx === ctx) {
          bucket = lastBucket;
        } else {
          bucket = bucketFor(state2, inst);
          lastCtx = ctx;
          lastBucket = bucket;
        }
        const hit = bucket.get(input);
        if (hit) {
          payload.value = hit.value;
          if (hit.issues) {
            if (hit.issues.length)
              payload.issues.push(...cloneIssues(hit.issues));
          } else {
            payload.memo = true;
            state2.backEdges ?? (state2.backEdges = /* @__PURE__ */ new WeakSet());
            state2.backEdges.add(hit.value);
          }
          return payload;
        }
        handoff = bucket;
        const depth = open.length;
        const result = base(payload, ctx);
        handoff = void 0;
        const entry = open.length > depth ? open.pop() : void 0;
        if (result instanceof Promise) {
          return result.then((r) => {
            if (entry)
              entry.issues = r.issues.length ? cloneIssues(r.issues) : NO_ISSUES;
            return r;
          });
        }
        if (entry)
          entry.issues = result.issues.length ? cloneIssues(result.issues) : NO_ISSUES;
        return result;
      };
      inst._zod.parse = wrapped;
      if (inst._zod.run === base)
        inst._zod.run = wrapped;
    });
  }
};
function memoizer() {
  return memo;
}
function isBackEdge(ctx, value) {
  const backEdges = ctx[STATE]?.backEdges;
  return backEdges !== void 0 && isRef(value) && backEdges.has(value);
}

// node_modules/zod/v4/locales/en.js
var error = () => {
  const Sizable = {
    string: { unit: "characters", verb: "to have" },
    file: { unit: "bytes", verb: "to have" },
    array: { unit: "items", verb: "to have" },
    set: { unit: "items", verb: "to have" },
    map: { unit: "entries", verb: "to have" }
  };
  function getSizing(origin) {
    return Sizable[origin] ?? null;
  }
  const FormatDictionary = {
    regex: "input",
    email: "email address",
    url: "URL",
    emoji: "emoji",
    uuid: "UUID",
    uuidv4: "UUIDv4",
    uuidv6: "UUIDv6",
    nanoid: "nanoid",
    guid: "GUID",
    cuid: "cuid",
    cuid2: "cuid2",
    ulid: "ULID",
    xid: "XID",
    ksuid: "KSUID",
    datetime: "ISO datetime",
    date: "ISO date",
    time: "ISO time",
    duration: "ISO duration",
    ipv4: "IPv4 address",
    ipv6: "IPv6 address",
    mac: "MAC address",
    cidrv4: "IPv4 range",
    cidrv6: "IPv6 range",
    base64: "base64-encoded string",
    base64url: "base64url-encoded string",
    json_string: "JSON string",
    e164: "E.164 number",
    currency_code: "currency code",
    credit_card: "credit card number",
    iban: "IBAN",
    jwt: "JWT",
    template_literal: "input"
  };
  const TypeDictionary = {
    // Compatibility: "nan" -> "NaN" for display
    nan: "NaN"
    // All other type names omitted - they fall back to raw values via ?? operator
  };
  function getTypeName(type, input) {
    if (type === "number" && typeof input === "number" && !Number.isFinite(input)) {
      return String(input);
    }
    return TypeDictionary[type] ?? type;
  }
  return (issue2) => {
    switch (issue2.code) {
      case "invalid_type": {
        const expected = getTypeName(issue2.expected);
        const receivedType = parsedType(issue2.input);
        const received = getTypeName(receivedType, issue2.input);
        return `Invalid input: expected ${expected}, received ${received}`;
      }
      case "invalid_value":
        if (issue2.values.length === 1)
          return `Invalid input: expected ${stringifyPrimitive(issue2.values[0])}`;
        return `Invalid option: expected one of ${joinValues(issue2.values, "|")}`;
      case "too_big": {
        const adj = issue2.exact ? "exactly " : issue2.inclusive ? "<=" : "<";
        const sizing = getSizing(issue2.origin);
        if (sizing)
          return `Too big: expected ${issue2.origin ?? "value"} to have ${adj}${issue2.maximum.toString()} ${sizing.unit ?? "elements"}`;
        return `Too big: expected ${issue2.origin ?? "value"} to be ${adj}${issue2.maximum.toString()}`;
      }
      case "too_small": {
        const adj = issue2.exact ? "exactly " : issue2.inclusive ? ">=" : ">";
        const sizing = getSizing(issue2.origin);
        if (sizing) {
          return `Too small: expected ${issue2.origin} to have ${adj}${issue2.minimum.toString()} ${sizing.unit}`;
        }
        return `Too small: expected ${issue2.origin} to be ${adj}${issue2.minimum.toString()}`;
      }
      case "invalid_format": {
        const _issue = issue2;
        if (_issue.format === "starts_with") {
          return `Invalid string: must start with "${_issue.prefix}"`;
        }
        if (_issue.format === "ends_with")
          return `Invalid string: must end with "${_issue.suffix}"`;
        if (_issue.format === "includes")
          return `Invalid string: must include "${_issue.includes}"`;
        if (_issue.format === "regex")
          return `Invalid string: must match pattern ${_issue.pattern}`;
        return `Invalid ${FormatDictionary[_issue.format] ?? issue2.format}`;
      }
      case "not_multiple_of":
        return `Invalid number: must be a multiple of ${issue2.divisor}`;
      case "unrecognized_keys":
        return `Unrecognized key${issue2.keys.length > 1 ? "s" : ""}: ${joinValues(issue2.keys, ", ")}`;
      case "invalid_key":
        return `Invalid key in ${issue2.origin}`;
      case "invalid_union":
        if (issue2.options && Array.isArray(issue2.options) && issue2.options.length > 0) {
          const opts = issue2.options.map((o) => `'${o}'`).join(" | ");
          return `Invalid discriminator value. Expected ${opts}`;
        }
        if (issue2.inclusive === false) {
          return "Invalid input: more than one option matched";
        }
        return "Invalid input";
      case "invalid_element":
        return `Invalid value in ${issue2.origin}`;
      default:
        return `Invalid input`;
    }
  };
};
function en_default() {
  return {
    localeError: error()
  };
}

// node_modules/zod/v4/core/registries.js
var _a2;
var $ZodRegistry = class {
  constructor() {
    this._map = /* @__PURE__ */ new WeakMap();
    this._idmap = /* @__PURE__ */ new Map();
  }
  add(schema, ..._meta) {
    const meta2 = _meta[0];
    this._map.set(schema, meta2);
    if (meta2 && typeof meta2 === "object" && "id" in meta2) {
      this._idmap.set(meta2.id, schema);
    }
    return this;
  }
  clear() {
    this._map = /* @__PURE__ */ new WeakMap();
    this._idmap = /* @__PURE__ */ new Map();
    return this;
  }
  remove(schema) {
    const meta2 = this._map.get(schema);
    if (meta2 && typeof meta2 === "object" && "id" in meta2) {
      this._idmap.delete(meta2.id);
    }
    this._map.delete(schema);
    return this;
  }
  get(schema) {
    const p = schema._zod.parent;
    if (p) {
      const pm = { ...this.get(p) ?? {} };
      delete pm.id;
      const f = { ...pm, ...this._map.get(schema) };
      return Object.keys(f).length ? f : void 0;
    }
    return this._map.get(schema);
  }
  has(schema) {
    return this._map.has(schema);
  }
};
function registry() {
  return new $ZodRegistry();
}
(_a2 = globalThis).__zod_globalRegistry ?? (_a2.__zod_globalRegistry = registry());
var globalRegistry = globalThis.__zod_globalRegistry;

// node_modules/zod/v4/core/api.js
function snapshotChecks(def2) {
  if (def2.checks)
    def2.checks = [...def2.checks];
  return def2;
}
// @__NO_SIDE_EFFECTS__
function _string(Class2, params) {
  return new Class2(snapshotChecks({ type: "string", ...normalizeParams(params) }));
}
// @__NO_SIDE_EFFECTS__
function _email(Class2, params) {
  return new Class2({
    type: "string",
    format: "email",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _guid(Class2, params) {
  return new Class2({
    type: "string",
    format: "guid",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _uuid(Class2, params) {
  return new Class2({
    type: "string",
    format: "uuid",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _uuidv4(Class2, params) {
  return new Class2({
    type: "string",
    format: "uuid",
    check: "string_format",
    abort: false,
    version: "v4",
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _uuidv6(Class2, params) {
  return new Class2({
    type: "string",
    format: "uuid",
    check: "string_format",
    abort: false,
    version: "v6",
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _uuidv7(Class2, params) {
  return new Class2({
    type: "string",
    format: "uuid",
    check: "string_format",
    abort: false,
    version: "v7",
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _url(Class2, params) {
  return new Class2({
    type: "string",
    format: "url",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _emoji2(Class2, params) {
  return new Class2({
    type: "string",
    format: "emoji",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _nanoid(Class2, params) {
  return new Class2({
    type: "string",
    format: "nanoid",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _cuid(Class2, params) {
  return new Class2({
    type: "string",
    format: "cuid",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _cuid2(Class2, params) {
  return new Class2({
    type: "string",
    format: "cuid2",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _ulid(Class2, params) {
  return new Class2({
    type: "string",
    format: "ulid",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _xid(Class2, params) {
  return new Class2({
    type: "string",
    format: "xid",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _ksuid(Class2, params) {
  return new Class2({
    type: "string",
    format: "ksuid",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _ipv4(Class2, params) {
  return new Class2({
    type: "string",
    format: "ipv4",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _ipv6(Class2, params) {
  return new Class2({
    type: "string",
    format: "ipv6",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _cidrv4(Class2, params) {
  return new Class2({
    type: "string",
    format: "cidrv4",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _cidrv6(Class2, params) {
  return new Class2({
    type: "string",
    format: "cidrv6",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _base64(Class2, params) {
  return new Class2({
    type: "string",
    format: "base64",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _base64url(Class2, params) {
  return new Class2({
    type: "string",
    format: "base64url",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _e164(Class2, params) {
  return new Class2({
    type: "string",
    format: "e164",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _jwt(Class2, params) {
  return new Class2({
    type: "string",
    format: "jwt",
    check: "string_format",
    abort: false,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _isoDateTime(Class2, params) {
  return new Class2({
    type: "string",
    format: "datetime",
    check: "string_format",
    offset: false,
    local: false,
    precision: null,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _isoDate(Class2, params) {
  return new Class2({
    type: "string",
    format: "date",
    check: "string_format",
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _isoTime(Class2, params) {
  return new Class2({
    type: "string",
    format: "time",
    check: "string_format",
    precision: null,
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _isoDuration(Class2, params) {
  return new Class2({
    type: "string",
    format: "duration",
    check: "string_format",
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _number(Class2, params) {
  return new Class2(snapshotChecks({ type: "number", checks: [], ...normalizeParams(params) }));
}
// @__NO_SIDE_EFFECTS__
function _int(Class2, params) {
  return new Class2({
    type: "number",
    check: "number_format",
    abort: false,
    format: "safeint",
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _boolean(Class2, params) {
  return new Class2({
    type: "boolean",
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _unknown(Class2) {
  return new Class2({
    type: "unknown"
  });
}
// @__NO_SIDE_EFFECTS__
function _never(Class2, params) {
  return new Class2({
    type: "never",
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _lt(value, params) {
  return new $ZodCheckLessThan({
    check: "less_than",
    ...normalizeParams(params),
    value,
    inclusive: false
  });
}
// @__NO_SIDE_EFFECTS__
function _lte(value, params) {
  return new $ZodCheckLessThan({
    check: "less_than",
    ...normalizeParams(params),
    value,
    inclusive: true
  });
}
// @__NO_SIDE_EFFECTS__
function _gt(value, params) {
  return new $ZodCheckGreaterThan({
    check: "greater_than",
    ...normalizeParams(params),
    value,
    inclusive: false
  });
}
// @__NO_SIDE_EFFECTS__
function _gte(value, params) {
  return new $ZodCheckGreaterThan({
    check: "greater_than",
    ...normalizeParams(params),
    value,
    inclusive: true
  });
}
// @__NO_SIDE_EFFECTS__
function _multipleOf(value, params) {
  return new $ZodCheckMultipleOf({
    check: "multiple_of",
    ...normalizeParams(params),
    value
  });
}
// @__NO_SIDE_EFFECTS__
function _maxLength(maximum, params) {
  const ch = new $ZodCheckMaxLength({
    check: "max_length",
    ...normalizeParams(params),
    maximum
  });
  return ch;
}
// @__NO_SIDE_EFFECTS__
function _minLength(minimum, params) {
  return new $ZodCheckMinLength({
    check: "min_length",
    ...normalizeParams(params),
    minimum
  });
}
// @__NO_SIDE_EFFECTS__
function _length(length, params) {
  return new $ZodCheckLengthEquals({
    check: "length_equals",
    ...normalizeParams(params),
    length
  });
}
// @__NO_SIDE_EFFECTS__
function _regex(pattern, params) {
  return new $ZodCheckRegex({
    check: "string_format",
    format: "regex",
    ...normalizeParams(params),
    pattern
  });
}
// @__NO_SIDE_EFFECTS__
function _lowercase(params) {
  return new $ZodCheckLowerCase({
    check: "string_format",
    format: "lowercase",
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _uppercase(params) {
  return new $ZodCheckUpperCase({
    check: "string_format",
    format: "uppercase",
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _includes(includes, params) {
  return new $ZodCheckIncludes({
    check: "string_format",
    format: "includes",
    ...normalizeParams(params),
    includes
  });
}
// @__NO_SIDE_EFFECTS__
function _startsWith(prefix, params) {
  return new $ZodCheckStartsWith({
    check: "string_format",
    format: "starts_with",
    ...normalizeParams(params),
    prefix
  });
}
// @__NO_SIDE_EFFECTS__
function _endsWith(suffix, params) {
  return new $ZodCheckEndsWith({
    check: "string_format",
    format: "ends_with",
    ...normalizeParams(params),
    suffix
  });
}
// @__NO_SIDE_EFFECTS__
function _overwrite(tx) {
  return new $ZodCheckOverwrite({
    check: "overwrite",
    tx
  });
}
// @__NO_SIDE_EFFECTS__
function _normalize(form) {
  return /* @__PURE__ */ _overwrite((input) => input.normalize(form));
}
// @__NO_SIDE_EFFECTS__
function _trim() {
  return /* @__PURE__ */ _overwrite((input) => input.trim());
}
// @__NO_SIDE_EFFECTS__
function _toLowerCase() {
  return /* @__PURE__ */ _overwrite((input) => input.toLowerCase());
}
// @__NO_SIDE_EFFECTS__
function _toUpperCase() {
  return /* @__PURE__ */ _overwrite((input) => input.toUpperCase());
}
// @__NO_SIDE_EFFECTS__
function _slugify() {
  return /* @__PURE__ */ _overwrite((input) => slugify(input));
}
// @__NO_SIDE_EFFECTS__
function _array(Class2, element, params) {
  return new Class2({
    type: "array",
    element,
    // get element() {
    //   return element;
    // },
    ...normalizeParams(params)
  });
}
// @__NO_SIDE_EFFECTS__
function _refine(Class2, fn, _params) {
  const schema = new Class2({
    type: "custom",
    check: "custom",
    fn,
    ...normalizeParams(_params)
  });
  return schema;
}
// @__NO_SIDE_EFFECTS__
function _superRefine(fn, params) {
  const ch = /* @__PURE__ */ _check((payload) => {
    payload.addIssue = (issue2) => {
      if (typeof issue2 === "string") {
        payload.issues.push(issue(issue2, payload.value, ch._zod.def));
      } else {
        const _issue = issue2;
        if (_issue.fatal)
          _issue.continue = false;
        _issue.code ?? (_issue.code = "custom");
        if (!("input" in _issue))
          _issue.input = payload.value;
        _issue.inst ?? (_issue.inst = ch);
        _issue.continue ?? (_issue.continue = !ch._zod.def.abort);
        payload.issues.push(issue(_issue));
      }
    };
    return fn(payload.value, payload);
  }, params);
  return ch;
}
// @__NO_SIDE_EFFECTS__
function _check(fn, params) {
  const ch = new $ZodCheck({
    check: "custom",
    ...normalizeParams(params)
  });
  ch._zod.check = fn;
  return ch;
}

// node_modules/zod/v4/core/to-json-schema.js
function assignProps(target, ...sources) {
  for (const source of sources) {
    for (const key of Reflect.ownKeys(source)) {
      if (Object.prototype.propertyIsEnumerable.call(source, key)) {
        assignProp(target, key, source[key]);
      }
    }
  }
  return target;
}
function initializeContext(params) {
  let target = params?.target ?? "draft-2020-12";
  if (target === "draft-4")
    target = "draft-04";
  if (target === "draft-7")
    target = "draft-07";
  return {
    processors: params.processors ?? {},
    metadataRegistry: params?.metadata ?? globalRegistry,
    target,
    unrepresentable: params?.unrepresentable ?? "throw",
    override: params?.override ?? (() => {
    }),
    io: params?.io ?? "output",
    counter: 0,
    seen: /* @__PURE__ */ new Map(),
    sharedDefsExtractedFor: void 0,
    sharedEmitDoneFor: void 0,
    cycles: params?.cycles ?? "ref",
    reused: params?.reused ?? "inline",
    intersections: [],
    deferred: [],
    external: params?.external ?? void 0
  };
}
function handleUnrepresentable(schema, ctx, json, params, message) {
  const result = typeof ctx.unrepresentable === "function" ? ctx.unrepresentable({ zodSchema: schema, path: params.path, message }) : ctx.unrepresentable;
  if (result === "any")
    return false;
  if (result === void 0 || result === "throw")
    throw new Error(message);
  Object.assign(json, result);
  return true;
}
function processSchema(schema, ctx, _params = { path: [], schemaPath: [] }) {
  var _a3;
  const def2 = schema._zod.def;
  const seen = ctx.seen.get(schema);
  if (seen) {
    seen.count++;
    const isCycle = _params.schemaPath.includes(schema);
    if (isCycle) {
      seen.cycle = _params.path;
    }
    return seen.schema;
  }
  const result = { schema: {}, count: 1, cycle: void 0, path: _params.path };
  ctx.seen.set(schema, result);
  ctx.sharedDefsExtractedFor = void 0;
  ctx.sharedEmitDoneFor = void 0;
  const overrideSchema = schema._zod.toJSONSchema?.();
  if (overrideSchema) {
    result.schema = overrideSchema;
  } else {
    const params = {
      ..._params,
      schemaPath: [..._params.schemaPath, schema],
      path: _params.path
    };
    if (schema._zod.processJSONSchema) {
      schema._zod.processJSONSchema(ctx, result.schema, params);
    } else {
      const _json = result.schema;
      const processor = ctx.processors[def2.type];
      if (!processor) {
        throw new Error(`[toJSONSchema]: Non-representable type encountered: ${def2.type}`);
      }
      processor(schema, ctx, _json, params);
    }
    const parent = schema._zod.parent;
    if (parent) {
      if (!result.ref)
        result.ref = parent;
      processSchema(parent, ctx, params);
      ctx.seen.get(parent).isParent = true;
    }
  }
  const meta2 = ctx.metadataRegistry.get(schema);
  if (meta2)
    assignProps(result.schema, meta2);
  if (ctx.io === "input" && isTransforming(schema)) {
    delete result.schema.examples;
    delete result.schema.default;
  }
  if (ctx.io === "input" && "_prefault" in result.schema)
    (_a3 = result.schema).default ?? (_a3.default = result.schema._prefault);
  delete result.schema._prefault;
  const _result = ctx.seen.get(schema);
  return _result.schema;
}
function encodeJSONPointerSegment(segment) {
  return segment.replace(/~/g, "~0").replace(/\//g, "~1");
}
function extractDefs(ctx, schema) {
  const root = ctx.seen.get(schema);
  if (!root)
    throw new Error("Unprocessed schema. This is a bug in Zod.");
  if (ctx.external && ctx.sharedDefsExtractedFor === ctx.external)
    return;
  const idToSchema = /* @__PURE__ */ new Map();
  for (const entry of ctx.seen.entries()) {
    const id = ctx.metadataRegistry.get(entry[0])?.id;
    if (id) {
      const existing = idToSchema.get(id);
      if (existing && existing !== entry[0]) {
        throw new Error(`Duplicate schema id "${id}" detected during JSON Schema conversion. Two different schemas cannot share the same id when converted together.`);
      }
      idToSchema.set(id, entry[0]);
    }
  }
  const makeURI = (entry) => {
    const defsSegment = ctx.target === "draft-2020-12" ? "$defs" : "definitions";
    if (ctx.external) {
      const externalId = ctx.external.registry.get(entry[0])?.id;
      const uriGenerator = ctx.external.uri ?? ((id2) => id2);
      if (externalId) {
        return { ref: uriGenerator(externalId) };
      }
      const id = entry[1].defId ?? entry[1].schema.id ?? `schema${ctx.counter++}`;
      entry[1].defId = id;
      return { defId: id, ref: `${uriGenerator("__shared")}#/${defsSegment}/${encodeJSONPointerSegment(id)}` };
    }
    const uriPrefix = `#`;
    const defUriPrefix = `${uriPrefix}/${defsSegment}/`;
    if (entry[1] === root && !entry[1].schema.id) {
      return { ref: uriPrefix };
    }
    const defId = entry[1].schema.id ?? `__schema${ctx.counter++}`;
    return { defId, ref: defUriPrefix + encodeJSONPointerSegment(defId) };
  };
  const extractToDef = (entry) => {
    if (entry[1].schema.$ref) {
      return;
    }
    const seen = entry[1];
    const { ref, defId } = makeURI(entry);
    seen.def = { ...seen.schema };
    if (defId)
      seen.defId = defId;
    const schema2 = seen.schema;
    for (const key in schema2) {
      delete schema2[key];
    }
    schema2.$ref = ref;
  };
  if (ctx.cycles === "throw") {
    for (const entry of ctx.seen.entries()) {
      const seen = entry[1];
      if (seen.cycle) {
        throw new Error(`Cycle detected: #/${seen.cycle?.join("/")}/<root>

Set the \`cycles\` parameter to \`"ref"\` to resolve cyclical schemas with defs.`);
      }
    }
  }
  for (const entry of ctx.seen.entries()) {
    const seen = entry[1];
    if (schema === entry[0]) {
      extractToDef(entry);
      continue;
    }
    if (ctx.external) {
      const ext = ctx.external.registry.get(entry[0])?.id;
      if (schema !== entry[0] && ext) {
        extractToDef(entry);
        continue;
      }
    }
    const id = ctx.metadataRegistry.get(entry[0])?.id;
    if (id) {
      extractToDef(entry);
      continue;
    }
    if (seen.cycle) {
      extractToDef(entry);
      continue;
    }
    if (seen.count > 1) {
      if (ctx.reused === "ref") {
        extractToDef(entry);
      }
    }
  }
  if (ctx.external)
    ctx.sharedDefsExtractedFor = ctx.external;
}
function compactTypeUnion(schema) {
  const options2 = schema.anyOf;
  if (!Array.isArray(options2) || options2.length === 0 || schema.type !== void 0)
    return;
  const types = [];
  for (const option of options2) {
    if (!option || typeof option !== "object")
      return;
    compactTypeUnion(option);
    const keys = Object.keys(option);
    if (keys.length !== 1 || keys[0] !== "type")
      return;
    const type = option.type;
    for (const member of Array.isArray(type) ? type : [type]) {
      if (typeof member !== "string")
        return;
      if (!types.includes(member))
        types.push(member);
    }
  }
  delete schema.anyOf;
  schema.type = types.length === 1 ? types[0] : types;
}
var FOLDABLE_KEYS = /* @__PURE__ */ new Set(["type", "properties", "required", "additionalProperties"]);
var UNION_KEYS = ["oneOf", "anyOf"];
function undeclaredConstraint(member) {
  const extra = member.additionalProperties;
  if (extra === void 0 || extra === false || typeof extra !== "object" || extra === null)
    return null;
  return Object.keys(extra).length ? extra : null;
}
function foldObjects(members2) {
  const objects = [];
  for (const member of members2) {
    if (typeof member !== "object" || member.type !== "object")
      return null;
    for (const key in member) {
      if (!FOLDABLE_KEYS.has(key))
        return null;
    }
    objects.push(member);
  }
  const properties = {};
  const required2 = /* @__PURE__ */ new Set();
  for (const object2 of objects) {
    for (const key in object2.properties) {
      if (Object.prototype.hasOwnProperty.call(properties, key))
        continue;
      const parts = [];
      for (const other of objects) {
        const part = other.properties?.[key] ?? undeclaredConstraint(other);
        if (part === null || part === void 0)
          continue;
        if (!parts.some((seen) => JSON.stringify(seen) === JSON.stringify(part)))
          parts.push(part);
      }
      const merged = parts.length === 1 ? parts[0] : foldObjects(parts) ?? { allOf: parts };
      assignProp(properties, key, merged);
    }
    for (const key of object2.required ?? [])
      required2.add(key);
  }
  const folded = { type: "object", properties };
  if (required2.size)
    folded.required = [...required2];
  if (objects.every((object2) => object2.additionalProperties === false)) {
    folded.additionalProperties = false;
  } else {
    const constraints = [];
    for (const object2 of objects) {
      const constraint = undeclaredConstraint(object2);
      if (constraint && !constraints.some((seen) => JSON.stringify(seen) === JSON.stringify(constraint)))
        constraints.push(constraint);
    }
    if (constraints.length === 1)
      folded.additionalProperties = constraints[0];
    else if (constraints.length > 1)
      folded.additionalProperties = { allOf: constraints };
  }
  return folded;
}
function foldIntersection(json) {
  const allOf = json.allOf;
  if (!Array.isArray(allOf) || allOf.length < 2)
    return;
  for (const key of FOLDABLE_KEYS)
    if (key in json)
      return;
  const unions = allOf.filter((m) => UNION_KEYS.some((k) => Array.isArray(m[k])));
  let folded = null;
  if (!unions.length) {
    folded = foldObjects(allOf);
  } else {
    const union2 = unions[0];
    const keyword = UNION_KEYS.find((k) => Array.isArray(union2[k]));
    if (Object.keys(union2).length !== 1)
      return;
    const rest = allOf.filter((m) => m !== union2);
    const branches = union2[keyword].map((branch) => foldObjects([...rest, branch]));
    if (branches.some((b) => !b))
      return;
    folded = { [keyword]: branches };
  }
  if (!folded)
    return;
  delete json.allOf;
  assignProps(json, folded);
}
function finalize(ctx, schema) {
  const root = ctx.seen.get(schema);
  if (!root)
    throw new Error("Unprocessed schema. This is a bug in Zod.");
  const flattenRef = (zodSchema) => {
    const seen = ctx.seen.get(zodSchema);
    if (seen.ref === null)
      return;
    const schema2 = seen.def ?? seen.schema;
    const _cached = { ...schema2 };
    const ref = seen.ref;
    seen.ref = null;
    if (ref) {
      flattenRef(ref);
      const refSeen = ctx.seen.get(ref);
      const refSchema = refSeen.schema;
      if (refSchema.$ref && (ctx.target === "draft-07" || ctx.target === "draft-04" || ctx.target === "openapi-3.0")) {
        schema2.allOf = schema2.allOf ?? [];
        schema2.allOf.push(refSchema);
      } else {
        assignProps(schema2, refSchema);
      }
      assignProps(schema2, _cached);
      const isParentRef = zodSchema._zod.parent === ref;
      if (isParentRef) {
        for (const key in schema2) {
          if (key === "$ref" || key === "allOf")
            continue;
          if (!(key in _cached)) {
            delete schema2[key];
          }
        }
      }
      if (refSchema.$ref && refSeen.def) {
        for (const key in schema2) {
          if (key === "$ref" || key === "allOf")
            continue;
          if (key in refSeen.def && JSON.stringify(schema2[key]) === JSON.stringify(refSeen.def[key])) {
            delete schema2[key];
          }
        }
      }
    }
    const parent = zodSchema._zod.parent;
    if (parent && parent !== ref) {
      flattenRef(parent);
      const parentSeen = ctx.seen.get(parent);
      if (parentSeen?.schema.$ref) {
        schema2.$ref = parentSeen.schema.$ref;
        if (parentSeen.def) {
          for (const key in schema2) {
            if (key === "$ref" || key === "allOf")
              continue;
            if (key in parentSeen.def && JSON.stringify(schema2[key]) === JSON.stringify(parentSeen.def[key])) {
              delete schema2[key];
            }
          }
        }
      }
    }
    ctx.override({
      zodSchema,
      jsonSchema: schema2,
      path: seen.path ?? []
    });
  };
  if (!ctx.external || ctx.sharedEmitDoneFor !== ctx.external) {
    for (const entry of [...ctx.seen.entries()].reverse()) {
      flattenRef(entry[0]);
    }
    if (ctx.target !== "openapi-3.0") {
      for (const entry of ctx.seen.entries()) {
        compactTypeUnion(entry[1].def ?? entry[1].schema);
      }
    }
    for (const rewrite of ctx.deferred)
      rewrite();
    if (ctx.intersections.length) {
      const carriers = /* @__PURE__ */ new Map();
      for (const seen of ctx.seen.values()) {
        for (const json of [seen.schema, seen.def]) {
          const allOf = json?.allOf;
          if (!Array.isArray(allOf))
            continue;
          const existing = carriers.get(allOf);
          if (existing)
            existing.push(json);
          else
            carriers.set(allOf, [json]);
        }
      }
      for (const allOf of ctx.intersections) {
        for (const json of carriers.get(allOf) ?? [])
          foldIntersection(json);
      }
    }
  }
  const result = {};
  if (ctx.target === "draft-2020-12") {
    result.$schema = "https://json-schema.org/draft/2020-12/schema";
  } else if (ctx.target === "draft-07") {
    result.$schema = "http://json-schema.org/draft-07/schema#";
  } else if (ctx.target === "draft-04") {
    result.$schema = "http://json-schema.org/draft-04/schema#";
  } else if (ctx.target === "openapi-3.0") {
  } else {
  }
  if (ctx.external?.uri) {
    const id = ctx.external.registry.get(schema)?.id;
    if (!id)
      throw new Error("Schema is missing an `id` property");
    result.$id = ctx.external.uri(id);
  }
  assignProps(result, root.defId ? root.schema : root.def ?? root.schema);
  const rootMetaId = ctx.metadataRegistry.get(schema)?.id;
  if (rootMetaId !== void 0 && result.id === rootMetaId)
    delete result.id;
  const defs = ctx.external?.defs ?? {};
  if (!ctx.external || ctx.sharedEmitDoneFor !== ctx.external) {
    for (const entry of ctx.seen.entries()) {
      const seen = entry[1];
      if (seen.def && seen.defId) {
        if (seen.def.id === seen.defId)
          delete seen.def.id;
        assignProp(defs, seen.defId, seen.def);
      }
    }
  }
  if (ctx.external)
    ctx.sharedEmitDoneFor = ctx.external;
  if (ctx.external) {
  } else {
    if (Object.keys(defs).length > 0) {
      if (ctx.target === "draft-2020-12") {
        result.$defs = defs;
      } else {
        result.definitions = defs;
      }
    }
  }
  try {
    const finalized = JSON.parse(JSON.stringify(result));
    Object.defineProperty(finalized, "~standard", {
      value: {
        ...schema["~standard"],
        jsonSchema: {
          input: createStandardJSONSchemaMethod(schema, "input", ctx.processors),
          output: createStandardJSONSchemaMethod(schema, "output", ctx.processors)
        }
      },
      enumerable: false,
      writable: false
    });
    return finalized;
  } catch (_err) {
    throw new Error("Error converting schema to JSON.");
  }
}
function isTransforming(_schema, _ctx) {
  const ctx = _ctx ?? { seen: /* @__PURE__ */ new Set() };
  if (ctx.seen.has(_schema))
    return false;
  ctx.seen.add(_schema);
  const def2 = _schema._zod.def;
  if (def2.type === "transform")
    return true;
  if (def2.type === "array")
    return isTransforming(def2.element, ctx);
  if (def2.type === "set")
    return isTransforming(def2.valueType, ctx);
  if (def2.type === "lazy")
    return isTransforming(def2.getter(), ctx);
  if (def2.type === "promise" || def2.type === "optional" || def2.type === "nonoptional" || def2.type === "nullable" || def2.type === "readonly" || def2.type === "default" || def2.type === "prefault" || def2.type === "catch") {
    return isTransforming(def2.innerType, ctx);
  }
  if (def2.type === "intersection") {
    return isTransforming(def2.left, ctx) || isTransforming(def2.right, ctx);
  }
  if (def2.type === "record" || def2.type === "map") {
    return isTransforming(def2.keyType, ctx) || isTransforming(def2.valueType, ctx);
  }
  if (def2.type === "pipe") {
    if (_schema._zod.traits.has("$ZodCodec"))
      return true;
    return isTransforming(def2.in, ctx) || isTransforming(def2.out, ctx);
  }
  if (def2.type === "object") {
    for (const key in def2.shape) {
      if (isTransforming(def2.shape[key], ctx))
        return true;
    }
    return false;
  }
  if (def2.type === "union") {
    for (const option of def2.options) {
      if (isTransforming(option, ctx))
        return true;
    }
    return false;
  }
  if (def2.type === "tuple") {
    for (const item of def2.items) {
      if (isTransforming(item, ctx))
        return true;
    }
    if (def2.rest && isTransforming(def2.rest, ctx))
      return true;
    return false;
  }
  return false;
}
var createToJSONSchemaMethod = (schema, processors = {}) => (params) => {
  const ctx = initializeContext({ ...params, processors });
  processSchema(schema, ctx);
  extractDefs(ctx, schema);
  return finalize(ctx, schema);
};
var createStandardJSONSchemaMethod = (schema, io, processors = {}) => (params) => {
  const { libraryOptions, target } = params ?? {};
  const ctx = initializeContext({ ...libraryOptions ?? {}, target, io, processors });
  processSchema(schema, ctx);
  extractDefs(ctx, schema);
  return finalize(ctx, schema);
};

// node_modules/zod/v4/core/json-schema-processors.js
var narrowMin = (agg, key, value) => {
  if (agg[key] === void 0 || value > agg[key])
    agg[key] = value;
};
var narrowMax = (agg, key, value) => {
  if (agg[key] === void 0 || value < agg[key])
    agg[key] = value;
};
var narrowBoth = (agg, value) => {
  narrowMin(agg, "minimum", value);
  narrowMax(agg, "maximum", value);
};
var addDivisor = (agg, value) => {
  agg.multipleOf ?? (agg.multipleOf = []);
  if (!agg.multipleOf.includes(value))
    agg.multipleOf.push(value);
};
var addPattern = (agg, pattern) => {
  agg.patterns ?? (agg.patterns = /* @__PURE__ */ new Set());
  agg.patterns.add(pattern);
};
var intersectMime = (agg, mime) => {
  agg.mime = agg.mime ? agg.mime.filter((m) => mime.includes(m)) : [...mime];
};
var setFormat = (agg, format) => {
  agg.format = format;
  if (format.includes("int"))
    agg.isInt = true;
};
var minContributor = (agg, def2) => narrowMin(agg, "minimum", def2.minimum);
var maxContributor = (agg, def2) => narrowMax(agg, "maximum", def2.maximum);
var formatContributor = (ranges) => (agg, def2) => {
  setFormat(agg, def2.format);
  const [minimum, maximum] = ranges[def2.format];
  narrowMin(agg, "minimum", minimum);
  narrowMax(agg, "maximum", maximum);
};
var contributors = {
  greater_than: (agg, def2) => narrowMin(agg, def2.inclusive ? "minimum" : "exclusiveMinimum", def2.value),
  less_than: (agg, def2) => narrowMax(agg, def2.inclusive ? "maximum" : "exclusiveMaximum", def2.value),
  multiple_of: (agg, def2) => addDivisor(agg, def2.value),
  number_format: formatContributor(NUMBER_FORMAT_RANGES),
  bigint_format: formatContributor(BIGINT_FORMAT_RANGES),
  min_length: minContributor,
  max_length: maxContributor,
  length_equals: (agg, def2) => narrowBoth(agg, def2.length),
  min_size: minContributor,
  max_size: maxContributor,
  size_equals: (agg, def2) => narrowBoth(agg, def2.size),
  string_format: (agg, def2) => {
    setFormat(agg, def2.format);
    if (def2.pattern)
      addPattern(agg, def2.pattern);
    if (def2.format === "base64" || def2.format === "base64url")
      agg.contentEncoding = def2.format;
    if (def2.local || def2.precision === -1)
      agg.laxFormat = true;
  },
  mime_type: (agg, def2) => intersectMime(agg, def2.mime)
};
function aggregateChecks(schema) {
  const agg = {};
  const def2 = schema._zod.def;
  const list2 = schema._zod.traits.has("$ZodCheck") ? [schema, ...def2.checks ?? []] : def2.checks ?? [];
  for (const ch of list2)
    contributors[ch._zod.def.check]?.(agg, ch._zod.def);
  const bag = schema._zod.bag;
  if (bag.minimum !== void 0)
    narrowMin(agg, "minimum", bag.minimum);
  if (bag.exclusiveMinimum !== void 0)
    narrowMin(agg, "exclusiveMinimum", bag.exclusiveMinimum);
  if (bag.maximum !== void 0)
    narrowMax(agg, "maximum", bag.maximum);
  if (bag.exclusiveMaximum !== void 0)
    narrowMax(agg, "exclusiveMaximum", bag.exclusiveMaximum);
  if (bag.multipleOf !== void 0)
    addDivisor(agg, bag.multipleOf);
  if (bag.format !== void 0) {
    agg.format ?? (agg.format = bag.format);
    if (bag.format.includes("int"))
      agg.isInt = true;
  }
  if (bag.mime)
    intersectMime(agg, bag.mime);
  for (const pattern of bag.patterns ?? [])
    addPattern(agg, pattern);
  return agg;
}
var formatMap = {
  guid: "uuid",
  url: "uri",
  datetime: "date-time",
  json_string: "json-string",
  regex: ""
  // do not set
};
var exactPatterns = /* @__PURE__ */ new Map([
  [base64Charset, base64],
  [base64urlCharset, base64url]
]);
var exactPattern = (p) => exactPatterns.get(p) ?? p;
var stringProcessor = (schema, ctx, _json, _params) => {
  const json = _json;
  json.type = "string";
  const { minimum, maximum, format, patterns, contentEncoding, laxFormat } = aggregateChecks(schema);
  if (typeof minimum === "number")
    json.minLength = minimum;
  if (typeof maximum === "number")
    json.maxLength = maximum;
  if (format) {
    json.format = formatMap[format] ?? format;
    if (json.format === "")
      delete json.format;
    if (format === "time" || laxFormat) {
      delete json.format;
    }
  }
  if (contentEncoding)
    json.contentEncoding = contentEncoding;
  if (patterns && patterns.size > 0) {
    const patternList = [...patterns].map(exactPattern);
    if (patternList.length === 1)
      json.pattern = patternList[0].source;
    else if (patternList.length > 1) {
      json.allOf = [
        ...patternList.map((regex) => ({
          ...ctx.target === "draft-07" || ctx.target === "draft-04" || ctx.target === "openapi-3.0" ? { type: "string" } : {},
          pattern: regex.source
        }))
      ];
    }
  }
};
var numberProcessor = (schema, ctx, _json, params) => {
  const json = _json;
  const { minimum, maximum, multipleOf, exclusiveMaximum, exclusiveMinimum, isInt } = aggregateChecks(schema);
  json.type = isInt ? "integer" : "number";
  const exMin = typeof exclusiveMinimum === "number" && exclusiveMinimum >= (minimum ?? Number.NEGATIVE_INFINITY);
  const exMax = typeof exclusiveMaximum === "number" && exclusiveMaximum <= (maximum ?? Number.POSITIVE_INFINITY);
  const legacy = ctx.target === "draft-04" || ctx.target === "openapi-3.0";
  if (exMin) {
    if (legacy) {
      json.minimum = exclusiveMinimum;
      json.exclusiveMinimum = true;
    } else {
      json.exclusiveMinimum = exclusiveMinimum;
    }
  } else if (typeof minimum === "number") {
    json.minimum = minimum;
  }
  if (exMax) {
    if (legacy) {
      json.maximum = exclusiveMaximum;
      json.exclusiveMaximum = true;
    } else {
      json.exclusiveMaximum = exclusiveMaximum;
    }
  } else if (typeof maximum === "number") {
    json.maximum = maximum;
  }
  if (multipleOf) {
    const divisors = /* @__PURE__ */ new Set();
    for (const divisor of multipleOf) {
      if (Number.isFinite(divisor) && divisor !== 0)
        divisors.add(Math.abs(divisor));
      else
        handleUnrepresentable(schema, ctx, json, params, `A multipleOf divisor of ${divisor} cannot be represented in JSON Schema`);
    }
    const [first, ...rest] = divisors;
    if (first !== void 0)
      json.multipleOf = first;
    if (rest.length)
      json.allOf = [...json.allOf ?? [], ...rest.map((m) => ({ multipleOf: m }))];
  }
};
var booleanProcessor = (_schema, _ctx, json, _params) => {
  json.type = "boolean";
};
var neverProcessor = (_schema, _ctx, json, _params) => {
  json.not = {};
};
var unknownProcessor = (_schema, _ctx, _json, _params) => {
};
var enumProcessor = (schema, _ctx, json, _params) => {
  const def2 = schema._zod.def;
  const values = getEnumValues(def2.entries);
  if (values.length === 0) {
    json.not = {};
    return;
  }
  if (values.every((v) => typeof v === "number"))
    json.type = "number";
  if (values.every((v) => typeof v === "string"))
    json.type = "string";
  json.enum = values;
};
var literalProcessor = (schema, ctx, json, params) => {
  const def2 = schema._zod.def;
  if (def2.values.length === 0) {
    json.not = {};
    return;
  }
  const vals = [];
  for (const val of def2.values) {
    if (val === void 0) {
      if (handleUnrepresentable(schema, ctx, json, params, "Literal `undefined` cannot be represented in JSON Schema"))
        return;
    } else if (typeof val === "bigint") {
      if (handleUnrepresentable(schema, ctx, json, params, "BigInt literals cannot be represented in JSON Schema"))
        return;
      vals.push(Number(val));
    } else {
      vals.push(val);
    }
  }
  if (vals.length === 0) {
  } else if (vals.length === 1) {
    const val = vals[0];
    json.type = val === null ? "null" : typeof val;
    if (ctx.target === "draft-04" || ctx.target === "openapi-3.0") {
      json.enum = [val];
    } else {
      json.const = val;
    }
  } else {
    if (vals.every((v) => typeof v === "number"))
      json.type = "number";
    if (vals.every((v) => typeof v === "string"))
      json.type = "string";
    if (vals.every((v) => typeof v === "boolean"))
      json.type = "boolean";
    if (vals.every((v) => v === null))
      json.type = "null";
    json.enum = vals;
  }
};
var customProcessor = (schema, ctx, json, params) => {
  handleUnrepresentable(schema, ctx, json, params, "Custom types cannot be represented in JSON Schema");
};
var transformProcessor = (schema, ctx, json, params) => {
  handleUnrepresentable(schema, ctx, json, params, "Transforms cannot be represented in JSON Schema");
};
var arrayProcessor = (schema, ctx, _json, params) => {
  const json = _json;
  const def2 = schema._zod.def;
  const { minimum, maximum } = aggregateChecks(schema);
  if (typeof minimum === "number")
    json.minItems = minimum;
  if (typeof maximum === "number")
    json.maxItems = maximum;
  json.type = "array";
  json.items = processSchema(def2.element, ctx, {
    ...params,
    path: [...params.path, "items"]
  });
};
function inputOptin(schema) {
  const def2 = schema._zod.def;
  if (def2.type === "pipe" && def2.in._zod.traits.has("$ZodTransform")) {
    return inputOptin(def2.out);
  }
  if (def2.type === "catch") {
    return inputOptin(def2.innerType);
  }
  return schema._zod.optin;
}
var objectProcessor = (schema, ctx, _json, params) => {
  const json = _json;
  const def2 = schema._zod.def;
  const shape = def2.shape;
  const symbolKeys = Object.getOwnPropertySymbols(shape);
  if (symbolKeys.length && handleUnrepresentable(schema, ctx, json, params, "Symbol keys cannot be represented in JSON Schema")) {
    return;
  }
  json.type = "object";
  json.properties = {};
  for (const key in shape) {
    assignProp(json.properties, key, processSchema(shape[key], ctx, {
      ...params,
      path: [...params.path, "properties", key]
    }));
  }
  const requiredKeys = [];
  for (const key of Object.keys(shape)) {
    const field = def2.shape[key];
    if (ctx.io === "input" ? inputOptin(field) === void 0 : field._zod.optout === void 0) {
      requiredKeys.push(key);
    }
  }
  if (requiredKeys.length > 0) {
    json.required = requiredKeys;
  }
  if (def2.catchall?._zod.def.type === "never") {
    json.additionalProperties = false;
  } else if (!def2.catchall) {
    if (ctx.io === "output")
      json.additionalProperties = false;
  } else if (def2.catchall) {
    json.additionalProperties = processSchema(def2.catchall, ctx, {
      ...params,
      path: [...params.path, "additionalProperties"]
    });
  }
};
var unionProcessor = (schema, ctx, json, params) => {
  const def2 = schema._zod.def;
  const isExclusive = def2.inclusive === false;
  const options2 = def2.options.map((x, i) => processSchema(x, ctx, {
    ...params,
    path: [...params.path, isExclusive ? "oneOf" : "anyOf", i]
  }));
  if (isExclusive) {
    json.oneOf = options2;
  } else {
    json.anyOf = options2;
  }
};
var intersectionProcessor = (schema, ctx, json, params) => {
  const def2 = schema._zod.def;
  const a = processSchema(def2.left, ctx, {
    ...params,
    path: [...params.path, "allOf", 0]
  });
  const b = processSchema(def2.right, ctx, {
    ...params,
    path: [...params.path, "allOf", 1]
  });
  const isSimpleIntersection = (val) => "allOf" in val && Object.keys(val).length === 1;
  const allOf = [
    ...isSimpleIntersection(a) ? a.allOf : [a],
    ...isSimpleIntersection(b) ? b.allOf : [b]
  ];
  json.allOf = allOf;
  ctx.intersections.push(allOf);
};
function stringifyKeyNames(bySchema, json, visited) {
  if (json.$ref) {
    if (visited.has(json))
      return json;
    visited.add(json);
    const def2 = bySchema.get(json)?.def;
    if (!def2)
      return json;
    const inlined = stringifyKeyNames(bySchema, def2, visited);
    return inlined === def2 ? json : inlined;
  }
  for (const keyword of ["anyOf", "oneOf"]) {
    const branches = json[keyword];
    if (!Array.isArray(branches))
      continue;
    const mapped = branches.map((branch) => stringifyKeyNames(bySchema, branch, visited));
    if (mapped.some((branch, i) => branch !== branches[i]))
      json = { ...json, [keyword]: mapped };
  }
  const types = Array.isArray(json.type) ? json.type : [json.type];
  const numericType = !types.includes("string") && types.some((t) => t === "number" || t === "integer");
  const values = json.enum ?? (json.const !== void 0 ? [json.const] : void 0);
  if (!numericType && !values?.some((v) => typeof v === "number"))
    return json;
  const { minimum, maximum, exclusiveMinimum, exclusiveMaximum, multipleOf, format, id, ...rest } = json;
  if (rest.enum)
    rest.enum = rest.enum.map((v) => typeof v === "number" ? String(v) : v);
  else if (typeof rest.const === "number")
    rest.const = String(rest.const);
  if (!numericType)
    return rest;
  rest.type = "string";
  if (!values)
    rest.pattern = (types.includes("number") ? number : integer).source;
  return rest;
}
var pendingRecords = /* @__PURE__ */ new WeakMap();
function rewriteKeyNames(ctx) {
  const bySchema = /* @__PURE__ */ new Map();
  for (const entry of ctx.seen.values()) {
    if (entry.def && !bySchema.has(entry.schema))
      bySchema.set(entry.schema, entry);
  }
  const rewrites = /* @__PURE__ */ new Map();
  for (const record2 of pendingRecords.get(ctx) ?? []) {
    const seen = ctx.seen.get(record2);
    const names = (seen?.def ?? seen?.schema)?.propertyNames;
    if (!names || names === true || rewrites.has(names))
      continue;
    const rewritten = stringifyKeyNames(bySchema, names, /* @__PURE__ */ new Set());
    if (rewritten !== names)
      rewrites.set(names, rewritten);
  }
  if (!rewrites.size)
    return;
  for (const entry of ctx.seen.values()) {
    for (const carrier of [entry.schema, entry.def]) {
      const rewritten = carrier && rewrites.get(carrier.propertyNames);
      if (rewritten)
        carrier.propertyNames = rewritten;
    }
  }
}
var recordProcessor = (schema, ctx, _json, params) => {
  const json = _json;
  const def2 = schema._zod.def;
  json.type = "object";
  const keyType = def2.keyType;
  const patterns = aggregateChecks(keyType).patterns;
  if (def2.mode === "loose" && patterns && patterns.size > 0) {
    const valueSchema = processSchema(def2.valueType, ctx, {
      ...params,
      path: [...params.path, "patternProperties", "*"]
    });
    json.patternProperties = {};
    for (const pattern of patterns) {
      assignProp(json.patternProperties, exactPattern(pattern).source, valueSchema);
    }
  } else {
    if (ctx.target === "draft-07" || ctx.target === "draft-2020-12") {
      json.propertyNames = processSchema(def2.keyType, ctx, {
        ...params,
        path: [...params.path, "propertyNames"]
      });
      let pending = pendingRecords.get(ctx);
      if (!pending) {
        pending = [];
        pendingRecords.set(ctx, pending);
        ctx.deferred.push(() => rewriteKeyNames(ctx));
      }
      pending.push(schema);
    }
    json.additionalProperties = processSchema(def2.valueType, ctx, {
      ...params,
      path: [...params.path, "additionalProperties"]
    });
  }
  const keyValues = keyType._zod.values;
  const omittableOnInput = ctx.io === "input" && inputOptin(def2.valueType) !== void 0;
  if (keyValues && !def2.partial && !omittableOnInput) {
    const validKeyValues = [...keyValues].filter((v) => typeof v === "string" || typeof v === "number");
    if (validKeyValues.length > 0) {
      json.required = validKeyValues.map(String);
    }
  }
};
var nullableProcessor = (schema, ctx, json, params) => {
  const def2 = schema._zod.def;
  const inner = processSchema(def2.innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  if (ctx.target === "openapi-3.0") {
    seen.ref = def2.innerType;
    json.nullable = true;
  } else {
    json.anyOf = [inner, { type: "null" }];
  }
};
var nonoptionalProcessor = (schema, ctx, _json, params) => {
  const def2 = schema._zod.def;
  processSchema(def2.innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  seen.ref = def2.innerType;
};
var UNREPRESENTABLE_DEFAULT = Symbol();
function serializeDefaultValue(value, schema, ctx, json, params) {
  let unrepresentable = false;
  const serialized = JSON.stringify(value, (_, val) => {
    if (typeof val !== "bigint")
      return val;
    unrepresentable = true;
    return null;
  });
  if (!unrepresentable)
    return JSON.parse(serialized);
  handleUnrepresentable(schema, ctx, json, params, "BigInt defaults cannot be represented in JSON Schema");
  return UNREPRESENTABLE_DEFAULT;
}
var defaultProcessor = (schema, ctx, json, params) => {
  const def2 = schema._zod.def;
  processSchema(def2.innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  seen.ref = def2.innerType;
  const value = serializeDefaultValue(def2.defaultValue, schema, ctx, json, params);
  if (value !== UNREPRESENTABLE_DEFAULT)
    json.default = value;
};
var prefaultProcessor = (schema, ctx, json, params) => {
  const def2 = schema._zod.def;
  processSchema(def2.innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  seen.ref = def2.innerType;
  if (ctx.io !== "input")
    return;
  const value = serializeDefaultValue(def2.defaultValue, schema, ctx, json, params);
  if (value !== UNREPRESENTABLE_DEFAULT)
    json._prefault = value;
};
var catchProcessor = (schema, ctx, json, params) => {
  const def2 = schema._zod.def;
  processSchema(def2.innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  seen.ref = def2.innerType;
  let catchValue;
  try {
    catchValue = def2.catchValue(void 0);
  } catch {
    handleUnrepresentable(schema, ctx, json, params, "Dynamic catch values are not supported in JSON Schema");
    return;
  }
  json.default = catchValue;
};
var pipeProcessor = (schema, ctx, _json, params) => {
  const def2 = schema._zod.def;
  const inIsTransform = def2.in._zod.traits.has("$ZodTransform");
  const innerType = ctx.io === "input" ? inIsTransform ? def2.out : def2.in : def2.out;
  processSchema(innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  seen.ref = innerType;
};
var readonlyProcessor = (schema, ctx, json, params) => {
  const def2 = schema._zod.def;
  processSchema(def2.innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  seen.ref = def2.innerType;
  json.readOnly = true;
};
var optionalProcessor = (schema, ctx, _json, params) => {
  const def2 = schema._zod.def;
  processSchema(def2.innerType, ctx, params);
  const seen = ctx.seen.get(schema);
  seen.ref = def2.innerType;
};

// node_modules/zod/v4/classic/errors.js
var _installedErrorProtos = /* @__PURE__ */ new WeakSet([Object.prototype, Error.prototype]);
function _lazyMethod(proto, key, make) {
  Object.defineProperty(proto, key, {
    configurable: true,
    enumerable: false,
    get() {
      const value = make(this);
      Object.defineProperty(this, key, { value, configurable: true, writable: true });
      return value;
    },
    set(value) {
      Object.defineProperty(this, key, { value, configurable: true, writable: true });
    }
  });
}
var initializer2 = (inst, issues) => {
  $ZodError.init(inst, issues);
  inst.name = "ZodError";
  const proto = Object.getPrototypeOf(inst);
  if (_installedErrorProtos.has(proto))
    return;
  _installedErrorProtos.add(proto);
  _lazyMethod(proto, "format", (self) => (mapper) => formatError(self, mapper));
  _lazyMethod(proto, "flatten", (self) => (mapper) => flattenError(self, mapper));
  _lazyMethod(proto, "addIssue", (self) => (issue2) => {
    self.issues.push(issue2);
    self.message = JSON.stringify(self.issues, jsonStringifyReplacer, 2);
  });
  _lazyMethod(proto, "addIssues", (self) => (issues2) => {
    self.issues.push(...issues2);
    self.message = JSON.stringify(self.issues, jsonStringifyReplacer, 2);
  });
  Object.defineProperty(proto, "isEmpty", {
    configurable: true,
    enumerable: false,
    get() {
      return this.issues.length === 0;
    }
  });
};
var ZodRealError = /* @__PURE__ */ $constructor("ZodError", initializer2, void 0, {
  Parent: Error
});

// node_modules/zod/v4/classic/parse.js
var parse = /* @__PURE__ */ _parse(ZodRealError);
var parseAsync = /* @__PURE__ */ _parseAsync(ZodRealError);
var safeParse = /* @__PURE__ */ _safeParse(ZodRealError);
var safeParseAsync = /* @__PURE__ */ _safeParseAsync(ZodRealError);
var encode = /* @__PURE__ */ _encode(ZodRealError);
var decode = /* @__PURE__ */ _decode(ZodRealError);
var encodeAsync = /* @__PURE__ */ _encodeAsync(ZodRealError);
var decodeAsync = /* @__PURE__ */ _decodeAsync(ZodRealError);
var safeEncode = /* @__PURE__ */ _safeEncode(ZodRealError);
var safeDecode = /* @__PURE__ */ _safeDecode(ZodRealError);
var safeEncodeAsync = /* @__PURE__ */ _safeEncodeAsync(ZodRealError);
var safeDecodeAsync = /* @__PURE__ */ _safeDecodeAsync(ZodRealError);

// node_modules/zod/v4/classic/schemas.js
function _ensureDefaultLocale() {
  if (!globalConfig.localeError)
    config(en_default());
}
function _ensureDefaultMemoizer() {
  if (!globalConfig.memoizer)
    config({ memoizer: memoizer() });
}
var ZodType = /* @__PURE__ */ $constructor("ZodType", (inst, def2) => {
  _ensureDefaultLocale();
  $ZodType.init(inst, def2);
  inst.def = def2;
  inst.type = def2.type;
  return inst;
}, {
  check(...chks) {
    const def2 = this.def;
    return this.clone(util_exports.mergeDefs(def2, {
      checks: [
        ...def2.checks ?? [],
        ...chks.map((ch) => typeof ch === "function" ? { _zod: { check: ch, def: { check: "custom" }, onattach: [] } } : ch)
      ]
    }), { parent: true });
  },
  with(...chks) {
    return this.check(...chks);
  },
  clone(def2, params) {
    return clone(this, def2, params);
  },
  brand() {
    return this;
  },
  register(reg, meta2) {
    reg.add(this, meta2);
    return this;
  },
  refine(check, params) {
    return this.check(refine(check, params));
  },
  superRefine(refinement, params) {
    return this.check(superRefine(refinement, params));
  },
  overwrite(fn) {
    return this.check(_overwrite(fn));
  },
  optional() {
    return optional(this);
  },
  exactOptional() {
    return exactOptional(this);
  },
  nullable() {
    return nullable(this);
  },
  nullish() {
    return optional(nullable(this));
  },
  nonoptional(params) {
    return nonoptional(this, params);
  },
  array() {
    return array(this);
  },
  or(arg) {
    return union([this, arg]);
  },
  and(arg) {
    return intersection(this, arg);
  },
  transform(tx) {
    return pipe(this, transform(tx));
  },
  default(d) {
    return _default(this, d);
  },
  prefault(d) {
    return prefault(this, d);
  },
  catch(params) {
    return _catch(this, params);
  },
  pipe(target) {
    return pipe(this, target);
  },
  readonly() {
    return readonly(this);
  },
  describe(description) {
    const cl = this.clone();
    globalRegistry.add(cl, { description });
    return cl;
  },
  meta(...args) {
    if (args.length === 0)
      return globalRegistry.get(this);
    const cl = this.clone();
    globalRegistry.add(cl, args[0]);
    return cl;
  },
  isOptional() {
    return this.safeParse(void 0).success;
  },
  isNullable() {
    return this.safeParse(null).success;
  },
  apply(fn, ...args) {
    return args.length === 0 ? fn(this) : fn(this, ...args);
  },
  // Overrides core's `~standard` to add `jsonSchema`. Must stay a prototype entry: redefining it per instance demotes instances to dictionary mode.
  get "~standard"() {
    return util_exports.hide(this, "~standard", {
      ...standardProps(this),
      jsonSchema: {
        input: createStandardJSONSchemaMethod(this, "input"),
        output: createStandardJSONSchemaMethod(this, "output")
      }
    });
  },
  set "~standard"(value) {
    util_exports.own(this, "~standard", value);
  },
  parse: function _parse2(data, params) {
    return parse(this, data, params, { callee: _parse2 });
  },
  parseAsync: async function _parseAsync2(data, params) {
    return await parseAsync(this, data, params, { callee: _parseAsync2 });
  },
  safeParse(data, params) {
    return safeParse(this, data, params);
  },
  async safeParseAsync(data, params) {
    return safeParseAsync(this, data, params);
  },
  // `spa` is an alias: same function object as `safeParseAsync`, as before.
  get spa() {
    return this?.safeParseAsync;
  },
  set spa(value) {
    util_exports.own(this, "spa", value);
  },
  validate(data, params) {
    return validate(this, data, params);
  },
  validateAsync(data, params) {
    return validateAsync(this, data, params);
  },
  encode: function _encode2(data, params) {
    return encode(this, data, params, { callee: _encode2 });
  },
  decode: function _decode2(data, params) {
    return decode(this, data, params, { callee: _decode2 });
  },
  encodeAsync: async function _encodeAsync2(data, params) {
    return await encodeAsync(this, data, params, { callee: _encodeAsync2 });
  },
  decodeAsync: async function _decodeAsync2(data, params) {
    return await decodeAsync(this, data, params, { callee: _decodeAsync2 });
  },
  safeEncode(data, params) {
    return safeEncode(this, data, params);
  },
  safeDecode(data, params) {
    return safeDecode(this, data, params);
  },
  async safeEncodeAsync(data, params) {
    return safeEncodeAsync(this, data, params);
  },
  async safeDecodeAsync(data, params) {
    return safeDecodeAsync(this, data, params);
  },
  toJSONSchema(params) {
    return createToJSONSchemaMethod(this, {})(params);
  },
  // Reads through to the registry on every access, so it must not cache.
  get description() {
    return globalRegistry.get(this)?.description;
  },
  // No setter: `schema._def = x` throws, as it did when `_def` was a non-writable own property.
  get _def() {
    return this._zod.def;
  }
});
var _ZodString = /* @__PURE__ */ $constructor(
  "_ZodString",
  (inst, def2) => {
    $ZodString.init(inst, def2);
    ZodType.init(inst, def2);
    inst._zod.processJSONSchema = (ctx, json, params) => stringProcessor(inst, ctx, json, params);
  },
  /* @__PURE__ */ util_exports.derived({
    format: (inst) => aggregateChecks(inst).format ?? null,
    minLength: (inst) => aggregateChecks(inst).minimum ?? null,
    maxLength: (inst) => aggregateChecks(inst).maximum ?? null
  }, {
    regex(...args) {
      return this.check(_regex(...args));
    },
    includes(...args) {
      return this.check(_includes(...args));
    },
    startsWith(...args) {
      return this.check(_startsWith(...args));
    },
    endsWith(...args) {
      return this.check(_endsWith(...args));
    },
    min(...args) {
      return this.check(_minLength(...args));
    },
    max(...args) {
      return this.check(_maxLength(...args));
    },
    length(...args) {
      return this.check(_length(...args));
    },
    nonempty(...args) {
      return this.check(_minLength(1, ...args));
    },
    lowercase(params) {
      return this.check(_lowercase(params));
    },
    uppercase(params) {
      return this.check(_uppercase(params));
    },
    trim() {
      return this.check(_trim());
    },
    normalize(...args) {
      return this.check(_normalize(...args));
    },
    toLowerCase() {
      return this.check(_toLowerCase());
    },
    toUpperCase() {
      return this.check(_toUpperCase());
    },
    slugify() {
      return this.check(_slugify());
    }
  })
);
var ZodString = /* @__PURE__ */ $constructor("ZodString", (inst, def2) => {
  $ZodString.init(inst, def2);
  _ZodString.init(inst, def2);
}, {
  email(params) {
    return this.check(_email(ZodEmail, params));
  },
  url(params) {
    return this.check(_url(ZodURL, params));
  },
  jwt(params) {
    return this.check(_jwt(ZodJWT, params));
  },
  emoji(params) {
    return this.check(_emoji2(ZodEmoji, params));
  },
  guid(params) {
    return this.check(_guid(ZodGUID, params));
  },
  uuid(params) {
    return this.check(_uuid(ZodUUID, params));
  },
  uuidv4(params) {
    return this.check(_uuidv4(ZodUUID, params));
  },
  uuidv6(params) {
    return this.check(_uuidv6(ZodUUID, params));
  },
  uuidv7(params) {
    return this.check(_uuidv7(ZodUUID, params));
  },
  nanoid(params) {
    return this.check(_nanoid(ZodNanoID, params));
  },
  cuid(params) {
    return this.check(_cuid(ZodCUID, params));
  },
  cuid2(params) {
    return this.check(_cuid2(ZodCUID2, params));
  },
  ulid(params) {
    return this.check(_ulid(ZodULID, params));
  },
  base64(params) {
    return this.check(_base64(ZodBase64, params));
  },
  base64url(params) {
    return this.check(_base64url(ZodBase64URL, params));
  },
  xid(params) {
    return this.check(_xid(ZodXID, params));
  },
  ksuid(params) {
    return this.check(_ksuid(ZodKSUID, params));
  },
  ipv4(params) {
    return this.check(_ipv4(ZodIPv4, params));
  },
  ipv6(params) {
    return this.check(_ipv6(ZodIPv6, params));
  },
  cidrv4(params) {
    return this.check(_cidrv4(ZodCIDRv4, params));
  },
  cidrv6(params) {
    return this.check(_cidrv6(ZodCIDRv6, params));
  },
  e164(params) {
    return this.check(_e164(ZodE164, params));
  },
  datetime(params) {
    return this.check(_isoDateTime(ZodISODateTime, params));
  },
  date(params) {
    return this.check(_isoDate(ZodISODate, params));
  },
  time(params) {
    return this.check(_isoTime(ZodISOTime, params));
  },
  duration(params) {
    return this.check(_isoDuration(ZodISODuration, params));
  }
});
function string2(params) {
  return _string(ZodString, params);
}
var ZodStringFormat = /* @__PURE__ */ $constructor("ZodStringFormat", (inst, def2) => {
  $ZodStringFormat.init(inst, def2);
  _ZodString.init(inst, def2);
});
var ZodISODateTime = /* @__PURE__ */ $constructor("ZodISODateTime", (inst, def2) => {
  $ZodISODateTime.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodISODate = /* @__PURE__ */ $constructor("ZodISODate", (inst, def2) => {
  $ZodISODate.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodISOTime = /* @__PURE__ */ $constructor("ZodISOTime", (inst, def2) => {
  $ZodISOTime.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodISODuration = /* @__PURE__ */ $constructor("ZodISODuration", (inst, def2) => {
  $ZodISODuration.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodEmail = /* @__PURE__ */ $constructor("ZodEmail", (inst, def2) => {
  $ZodEmail.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodGUID = /* @__PURE__ */ $constructor("ZodGUID", (inst, def2) => {
  $ZodGUID.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodUUID = /* @__PURE__ */ $constructor("ZodUUID", (inst, def2) => {
  $ZodUUID.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodURL = /* @__PURE__ */ $constructor("ZodURL", (inst, def2) => {
  $ZodURL.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
function url(params) {
  return _url(ZodURL, params);
}
var ZodEmoji = /* @__PURE__ */ $constructor("ZodEmoji", (inst, def2) => {
  $ZodEmoji.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodNanoID = /* @__PURE__ */ $constructor("ZodNanoID", (inst, def2) => {
  $ZodNanoID.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodCUID = /* @__PURE__ */ $constructor("ZodCUID", (inst, def2) => {
  $ZodCUID.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodCUID2 = /* @__PURE__ */ $constructor("ZodCUID2", (inst, def2) => {
  $ZodCUID2.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodULID = /* @__PURE__ */ $constructor("ZodULID", (inst, def2) => {
  $ZodULID.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodXID = /* @__PURE__ */ $constructor("ZodXID", (inst, def2) => {
  $ZodXID.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodKSUID = /* @__PURE__ */ $constructor("ZodKSUID", (inst, def2) => {
  $ZodKSUID.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodIPv4 = /* @__PURE__ */ $constructor("ZodIPv4", (inst, def2) => {
  $ZodIPv4.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodIPv6 = /* @__PURE__ */ $constructor("ZodIPv6", (inst, def2) => {
  $ZodIPv6.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodCIDRv4 = /* @__PURE__ */ $constructor("ZodCIDRv4", (inst, def2) => {
  $ZodCIDRv4.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodCIDRv6 = /* @__PURE__ */ $constructor("ZodCIDRv6", (inst, def2) => {
  $ZodCIDRv6.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodBase64 = /* @__PURE__ */ $constructor("ZodBase64", (inst, def2) => {
  $ZodBase64.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodBase64URL = /* @__PURE__ */ $constructor("ZodBase64URL", (inst, def2) => {
  $ZodBase64URL.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodE164 = /* @__PURE__ */ $constructor("ZodE164", (inst, def2) => {
  $ZodE164.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodJWT = /* @__PURE__ */ $constructor("ZodJWT", (inst, def2) => {
  $ZodJWT.init(inst, def2);
  ZodStringFormat.init(inst, def2);
});
var ZodNumber = /* @__PURE__ */ $constructor(
  "ZodNumber",
  (inst, def2) => {
    $ZodNumber.init(inst, def2);
    ZodType.init(inst, def2);
    inst._zod.processJSONSchema = (ctx, json, params) => numberProcessor(inst, ctx, json, params);
    inst.isFinite = true;
  },
  /* @__PURE__ */ util_exports.derived({
    minValue: (inst) => {
      const { minimum, exclusiveMinimum } = aggregateChecks(inst);
      return Math.max(minimum ?? Number.NEGATIVE_INFINITY, exclusiveMinimum ?? Number.NEGATIVE_INFINITY);
    },
    maxValue: (inst) => {
      const { maximum, exclusiveMaximum } = aggregateChecks(inst);
      return Math.min(maximum ?? Number.POSITIVE_INFINITY, exclusiveMaximum ?? Number.POSITIVE_INFINITY);
    },
    isInt: (inst) => {
      const { isInt, multipleOf } = aggregateChecks(inst);
      return !!isInt || !!multipleOf?.some(Number.isSafeInteger);
    },
    format: (inst) => aggregateChecks(inst).format ?? null
  }, {
    gt(value, params) {
      return this.check(_gt(value, params));
    },
    gte(value, params) {
      return this.check(_gte(value, params));
    },
    min(value, params) {
      return this.check(_gte(value, params));
    },
    lt(value, params) {
      return this.check(_lt(value, params));
    },
    lte(value, params) {
      return this.check(_lte(value, params));
    },
    max(value, params) {
      return this.check(_lte(value, params));
    },
    int(params) {
      return this.check(int(params));
    },
    safe(params) {
      return this.check(int(params));
    },
    positive(params) {
      return this.check(_gt(0, params));
    },
    nonnegative(params) {
      return this.check(_gte(0, params));
    },
    negative(params) {
      return this.check(_lt(0, params));
    },
    nonpositive(params) {
      return this.check(_lte(0, params));
    },
    multipleOf(value, params) {
      return this.check(_multipleOf(value, params));
    },
    step(value, params) {
      return this.check(_multipleOf(value, params));
    },
    finite() {
      return this;
    }
  })
);
function number2(params) {
  return _number(ZodNumber, params);
}
var ZodNumberFormat = /* @__PURE__ */ $constructor("ZodNumberFormat", (inst, def2) => {
  $ZodNumberFormat.init(inst, def2);
  ZodNumber.init(inst, def2);
});
function int(params) {
  return _int(ZodNumberFormat, params);
}
var ZodBoolean = /* @__PURE__ */ $constructor("ZodBoolean", (inst, def2) => {
  $ZodBoolean.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => booleanProcessor(inst, ctx, json, params);
});
function boolean2(params) {
  return _boolean(ZodBoolean, params);
}
var ZodUnknown = /* @__PURE__ */ $constructor("ZodUnknown", (inst, def2) => {
  $ZodUnknown.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => unknownProcessor(inst, ctx, json, params);
});
function unknown() {
  return _unknown(ZodUnknown);
}
var ZodNever = /* @__PURE__ */ $constructor("ZodNever", (inst, def2) => {
  $ZodNever.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => neverProcessor(inst, ctx, json, params);
});
function never(params) {
  return _never(ZodNever, params);
}
var ZodArray = /* @__PURE__ */ $constructor("ZodArray", (inst, def2) => {
  _ensureDefaultMemoizer();
  $ZodArray.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => arrayProcessor(inst, ctx, json, params);
  inst.element = def2.element;
}, {
  min(n, params) {
    return this.check(_minLength(n, params));
  },
  nonempty(params) {
    return this.check(_minLength(1, params));
  },
  max(n, params) {
    return this.check(_maxLength(n, params));
  },
  length(n, params) {
    return this.check(_length(n, params));
  },
  unwrap() {
    return this.element;
  }
});
function array(element, params) {
  return _array(ZodArray, element, params);
}
var ZodObject = /* @__PURE__ */ $constructor("ZodObject", (inst, def2) => {
  _ensureDefaultMemoizer();
  $ZodObjectJIT.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => objectProcessor(inst, ctx, json, params);
  util_exports.installLazyProp(inst, "shape", (self) => self._zod.def.shape, false);
}, {
  keyof() {
    return _enum(Object.keys(this._zod.def.shape));
  },
  catchall(catchall) {
    return this.clone(util_exports.mergeDefs(this._zod.def, { catchall }));
  },
  passthrough() {
    return this.clone(util_exports.mergeDefs(this._zod.def, { catchall: unknown() }));
  },
  loose() {
    return this.clone(util_exports.mergeDefs(this._zod.def, { catchall: unknown() }));
  },
  strict() {
    return this.clone(util_exports.mergeDefs(this._zod.def, { catchall: never() }));
  },
  strip() {
    return this.clone(util_exports.mergeDefs(this._zod.def, { catchall: void 0 }));
  },
  extend(incoming) {
    return util_exports.extend(this, incoming);
  },
  safeExtend(incoming) {
    return util_exports.safeExtend(this, incoming);
  },
  merge(other) {
    return util_exports.merge(this, other);
  },
  pick(mask) {
    return util_exports.pick(this, mask);
  },
  omit(mask) {
    return util_exports.omit(this, mask);
  },
  partial(...args) {
    return util_exports.partial(ZodOptional, this, args[0]);
  },
  exactPartial(...args) {
    return util_exports.partial(ZodExactOptional, this, args[0], "exactPartial");
  },
  required(...args) {
    return util_exports.required(ZodNonOptional, this, args[0]);
  }
});
function object(shape, params) {
  const def2 = {
    type: "object",
    shape: shape ?? {},
    ...util_exports.normalizeParams(params)
  };
  return new ZodObject(def2);
}
var ZodUnion = /* @__PURE__ */ $constructor("ZodUnion", (inst, def2) => {
  $ZodUnion.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => unionProcessor(inst, ctx, json, params);
  inst.options = def2.options;
});
function union(options2, params) {
  return new ZodUnion({
    type: "union",
    options: options2,
    ...util_exports.normalizeParams(params)
  });
}
var ZodIntersection = /* @__PURE__ */ $constructor("ZodIntersection", (inst, def2) => {
  $ZodIntersection.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => intersectionProcessor(inst, ctx, json, params);
});
function intersection(left, right) {
  return new ZodIntersection({
    type: "intersection",
    left,
    right
  });
}
var ZodRecord = /* @__PURE__ */ $constructor("ZodRecord", (inst, def2) => {
  _ensureDefaultMemoizer();
  $ZodRecord.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => recordProcessor(inst, ctx, json, params);
  inst.keyType = def2.keyType;
  inst.valueType = def2.valueType;
});
function record(keyType, valueType, params) {
  if (!valueType || !valueType._zod) {
    return new ZodRecord({
      type: "record",
      keyType: string2(),
      valueType: keyType,
      ...util_exports.normalizeParams(valueType)
    });
  }
  return new ZodRecord({
    type: "record",
    keyType,
    valueType,
    ...util_exports.normalizeParams(params)
  });
}
var ZodEnum = /* @__PURE__ */ $constructor("ZodEnum", (inst, def2) => {
  $ZodEnum.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => enumProcessor(inst, ctx, json, params);
  inst.enum = def2.entries;
  inst.options = [...inst._zod.values];
  const keys = new Set(Object.keys(def2.entries));
  inst.extract = (values, params) => {
    const newEntries = {};
    for (const value of values) {
      if (keys.has(value)) {
        newEntries[value] = def2.entries[value];
      } else
        throw new Error(`Key ${value} not found in enum`);
    }
    return new ZodEnum({
      ...def2,
      checks: [],
      ...util_exports.normalizeParams(params),
      entries: newEntries
    });
  };
  inst.exclude = (values, params) => {
    const newEntries = { ...def2.entries };
    for (const value of values) {
      if (keys.has(value)) {
        delete newEntries[value];
      } else
        throw new Error(`Key ${value} not found in enum`);
    }
    return new ZodEnum({
      ...def2,
      checks: [],
      ...util_exports.normalizeParams(params),
      entries: newEntries
    });
  };
});
function _enum(values, params) {
  const entries = Array.isArray(values) ? Object.fromEntries(values.map((v) => [v, v])) : values;
  return new ZodEnum({
    type: "enum",
    entries,
    ...util_exports.normalizeParams(params)
  });
}
var ZodLiteral = /* @__PURE__ */ $constructor("ZodLiteral", (inst, def2) => {
  $ZodLiteral.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => literalProcessor(inst, ctx, json, params);
  inst.values = new Set(def2.values);
  Object.defineProperty(inst, "value", {
    get() {
      if (def2.values.length > 1) {
        throw new Error("This schema contains multiple valid literal values. Use `.values` instead.");
      }
      return def2.values[0];
    }
  });
});
function literal(value, params) {
  return new ZodLiteral({
    type: "literal",
    values: Array.isArray(value) ? value : [value],
    ...util_exports.normalizeParams(params)
  });
}
var ZodTransform = /* @__PURE__ */ $constructor("ZodTransform", (inst, def2) => {
  _ensureDefaultMemoizer();
  $ZodTransform.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => transformProcessor(inst, ctx, json, params);
  inst._zod.parse = (payload, _ctx) => {
    if (_ctx.direction === "backward") {
      throw new $ZodEncodeError(inst.constructor.name);
    }
    payload.addIssue = (issue2) => {
      if (typeof issue2 === "string") {
        payload.issues.push(util_exports.issue(issue2, payload.value, def2));
      } else {
        const _issue = issue2;
        if (_issue.fatal)
          _issue.continue = false;
        _issue.code ?? (_issue.code = "custom");
        if (!("input" in _issue))
          _issue.input = payload.value;
        _issue.inst ?? (_issue.inst = inst);
        payload.issues.push(util_exports.issue(_issue));
      }
    };
    const output = def2.transform(payload.value, payload);
    if (output instanceof Promise) {
      return output.then((output2) => {
        payload.value = output2;
        return payload;
      });
    }
    payload.value = output;
    return payload;
  };
});
function transform(fn) {
  return new ZodTransform({
    type: "transform",
    transform: fn
  });
}
var ZodOptional = /* @__PURE__ */ $constructor("ZodOptional", (inst, def2) => {
  $ZodOptional.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => optionalProcessor(inst, ctx, json, params);
  inst.unwrap = () => inst._zod.def.innerType;
});
function optional(innerType) {
  return new ZodOptional({
    type: "optional",
    innerType
  });
}
var ZodExactOptional = /* @__PURE__ */ $constructor("ZodExactOptional", (inst, def2) => {
  $ZodExactOptional.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => optionalProcessor(inst, ctx, json, params);
  inst.unwrap = () => inst._zod.def.innerType;
});
function exactOptional(innerType) {
  return new ZodExactOptional({
    type: "optional",
    innerType
  });
}
var ZodNullable = /* @__PURE__ */ $constructor("ZodNullable", (inst, def2) => {
  $ZodNullable.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => nullableProcessor(inst, ctx, json, params);
  inst.unwrap = () => inst._zod.def.innerType;
});
function nullable(innerType) {
  return new ZodNullable({
    type: "nullable",
    innerType
  });
}
var ZodDefault = /* @__PURE__ */ $constructor("ZodDefault", (inst, def2) => {
  $ZodDefault.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => defaultProcessor(inst, ctx, json, params);
  inst.unwrap = () => inst._zod.def.innerType;
  inst.removeDefault = inst.unwrap;
});
function _default(innerType, defaultValue) {
  return new ZodDefault({
    type: "default",
    innerType,
    get defaultValue() {
      return typeof defaultValue === "function" ? defaultValue() : util_exports.shallowClone(defaultValue);
    }
  });
}
var ZodPrefault = /* @__PURE__ */ $constructor("ZodPrefault", (inst, def2) => {
  $ZodPrefault.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => prefaultProcessor(inst, ctx, json, params);
  inst.unwrap = () => inst._zod.def.innerType;
});
function prefault(innerType, defaultValue) {
  return new ZodPrefault({
    type: "prefault",
    innerType,
    get defaultValue() {
      return typeof defaultValue === "function" ? defaultValue() : util_exports.shallowClone(defaultValue);
    }
  });
}
var ZodNonOptional = /* @__PURE__ */ $constructor("ZodNonOptional", (inst, def2) => {
  $ZodNonOptional.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => nonoptionalProcessor(inst, ctx, json, params);
  inst.unwrap = () => inst._zod.def.innerType;
});
function nonoptional(innerType, params) {
  return new ZodNonOptional({
    type: "nonoptional",
    innerType,
    ...util_exports.normalizeParams(params)
  });
}
var ZodCatch = /* @__PURE__ */ $constructor("ZodCatch", (inst, def2) => {
  $ZodCatch.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => catchProcessor(inst, ctx, json, params);
  inst.unwrap = () => inst._zod.def.innerType;
  inst.removeCatch = inst.unwrap;
});
function _catch(innerType, catchValue) {
  return new ZodCatch({
    type: "catch",
    innerType,
    catchValue: typeof catchValue === "function" ? catchValue : util_exports.constantCatch(catchValue)
  });
}
var ZodPipe = /* @__PURE__ */ $constructor("ZodPipe", (inst, def2) => {
  $ZodPipe.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => pipeProcessor(inst, ctx, json, params);
  inst.in = def2.in;
  inst.out = def2.out;
});
function pipe(in_, out) {
  return new ZodPipe({
    type: "pipe",
    in: in_,
    out
    // ...util.normalizeParams(params),
  });
}
var ZodPreprocess = /* @__PURE__ */ $constructor("ZodPreprocess", (inst, def2) => {
  ZodPipe.init(inst, def2);
  $ZodPreprocess.init(inst, def2);
});
var ZodReadonly = /* @__PURE__ */ $constructor("ZodReadonly", (inst, def2) => {
  $ZodReadonly.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => readonlyProcessor(inst, ctx, json, params);
  inst.unwrap = () => inst._zod.def.innerType;
});
function readonly(innerType) {
  return new ZodReadonly({
    type: "readonly",
    innerType
  });
}
var ZodCustom = /* @__PURE__ */ $constructor("ZodCustom", (inst, def2) => {
  $ZodCustom.init(inst, def2);
  ZodType.init(inst, def2);
  inst._zod.processJSONSchema = (ctx, json, params) => customProcessor(inst, ctx, json, params);
});
function refine(fn, _params = {}) {
  return _refine(ZodCustom, fn, _params);
}
function superRefine(fn, params) {
  return _superRefine(fn, params);
}
function preprocess(fn, schema) {
  return new ZodPreprocess({
    type: "pipe",
    in: transform(fn),
    out: schema
  });
}

// node_modules/@agentclientprotocol/sdk/dist/core/schema-deserialize.js
var skippedItem = Symbol("skippedItem");
function defaultOnError(schema, fallback) {
  return schema.catch(fallback);
}
function requiredDefaultOnError(schema, fallback) {
  const schemaWithCatch = schema.catch(fallback);
  return unknown().transform((value, context) => {
    if (value !== void 0)
      return schemaWithCatch.parse(value);
    context.addIssue({
      code: "custom",
      message: "Required value is missing"
    });
    return NEVER;
  });
}
function stringTag(value, key) {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    return void 0;
  }
  const tag2 = value[key];
  return typeof tag2 === "string" ? tag2 : void 0;
}
function excludeKnownTags(schema, key, knownTags) {
  return schema.superRefine((value, context) => {
    const tag2 = stringTag(value, key);
    if (tag2 !== void 0 && knownTags.includes(tag2)) {
      context.addIssue({
        code: "custom",
        path: [key],
        message: `${key} ${JSON.stringify(tag2)} is reserved by a known variant, but the value does not match that variant's schema`
      });
    }
  });
}
function preserveCustomPayload(schema, key, knownTags) {
  return unknown().transform((value, context) => {
    const result = schema.safeParse(value);
    if (!result.success) {
      for (const issue2 of result.error.issues) {
        context.addIssue({ ...issue2, input: value });
      }
      return NEVER;
    }
    const output = result.data;
    const tag2 = stringTag(value, key);
    if (tag2 !== void 0 && !knownTags.includes(tag2)) {
      const raw = value;
      for (const [property, rawValue] of Object.entries(raw)) {
        if (property === "__proto__")
          continue;
        if (!Object.hasOwn(output, property))
          output[property] = rawValue;
      }
    }
    return output;
  });
}
function vecSkipError(itemSchema) {
  return array(itemSchema.catch(skippedItem)).transform((items) => items.filter((item) => item !== skippedItem));
}
function defaultOnNull(schema, empty) {
  return preprocess((value) => value === null ? empty() : value, schema);
}
function wideInt() {
  return number2().refine(Number.isInteger, {
    error: "Invalid input: expected int, received number"
  });
}

// node_modules/@agentclientprotocol/sdk/dist/v1/schema/zod.gen.js
var zRequestId = union([wideInt(), string2()]).nullable();
var zSessionId = string2();
var zWriteTextFileRequest = object({
  sessionId: zSessionId,
  path: string2(),
  content: string2(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zReadTextFileRequest = object({
  sessionId: zSessionId,
  path: string2(),
  line: int().gte(0).max(4294967295, {
    error: "Invalid value: Expected uint32 to be <= 4294967295"
  }).nullish(),
  limit: int().gte(0).max(4294967295, {
    error: "Invalid value: Expected uint32 to be <= 4294967295"
  }).nullish(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zToolCallId = string2();
var zToolKind = union([
  literal("read"),
  literal("edit"),
  literal("delete"),
  literal("move"),
  literal("search"),
  literal("execute"),
  literal("think"),
  literal("fetch"),
  literal("switch_mode"),
  literal("other")
]);
var zToolCallStatus = union([
  literal("pending"),
  literal("in_progress"),
  literal("completed"),
  literal("failed")
]);
var zRole = union([literal("assistant"), literal("user")]);
var zAnnotations = object({
  audience: defaultOnError(vecSkipError(zRole).nullish(), () => void 0),
  lastModified: defaultOnError(string2().nullish(), () => void 0),
  priority: defaultOnError(number2().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zTextContent = object({
  annotations: defaultOnError(zAnnotations.nullish(), () => void 0),
  text: string2(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zImageContent = object({
  annotations: defaultOnError(zAnnotations.nullish(), () => void 0),
  data: string2(),
  mimeType: string2(),
  uri: defaultOnError(string2().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zAudioContent = object({
  annotations: defaultOnError(zAnnotations.nullish(), () => void 0),
  data: string2(),
  mimeType: string2(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zResourceLink = object({
  annotations: defaultOnError(zAnnotations.nullish(), () => void 0),
  description: defaultOnError(string2().nullish(), () => void 0),
  mimeType: defaultOnError(string2().nullish(), () => void 0),
  name: string2(),
  size: defaultOnError(wideInt().nullish(), () => void 0),
  title: defaultOnError(string2().nullish(), () => void 0),
  uri: string2(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zTextResourceContents = object({
  mimeType: defaultOnError(string2().nullish(), () => void 0),
  text: string2(),
  uri: string2(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zBlobResourceContents = object({
  blob: string2(),
  mimeType: defaultOnError(string2().nullish(), () => void 0),
  uri: string2(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zEmbeddedResourceResource = union([
  zTextResourceContents,
  zBlobResourceContents
]);
var zEmbeddedResource = object({
  annotations: defaultOnError(zAnnotations.nullish(), () => void 0),
  resource: zEmbeddedResourceResource,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zContentBlock = union([
  zTextContent.and(object({
    type: literal("text")
  })),
  zImageContent.and(object({
    type: literal("image")
  })),
  zAudioContent.and(object({
    type: literal("audio")
  })),
  zResourceLink.and(object({
    type: literal("resource_link")
  })),
  zEmbeddedResource.and(object({
    type: literal("resource")
  }))
]);
var zContent = object({
  content: zContentBlock,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zDiff = object({
  path: string2(),
  oldText: defaultOnError(string2().nullish(), () => void 0),
  newText: string2(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zTerminalId = string2();
var zTerminal = object({
  terminalId: zTerminalId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zToolCallContent = union([
  zContent.and(object({
    type: literal("content")
  })),
  zDiff.and(object({
    type: literal("diff")
  })),
  zTerminal.and(object({
    type: literal("terminal")
  }))
]);
var zToolCallLocation = object({
  path: string2(),
  line: defaultOnError(int().gte(0).max(4294967295, {
    error: "Invalid value: Expected uint32 to be <= 4294967295"
  }).nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zToolCallUpdate = object({
  toolCallId: zToolCallId,
  kind: defaultOnError(zToolKind.nullish(), () => void 0),
  status: defaultOnError(zToolCallStatus.nullish(), () => void 0),
  title: defaultOnError(string2().nullish(), () => void 0),
  name: defaultOnError(string2().nullish(), () => void 0),
  content: defaultOnError(vecSkipError(zToolCallContent).nullish(), () => void 0),
  locations: defaultOnError(vecSkipError(zToolCallLocation).nullish(), () => void 0),
  rawInput: defaultOnError(unknown().optional(), () => void 0),
  rawOutput: defaultOnError(unknown().optional(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zPermissionOptionId = string2();
var zPermissionOptionKind = union([
  literal("allow_once"),
  literal("allow_always"),
  literal("reject_once"),
  literal("reject_always")
]);
var zPermissionOption = object({
  optionId: zPermissionOptionId,
  name: string2(),
  kind: zPermissionOptionKind,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zRequestPermissionRequest = object({
  sessionId: zSessionId,
  toolCall: zToolCallUpdate,
  options: array(zPermissionOption),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zEnvVariable = object({
  name: string2(),
  value: string2(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zCreateTerminalRequest = object({
  sessionId: zSessionId,
  command: string2(),
  args: defaultOnNull(array(string2()), () => []).optional(),
  env: defaultOnNull(array(zEnvVariable), () => []).optional(),
  cwd: string2().nullish(),
  outputByteLimit: wideInt().gte(0).nullish(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zTerminalOutputRequest = object({
  sessionId: zSessionId,
  terminalId: zTerminalId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zReleaseTerminalRequest = object({
  sessionId: zSessionId,
  terminalId: zTerminalId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zWaitForTerminalExitRequest = object({
  sessionId: zSessionId,
  terminalId: zTerminalId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zKillTerminalRequest = object({
  sessionId: zSessionId,
  terminalId: zTerminalId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zElicitationSessionScope = object({
  sessionId: zSessionId,
  toolCallId: defaultOnError(zToolCallId.nullish(), () => void 0)
});
var zElicitationRequestScope = object({
  requestId: zRequestId
});
var zElicitationSchemaType = literal("object");
var zStringFormat = union([
  literal("email"),
  literal("uri"),
  literal("date"),
  literal("date-time")
]);
var zEnumOption = object({
  const: string2(),
  title: string2(),
  description: defaultOnError(string2().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zStringPropertySchema = object({
  title: defaultOnError(string2().nullish(), () => void 0),
  description: defaultOnError(string2().nullish(), () => void 0),
  minLength: int().gte(0).max(4294967295, {
    error: "Invalid value: Expected uint32 to be <= 4294967295"
  }).nullish(),
  maxLength: int().gte(0).max(4294967295, {
    error: "Invalid value: Expected uint32 to be <= 4294967295"
  }).nullish(),
  pattern: string2().nullish(),
  format: zStringFormat.nullish(),
  default: defaultOnError(string2().nullish(), () => void 0),
  enum: array(string2()).nullish(),
  oneOf: array(zEnumOption).nullish(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zNumberPropertySchema = object({
  title: defaultOnError(string2().nullish(), () => void 0),
  description: defaultOnError(string2().nullish(), () => void 0),
  minimum: number2().nullish(),
  maximum: number2().nullish(),
  default: defaultOnError(number2().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zIntegerPropertySchema = object({
  title: defaultOnError(string2().nullish(), () => void 0),
  description: defaultOnError(string2().nullish(), () => void 0),
  minimum: wideInt().nullish(),
  maximum: wideInt().nullish(),
  default: defaultOnError(wideInt().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zBooleanPropertySchema = object({
  title: defaultOnError(string2().nullish(), () => void 0),
  description: defaultOnError(string2().nullish(), () => void 0),
  default: defaultOnError(boolean2().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zStringMultiSelectItems = object({
  enum: array(string2()),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zTitledMultiSelectItems = object({
  anyOf: array(zEnumOption),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zMultiSelectItems = preserveCustomPayload(union([
  zStringMultiSelectItems.and(object({
    type: literal("string")
  })),
  excludeKnownTags(object({
    type: string2()
  }), "type", ["string"]),
  zTitledMultiSelectItems
]), "type", ["string"]);
var zMultiSelectPropertySchema = object({
  title: defaultOnError(string2().nullish(), () => void 0),
  description: defaultOnError(string2().nullish(), () => void 0),
  minItems: wideInt().gte(0).nullish(),
  maxItems: wideInt().gte(0).nullish(),
  items: zMultiSelectItems,
  default: defaultOnError(vecSkipError(string2()).nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zElicitationPropertySchema = preserveCustomPayload(union([
  zStringPropertySchema.and(object({
    type: literal("string")
  })),
  zNumberPropertySchema.and(object({
    type: literal("number")
  })),
  zIntegerPropertySchema.and(object({
    type: literal("integer")
  })),
  zBooleanPropertySchema.and(object({
    type: literal("boolean")
  })),
  zMultiSelectPropertySchema.and(object({
    type: literal("array")
  })),
  excludeKnownTags(object({
    type: string2()
  }), "type", ["array", "boolean", "integer", "number", "string"])
]), "type", ["array", "boolean", "integer", "number", "string"]);
var zElicitationSchema = object({
  type: defaultOnError(zElicitationSchemaType.optional().default("object"), () => "object"),
  title: defaultOnError(string2().nullish(), () => void 0),
  properties: record(string2(), zElicitationPropertySchema).optional().default({}),
  required: array(string2()).nullish(),
  description: defaultOnError(string2().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zElicitationFormMode = intersection(union([zElicitationSessionScope, zElicitationRequestScope]), object({
  requestedSchema: zElicitationSchema
}));
var zElicitationId = string2();
var zElicitationUrlMode = intersection(union([zElicitationSessionScope, zElicitationRequestScope]), object({
  elicitationId: zElicitationId,
  url: url()
}));
var zCreateElicitationRequest = preserveCustomPayload(intersection(union([
  zElicitationFormMode.and(object({
    mode: literal("form")
  })),
  zElicitationUrlMode.and(object({
    mode: literal("url")
  })),
  excludeKnownTags(intersection(union([zElicitationSessionScope, zElicitationRequestScope]), object({
    mode: string2()
  })), "mode", ["form", "url"])
]), object({
  message: string2(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
})), "mode", ["form", "url"]);
var zExtRequest = unknown();
var zAgentRequest = object({
  id: zRequestId,
  method: string2(),
  params: union([
    zWriteTextFileRequest,
    zReadTextFileRequest,
    zRequestPermissionRequest,
    zCreateTerminalRequest,
    zTerminalOutputRequest,
    zReleaseTerminalRequest,
    zWaitForTerminalExitRequest,
    zKillTerminalRequest,
    zCreateElicitationRequest,
    zExtRequest
  ]).nullish()
});
var zProtocolVersion = int().gte(0).lte(65535);
var zPromptCapabilities = object({
  image: defaultOnError(boolean2().optional().default(false), () => false),
  audio: defaultOnError(boolean2().optional().default(false), () => false),
  embeddedContext: defaultOnError(boolean2().optional().default(false), () => false),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zMcpCapabilities = object({
  http: defaultOnError(boolean2().optional().default(false), () => false),
  sse: defaultOnError(boolean2().optional().default(false), () => false),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSessionListCapabilities = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSessionDeleteCapabilities = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSessionAdditionalDirectoriesCapabilities = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSessionResumeCapabilities = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSessionCloseCapabilities = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSessionCapabilities = object({
  list: defaultOnError(zSessionListCapabilities.nullish(), () => void 0),
  delete: defaultOnError(zSessionDeleteCapabilities.nullish(), () => void 0),
  additionalDirectories: defaultOnError(zSessionAdditionalDirectoriesCapabilities.nullish(), () => void 0),
  resume: defaultOnError(zSessionResumeCapabilities.nullish(), () => void 0),
  close: defaultOnError(zSessionCloseCapabilities.nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zLogoutCapabilities = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zAgentAuthCapabilities = object({
  logout: defaultOnError(zLogoutCapabilities.nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zAgentCapabilities = object({
  loadSession: defaultOnError(boolean2().optional().default(false), () => false),
  promptCapabilities: defaultOnError(zPromptCapabilities.optional().default({
    image: false,
    audio: false,
    embeddedContext: false
  }), () => ({
    image: false,
    audio: false,
    embeddedContext: false
  })),
  mcpCapabilities: defaultOnError(zMcpCapabilities.optional().default({ http: false, sse: false }), () => ({ http: false, sse: false })),
  sessionCapabilities: defaultOnError(zSessionCapabilities.optional().default({}), () => ({})),
  auth: defaultOnError(zAgentAuthCapabilities.optional().default({}), () => ({})),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zAuthMethodId = string2();
var zAuthMethodTerminal = object({
  id: zAuthMethodId,
  name: string2(),
  description: defaultOnError(string2().nullish(), () => void 0),
  args: defaultOnNull(array(string2()), () => []).optional(),
  env: defaultOnNull(record(string2(), string2()), () => ({})).optional(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zAuthMethodAgent = object({
  id: zAuthMethodId,
  name: string2(),
  description: defaultOnError(string2().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zAuthMethod = union([
  zAuthMethodTerminal.and(object({
    type: literal("terminal")
  })),
  zAuthMethodAgent
]);
var zImplementation = object({
  name: string2(),
  title: defaultOnError(string2().nullish(), () => void 0),
  version: string2(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zInitializeResponse = object({
  protocolVersion: zProtocolVersion,
  agentCapabilities: defaultOnError(zAgentCapabilities.optional().default({
    loadSession: false,
    promptCapabilities: {
      image: false,
      audio: false,
      embeddedContext: false
    },
    mcpCapabilities: { http: false, sse: false },
    sessionCapabilities: {},
    auth: {}
  }), () => ({
    loadSession: false,
    promptCapabilities: {
      image: false,
      audio: false,
      embeddedContext: false
    },
    mcpCapabilities: { http: false, sse: false },
    sessionCapabilities: {},
    auth: {}
  })),
  authMethods: defaultOnError(vecSkipError(zAuthMethod).optional().default([]), () => []),
  agentInfo: defaultOnError(zImplementation.nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zAuthenticateResponse = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zLogoutResponse = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSessionModeId = string2();
var zSessionMode = object({
  id: zSessionModeId,
  name: string2(),
  description: defaultOnError(string2().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSessionModeState = object({
  currentModeId: zSessionModeId,
  availableModes: requiredDefaultOnError(vecSkipError(zSessionMode), () => []),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSessionConfigId = string2();
var zSessionConfigOptionCategory = union([
  literal("mode"),
  literal("model"),
  literal("model_config"),
  literal("thought_level"),
  string2()
]);
var zSessionConfigValueId = string2();
var zSessionConfigSelectOption = object({
  value: zSessionConfigValueId,
  name: string2(),
  description: defaultOnError(string2().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSessionConfigGroupId = string2();
var zSessionConfigSelectGroup = object({
  group: zSessionConfigGroupId,
  name: string2(),
  options: requiredDefaultOnError(vecSkipError(zSessionConfigSelectOption), () => []),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSessionConfigSelectOptions = union([
  array(zSessionConfigSelectOption),
  array(zSessionConfigSelectGroup)
]);
var zSessionConfigSelect = object({
  currentValue: zSessionConfigValueId,
  options: zSessionConfigSelectOptions
});
var zSessionConfigBoolean = object({
  currentValue: boolean2()
});
var zSessionConfigOption = intersection(union([
  zSessionConfigSelect.and(object({
    type: literal("select")
  })),
  zSessionConfigBoolean.and(object({
    type: literal("boolean")
  }))
]), object({
  id: zSessionConfigId,
  name: string2(),
  description: defaultOnError(string2().nullish(), () => void 0),
  category: defaultOnError(zSessionConfigOptionCategory.nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
}));
var zNewSessionResponse = object({
  sessionId: zSessionId,
  modes: defaultOnError(zSessionModeState.nullish(), () => void 0),
  configOptions: defaultOnError(vecSkipError(zSessionConfigOption).nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zLoadSessionResponse = object({
  modes: defaultOnError(zSessionModeState.nullish(), () => void 0),
  configOptions: defaultOnError(vecSkipError(zSessionConfigOption).nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSessionInfo = object({
  sessionId: zSessionId,
  cwd: string2(),
  additionalDirectories: defaultOnError(vecSkipError(string2()).optional(), () => []),
  title: defaultOnError(string2().nullish(), () => void 0),
  updatedAt: defaultOnError(string2().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zListSessionsResponse = object({
  sessions: requiredDefaultOnError(vecSkipError(zSessionInfo), () => []),
  nextCursor: defaultOnError(string2().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zDeleteSessionResponse = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zResumeSessionResponse = object({
  modes: defaultOnError(zSessionModeState.nullish(), () => void 0),
  configOptions: defaultOnError(vecSkipError(zSessionConfigOption).nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zCloseSessionResponse = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSetSessionModeResponse = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSetSessionConfigOptionResponse = object({
  configOptions: requiredDefaultOnError(vecSkipError(zSessionConfigOption), () => []),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zStopReason = union([
  literal("end_turn"),
  literal("max_tokens"),
  literal("max_turn_requests"),
  literal("refusal"),
  literal("cancelled")
]);
var zPromptResponse = object({
  stopReason: zStopReason,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zExtResponse = unknown();
var zErrorCode = union([
  literal(-32700),
  literal(-32600),
  literal(-32601),
  literal(-32602),
  literal(-32603),
  literal(-32800),
  literal(-32e3),
  literal(-32002),
  int().min(-2147483648, {
    error: "Invalid value: Expected int32 to be >= -2147483648"
  }).max(2147483647, {
    error: "Invalid value: Expected int32 to be <= 2147483647"
  })
]);
var zError = object({
  code: zErrorCode,
  message: string2(),
  data: defaultOnError(unknown().optional(), () => void 0)
});
var zAgentResponse = union([
  object({
    id: zRequestId,
    result: union([
      zInitializeResponse,
      zAuthenticateResponse,
      zLogoutResponse,
      zNewSessionResponse,
      zLoadSessionResponse,
      zListSessionsResponse,
      zDeleteSessionResponse,
      zResumeSessionResponse,
      zCloseSessionResponse,
      zSetSessionModeResponse,
      zSetSessionConfigOptionResponse,
      zPromptResponse,
      zExtResponse
    ])
  }),
  object({
    id: zRequestId,
    error: zError
  })
]);
var zMessageId = string2();
var zContentChunk = object({
  content: zContentBlock,
  messageId: defaultOnError(zMessageId.nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zToolCall = object({
  toolCallId: zToolCallId,
  title: string2(),
  name: defaultOnError(string2().nullish(), () => void 0),
  kind: defaultOnError(zToolKind.optional(), () => void 0),
  status: defaultOnError(zToolCallStatus.optional(), () => void 0),
  content: defaultOnError(vecSkipError(zToolCallContent).optional(), () => []),
  locations: defaultOnError(vecSkipError(zToolCallLocation).optional(), () => []),
  rawInput: defaultOnError(unknown().optional(), () => void 0),
  rawOutput: defaultOnError(unknown().optional(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zPlanEntryPriority = union([
  literal("high"),
  literal("medium"),
  literal("low")
]);
var zPlanEntryStatus = union([
  literal("pending"),
  literal("in_progress"),
  literal("completed")
]);
var zPlanEntry = object({
  content: string2(),
  priority: zPlanEntryPriority,
  status: zPlanEntryStatus,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zPlan = object({
  entries: requiredDefaultOnError(vecSkipError(zPlanEntry), () => []),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zUnstructuredCommandInput = object({
  hint: string2(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zAvailableCommandInput = zUnstructuredCommandInput;
var zAvailableCommand = object({
  name: string2(),
  description: string2(),
  input: defaultOnError(zAvailableCommandInput.nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zAvailableCommandsUpdate = object({
  availableCommands: requiredDefaultOnError(vecSkipError(zAvailableCommand), () => []),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zCurrentModeUpdate = object({
  currentModeId: zSessionModeId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zConfigOptionUpdate = object({
  configOptions: requiredDefaultOnError(vecSkipError(zSessionConfigOption), () => []),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSessionInfoUpdate = object({
  title: defaultOnError(string2().nullish(), () => void 0),
  updatedAt: defaultOnError(string2().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zCost = object({
  amount: number2(),
  currency: string2(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zUsageUpdate = object({
  used: wideInt().gte(0),
  size: wideInt().gte(0),
  cost: defaultOnError(zCost.nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zNoticeSeverity = union([
  literal("info"),
  literal("warning"),
  literal("error"),
  string2()
]);
var zNotice = object({
  severity: zNoticeSeverity,
  title: string2().min(1),
  description: defaultOnError(string2().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zCompactionId = string2();
var zCompactionStatus = union([
  literal("in_progress"),
  literal("completed"),
  literal("failed"),
  literal("cancelled"),
  string2()
]);
var zCompactionUpdate = object({
  compactionId: zCompactionId,
  status: zCompactionStatus,
  summary: defaultOnError(vecSkipError(zContentBlock).nullish(), () => void 0),
  error: defaultOnError(string2().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zCompactionSummaryChunk = object({
  compactionId: zCompactionId,
  content: zContentBlock,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSessionUpdate = union([
  zContentChunk.and(object({
    sessionUpdate: literal("user_message_chunk")
  })),
  zContentChunk.and(object({
    sessionUpdate: literal("agent_message_chunk")
  })),
  zContentChunk.and(object({
    sessionUpdate: literal("agent_thought_chunk")
  })),
  zToolCall.and(object({
    sessionUpdate: literal("tool_call")
  })),
  zToolCallUpdate.and(object({
    sessionUpdate: literal("tool_call_update")
  })),
  zPlan.and(object({
    sessionUpdate: literal("plan")
  })),
  zAvailableCommandsUpdate.and(object({
    sessionUpdate: literal("available_commands_update")
  })),
  zCurrentModeUpdate.and(object({
    sessionUpdate: literal("current_mode_update")
  })),
  zConfigOptionUpdate.and(object({
    sessionUpdate: literal("config_option_update")
  })),
  zSessionInfoUpdate.and(object({
    sessionUpdate: literal("session_info_update")
  })),
  zUsageUpdate.and(object({
    sessionUpdate: literal("usage_update")
  })),
  zNotice.and(object({
    sessionUpdate: literal("notice")
  })),
  zCompactionUpdate.and(object({
    sessionUpdate: literal("compaction_update")
  })),
  zCompactionSummaryChunk.and(object({
    sessionUpdate: literal("compaction_summary_chunk")
  }))
]);
var zSessionNotification = object({
  sessionId: zSessionId,
  update: zSessionUpdate,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zCompleteElicitationNotification = object({
  elicitationId: zElicitationId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zExtNotification = unknown();
var zAgentNotification = object({
  method: string2(),
  params: union([
    zSessionNotification,
    zCompleteElicitationNotification,
    zExtNotification
  ]).nullish()
});
var zFileSystemCapabilities = object({
  readTextFile: defaultOnError(boolean2().optional().default(false), () => false),
  writeTextFile: defaultOnError(boolean2().optional().default(false), () => false),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zCompactionCapabilities = record(string2(), unknown());
var zBooleanConfigOptionCapabilities = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSessionConfigOptionsCapabilities = object({
  boolean: defaultOnError(zBooleanConfigOptionCapabilities.nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zNoticeCapabilities = record(string2(), unknown());
var zClientSessionCapabilities = object({
  compaction: defaultOnError(zCompactionCapabilities.nullish(), () => void 0),
  configOptions: defaultOnError(zSessionConfigOptionsCapabilities.nullish(), () => void 0),
  notices: defaultOnError(zNoticeCapabilities.nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zAuthCapabilities = object({
  terminal: defaultOnError(boolean2().optional().default(false), () => false),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zElicitationFormCapabilities = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zElicitationUrlCapabilities = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zElicitationCapabilities = object({
  form: defaultOnError(zElicitationFormCapabilities.nullish(), () => void 0),
  url: defaultOnError(zElicitationUrlCapabilities.nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zClientCapabilities = object({
  fs: defaultOnError(zFileSystemCapabilities.optional().default({ readTextFile: false, writeTextFile: false }), () => ({ readTextFile: false, writeTextFile: false })),
  terminal: defaultOnError(boolean2().optional().default(false), () => false),
  session: defaultOnError(zClientSessionCapabilities.nullish(), () => void 0),
  auth: defaultOnError(zAuthCapabilities.optional().default({ terminal: false }), () => ({ terminal: false })),
  elicitation: defaultOnError(zElicitationCapabilities.nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zInitializeRequest = object({
  protocolVersion: zProtocolVersion,
  clientCapabilities: defaultOnError(zClientCapabilities.optional().default({
    fs: { readTextFile: false, writeTextFile: false },
    terminal: false,
    auth: { terminal: false }
  }), () => ({
    fs: { readTextFile: false, writeTextFile: false },
    terminal: false,
    auth: { terminal: false }
  })),
  clientInfo: defaultOnError(zImplementation.nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zAuthenticateRequest = object({
  methodId: zAuthMethodId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zLogoutRequest = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zHttpHeader = object({
  name: string2(),
  value: string2(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zMcpServerHttp = object({
  name: string2(),
  url: string2(),
  headers: defaultOnNull(array(zHttpHeader), () => []),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zMcpServerSse = object({
  name: string2(),
  url: string2(),
  headers: defaultOnNull(array(zHttpHeader), () => []),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zMcpServerStdio = object({
  name: string2(),
  command: string2(),
  args: defaultOnNull(array(string2()), () => []),
  env: defaultOnNull(array(zEnvVariable), () => []),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zMcpServer = union([
  zMcpServerHttp.and(object({
    type: literal("http")
  })),
  zMcpServerSse.and(object({
    type: literal("sse")
  })),
  zMcpServerStdio
]);
var zNewSessionRequest = object({
  cwd: string2(),
  additionalDirectories: defaultOnNull(array(string2()), () => []).optional(),
  mcpServers: defaultOnNull(array(zMcpServer), () => []),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zLoadSessionRequest = object({
  mcpServers: defaultOnNull(array(zMcpServer), () => []),
  cwd: string2(),
  additionalDirectories: defaultOnNull(array(string2()), () => []).optional(),
  sessionId: zSessionId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zListSessionsRequest = object({
  cwd: string2().nullish(),
  cursor: string2().nullish(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zDeleteSessionRequest = object({
  sessionId: zSessionId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zResumeSessionRequest = object({
  sessionId: zSessionId,
  cwd: string2(),
  additionalDirectories: defaultOnNull(array(string2()), () => []).optional(),
  mcpServers: defaultOnNull(array(zMcpServer), () => []).optional(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zCloseSessionRequest = object({
  sessionId: zSessionId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSetSessionModeRequest = object({
  sessionId: zSessionId,
  modeId: zSessionModeId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSetSessionConfigOptionRequest = intersection(union([
  object({
    type: literal("boolean"),
    value: boolean2()
  }),
  object({
    value: zSessionConfigValueId
  })
]), object({
  sessionId: zSessionId,
  configId: zSessionConfigId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
}));
var zPromptRequest = object({
  sessionId: zSessionId,
  prompt: array(zContentBlock),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zClientRequest = object({
  id: zRequestId,
  method: string2(),
  params: union([
    zInitializeRequest,
    zAuthenticateRequest,
    zLogoutRequest,
    zNewSessionRequest,
    zLoadSessionRequest,
    zListSessionsRequest,
    zDeleteSessionRequest,
    zResumeSessionRequest,
    zCloseSessionRequest,
    zSetSessionModeRequest,
    zSetSessionConfigOptionRequest,
    zPromptRequest,
    zExtRequest
  ]).nullish()
});
var zWriteTextFileResponse = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zReadTextFileResponse = object({
  content: string2(),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zSelectedPermissionOutcome = object({
  optionId: zPermissionOptionId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zRequestPermissionOutcome = union([
  object({
    outcome: literal("cancelled")
  }),
  zSelectedPermissionOutcome.and(object({
    outcome: literal("selected")
  }))
]);
var zRequestPermissionResponse = object({
  outcome: zRequestPermissionOutcome,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zCreateTerminalResponse = object({
  terminalId: zTerminalId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zTerminalExitStatus = object({
  exitCode: defaultOnError(int().gte(0).max(4294967295, {
    error: "Invalid value: Expected uint32 to be <= 4294967295"
  }).nullish(), () => void 0),
  signal: defaultOnError(string2().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zTerminalOutputResponse = object({
  output: string2(),
  truncated: boolean2(),
  exitStatus: defaultOnError(zTerminalExitStatus.nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zReleaseTerminalResponse = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zWaitForTerminalExitResponse = object({
  exitCode: defaultOnError(int().gte(0).max(4294967295, {
    error: "Invalid value: Expected uint32 to be <= 4294967295"
  }).nullish(), () => void 0),
  signal: defaultOnError(string2().nullish(), () => void 0),
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zKillTerminalResponse = object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zElicitationContentValue = union([
  string2(),
  wideInt(),
  number2(),
  boolean2(),
  array(string2())
]);
var zElicitationAcceptAction = object({
  content: record(string2(), zElicitationContentValue).nullish()
});
var zCreateElicitationResponse = preserveCustomPayload(intersection(union([
  zElicitationAcceptAction.and(object({
    action: literal("accept")
  })),
  object({
    action: literal("decline")
  }),
  object({
    action: literal("cancel")
  }),
  excludeKnownTags(object({
    action: string2()
  }), "action", ["accept", "cancel", "decline"])
]), object({
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
})), "action", ["accept", "cancel", "decline"]);
var zClientResponse = union([
  object({
    id: zRequestId,
    result: union([
      zWriteTextFileResponse,
      zReadTextFileResponse,
      zRequestPermissionResponse,
      zCreateTerminalResponse,
      zTerminalOutputResponse,
      zReleaseTerminalResponse,
      zWaitForTerminalExitResponse,
      zKillTerminalResponse,
      zCreateElicitationResponse,
      zExtResponse
    ])
  }),
  object({
    id: zRequestId,
    error: zError
  })
]);
var zCancelNotification = object({
  sessionId: zSessionId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});
var zClientNotification = object({
  method: string2(),
  params: union([zCancelNotification, zExtNotification]).nullish()
});
var zCancelRequestNotification = object({
  requestId: zRequestId,
  _meta: defaultOnError(record(string2(), unknown()).nullish(), () => void 0)
});

// node_modules/@agentclientprotocol/sdk/dist/core/jsonrpc.js
var CANCEL_REQUEST_METHOD = "$/cancel_request";
function treatSessionNotFoundAsSuccess(method, response, success) {
  if (method !== "session/close" && method !== "session/delete") {
    return response;
  }
  return response.catch((error2) => {
    if (error2 instanceof RequestError && error2.code === -32002) {
      return success();
    }
    throw error2;
  });
}
function isRequestMessage(value) {
  return isJsonRpcEnvelope(value) && "id" in value && typeof value["method"] === "string" && isJsonRpcId(value["id"]);
}
function isResponseMessage(value) {
  if (!isJsonRpcEnvelope(value) || "method" in value) {
    return false;
  }
  if (!("id" in value) || !isJsonRpcId(value["id"])) {
    return false;
  }
  const hasResult = Object.hasOwn(value, "result");
  const hasError = Object.hasOwn(value, "error");
  if (hasResult === hasError) {
    return false;
  }
  return !hasError || isErrorResponse(value["error"]);
}
function isNotificationMessage(value) {
  return isJsonRpcEnvelope(value) && !("id" in value) && typeof value["method"] === "string";
}
function isRecord(value) {
  return typeof value === "object" && value !== null;
}
function isJsonRpcEnvelope(value) {
  return isRecord(value) && value["jsonrpc"] === "2.0";
}
function isJsonRpcId(value) {
  return value === null || typeof value === "string" || typeof value === "number" && Number.isFinite(value);
}
function isResponseShapedMessage(value) {
  return isRecord(value) && !("method" in value) && ("id" in value || "result" in value || "error" in value);
}
function cancelRequestId(params) {
  if (!isRecord(params) || !isJsonRpcId(params["requestId"])) {
    return void 0;
  }
  return params["requestId"];
}
function isErrorResponse(value) {
  return isRecord(value) && typeof value["code"] === "number" && Number.isInteger(value["code"]) && typeof value["message"] === "string";
}
var Handled = {
  /**
   * Marks a message as handled.
   */
  yes() {
    return { handled: true };
  },
  /**
   * Leaves a message unhandled so later handlers can process it.
   */
  no(message, retry = false) {
    return { handled: false, message, retry };
  }
};
function rejectedPromise(error2) {
  const promise = Promise.reject(error2);
  promise.catch(() => {
  });
  return promise;
}
function errorDetails(error2) {
  if (error2 instanceof Error) {
    return error2.message;
  }
  if (typeof error2 === "object" && error2 != null && "message" in error2 && typeof error2.message === "string") {
    return error2.message;
  }
  return void 0;
}
function isZodError(error2) {
  return typeof error2 === "object" && error2 !== null && "name" in error2 && error2.name === "ZodError" && "issues" in error2 && Array.isArray(error2.issues) && "format" in error2 && typeof error2.format === "function";
}
function describeInvalidMessage(message) {
  if (Array.isArray(message)) {
    return { type: "array" };
  }
  if (isRecord(message)) {
    return { type: "object", keys: Object.keys(message) };
  }
  return { type: message === null ? "null" : typeof message };
}
function errorToResult(error2) {
  if (error2 instanceof RequestError) {
    return error2.toResult();
  }
  if (isZodError(error2)) {
    return RequestError.invalidParams(error2.format()).toResult();
  }
  const details = errorDetails(error2);
  try {
    return RequestError.internalError(details ? JSON.parse(details) : {}).toResult();
  } catch {
    return RequestError.internalError({ details }).toResult();
  }
}
function requestCancelledError(reason) {
  if (reason instanceof RequestError && reason.code === -32800) {
    return reason;
  }
  return RequestError.requestCancelled(reason);
}
function errorToRequestResult(error2, signal) {
  const requestCancelled = abortErrorToRequestCancelled(error2, signal);
  return requestCancelled ? requestCancelled.toResult() : errorToResult(error2);
}
function abortErrorToRequestCancelled(error2, signal) {
  if (!signal.aborted || !isAbortError(error2)) {
    return void 0;
  }
  return requestCancelledError(signal.reason);
}
function isAbortError(error2) {
  if (typeof error2 !== "object" || error2 === null) {
    return false;
  }
  const maybeAbortError = error2;
  return maybeAbortError.name === "AbortError" || maybeAbortError.code === "ABORT_ERR";
}
var RequestResponder = class {
  id;
  sendResult;
  signal;
  finishRequest;
  didRespond = false;
  constructor(id, sendResult, signal = new AbortController().signal, finishRequest) {
    this.id = id;
    this.sendResult = sendResult;
    this.signal = signal;
    this.finishRequest = finishRequest;
  }
  /**
   * Whether this request has already received a response.
   */
  get responded() {
    return this.didRespond;
  }
  /**
   * Sends a successful JSON-RPC response.
   */
  respond(response) {
    return this.respondWithResult({ result: response ?? null });
  }
  /**
   * Sends an error JSON-RPC response.
   */
  respondWithError(error2) {
    const errorResponse = error2 instanceof RequestError ? error2.toErrorResponse() : error2;
    return this.respondWithResult({ error: errorResponse });
  }
  /**
   * Sends a complete JSON-RPC result payload.
   */
  respondWithResult(result) {
    if (this.didRespond) {
      return rejectedPromise(new Error("JSON-RPC request already responded"));
    }
    this.didRespond = true;
    return this.sendResult(result).finally(() => {
      this.finishRequest?.();
    });
  }
};
var nullIdResponders = /* @__PURE__ */ new WeakSet();
var HandlerRegistration = class {
  disposeHandler;
  active = true;
  constructor(disposeHandler) {
    this.disposeHandler = disposeHandler;
  }
  /**
   * Unregisters the associated handler.
   */
  dispose() {
    if (!this.active) {
      return;
    }
    this.active = false;
    this.disposeHandler();
  }
  /**
   * Supports explicit resource management with `using`.
   */
  [Symbol.dispose]() {
    this.dispose();
  }
  /**
   * Returns this registration for call sites that intentionally keep it active.
   */
  runIndefinitely() {
    return this;
  }
};
var ConnectionContext = class {
  connection;
  constructor(connection) {
    this.connection = connection;
  }
  /**
   * Sends a request over the connection.
   */
  sendRequest(method, params, mapResponse, options2) {
    return this.connection.sendRequest(method, params, mapResponse, options2);
  }
  /**
   * Sends a notification over the connection.
   */
  sendNotification(method, params) {
    return this.connection.sendNotification(method, params);
  }
  /**
   * Sends a protocol-level request cancellation notification.
   */
  sendCancelRequest(requestId) {
    return this.connection.sendCancelRequest(requestId);
  }
  /**
   * Registers a handler that can be disposed independently.
   */
  addDynamicHandler(handler) {
    return this.connection.addDynamicHandler(handler);
  }
  /**
   * AbortSignal that aborts when the connection closes.
   */
  get signal() {
    return this.connection.signal;
  }
  /**
   * Promise that resolves when the connection closes.
   */
  get closed() {
    return this.connection.closed;
  }
};
var Connection = class {
  pendingResponses = /* @__PURE__ */ new Map();
  incomingRequests = /* @__PURE__ */ new Map();
  nextRequestId = 0;
  staticHandlers = [];
  dynamicHandlers = /* @__PURE__ */ new Set();
  fallbackHandlers = [];
  stream;
  writeQueue = Promise.resolve();
  abortController = new AbortController();
  closedPromise;
  retryQueue = [];
  context = new ConnectionContext(this);
  receiveReader;
  constructor(requestHandlerOrStream, notificationHandlerOrHandlers, streamOrOptions, options2) {
    if (typeof requestHandlerOrStream === "function") {
      const requestHandler = requestHandlerOrStream;
      const notificationHandler = notificationHandlerOrHandlers;
      const stream2 = streamOrOptions;
      this.initialize(stream2, [
        ...options2?.handlers ?? [],
        this.legacyHandler(requestHandler, notificationHandler)
      ], options2);
      return;
    }
    const stream = requestHandlerOrStream;
    const handlers = notificationHandlerOrHandlers;
    const connectionOptions = streamOrOptions;
    this.initialize(stream, [...connectionOptions?.handlers ?? [], ...handlers], connectionOptions);
  }
  /**
   * Creates a builder for configuring a handler-based connection.
   */
  static builder() {
    return new ConnectionBuilder();
  }
  /**
   * Runs an operation while the connection is open, then closes the connection.
   *
   * If the stream closes before `op` settles, the returned promise rejects with
   * the connection close reason.
   */
  runUntil(op) {
    let opSettled = false;
    const opPromise = Promise.resolve().then(() => op(this.context)).finally(() => {
      opSettled = true;
    });
    const closedPromise = this.closed.then(() => {
      if (opSettled) {
        return new Promise(() => {
        });
      }
      throw this.closedReason();
    });
    return Promise.race([opPromise, closedPromise]).finally(() => {
      opSettled = true;
      this.close();
    });
  }
  /**
   * Adds a handler after the connection has started.
   *
   * Any messages queued with `Handled.no(message, true)` are retried after the
   * handler is added.
   */
  addDynamicHandler(handler) {
    this.dynamicHandlers.add(handler);
    if (this.retryQueue.length > 0) {
      for (const message of this.retryQueue.splice(0)) {
        void this.processIncomingMessage(message).catch((error2) => this.close(error2));
      }
    }
    return new HandlerRegistration(() => {
      this.dynamicHandlers.delete(handler);
    });
  }
  /**
   * AbortSignal that aborts when the connection closes.
   */
  get signal() {
    return this.abortController.signal;
  }
  /**
   * Promise that resolves when the connection closes.
   */
  get closed() {
    return this.closedPromise;
  }
  /** @internal */
  getContext() {
    return this.context;
  }
  /**
   * Sends a JSON-RPC request.
   *
   * `mapResponse` can convert the raw result before the returned promise
   * resolves.
   */
  sendRequest(method, params, mapResponse, options2 = {}) {
    if (this.abortController.signal.aborted) {
      return rejectedPromise(this.closedReason());
    }
    const request = this.prepareRequest(method, params, mapResponse, options2);
    const requestSent = this.sendMessage(request.message);
    void requestSent.catch(() => {
    });
    if (options2.cancellationSignal?.aborted) {
      request.cancel();
    }
    return request.response;
  }
  /**
   * Sends a protocol-level request cancellation notification.
   */
  sendCancelRequest(requestId) {
    return this.sendNotification(CANCEL_REQUEST_METHOD, { requestId });
  }
  /**
   * Sends a JSON-RPC notification.
   */
  sendNotification(method, params) {
    if (this.abortController.signal.aborted) {
      return rejectedPromise(this.closedReason());
    }
    return this.sendMessage({ jsonrpc: "2.0", method, params });
  }
  prepareRequest(method, params, mapResponse, options2 = {}) {
    const id = this.nextRequestId++;
    let cancel = () => {
    };
    const response = new Promise((resolve, reject) => {
      const pendingResponse = {
        resolve: (value) => {
          try {
            resolve(mapResponse ? mapResponse(value) : value);
          } catch (error2) {
            reject(error2);
          }
        },
        reject
      };
      cancel = () => {
        if (pendingResponse.cancellationSent) {
          return;
        }
        pendingResponse.cancellationSent = true;
        pendingResponse.cleanup?.();
        void this.sendCancelRequest(id).catch(() => {
        });
        if (options2.stopWaitingOnCancel) {
          pendingResponse.abandoned = true;
          reject(requestCancelledError(options2.cancellationSignal?.reason));
        }
      };
      options2.cancellationSignal?.addEventListener("abort", cancel, {
        once: true
      });
      pendingResponse.cleanup = () => {
        options2.cancellationSignal?.removeEventListener("abort", cancel);
      };
      this.pendingResponses.set(id, pendingResponse);
    });
    response.catch(() => {
    });
    return {
      message: { jsonrpc: "2.0", id, method, params },
      response,
      cancel: () => cancel()
    };
  }
  /**
   * Closes the connection and rejects pending requests.
   */
  close(error2) {
    if (this.abortController.signal.aborted) {
      return;
    }
    const closeError = error2 ?? new Error("ACP connection closed");
    this.abortController.abort(closeError);
    for (const pendingResponse of this.pendingResponses.values()) {
      pendingResponse.cleanup?.();
      pendingResponse.reject(closeError);
    }
    this.pendingResponses.clear();
    for (const controller of this.incomingRequests.values()) {
      controller.abort(closeError);
    }
    this.incomingRequests.clear();
    void this.receiveReader?.cancel(closeError).catch(() => {
    });
  }
  initialize(stream, handlers, options2) {
    this.stream = stream;
    this.staticHandlers = handlers;
    this.fallbackHandlers = [...options2?.fallbackHandlers ?? []];
    this.closedPromise = new Promise((resolve) => {
      this.abortController.signal.addEventListener("abort", () => resolve());
    });
    void this.receive();
  }
  legacyHandler(requestHandler, notificationHandler) {
    return {
      handleMessage: async (message, cx) => {
        if (message.kind === "request") {
          const result = await requestHandler(message.method, message.params, cx);
          await message.responder.respond(result);
        } else {
          await notificationHandler(message.method, message.params, cx);
        }
        return Handled.yes();
      }
    };
  }
  async receive() {
    let closeError = void 0;
    try {
      const reader = this.stream.readable.getReader();
      this.receiveReader = reader;
      try {
        while (!this.abortController.signal.aborted) {
          const { value: message, done } = await reader.read();
          if (this.abortController.signal.aborted) {
            break;
          }
          if (done) {
            break;
          }
          this.receiveUnvalidatedMessage(message);
        }
      } finally {
        if (this.receiveReader === reader) {
          this.receiveReader = void 0;
        }
        reader.releaseLock();
      }
    } catch (error2) {
      closeError = error2;
    } finally {
      this.close(closeError);
    }
  }
  receiveUnvalidatedMessage(message) {
    if (Array.isArray(message)) {
      void this.sendMessage(protocolErrorResponse(RequestError.invalidRequest(void 0, "JSON-RPC batches are not supported"))).catch(() => {
      });
      return;
    }
    if (!isRequestMessage(message) && !isNotificationMessage(message) && !isResponseShapedMessage(message)) {
      void this.sendMessage(protocolErrorResponse(RequestError.invalidRequest(message))).catch(() => {
      });
      return;
    }
    this.receiveMessage(message);
  }
  receiveMessage(message) {
    if (this.abortController.signal.aborted) {
      return Promise.resolve();
    }
    if (!isRecord(message)) {
      console.error("Invalid message", describeInvalidMessage(message));
      return Promise.resolve();
    }
    if ("method" in message) {
      if (!("id" in message)) {
        this.handleProtocolNotification(message);
      }
      return this.processIncomingMessage(this.toIncomingMessage(message)).catch((error2) => this.close(error2));
    } else if ("id" in message) {
      this.handleResponse(message);
    } else {
      console.error("Invalid message", describeInvalidMessage(message));
    }
    return Promise.resolve();
  }
  async processIncomingMessage(message) {
    if (this.abortController.signal.aborted) {
      return;
    }
    let current = message;
    let retry = false;
    try {
      for (const handler of [
        ...this.staticHandlers,
        ...this.dynamicHandlers.values(),
        ...this.fallbackHandlers
      ]) {
        if (this.abortController.signal.aborted) {
          return;
        }
        const result = await handler.handleMessage(current, this.context) ?? {
          handled: true
        };
        if (result.handled) {
          return;
        }
        current = result.message ?? current;
        retry = retry || Boolean(result.retry);
      }
      if (retry) {
        this.retryQueue.push(current);
      } else if (current.kind === "request") {
        await current.responder.respondWithError(RequestError.methodNotFound(current.method));
      }
    } catch (error2) {
      if (this.abortController.signal.aborted) {
        return;
      }
      if (current.kind === "request" && !current.responder.responded) {
        await current.responder.respondWithResult(errorToRequestResult(error2, current.responder.signal));
      } else {
        const response = errorToResult(error2);
        if ("error" in response) {
          console.error("Error handling notification", {
            method: message.method,
            code: response.error.code,
            message: response.error.message
          });
        }
      }
    }
  }
  toIncomingMessage(message) {
    if ("id" in message) {
      const abortController = new AbortController();
      this.incomingRequests.set(message.id, abortController);
      const finishRequest = () => {
        if (this.incomingRequests.get(message.id) === abortController) {
          this.incomingRequests.delete(message.id);
        }
      };
      const responder = new RequestResponder(message.id, (result) => {
        const response = {
          jsonrpc: "2.0",
          id: nullIdResponders.has(responder) ? null : message.id,
          ...result
        };
        return this.sendMessage(response);
      }, abortController.signal, finishRequest);
      return {
        kind: "request",
        method: message.method,
        params: message.params,
        raw: message,
        signal: abortController.signal,
        responder
      };
    }
    return {
      kind: "notification",
      method: message.method,
      params: message.params,
      raw: message
    };
  }
  handleResponse(response) {
    const pendingResponse = this.pendingResponses.get(response.id);
    if (pendingResponse) {
      this.pendingResponses.delete(response.id);
      pendingResponse.cleanup?.();
      if (pendingResponse.abandoned) {
        return;
      }
      if (!isResponseMessage(response)) {
        pendingResponse.reject(RequestError.invalidRequest(response));
      } else if ("result" in response) {
        pendingResponse.resolve(response.result);
      } else {
        const { code, message, data } = response.error;
        pendingResponse.reject(new RequestError(code, message, data));
      }
    } else {
      console.error("Got response to unknown request", response.id);
    }
  }
  handleProtocolNotification(message) {
    if (message.method !== CANCEL_REQUEST_METHOD) {
      return;
    }
    const requestId = cancelRequestId(message.params);
    if (requestId === void 0) {
      return;
    }
    const controller = this.incomingRequests.get(requestId);
    if (!controller || controller.signal.aborted) {
      return;
    }
    controller.abort(RequestError.requestCancelled({ requestId }));
  }
  closedReason() {
    return this.abortController.signal.reason ?? new Error("ACP connection closed");
  }
  async sendMessage(message) {
    if (this.abortController.signal.aborted) {
      return rejectedPromise(this.closedReason());
    }
    this.writeQueue = this.writeQueue.then(async () => {
      if (this.abortController.signal.aborted) {
        throw this.closedReason();
      }
      const writer = this.stream.writable.getWriter();
      try {
        await writer.write(message);
      } finally {
        writer.releaseLock();
      }
    }).catch((error2) => {
      this.close(error2);
      throw error2;
    });
    return this.writeQueue;
  }
};
var ConnectionBuilder = class {
  handlers = [];
  fallbackHandlers = [];
  connectionName;
  /**
   * Sets a diagnostic name used by handlers created from this builder.
   */
  name(name) {
    this.connectionName = name;
    return this;
  }
  /**
   * Adds a raw JSON-RPC handler to the handler chain.
   */
  withHandler(handler) {
    this.handlers.push(handler);
    return this;
  }
  /**
   * Adds a fallback after static and dynamic handlers. Each connection snapshots
   * this registration alongside the builder's other handlers.
   */
  withFallbackHandler(handler) {
    this.fallbackHandlers.push(handler);
    return this;
  }
  /**
   * Adds a handler that can inspect every incoming request or notification.
   *
   * Observer callbacks that return void pass the message through to later
   * handlers. Return `Handled.yes()` to stop dispatch explicitly.
   */
  onReceiveMessage(handler) {
    return this.withHandler({
      handleMessage: async (message, cx) => await handler(message, cx) ?? Handled.no(message),
      describe: () => this.connectionName ?? "onReceiveMessage"
    });
  }
  /**
   * Adds a typed request handler for one method.
   */
  onReceiveRequest(method, parse2, handler) {
    return this.withHandler({
      handleMessage: async (message, cx) => {
        if (message.kind !== "request" || message.method !== method) {
          return Handled.no(message);
        }
        const request = parse2(message.params);
        return await handler(request, message.responder, cx) ?? Handled.yes();
      },
      describe: () => `${this.connectionName ?? "request"}:${method}`
    });
  }
  /**
   * Adds a typed notification handler for one method.
   */
  onReceiveNotification(method, parse2, handler) {
    return this.withHandler({
      handleMessage: async (message, cx) => {
        if (message.kind !== "notification" || message.method !== method) {
          return Handled.no(message);
        }
        const notification = parse2(message.params);
        return await handler(notification, cx) ?? Handled.yes();
      },
      describe: () => `${this.connectionName ?? "notification"}:${method}`
    });
  }
  /**
   * Connects the configured handlers to a stream.
   */
  connect(stream, options2) {
    return new Connection(stream, this.handlers, {
      ...options2,
      fallbackHandlers: [
        ...options2?.fallbackHandlers ?? [],
        ...this.fallbackHandlers
      ]
    });
  }
  /**
   * Connects to a stream for the lifetime of `op`, then closes the connection.
   */
  connectWith(stream, op, options2) {
    return this.connect(stream, options2).runUntil(op);
  }
};
var RequestError = class _RequestError extends Error {
  code;
  /**
   * Additional JSON-RPC error data.
   */
  data;
  constructor(code, message, data) {
    super(message);
    this.code = code;
    this.name = "RequestError";
    this.data = data;
  }
  /**
   * Invalid JSON was received by the server. An error occurred on the server while parsing the JSON text.
   */
  static parseError(data, additionalMessage) {
    return new _RequestError(-32700, `Parse error${additionalMessage ? `: ${additionalMessage}` : ""}`, data);
  }
  /**
   * The JSON sent is not a valid Request object.
   */
  static invalidRequest(data, additionalMessage) {
    return new _RequestError(-32600, `Invalid request${additionalMessage ? `: ${additionalMessage}` : ""}`, data);
  }
  /**
   * The method does not exist / is not available.
   */
  static methodNotFound(method) {
    return new _RequestError(-32601, `"Method not found": ${method}`, {
      method
    });
  }
  /**
   * Invalid method parameter(s).
   */
  static invalidParams(data, additionalMessage) {
    return new _RequestError(-32602, `Invalid params${additionalMessage ? `: ${additionalMessage}` : ""}`, data);
  }
  /**
   * Internal JSON-RPC error.
   */
  static internalError(data, additionalMessage) {
    return new _RequestError(-32603, `Internal error${additionalMessage ? `: ${additionalMessage}` : ""}`, data);
  }
  /**
   * Execution of the request was aborted.
   */
  static requestCancelled(data, additionalMessage) {
    return new _RequestError(-32800, `Request cancelled${additionalMessage ? `: ${additionalMessage}` : ""}`, data);
  }
  /**
   * Authentication required.
   */
  static authRequired(data, additionalMessage) {
    return new _RequestError(-32e3, `Authentication required${additionalMessage ? `: ${additionalMessage}` : ""}`, data);
  }
  /**
   * A referenced resource, such as a session or file, does not exist.
   *
   * For an unknown or inactive session, prefer
   * {@link RequestError.sessionNotFound}, which sets `data.sessionId` as the
   * protocol expects.
   */
  static resourceNotFound(uri) {
    return new _RequestError(-32002, `Resource not found${uri ? `: ${uri}` : ""}`, uri && { uri });
  }
  /**
   * The session named by a session-scoped request does not exist or is not
   * active on this connection.
   *
   * Produces a `-32002` (Resource not found) error with `data.sessionId` set
   * to the requested session ID.
   */
  static sessionNotFound(sessionId) {
    return new _RequestError(-32002, "Resource not found", {
      sessionId
    });
  }
  /**
   * No supported protocol version is at or below the requested version.
   *
   * Produces a `-32022` error whose data lists the `requested` version and the
   * `supported` versions.
   */
  static unsupportedProtocolVersion(requested, supported) {
    return new _RequestError(-32022, "Unsupported protocol version", {
      requested,
      supported: [...supported]
    });
  }
  /**
   * Converts this error to a JSON-RPC result object.
   */
  toResult() {
    return {
      error: {
        code: this.code,
        message: this.message,
        data: this.data
      }
    };
  }
  /**
   * Converts this error to a JSON-RPC error response payload.
   */
  toErrorResponse() {
    return {
      code: this.code,
      message: this.message,
      data: this.data
    };
  }
};
function protocolErrorResponse(error2) {
  return {
    jsonrpc: "2.0",
    id: null,
    error: error2.toErrorResponse()
  };
}

// node_modules/@agentclientprotocol/sdk/dist/core/stream-limits.js
var DEFAULT_MAX_MESSAGE_BYTES = 32 * 1024 * 1024;

// node_modules/@agentclientprotocol/sdk/dist/v1/schema/guards.gen.js
var zGuardCreateElicitationRequestForm = zElicitationFormMode.and(object({ mode: literal("form") })).and(object({ message: string2() }));
var zGuardCreateElicitationRequestUrl = zElicitationUrlMode.and(object({ mode: literal("url") })).and(object({ message: string2() }));
var zGuardCreateElicitationRequestCustom = union([zElicitationSessionScope, zElicitationRequestScope]).and(object({ message: string2() }));
var zGuardElicitationPropertySchemaString = zStringPropertySchema.and(object({ type: literal("string") }));
var zGuardElicitationPropertySchemaNumber = zNumberPropertySchema.and(object({ type: literal("number") }));
var zGuardElicitationPropertySchemaInteger = zIntegerPropertySchema.and(object({ type: literal("integer") }));
var zGuardElicitationPropertySchemaBoolean = zBooleanPropertySchema.and(object({ type: literal("boolean") }));
var zGuardElicitationPropertySchemaArray = zMultiSelectPropertySchema.and(object({ type: literal("array") }));
var zGuardMultiSelectItemsString = zStringMultiSelectItems.and(object({ type: literal("string") }));
var zGuardCreateElicitationResponseAccept = zElicitationAcceptAction.and(object({ action: literal("accept") }));
var zGuardCreateElicitationResponseDecline = object({
  action: literal("decline")
});
var zGuardCreateElicitationResponseCancel = object({
  action: literal("cancel")
});

// node_modules/@agentclientprotocol/sdk/dist/core/app-handlers.js
function parseParams(parser2, params) {
  if (!parser2) {
    return params;
  }
  if (typeof parser2 === "function") {
    return parser2(params);
  }
  return parser2.parse(params);
}
function registerAppRequest(builder, spec, context, handler, serializeResponse, lifecycle) {
  const handle = async (params, responder, cx, method) => {
    try {
      const response = await handler(context(params, cx, responder.signal, responder.id, method));
      const sentResponse = serializeResponse ? serializeResponse(response) : response;
      lifecycle?.beforeResponse(params, sentResponse, cx);
      await responder.respond(sentResponse);
    } catch (error2) {
      lifecycle?.onError(cx, error2);
      throw error2;
    }
  };
  if (spec.method === "*") {
    builder.withFallbackHandler({
      handleMessage: async (message, cx) => {
        if (message.kind !== "request")
          return Handled.no(message);
        await handle(message.params, message.responder, cx, message.method);
        return Handled.yes();
      }
    });
  } else {
    builder.onReceiveRequest(spec.method, (params) => parseParams(spec.params, params), (params, responder, cx) => handle(params, responder, cx, spec.method));
  }
}
function registerAppNotification(builder, spec, context, handler) {
  if (spec.method === "*") {
    builder.withFallbackHandler({
      handleMessage: async (message, cx) => {
        if (message.kind !== "notification")
          return Handled.no(message);
        await handler(context(message.params, cx, cx.signal, message.method));
        return Handled.yes();
      }
    });
  } else {
    builder.onReceiveNotification(spec.method, (params) => parseParams(spec.params, params), (params, cx) => handler(context(params, cx, cx.signal, spec.method)));
  }
}

// node_modules/@agentclientprotocol/sdk/dist/core/connection-access.js
var connectionContextOf = Symbol("connectionContextOf");

// node_modules/@agentclientprotocol/sdk/dist/v1/stable/acp.js
var initializedConnections = /* @__PURE__ */ new WeakSet();
function agentInitializationGate(enabled) {
  return {
    handleMessage(message, cx) {
      if (message.kind !== "request" || !enabled() || initializedConnections.has(cx)) {
        return Handled.no(message);
      }
      if (message.method !== AGENT_METHODS.initialize) {
        return message.responder.respondWithError(RequestError.invalidRequest(void 0, `ACP connection must be initialized before '${message.method}'`)).then(() => Handled.yes());
      }
      const { responder } = message;
      return Handled.no({
        ...message,
        responder: new RequestResponder(responder.id, (result) => {
          if ("result" in result) {
            initializedConnections.add(cx);
          }
          return responder.respondWithResult(result);
        }, responder.signal)
      });
    },
    describe: () => "agent-initialization"
  };
}
function emptyObjectResponse(response) {
  return response ?? {};
}
function isStream(value) {
  return typeof value === "object" && value !== null && "readable" in value && "writable" in value;
}
function memoryStreamPair() {
  const leftToRight = new TransformStream();
  const rightToLeft = new TransformStream();
  return [
    {
      readable: rightToLeft.readable,
      writable: leftToRight.writable
    },
    {
      readable: leftToRight.readable,
      writable: rightToLeft.writable
    }
  ];
}
var methods = {
  agent: {
    initialize: AGENT_METHODS.initialize,
    authenticate: AGENT_METHODS.authenticate,
    logout: AGENT_METHODS.logout,
    session: {
      new: AGENT_METHODS.session_new,
      load: AGENT_METHODS.session_load,
      list: AGENT_METHODS.session_list,
      delete: AGENT_METHODS.session_delete,
      resume: AGENT_METHODS.session_resume,
      close: AGENT_METHODS.session_close,
      setMode: AGENT_METHODS.session_set_mode,
      setConfigOption: AGENT_METHODS.session_set_config_option,
      prompt: AGENT_METHODS.session_prompt,
      cancel: AGENT_METHODS.session_cancel
    }
  },
  client: {
    session: {
      requestPermission: CLIENT_METHODS.session_request_permission,
      update: CLIENT_METHODS.session_update
    },
    fs: {
      writeTextFile: CLIENT_METHODS.fs_write_text_file,
      readTextFile: CLIENT_METHODS.fs_read_text_file
    },
    terminal: {
      create: CLIENT_METHODS.terminal_create,
      output: CLIENT_METHODS.terminal_output,
      release: CLIENT_METHODS.terminal_release,
      waitForExit: CLIENT_METHODS.terminal_wait_for_exit,
      kill: CLIENT_METHODS.terminal_kill
    },
    elicitation: {
      create: CLIENT_METHODS.elicitation_create,
      complete: CLIENT_METHODS.elicitation_complete
    }
  },
  protocol: {
    cancelRequest: PROTOCOL_METHODS.cancel_request
  }
};
var startActiveSession = Symbol("startActiveSession");
var AcpContext = class {
  cx;
  currentRequestId;
  /** @internal */
  constructor(cx, currentRequestId) {
    this.cx = cx;
    this.currentRequestId = currentRequestId;
  }
  /**
   * JSON-RPC id of the request currently being handled.
   *
   * This is `undefined` for notification handlers and for contexts created
   * outside an inbound request, such as `connect(...)` and `connectWith(...)`.
   */
  get requestId() {
    return this.currentRequestId;
  }
  /** @internal */
  get connectionContext() {
    return this.cx;
  }
  /** @hidden SDK helpers use this to identify the connection. */
  [connectionContextOf]() {
    return this.cx;
  }
  /** @internal */
  sendRequest(method, params, mapResponse, options2) {
    return this.cx.sendRequest(method, params, mapResponse, options2);
  }
  /** @internal */
  sendNotification(method, params) {
    return this.cx.sendNotification(method, params);
  }
  /** @internal */
  addDynamicHandler(handler) {
    return this.cx.addDynamicHandler(handler);
  }
};
var AgentContext = class _AgentContext extends AcpContext {
  constructor(cx, requestId) {
    super(cx, requestId);
  }
  /** @internal */
  static create(cx, requestId) {
    return new _AgentContext(cx, requestId);
  }
  /**
   * Executes a command in a new client terminal.
   *
   * Returns a handle for reading output, waiting for exit, killing the command,
   * and releasing the terminal. Check the client's terminal capability first.
   * Use `await using` or call `release()` when finished.
   */
  async createTerminal(params, options2) {
    const response = await this.request(CLIENT_METHODS.terminal_create, params, options2);
    return TerminalHandle.create(response.terminalId, params.sessionId, this.connectionContext);
  }
  request(method, params, options2) {
    const spec = clientRequestSpecsByMethod[method];
    return this.sendRequest(method, params, spec?.mapResponse, options2);
  }
  notify(method, params) {
    return this.sendNotification(method, params);
  }
};
var ClientContext = class _ClientContext extends AcpContext {
  constructor(cx, requestId) {
    super(cx, requestId);
  }
  /** @internal */
  static create(cx, requestId) {
    return new _ClientContext(cx, requestId);
  }
  /** @internal */
  [startActiveSession](params, options2) {
    return this.sendRequest(AGENT_METHODS.session_new, params, (response) => this.attachSession(response), options2);
  }
  buildSession(cwdOrRequest) {
    if (typeof cwdOrRequest === "string") {
      return SessionBuilder.create(this, {
        cwd: cwdOrRequest,
        mcpServers: []
      });
    }
    return SessionBuilder.create(this, cwdOrRequest);
  }
  /**
   * Builds active-session helpers around a `session/new` response.
   */
  attachSession(response) {
    const updates = new AsyncQueue();
    const closeSignal = this.connectionContext.signal;
    const failUpdatesOnClose = () => {
      updates.fail(closeSignal.reason ?? new Error("ACP connection closed"));
    };
    if (closeSignal.aborted) {
      failUpdatesOnClose();
    } else {
      closeSignal.addEventListener("abort", failUpdatesOnClose);
    }
    const sessionRegistration = sessionUpdateRouter(this.connectionContext).attach(response, updates);
    const closeRegistration = new HandlerRegistration(() => {
      closeSignal.removeEventListener("abort", failUpdatesOnClose);
    });
    return ActiveSession.create(this, response, updates, [
      sessionRegistration,
      closeRegistration
    ]);
  }
  request(method, params, options2) {
    const spec = agentRequestSpecsByMethod[method];
    return treatSessionNotFoundAsSuccess(method, this.sendRequest(method, params, spec?.mapResponse, options2), () => spec?.mapResponse ? spec.mapResponse({}) : {});
  }
  notify(method, params) {
    return this.sendNotification(method, params);
  }
};
var AcpConnectionHandle = class {
  connection;
  constructor(connection) {
    this.connection = connection;
  }
  get signal() {
    return this.connection.signal;
  }
  get closed() {
    return this.connection.closed;
  }
  close(error2) {
    this.connection.close(error2);
  }
};
var AgentConnectionHandle = class extends AcpConnectionHandle {
  connectHandlers;
  client;
  didStartConnectHandlers = false;
  constructor(connection, connectHandlers = []) {
    super(connection);
    this.connectHandlers = connectHandlers;
    this.client = AgentContext.create(connection.getContext());
  }
  /** @internal */
  startConnectHandlers() {
    if (this.didStartConnectHandlers) {
      return;
    }
    this.didStartConnectHandlers = true;
    runConnectHandlers(this, this.connectHandlers);
  }
};
var ClientConnectionHandle = class extends AcpConnectionHandle {
  connectHandlers;
  agent;
  didStartConnectHandlers = false;
  constructor(connection, connectHandlers = []) {
    super(connection);
    this.connectHandlers = connectHandlers;
    this.agent = ClientContext.create(connection.getContext());
  }
  /** @internal */
  startConnectHandlers() {
    if (this.didStartConnectHandlers) {
      return;
    }
    this.didStartConnectHandlers = true;
    runConnectHandlers(this, this.connectHandlers);
  }
};
function agentConnection(connection, connectHandlers = []) {
  return new AgentConnectionHandle(connection, connectHandlers);
}
function clientConnection(connection, connectHandlers = []) {
  return new ClientConnectionHandle(connection, connectHandlers);
}
var AsyncQueue = class {
  values = [];
  waiters = [];
  failed = false;
  failure;
  enqueue(value) {
    if (this.failed) {
      return;
    }
    const waiter = this.waiters.shift();
    if (waiter) {
      waiter.resolve(value);
    } else {
      this.values.push({ kind: "value", value });
    }
  }
  reject(error2) {
    if (this.failed) {
      return;
    }
    if (this.waiters.length > 0) {
      for (const waiter of this.waiters.splice(0)) {
        waiter.reject(error2);
      }
      return;
    }
    this.values.push({ kind: "error", error: error2 });
  }
  clearErrors() {
    this.values = this.values.filter((entry) => entry.kind === "value");
  }
  fail(error2) {
    if (this.failed) {
      return;
    }
    this.failed = true;
    this.failure = error2;
    for (const waiter of this.waiters.splice(0)) {
      waiter.reject(error2);
    }
  }
  next() {
    if (this.values.length > 0) {
      const entry = this.values.shift();
      if (entry.kind === "error") {
        return Promise.reject(entry.error);
      }
      return Promise.resolve(entry.value);
    }
    if (this.failed) {
      return Promise.reject(this.failure);
    }
    return new Promise((resolve, reject) => {
      this.waiters.push({ resolve, reject });
    });
  }
};
function cloneNewSessionRequest(request) {
  return {
    ...request,
    additionalDirectories: request.additionalDirectories ? [...request.additionalDirectories] : void 0,
    mcpServers: [...request.mcpServers]
  };
}
var SessionBuilder = class _SessionBuilder {
  cx;
  request;
  constructor(cx, request) {
    this.cx = cx;
    this.request = cloneNewSessionRequest(request);
  }
  /** @internal */
  static create(cx, request) {
    return new _SessionBuilder(cx, request);
  }
  /**
   * Returns the `session/new` request that will be sent.
   *
   * The returned object is a defensive copy, so mutating it does not change the
   * builder.
   */
  toRequest() {
    return cloneNewSessionRequest(this.request);
  }
  /**
   * Replaces the additional workspace roots for this session.
   *
   * `additionalDirectories` expand the session's file-system scope without
   * changing `cwd`. Each path should be absolute.
   */
  withAdditionalDirectories(additionalDirectories) {
    this.request = {
      ...this.request,
      additionalDirectories: [...additionalDirectories]
    };
    return this;
  }
  /**
   * Adds one MCP server to the `session/new` request.
   */
  withMcpServer(mcpServer) {
    this.request = {
      ...this.request,
      mcpServers: [...this.request.mcpServers, mcpServer]
    };
    return this;
  }
  /**
   * Starts the session and returns an `ActiveSession` for prompting and reading
   * updates.
   *
   * Call `dispose()` on the returned session when you no longer need update
   * routing, or use `withSession(...)` to scope disposal automatically.
   */
  async start(options2) {
    return this.cx[startActiveSession](this.toRequest(), options2);
  }
  /**
   * Starts the session, runs `op`, and disposes the active-session update
   * routing when `op` finishes or throws.
   */
  async withSession(op) {
    const session = await this.start();
    try {
      return await op(session);
    } finally {
      session.dispose();
    }
  }
};
var ActiveSession = class _ActiveSession {
  cx;
  sessionResponse;
  updates;
  registrations;
  constructor(cx, sessionResponse, updates, registrations) {
    this.cx = cx;
    this.sessionResponse = sessionResponse;
    this.updates = updates;
    this.registrations = registrations;
  }
  /** @internal */
  static create(cx, sessionResponse, updates, registrations) {
    return new _ActiveSession(cx, sessionResponse, updates, registrations);
  }
  /**
   * Session ID returned by `session/new`.
   */
  get sessionId() {
    return this.sessionResponse.sessionId;
  }
  /**
   * Mode state returned when the session was created, if the agent provided it.
   */
  get modes() {
    return this.sessionResponse.modes;
  }
  /**
   * Metadata returned when the session was created.
   */
  get meta() {
    return this.sessionResponse._meta;
  }
  /**
   * Full response returned by `session/new`.
   */
  get newSessionResponse() {
    return this.sessionResponse;
  }
  /**
   * Sends a prompt to this session.
   *
   * Strings are converted to one text content block. A single content block is
   * wrapped in an array. The returned promise resolves with the final
   * `PromptResponse`, and the same completion is also queued as a `stop`
   * message for `nextUpdate()`.
   */
  prompt(prompt, options2) {
    this.updates.clearErrors();
    const response = this.cx.request(AGENT_METHODS.session_prompt, {
      sessionId: this.sessionId,
      prompt: this.promptBlocks(prompt)
    }, options2);
    void response.then((value) => {
      this.updates.enqueue({
        kind: "stop",
        response: value,
        stopReason: value.stopReason
      });
    }, (error2) => {
      this.updates.reject(error2);
    });
    return response;
  }
  /**
   * Reads the next update or stop message for this session.
   */
  nextUpdate() {
    return this.updates.next();
  }
  /**
   * Reads text chunks until the current prompt turn stops.
   *
   * Only `agent_message_chunk` updates with text content are appended. Other
   * update types are ignored by this helper; use `nextUpdate()` when you need
   * tool calls, plans, or the final `PromptResponse`.
   */
  async readText() {
    let output = "";
    for (; ; ) {
      const message = await this.nextUpdate();
      if (message.kind === "stop") {
        return output;
      }
      const { update } = message;
      if (update.sessionUpdate === "agent_message_chunk" && update.content.type === "text") {
        output += update.content.text;
      }
    }
  }
  /**
   * Stops routing updates to this active-session helper.
   *
   * This does not close the ACP session on the agent. Use `ClientContext`
   * session lifecycle methods when the protocol session itself should be closed
   * or deleted.
   */
  dispose() {
    for (const registration of this.registrations.splice(0)) {
      registration.dispose();
    }
    this.updates.fail(new Error("Active session disposed"));
  }
  /**
   * Supports explicit resource management with `using`.
   */
  [Symbol.dispose]() {
    this.dispose();
  }
  promptBlocks(prompt) {
    if (typeof prompt === "string") {
      return [{ type: "text", text: prompt }];
    }
    if (Array.isArray(prompt)) {
      return prompt;
    }
    return [prompt];
  }
};
function requestSpec(method, params, mapResponse) {
  return {
    method,
    // Match Rust's empty-payload handling without bypassing required fields.
    params: (value) => parseParams(params, value ?? {}),
    mapResponse
  };
}
function notificationSpec(method, params) {
  return { method, params };
}
function specsByMethod(specs) {
  const byMethod = {};
  for (const spec of Object.values(specs)) {
    byMethod[spec.method] = spec;
  }
  return byMethod;
}
var agentRequestSpecs = {
  initialize: requestSpec(AGENT_METHODS.initialize, zInitializeRequest),
  newSession: requestSpec(AGENT_METHODS.session_new, zNewSessionRequest),
  loadSession: requestSpec(AGENT_METHODS.session_load, zLoadSessionRequest, emptyObjectResponse),
  listSessions: requestSpec(AGENT_METHODS.session_list, zListSessionsRequest),
  deleteSession: requestSpec(AGENT_METHODS.session_delete, zDeleteSessionRequest, emptyObjectResponse),
  resumeSession: requestSpec(AGENT_METHODS.session_resume, zResumeSessionRequest),
  closeSession: requestSpec(AGENT_METHODS.session_close, zCloseSessionRequest, emptyObjectResponse),
  setSessionMode: requestSpec(AGENT_METHODS.session_set_mode, zSetSessionModeRequest, emptyObjectResponse),
  setSessionConfigOption: requestSpec(AGENT_METHODS.session_set_config_option, zSetSessionConfigOptionRequest),
  authenticate: requestSpec(AGENT_METHODS.authenticate, zAuthenticateRequest, emptyObjectResponse),
  logout: requestSpec(AGENT_METHODS.logout, zLogoutRequest, emptyObjectResponse),
  prompt: requestSpec(AGENT_METHODS.session_prompt, zPromptRequest)
};
var agentNotificationSpecs = {
  cancel: notificationSpec(AGENT_METHODS.session_cancel, zCancelNotification)
};
var clientRequestSpecs = {
  requestPermission: requestSpec(CLIENT_METHODS.session_request_permission, zRequestPermissionRequest),
  writeTextFile: requestSpec(CLIENT_METHODS.fs_write_text_file, zWriteTextFileRequest, emptyObjectResponse),
  readTextFile: requestSpec(CLIENT_METHODS.fs_read_text_file, zReadTextFileRequest),
  createTerminal: requestSpec(CLIENT_METHODS.terminal_create, zCreateTerminalRequest),
  terminalOutput: requestSpec(CLIENT_METHODS.terminal_output, zTerminalOutputRequest),
  releaseTerminal: requestSpec(CLIENT_METHODS.terminal_release, zReleaseTerminalRequest, emptyObjectResponse),
  waitForTerminalExit: requestSpec(CLIENT_METHODS.terminal_wait_for_exit, zWaitForTerminalExitRequest),
  killTerminal: requestSpec(CLIENT_METHODS.terminal_kill, zKillTerminalRequest, emptyObjectResponse),
  createElicitation: requestSpec(CLIENT_METHODS.elicitation_create, zCreateElicitationRequest)
};
var clientNotificationSpecs = {
  sessionUpdate: notificationSpec(CLIENT_METHODS.session_update, zSessionNotification),
  completeElicitation: notificationSpec(CLIENT_METHODS.elicitation_complete, zCompleteElicitationNotification)
};
var agentRequestSpecsByMethod = specsByMethod(agentRequestSpecs);
var agentNotificationSpecsByMethod = specsByMethod(agentNotificationSpecs);
var clientRequestSpecsByMethod = specsByMethod(clientRequestSpecs);
var clientNotificationSpecsByMethod = specsByMethod(clientNotificationSpecs);
function agentRequestContext(params, client2, signal, requestId, method) {
  return {
    method,
    params,
    requestId,
    signal,
    client: client2
  };
}
function agentNotificationContext(params, client2, signal, method) {
  return {
    method,
    params,
    signal,
    client: client2
  };
}
function clientRequestContext(params, agent, signal, requestId, method) {
  return {
    method,
    params,
    requestId,
    signal,
    agent
  };
}
function clientNotificationContext(params, agent, signal, method) {
  return {
    method,
    params,
    signal,
    agent
  };
}
var SessionUpdateRouter = class {
  activeSessions = /* @__PURE__ */ new Map();
  handleMessage(message) {
    if (message.kind !== "notification" || message.method !== CLIENT_METHODS.session_update) {
      return Handled.no(message);
    }
    const notification = zSessionNotification.parse(message.params);
    const update = {
      kind: "session_update",
      notification,
      update: notification.update
    };
    const activeSessions = this.activeSessions.get(notification.sessionId);
    if (activeSessions && activeSessions.size > 0) {
      for (const session of activeSessions) {
        session.enqueue(update);
      }
    }
    return Handled.no(message);
  }
  attach(response, updates) {
    const sessions = this.activeSessions.get(response.sessionId) ?? /* @__PURE__ */ new Set();
    sessions.add(updates);
    this.activeSessions.set(response.sessionId, sessions);
    return new HandlerRegistration(() => {
      sessions.delete(updates);
      if (sessions.size === 0) {
        this.activeSessions.delete(response.sessionId);
      }
    });
  }
};
var sessionUpdateRouters = /* @__PURE__ */ new WeakMap();
function sessionUpdateRouter(cx) {
  let router = sessionUpdateRouters.get(cx);
  if (!router) {
    router = new SessionUpdateRouter();
    sessionUpdateRouters.set(cx, router);
  }
  return router;
}
function runConnectHandlers(connection, handlers) {
  for (const handler of handlers) {
    let result;
    try {
      result = handler(connection);
    } catch (error2) {
      connection.close(error2);
      throw error2;
    }
    void Promise.resolve(result).catch((error2) => {
      connection.close(error2);
    });
  }
}
var appBuilder = Symbol("appBuilder");
var runAgentConnectHandlers = Symbol("runAgentConnectHandlers");
var runClientConnectHandlers = Symbol("runClientConnectHandlers");
var AgentApp = class {
  builder = Connection.builder();
  connectHandlers = [];
  fallbackKinds = /* @__PURE__ */ new Set();
  hasInitializeHandler = false;
  constructor(options2 = {}) {
    this.builder.withHandler(agentInitializationGate(() => this.hasInitializeHandler));
    if (options2.name) {
      this.builder.name(options2.name);
    }
  }
  /** @internal */
  [appBuilder]() {
    return this.builder;
  }
  /** @internal */
  [runAgentConnectHandlers](connection) {
    runConnectHandlers(connection, this.connectHandlers);
  }
  connect(target, options2 = {}) {
    return this.connectConnection(target, options2).connection;
  }
  /**
   * Connects this agent app to a stream or client app for the lifetime of `op`.
   *
   * The callback receives an `AgentContext` for calling client-side methods.
   * When `op` resolves or rejects, the connection is closed.
   */
  connectWith(target, op) {
    const { rawConnection, connection } = this.connectConnection(target);
    return rawConnection.runUntil(() => op(connection.client));
  }
  /**
   * Registers a handler that runs when this agent app opens a connection.
   *
   * Use this for connection-scoped work that needs to call client-side ACP
   * methods outside an inbound request handler, or to set up per-connection
   * state that the request handlers use.
   *
   * Handlers are called synchronously in registration order and are not
   * awaited. A handler that throws or rejects closes the connection. When they
   * are called depends on who opens the connection:
   *
   * - `connect(...)` and `connectWith(...)` call them before the connection
   *   handles its first inbound message, and so does the experimental
   *   `AgentProtocolRouter`'s `connect(...)` when it routes to this app. Only
   *   the work a handler does before its first `await` is guaranteed to come
   *   first.
   * - The experimental `AcpServer` calls them after the client has completed
   *   `initialize`, so messages they send cannot replace the initialize
   *   response. Request handlers, including `initialize`, can run before them.
   */
  onConnect(handler) {
    this.connectHandlers.push(handler);
    return this;
  }
  onRequest(method, handlerOrParams, handler) {
    if (method === "*") {
      if (handler)
        throw new Error("Wildcard handlers receive raw params; omit the parser.");
      if (this.fallbackKinds.has("request"))
        throw new Error("A request fallback is already registered.");
      this.fallbackKinds.add("request");
      return this.request({ method }, handlerOrParams);
    }
    if (handler) {
      return this.request({ method, params: handlerOrParams }, handler);
    }
    const spec = agentRequestSpecsByMethod[method];
    if (!spec) {
      throw new Error(`Unknown ACP request method '${method}'. Pass a params parser for custom methods.`);
    }
    if (method === AGENT_METHODS.initialize) {
      this.hasInitializeHandler = true;
    }
    return this.request(spec, handlerOrParams);
  }
  onNotification(method, handlerOrParams, handler) {
    if (method === "*") {
      if (handler)
        throw new Error("Wildcard handlers receive raw params; omit the parser.");
      if (this.fallbackKinds.has("notification"))
        throw new Error("A notification fallback is already registered.");
      this.fallbackKinds.add("notification");
      return this.notification({ method }, handlerOrParams);
    }
    if (handler) {
      return this.notification({ method, params: handlerOrParams }, handler);
    }
    const spec = agentNotificationSpecsByMethod[method];
    if (!spec) {
      throw new Error(`Unknown ACP notification method '${method}'. Pass a params parser for custom methods.`);
    }
    return this.notification(spec, handlerOrParams);
  }
  request(spec, handler) {
    registerAppRequest(this.builder, spec, (params, cx, signal, requestId, method) => agentRequestContext(params, AgentContext.create(cx, requestId), signal, requestId, method), handler, spec.mapResponse);
    return this;
  }
  notification(spec, handler) {
    registerAppNotification(this.builder, spec, (params, cx, signal, method) => agentNotificationContext(params, AgentContext.create(cx), signal, method), handler);
    return this;
  }
  connectConnection(target, options2 = {}) {
    if (isStream(target)) {
      const state3 = this.openStreamConnection(target);
      if (!options2.deferConnectHandlers) {
        this[runAgentConnectHandlers](state3.connection);
      }
      return state3;
    }
    const [thisStream, peerStream] = memoryStreamPair();
    const peerRawConnection = target[appBuilder]().connect(peerStream);
    const peerConnection = clientConnection(peerRawConnection);
    const state2 = this.openStreamConnection(thisStream);
    void state2.rawConnection.closed.then(() => peerConnection.close());
    void peerRawConnection.closed.then(() => state2.connection.close());
    try {
      target[runClientConnectHandlers](peerConnection);
      this[runAgentConnectHandlers](state2.connection);
    } catch (error2) {
      peerConnection.close(error2);
      state2.connection.close(error2);
      throw error2;
    }
    return state2;
  }
  openStreamConnection(stream) {
    const rawConnection = this.builder.connect(stream);
    return {
      rawConnection,
      connection: agentConnection(rawConnection, this.connectHandlers)
    };
  }
};
function client(options2) {
  return new ClientApp(options2);
}
var ClientApp = class {
  builder = Connection.builder();
  connectHandlers = [];
  fallbackKinds = /* @__PURE__ */ new Set();
  constructor(options2 = {}) {
    if (options2.name) {
      this.builder.name(options2.name);
    }
    this.builder.withHandler({
      handleMessage: (message, cx) => sessionUpdateRouter(cx).handleMessage(message),
      describe: () => "client-session-update-router"
    });
  }
  /** @internal */
  [appBuilder]() {
    return this.builder;
  }
  /** @internal */
  [runClientConnectHandlers](connection) {
    runConnectHandlers(connection, this.connectHandlers);
  }
  connect(target) {
    return this.connectConnection(target).connection;
  }
  /**
   * Connects this client app to a stream or agent app for the lifetime of `op`.
   *
   * The callback receives a `ClientContext` for calling agent-side methods.
   * When `op` resolves or rejects, the connection is closed.
   */
  connectWith(target, op) {
    const { rawConnection, connection } = this.connectConnection(target);
    return rawConnection.runUntil(() => op(connection.agent));
  }
  /**
   * Registers a handler that runs when this client app opens a connection.
   *
   * Use this for connection-scoped work that needs to call agent-side ACP
   * methods outside an inbound request handler.
   */
  onConnect(handler) {
    this.connectHandlers.push(handler);
    return this;
  }
  onRequest(method, handlerOrParams, handler) {
    if (method === "*") {
      if (handler)
        throw new Error("Wildcard handlers receive raw params; omit the parser.");
      if (this.fallbackKinds.has("request"))
        throw new Error("A request fallback is already registered.");
      this.fallbackKinds.add("request");
      return this.request({ method }, handlerOrParams);
    }
    if (handler) {
      return this.request({ method, params: handlerOrParams }, handler);
    }
    const spec = clientRequestSpecsByMethod[method];
    if (!spec) {
      throw new Error(`Unknown ACP request method '${method}'. Pass a params parser for custom methods.`);
    }
    return this.request(spec, handlerOrParams);
  }
  onNotification(method, handlerOrParams, handler) {
    if (method === "*") {
      if (handler)
        throw new Error("Wildcard handlers receive raw params; omit the parser.");
      if (this.fallbackKinds.has("notification"))
        throw new Error("A notification fallback is already registered.");
      this.fallbackKinds.add("notification");
      return this.notification({ method }, handlerOrParams);
    }
    if (handler) {
      return this.notification({ method, params: handlerOrParams }, handler);
    }
    const spec = clientNotificationSpecsByMethod[method];
    if (!spec) {
      throw new Error(`Unknown ACP notification method '${method}'. Pass a params parser for custom methods.`);
    }
    return this.notification(spec, handlerOrParams);
  }
  request(spec, handler) {
    registerAppRequest(this.builder, spec, (params, cx, signal, requestId, method) => clientRequestContext(params, ClientContext.create(cx, requestId), signal, requestId, method), handler, spec.mapResponse);
    return this;
  }
  notification(spec, handler) {
    registerAppNotification(this.builder, spec, (params, cx, signal, method) => clientNotificationContext(params, ClientContext.create(cx), signal, method), handler);
    return this;
  }
  connectConnection(target) {
    if (isStream(target)) {
      const state3 = this.openStreamConnection(target);
      this[runClientConnectHandlers](state3.connection);
      return state3;
    }
    const [thisStream, peerStream] = memoryStreamPair();
    const peerRawConnection = target[appBuilder]().connect(peerStream);
    const peerConnection = agentConnection(peerRawConnection);
    const state2 = this.openStreamConnection(thisStream);
    void state2.rawConnection.closed.then(() => peerConnection.close());
    void peerRawConnection.closed.then(() => state2.connection.close());
    try {
      target[runAgentConnectHandlers](peerConnection);
      this[runClientConnectHandlers](state2.connection);
    } catch (error2) {
      peerConnection.close(error2);
      state2.connection.close(error2);
      throw error2;
    }
    return state2;
  }
  openStreamConnection(stream) {
    const rawConnection = this.builder.connect(stream);
    return {
      rawConnection,
      connection: clientConnection(rawConnection, this.connectHandlers)
    };
  }
};
var TerminalHandle = class _TerminalHandle {
  /**
   * Terminal identifier returned by `terminal/create`.
   */
  id;
  sessionId;
  connection;
  constructor(id, sessionId, conn) {
    this.id = id;
    this.sessionId = sessionId;
    this.connection = conn;
  }
  /** @internal */
  static create(id, sessionId, conn) {
    return new _TerminalHandle(id, sessionId, conn);
  }
  /**
   * Gets the current terminal output without waiting for the command to exit.
   */
  currentOutput() {
    return this.connection.sendRequest(CLIENT_METHODS.terminal_output, {
      sessionId: this.sessionId,
      terminalId: this.id
    });
  }
  /**
   * Waits for the terminal command to complete and returns its exit status.
   */
  waitForExit() {
    return this.connection.sendRequest(CLIENT_METHODS.terminal_wait_for_exit, {
      sessionId: this.sessionId,
      terminalId: this.id
    });
  }
  /**
   * Kills the terminal command without releasing the terminal.
   *
   * The terminal remains valid after killing, allowing you to:
   * - Get the final output with `currentOutput()`
   * - Check the exit status
   * - Release the terminal when done
   *
   * Useful for implementing timeouts or cancellation.
   */
  kill() {
    return this.connection.sendRequest(CLIENT_METHODS.terminal_kill, {
      sessionId: this.sessionId,
      terminalId: this.id
    }, emptyObjectResponse);
  }
  /**
   * Releases the terminal and frees all associated resources.
   *
   * If the command is still running, it will be killed.
   * After release, the terminal ID becomes invalid and cannot be used
   * with other terminal methods.
   *
   * Tool calls that already reference this terminal will continue to
   * display its output.
   *
   * **Important:** Always call this method when done with the terminal.
   */
  release() {
    return this.connection.sendRequest(CLIENT_METHODS.terminal_release, {
      sessionId: this.sessionId,
      terminalId: this.id
    }, emptyObjectResponse);
  }
  /**
   * Releases the terminal when used with `await using`.
   */
  async [Symbol.asyncDispose]() {
    await this.release();
  }
};

// node_modules/@agentclientprotocol/sdk/dist/transports/cookie-store.js
var MemoryAcpCookieStore = class {
  cookies = /* @__PURE__ */ new Map();
  store(headers) {
    for (const value of setCookieHeaders(headers)) {
      const cookie = parseSetCookie(value);
      if (!cookie) {
        continue;
      }
      this.cookies.set(cookie.name, cookie.value);
    }
  }
  apply(headers) {
    const merged = mergeCookieHeaders(this.cookieHeader(), headers.get("Cookie"));
    if (merged) {
      headers.set("Cookie", merged);
    }
  }
  clear() {
    this.cookies.clear();
  }
  cookieHeader() {
    return this.cookies.size === 0 ? void 0 : Array.from(this.cookies).map(([name, value]) => `${name}=${value}`).join("; ");
  }
};
function setCookieHeaders(headers) {
  const getSetCookie = headers.getSetCookie;
  if (typeof getSetCookie === "function") {
    return getSetCookie.call(headers).flatMap(splitSetCookieHeader);
  }
  const setCookie = headers.get("Set-Cookie");
  return setCookie ? splitSetCookieHeader(setCookie) : [];
}
function splitSetCookieHeader(header) {
  return header.split(/,(?=\s*[^;,\s]+=)/).map((value) => value.trim()).filter((value) => value.length > 0);
}
function parseSetCookie(header) {
  const pair = header.split(";", 1)[0];
  const separator = pair.indexOf("=");
  if (separator <= 0) {
    return void 0;
  }
  const name = pair.slice(0, separator).trim();
  if (!name) {
    return void 0;
  }
  return {
    name,
    value: pair.slice(separator + 1).trim()
  };
}
function mergeCookieHeaders(managedCookieHeader, callerCookieHeader) {
  const cookies = /* @__PURE__ */ new Map();
  for (const cookie of parseCookieHeader(managedCookieHeader)) {
    cookies.set(cookie.name, cookie.value);
  }
  for (const cookie of parseCookieHeader(callerCookieHeader ?? void 0)) {
    cookies.set(cookie.name, cookie.value);
  }
  return cookies.size === 0 ? void 0 : Array.from(cookies).map(([name, value]) => `${name}=${value}`).join("; ");
}
function parseCookieHeader(header) {
  if (!header) {
    return [];
  }
  return header.split(";").map(parseCookiePair).filter((cookie) => cookie !== void 0);
}
function parseCookiePair(value) {
  const separator = value.indexOf("=");
  if (separator <= 0) {
    return void 0;
  }
  const name = value.slice(0, separator).trim();
  if (!name) {
    return void 0;
  }
  return {
    name,
    value: value.slice(separator + 1).trim()
  };
}

// node_modules/@agentclientprotocol/sdk/dist/transports/ws-utils.js
function onWebSocket(socket, type, listener) {
  if (socket.on) {
    const eventListener = (...args) => {
      listener(...normalizeEventEmitterMessageArgs(type, args));
    };
    socket.on(type, eventListener);
    return () => {
      if (socket.off) {
        socket.off(type, eventListener);
        return;
      }
      socket.removeListener?.(type, eventListener);
    };
  }
  if (socket.addEventListener) {
    const eventListener = (event) => listener(event);
    socket.addEventListener(type, eventListener);
    return () => {
      socket.removeEventListener?.(type, eventListener);
    };
  }
  throw new Error("WebSocket object does not support event listeners");
}
function webSocketMessageToString(args) {
  const data = extractMessageData(args);
  if (typeof data === "string") {
    return data;
  }
  return void 0;
}
function normalizeEventEmitterMessageArgs(type, args) {
  if (type !== "message" || typeof args[1] !== "boolean") {
    return args;
  }
  if (args[1]) {
    return [void 0];
  }
  return [decodeWebSocketTextData(args[0])];
}
function decodeWebSocketTextData(data) {
  if (typeof data === "string") {
    return data;
  }
  if (data instanceof ArrayBuffer) {
    return new TextDecoder().decode(data);
  }
  if (ArrayBuffer.isView(data)) {
    return new TextDecoder().decode(data);
  }
  if (isArrayBufferViewArray(data)) {
    return decodeArrayBufferViews(data);
  }
  return void 0;
}
function extractMessageData(args) {
  const [first] = args;
  if (isMessageEventLike(first)) {
    return first.data;
  }
  return first;
}
function isMessageEventLike(value) {
  return typeof value === "object" && value !== null && "data" in value;
}
function isArrayBufferViewArray(value) {
  return Array.isArray(value) && value.every(ArrayBuffer.isView);
}
function decodeArrayBufferViews(views) {
  const totalLength = views.reduce((sum, view) => sum + view.byteLength, 0);
  const combined = new Uint8Array(totalLength);
  let offset = 0;
  for (const view of views) {
    combined.set(new Uint8Array(view.buffer, view.byteOffset, view.byteLength), offset);
    offset += view.byteLength;
  }
  return new TextDecoder().decode(combined);
}

// node_modules/@agentclientprotocol/sdk/dist/transports/ws-stream.js
var SOCKET_OPEN = 1;
function createWebSocketStream(serverUrl, options2 = {}) {
  return new WebSocketStreamTransport(serverUrl, options2).stream;
}
var WebSocketStreamTransport = class {
  stream;
  socket;
  cookieStore;
  ownsCookieStore;
  readableController;
  isClosed = false;
  openPromise;
  resolveOpen;
  rejectOpen;
  detachListeners = [];
  sendQueue = Promise.resolve();
  constructor(serverUrl, options2) {
    const WebSocketCtor = resolveWebSocket(options2.WebSocket);
    const cookiePolicy = options2.cookies ?? "include";
    this.cookieStore = options2.cookieStore ?? new MemoryAcpCookieStore();
    this.ownsCookieStore = options2.cookieStore === void 0;
    this.socket = new WebSocketCtor(serverUrl, options2.protocols, {
      headers: createConstructorHeaders(options2.headers, cookiePolicy, this.cookieStore)
    });
    this.openPromise = new Promise((resolve, reject) => {
      this.resolveOpen = resolve;
      this.rejectOpen = reject;
    });
    this.openPromise.catch(() => void 0);
    this.detachListeners.push(onWebSocket(this.socket, "open", () => {
      this.resolveOpen?.();
      this.resolveOpen = void 0;
      this.rejectOpen = void 0;
      this.openPromise = void 0;
    }));
    this.detachListeners.push(onWebSocket(this.socket, "message", (...args) => {
      this.handleSocketMessage(args);
    }));
    this.detachListeners.push(onWebSocket(this.socket, "close", () => {
      this.closeReadable();
    }));
    this.detachListeners.push(onWebSocket(this.socket, "error", (error2) => {
      this.errorReadable(error2);
    }));
    if (cookiePolicy === "include") {
      this.detachListeners.push(onWebSocket(this.socket, "upgrade", (message) => {
        const headers = upgradeHeaders(message);
        if (headers) {
          this.cookieStore.store(headers);
        }
      }));
    }
    this.stream = {
      readable: new ReadableStream({
        start: (controller) => {
          this.readableController = controller;
        },
        cancel: () => {
          this.close();
        }
      }),
      writable: new WritableStream({
        write: (message) => this.queueMessage(message),
        close: () => {
          this.close();
        },
        abort: () => {
          this.close();
        }
      })
    };
  }
  queueMessage(message) {
    const send = this.sendQueue.then(() => this.sendMessage(message));
    this.sendQueue = send.catch(() => {
    });
    return send;
  }
  async sendMessage(message) {
    if (this.isClosed) {
      throw new Error("ACP WebSocket stream is closed");
    }
    await this.waitForOpen();
    if (this.isClosed) {
      throw new Error("ACP WebSocket stream is closed");
    }
    this.socket.send(JSON.stringify(message));
  }
  async waitForOpen() {
    if (this.socket.readyState === void 0 || this.socket.readyState === SOCKET_OPEN) {
      return;
    }
    await this.openPromise;
  }
  handleSocketMessage(args) {
    if (this.isClosed) {
      return;
    }
    const text = webSocketMessageToString(args);
    if (text === void 0) {
      return;
    }
    let value;
    try {
      value = JSON.parse(text);
    } catch {
      this.sendProtocolError(RequestError.parseError());
      return;
    }
    if (Array.isArray(value)) {
      this.sendProtocolError(RequestError.invalidRequest(void 0, "JSON-RPC batches are not supported"));
      return;
    }
    if (!isRecord(value)) {
      this.sendProtocolError(RequestError.invalidRequest(value));
      return;
    }
    this.readableController?.enqueue(value);
  }
  sendProtocolError(error2) {
    void this.queueMessage(protocolErrorResponse(error2)).catch((sendError) => {
      this.errorReadable(sendError);
    });
  }
  close() {
    this.closeSocket();
    this.closeReadable();
  }
  closeSocket() {
    try {
      this.socket.close();
    } catch (error2) {
      console.warn("Failed to close ACP WebSocket:", error2);
    }
  }
  clearOwnedCookieStore() {
    if (this.ownsCookieStore) {
      this.cookieStore.clear();
    }
  }
  closeReadable() {
    if (this.isClosed) {
      return;
    }
    this.isClosed = true;
    this.clearOwnedCookieStore();
    for (const detach of this.detachListeners.splice(0)) {
      detach();
    }
    this.rejectOpen?.(new Error("ACP WebSocket stream closed before open"));
    this.rejectOpen = void 0;
    this.resolveOpen = void 0;
    this.openPromise = void 0;
    try {
      this.readableController?.close();
    } catch {
    }
  }
  errorReadable(error2) {
    if (this.isClosed) {
      return;
    }
    this.isClosed = true;
    this.clearOwnedCookieStore();
    for (const detach of this.detachListeners.splice(0)) {
      detach();
    }
    this.rejectOpen?.(error2);
    this.rejectOpen = void 0;
    this.resolveOpen = void 0;
    this.openPromise = void 0;
    this.readableController?.error(error2);
  }
};
function createConstructorHeaders(headers, cookiePolicy, cookieStore) {
  const result = headers ? { ...headers } : {};
  if (cookiePolicy === "include") {
    const requestHeaders = new Headers(headers);
    cookieStore.apply(requestHeaders);
    const cookieHeader = requestHeaders.get("Cookie");
    if (cookieHeader) {
      result[findHeaderName(result, "Cookie") ?? "Cookie"] = cookieHeader;
    }
  }
  return Object.keys(result).length > 0 ? result : void 0;
}
function findHeaderName(headers, name) {
  return Object.keys(headers).find((candidate) => candidate.toLowerCase() === name.toLowerCase());
}
function upgradeHeaders(message) {
  if (message instanceof Headers) {
    return message;
  }
  if (!isRecord(message) || !("headers" in message)) {
    return void 0;
  }
  return headersFromRecord(message.headers);
}
function headersFromRecord(value) {
  if (value instanceof Headers) {
    return value;
  }
  if (!isRecord(value)) {
    return void 0;
  }
  const headers = new Headers();
  for (const [name, headerValue] of Object.entries(value)) {
    if (Array.isArray(headerValue)) {
      for (const item of headerValue) {
        headers.append(name, String(item));
      }
      continue;
    }
    if (headerValue !== void 0) {
      headers.set(name, String(headerValue));
    }
  }
  return headers;
}
function resolveWebSocket(WebSocketCtor) {
  if (WebSocketCtor) {
    return WebSocketCtor;
  }
  if (typeof globalThis.WebSocket === "function") {
    return globalThis.WebSocket;
  }
  throw new Error("createWebSocketStream requires globalThis.WebSocket or options.WebSocket");
}

// node_modules/marked/lib/marked.esm.js
function _getDefaults() {
  return {
    async: false,
    breaks: false,
    extensions: null,
    gfm: true,
    hooks: null,
    pedantic: false,
    renderer: null,
    silent: false,
    tokenizer: null,
    walkTokens: null
  };
}
var _defaults = _getDefaults();
function changeDefaults(newDefaults) {
  _defaults = newDefaults;
}
var escapeTest = /[&<>"']/;
var escapeReplace = new RegExp(escapeTest.source, "g");
var escapeTestNoEncode = /[<>"']|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/;
var escapeReplaceNoEncode = new RegExp(escapeTestNoEncode.source, "g");
var escapeReplacements = {
  "&": "&amp;",
  "<": "&lt;",
  ">": "&gt;",
  '"': "&quot;",
  "'": "&#39;"
};
var getEscapeReplacement = (ch) => escapeReplacements[ch];
function escape$1(html2, encode2) {
  if (encode2) {
    if (escapeTest.test(html2)) {
      return html2.replace(escapeReplace, getEscapeReplacement);
    }
  } else {
    if (escapeTestNoEncode.test(html2)) {
      return html2.replace(escapeReplaceNoEncode, getEscapeReplacement);
    }
  }
  return html2;
}
var unescapeTest = /&(#(?:\d+)|(?:#x[0-9A-Fa-f]+)|(?:\w+));?/ig;
function unescape(html2) {
  return html2.replace(unescapeTest, (_, n) => {
    n = n.toLowerCase();
    if (n === "colon")
      return ":";
    if (n.charAt(0) === "#") {
      return n.charAt(1) === "x" ? String.fromCharCode(parseInt(n.substring(2), 16)) : String.fromCharCode(+n.substring(1));
    }
    return "";
  });
}
var caret = /(^|[^\[])\^/g;
function edit(regex, opt) {
  let source = typeof regex === "string" ? regex : regex.source;
  opt = opt || "";
  const obj = {
    replace: (name, val) => {
      let valSource = typeof val === "string" ? val : val.source;
      valSource = valSource.replace(caret, "$1");
      source = source.replace(name, valSource);
      return obj;
    },
    getRegex: () => {
      return new RegExp(source, opt);
    }
  };
  return obj;
}
function cleanUrl(href) {
  try {
    href = encodeURI(href).replace(/%25/g, "%");
  } catch (e) {
    return null;
  }
  return href;
}
var noopTest = { exec: () => null };
function splitCells(tableRow, count) {
  const row = tableRow.replace(/\|/g, (match, offset, str) => {
    let escaped = false;
    let curr = offset;
    while (--curr >= 0 && str[curr] === "\\")
      escaped = !escaped;
    if (escaped) {
      return "|";
    } else {
      return " |";
    }
  }), cells = row.split(/ \|/);
  let i = 0;
  if (!cells[0].trim()) {
    cells.shift();
  }
  if (cells.length > 0 && !cells[cells.length - 1].trim()) {
    cells.pop();
  }
  if (count) {
    if (cells.length > count) {
      cells.splice(count);
    } else {
      while (cells.length < count)
        cells.push("");
    }
  }
  for (; i < cells.length; i++) {
    cells[i] = cells[i].trim().replace(/\\\|/g, "|");
  }
  return cells;
}
function rtrim(str, c, invert) {
  const l = str.length;
  if (l === 0) {
    return "";
  }
  let suffLen = 0;
  while (suffLen < l) {
    const currChar = str.charAt(l - suffLen - 1);
    if (currChar === c && !invert) {
      suffLen++;
    } else if (currChar !== c && invert) {
      suffLen++;
    } else {
      break;
    }
  }
  return str.slice(0, l - suffLen);
}
function findClosingBracket(str, b) {
  if (str.indexOf(b[1]) === -1) {
    return -1;
  }
  let level = 0;
  for (let i = 0; i < str.length; i++) {
    if (str[i] === "\\") {
      i++;
    } else if (str[i] === b[0]) {
      level++;
    } else if (str[i] === b[1]) {
      level--;
      if (level < 0) {
        return i;
      }
    }
  }
  return -1;
}
function outputLink(cap, link2, raw, lexer2) {
  const href = link2.href;
  const title = link2.title ? escape$1(link2.title) : null;
  const text = cap[1].replace(/\\([\[\]])/g, "$1");
  if (cap[0].charAt(0) !== "!") {
    lexer2.state.inLink = true;
    const token = {
      type: "link",
      raw,
      href,
      title,
      text,
      tokens: lexer2.inlineTokens(text)
    };
    lexer2.state.inLink = false;
    return token;
  }
  return {
    type: "image",
    raw,
    href,
    title,
    text: escape$1(text)
  };
}
function indentCodeCompensation(raw, text) {
  const matchIndentToCode = raw.match(/^(\s+)(?:```)/);
  if (matchIndentToCode === null) {
    return text;
  }
  const indentToCode = matchIndentToCode[1];
  return text.split("\n").map((node2) => {
    const matchIndentInNode = node2.match(/^\s+/);
    if (matchIndentInNode === null) {
      return node2;
    }
    const [indentInNode] = matchIndentInNode;
    if (indentInNode.length >= indentToCode.length) {
      return node2.slice(indentToCode.length);
    }
    return node2;
  }).join("\n");
}
var _Tokenizer = class {
  options;
  rules;
  // set by the lexer
  lexer;
  // set by the lexer
  constructor(options2) {
    this.options = options2 || _defaults;
  }
  space(src) {
    const cap = this.rules.block.newline.exec(src);
    if (cap && cap[0].length > 0) {
      return {
        type: "space",
        raw: cap[0]
      };
    }
  }
  code(src) {
    const cap = this.rules.block.code.exec(src);
    if (cap) {
      const text = cap[0].replace(/^ {1,4}/gm, "");
      return {
        type: "code",
        raw: cap[0],
        codeBlockStyle: "indented",
        text: !this.options.pedantic ? rtrim(text, "\n") : text
      };
    }
  }
  fences(src) {
    const cap = this.rules.block.fences.exec(src);
    if (cap) {
      const raw = cap[0];
      const text = indentCodeCompensation(raw, cap[3] || "");
      return {
        type: "code",
        raw,
        lang: cap[2] ? cap[2].trim().replace(this.rules.inline.anyPunctuation, "$1") : cap[2],
        text
      };
    }
  }
  heading(src) {
    const cap = this.rules.block.heading.exec(src);
    if (cap) {
      let text = cap[2].trim();
      if (/#$/.test(text)) {
        const trimmed = rtrim(text, "#");
        if (this.options.pedantic) {
          text = trimmed.trim();
        } else if (!trimmed || / $/.test(trimmed)) {
          text = trimmed.trim();
        }
      }
      return {
        type: "heading",
        raw: cap[0],
        depth: cap[1].length,
        text,
        tokens: this.lexer.inline(text)
      };
    }
  }
  hr(src) {
    const cap = this.rules.block.hr.exec(src);
    if (cap) {
      return {
        type: "hr",
        raw: cap[0]
      };
    }
  }
  blockquote(src) {
    const cap = this.rules.block.blockquote.exec(src);
    if (cap) {
      let text = cap[0].replace(/\n {0,3}((?:=+|-+) *)(?=\n|$)/g, "\n    $1");
      text = rtrim(text.replace(/^ *>[ \t]?/gm, ""), "\n");
      const top = this.lexer.state.top;
      this.lexer.state.top = true;
      const tokens = this.lexer.blockTokens(text);
      this.lexer.state.top = top;
      return {
        type: "blockquote",
        raw: cap[0],
        tokens,
        text
      };
    }
  }
  list(src) {
    let cap = this.rules.block.list.exec(src);
    if (cap) {
      let bull = cap[1].trim();
      const isordered = bull.length > 1;
      const list2 = {
        type: "list",
        raw: "",
        ordered: isordered,
        start: isordered ? +bull.slice(0, -1) : "",
        loose: false,
        items: []
      };
      bull = isordered ? `\\d{1,9}\\${bull.slice(-1)}` : `\\${bull}`;
      if (this.options.pedantic) {
        bull = isordered ? bull : "[*+-]";
      }
      const itemRegex = new RegExp(`^( {0,3}${bull})((?:[	 ][^\\n]*)?(?:\\n|$))`);
      let raw = "";
      let itemContents = "";
      let endsWithBlankLine = false;
      while (src) {
        let endEarly = false;
        if (!(cap = itemRegex.exec(src))) {
          break;
        }
        if (this.rules.block.hr.test(src)) {
          break;
        }
        raw = cap[0];
        src = src.substring(raw.length);
        let line = cap[2].split("\n", 1)[0].replace(/^\t+/, (t) => " ".repeat(3 * t.length));
        let nextLine = src.split("\n", 1)[0];
        let indent = 0;
        if (this.options.pedantic) {
          indent = 2;
          itemContents = line.trimStart();
        } else {
          indent = cap[2].search(/[^ ]/);
          indent = indent > 4 ? 1 : indent;
          itemContents = line.slice(indent);
          indent += cap[1].length;
        }
        let blankLine = false;
        if (!line && /^ *$/.test(nextLine)) {
          raw += nextLine + "\n";
          src = src.substring(nextLine.length + 1);
          endEarly = true;
        }
        if (!endEarly) {
          const nextBulletRegex = new RegExp(`^ {0,${Math.min(3, indent - 1)}}(?:[*+-]|\\d{1,9}[.)])((?:[ 	][^\\n]*)?(?:\\n|$))`);
          const hrRegex = new RegExp(`^ {0,${Math.min(3, indent - 1)}}((?:- *){3,}|(?:_ *){3,}|(?:\\* *){3,})(?:\\n+|$)`);
          const fencesBeginRegex = new RegExp(`^ {0,${Math.min(3, indent - 1)}}(?:\`\`\`|~~~)`);
          const headingBeginRegex = new RegExp(`^ {0,${Math.min(3, indent - 1)}}#`);
          while (src) {
            const rawLine = src.split("\n", 1)[0];
            nextLine = rawLine;
            if (this.options.pedantic) {
              nextLine = nextLine.replace(/^ {1,4}(?=( {4})*[^ ])/g, "  ");
            }
            if (fencesBeginRegex.test(nextLine)) {
              break;
            }
            if (headingBeginRegex.test(nextLine)) {
              break;
            }
            if (nextBulletRegex.test(nextLine)) {
              break;
            }
            if (hrRegex.test(src)) {
              break;
            }
            if (nextLine.search(/[^ ]/) >= indent || !nextLine.trim()) {
              itemContents += "\n" + nextLine.slice(indent);
            } else {
              if (blankLine) {
                break;
              }
              if (line.search(/[^ ]/) >= 4) {
                break;
              }
              if (fencesBeginRegex.test(line)) {
                break;
              }
              if (headingBeginRegex.test(line)) {
                break;
              }
              if (hrRegex.test(line)) {
                break;
              }
              itemContents += "\n" + nextLine;
            }
            if (!blankLine && !nextLine.trim()) {
              blankLine = true;
            }
            raw += rawLine + "\n";
            src = src.substring(rawLine.length + 1);
            line = nextLine.slice(indent);
          }
        }
        if (!list2.loose) {
          if (endsWithBlankLine) {
            list2.loose = true;
          } else if (/\n *\n *$/.test(raw)) {
            endsWithBlankLine = true;
          }
        }
        let istask = null;
        let ischecked;
        if (this.options.gfm) {
          istask = /^\[[ xX]\] /.exec(itemContents);
          if (istask) {
            ischecked = istask[0] !== "[ ] ";
            itemContents = itemContents.replace(/^\[[ xX]\] +/, "");
          }
        }
        list2.items.push({
          type: "list_item",
          raw,
          task: !!istask,
          checked: ischecked,
          loose: false,
          text: itemContents,
          tokens: []
        });
        list2.raw += raw;
      }
      list2.items[list2.items.length - 1].raw = raw.trimEnd();
      list2.items[list2.items.length - 1].text = itemContents.trimEnd();
      list2.raw = list2.raw.trimEnd();
      for (let i = 0; i < list2.items.length; i++) {
        this.lexer.state.top = false;
        list2.items[i].tokens = this.lexer.blockTokens(list2.items[i].text, []);
        if (!list2.loose) {
          const spacers = list2.items[i].tokens.filter((t) => t.type === "space");
          const hasMultipleLineBreaks = spacers.length > 0 && spacers.some((t) => /\n.*\n/.test(t.raw));
          list2.loose = hasMultipleLineBreaks;
        }
      }
      if (list2.loose) {
        for (let i = 0; i < list2.items.length; i++) {
          list2.items[i].loose = true;
        }
      }
      return list2;
    }
  }
  html(src) {
    const cap = this.rules.block.html.exec(src);
    if (cap) {
      const token = {
        type: "html",
        block: true,
        raw: cap[0],
        pre: cap[1] === "pre" || cap[1] === "script" || cap[1] === "style",
        text: cap[0]
      };
      return token;
    }
  }
  def(src) {
    const cap = this.rules.block.def.exec(src);
    if (cap) {
      const tag2 = cap[1].toLowerCase().replace(/\s+/g, " ");
      const href = cap[2] ? cap[2].replace(/^<(.*)>$/, "$1").replace(this.rules.inline.anyPunctuation, "$1") : "";
      const title = cap[3] ? cap[3].substring(1, cap[3].length - 1).replace(this.rules.inline.anyPunctuation, "$1") : cap[3];
      return {
        type: "def",
        tag: tag2,
        raw: cap[0],
        href,
        title
      };
    }
  }
  table(src) {
    const cap = this.rules.block.table.exec(src);
    if (!cap) {
      return;
    }
    if (!/[:|]/.test(cap[2])) {
      return;
    }
    const headers = splitCells(cap[1]);
    const aligns = cap[2].replace(/^\||\| *$/g, "").split("|");
    const rows = cap[3] && cap[3].trim() ? cap[3].replace(/\n[ \t]*$/, "").split("\n") : [];
    const item = {
      type: "table",
      raw: cap[0],
      header: [],
      align: [],
      rows: []
    };
    if (headers.length !== aligns.length) {
      return;
    }
    for (const align of aligns) {
      if (/^ *-+: *$/.test(align)) {
        item.align.push("right");
      } else if (/^ *:-+: *$/.test(align)) {
        item.align.push("center");
      } else if (/^ *:-+ *$/.test(align)) {
        item.align.push("left");
      } else {
        item.align.push(null);
      }
    }
    for (const header of headers) {
      item.header.push({
        text: header,
        tokens: this.lexer.inline(header)
      });
    }
    for (const row of rows) {
      item.rows.push(splitCells(row, item.header.length).map((cell) => {
        return {
          text: cell,
          tokens: this.lexer.inline(cell)
        };
      }));
    }
    return item;
  }
  lheading(src) {
    const cap = this.rules.block.lheading.exec(src);
    if (cap) {
      return {
        type: "heading",
        raw: cap[0],
        depth: cap[2].charAt(0) === "=" ? 1 : 2,
        text: cap[1],
        tokens: this.lexer.inline(cap[1])
      };
    }
  }
  paragraph(src) {
    const cap = this.rules.block.paragraph.exec(src);
    if (cap) {
      const text = cap[1].charAt(cap[1].length - 1) === "\n" ? cap[1].slice(0, -1) : cap[1];
      return {
        type: "paragraph",
        raw: cap[0],
        text,
        tokens: this.lexer.inline(text)
      };
    }
  }
  text(src) {
    const cap = this.rules.block.text.exec(src);
    if (cap) {
      return {
        type: "text",
        raw: cap[0],
        text: cap[0],
        tokens: this.lexer.inline(cap[0])
      };
    }
  }
  escape(src) {
    const cap = this.rules.inline.escape.exec(src);
    if (cap) {
      return {
        type: "escape",
        raw: cap[0],
        text: escape$1(cap[1])
      };
    }
  }
  tag(src) {
    const cap = this.rules.inline.tag.exec(src);
    if (cap) {
      if (!this.lexer.state.inLink && /^<a /i.test(cap[0])) {
        this.lexer.state.inLink = true;
      } else if (this.lexer.state.inLink && /^<\/a>/i.test(cap[0])) {
        this.lexer.state.inLink = false;
      }
      if (!this.lexer.state.inRawBlock && /^<(pre|code|kbd|script)(\s|>)/i.test(cap[0])) {
        this.lexer.state.inRawBlock = true;
      } else if (this.lexer.state.inRawBlock && /^<\/(pre|code|kbd|script)(\s|>)/i.test(cap[0])) {
        this.lexer.state.inRawBlock = false;
      }
      return {
        type: "html",
        raw: cap[0],
        inLink: this.lexer.state.inLink,
        inRawBlock: this.lexer.state.inRawBlock,
        block: false,
        text: cap[0]
      };
    }
  }
  link(src) {
    const cap = this.rules.inline.link.exec(src);
    if (cap) {
      const trimmedUrl = cap[2].trim();
      if (!this.options.pedantic && /^</.test(trimmedUrl)) {
        if (!/>$/.test(trimmedUrl)) {
          return;
        }
        const rtrimSlash = rtrim(trimmedUrl.slice(0, -1), "\\");
        if ((trimmedUrl.length - rtrimSlash.length) % 2 === 0) {
          return;
        }
      } else {
        const lastParenIndex = findClosingBracket(cap[2], "()");
        if (lastParenIndex > -1) {
          const start = cap[0].indexOf("!") === 0 ? 5 : 4;
          const linkLen = start + cap[1].length + lastParenIndex;
          cap[2] = cap[2].substring(0, lastParenIndex);
          cap[0] = cap[0].substring(0, linkLen).trim();
          cap[3] = "";
        }
      }
      let href = cap[2];
      let title = "";
      if (this.options.pedantic) {
        const link2 = /^([^'"]*[^\s])\s+(['"])(.*)\2/.exec(href);
        if (link2) {
          href = link2[1];
          title = link2[3];
        }
      } else {
        title = cap[3] ? cap[3].slice(1, -1) : "";
      }
      href = href.trim();
      if (/^</.test(href)) {
        if (this.options.pedantic && !/>$/.test(trimmedUrl)) {
          href = href.slice(1);
        } else {
          href = href.slice(1, -1);
        }
      }
      return outputLink(cap, {
        href: href ? href.replace(this.rules.inline.anyPunctuation, "$1") : href,
        title: title ? title.replace(this.rules.inline.anyPunctuation, "$1") : title
      }, cap[0], this.lexer);
    }
  }
  reflink(src, links) {
    let cap;
    if ((cap = this.rules.inline.reflink.exec(src)) || (cap = this.rules.inline.nolink.exec(src))) {
      const linkString = (cap[2] || cap[1]).replace(/\s+/g, " ");
      const link2 = links[linkString.toLowerCase()];
      if (!link2) {
        const text = cap[0].charAt(0);
        return {
          type: "text",
          raw: text,
          text
        };
      }
      return outputLink(cap, link2, cap[0], this.lexer);
    }
  }
  emStrong(src, maskedSrc, prevChar = "") {
    let match = this.rules.inline.emStrongLDelim.exec(src);
    if (!match)
      return;
    if (match[3] && prevChar.match(/[\p{L}\p{N}]/u))
      return;
    const nextChar = match[1] || match[2] || "";
    if (!nextChar || !prevChar || this.rules.inline.punctuation.exec(prevChar)) {
      const lLength = [...match[0]].length - 1;
      let rDelim, rLength, delimTotal = lLength, midDelimTotal = 0;
      const endReg = match[0][0] === "*" ? this.rules.inline.emStrongRDelimAst : this.rules.inline.emStrongRDelimUnd;
      endReg.lastIndex = 0;
      maskedSrc = maskedSrc.slice(-1 * src.length + lLength);
      while ((match = endReg.exec(maskedSrc)) != null) {
        rDelim = match[1] || match[2] || match[3] || match[4] || match[5] || match[6];
        if (!rDelim)
          continue;
        rLength = [...rDelim].length;
        if (match[3] || match[4]) {
          delimTotal += rLength;
          continue;
        } else if (match[5] || match[6]) {
          if (lLength % 3 && !((lLength + rLength) % 3)) {
            midDelimTotal += rLength;
            continue;
          }
        }
        delimTotal -= rLength;
        if (delimTotal > 0)
          continue;
        rLength = Math.min(rLength, rLength + delimTotal + midDelimTotal);
        const lastCharLength = [...match[0]][0].length;
        const raw = src.slice(0, lLength + match.index + lastCharLength + rLength);
        if (Math.min(lLength, rLength) % 2) {
          const text2 = raw.slice(1, -1);
          return {
            type: "em",
            raw,
            text: text2,
            tokens: this.lexer.inlineTokens(text2)
          };
        }
        const text = raw.slice(2, -2);
        return {
          type: "strong",
          raw,
          text,
          tokens: this.lexer.inlineTokens(text)
        };
      }
    }
  }
  codespan(src) {
    const cap = this.rules.inline.code.exec(src);
    if (cap) {
      let text = cap[2].replace(/\n/g, " ");
      const hasNonSpaceChars = /[^ ]/.test(text);
      const hasSpaceCharsOnBothEnds = /^ /.test(text) && / $/.test(text);
      if (hasNonSpaceChars && hasSpaceCharsOnBothEnds) {
        text = text.substring(1, text.length - 1);
      }
      text = escape$1(text, true);
      return {
        type: "codespan",
        raw: cap[0],
        text
      };
    }
  }
  br(src) {
    const cap = this.rules.inline.br.exec(src);
    if (cap) {
      return {
        type: "br",
        raw: cap[0]
      };
    }
  }
  del(src) {
    const cap = this.rules.inline.del.exec(src);
    if (cap) {
      return {
        type: "del",
        raw: cap[0],
        text: cap[2],
        tokens: this.lexer.inlineTokens(cap[2])
      };
    }
  }
  autolink(src) {
    const cap = this.rules.inline.autolink.exec(src);
    if (cap) {
      let text, href;
      if (cap[2] === "@") {
        text = escape$1(cap[1]);
        href = "mailto:" + text;
      } else {
        text = escape$1(cap[1]);
        href = text;
      }
      return {
        type: "link",
        raw: cap[0],
        text,
        href,
        tokens: [
          {
            type: "text",
            raw: text,
            text
          }
        ]
      };
    }
  }
  url(src) {
    let cap;
    if (cap = this.rules.inline.url.exec(src)) {
      let text, href;
      if (cap[2] === "@") {
        text = escape$1(cap[0]);
        href = "mailto:" + text;
      } else {
        let prevCapZero;
        do {
          prevCapZero = cap[0];
          cap[0] = this.rules.inline._backpedal.exec(cap[0])?.[0] ?? "";
        } while (prevCapZero !== cap[0]);
        text = escape$1(cap[0]);
        if (cap[1] === "www.") {
          href = "http://" + cap[0];
        } else {
          href = cap[0];
        }
      }
      return {
        type: "link",
        raw: cap[0],
        text,
        href,
        tokens: [
          {
            type: "text",
            raw: text,
            text
          }
        ]
      };
    }
  }
  inlineText(src) {
    const cap = this.rules.inline.text.exec(src);
    if (cap) {
      let text;
      if (this.lexer.state.inRawBlock) {
        text = cap[0];
      } else {
        text = escape$1(cap[0]);
      }
      return {
        type: "text",
        raw: cap[0],
        text
      };
    }
  }
};
var newline = /^(?: *(?:\n|$))+/;
var blockCode = /^( {4}[^\n]+(?:\n(?: *(?:\n|$))*)?)+/;
var fences = /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}\1[~`]* *(?=\n|$)|$)/;
var hr = /^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/;
var heading = /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/;
var bullet = /(?:[*+-]|\d{1,9}[.)])/;
var lheading = edit(/^(?!bull |blockCode|fences|blockquote|heading|html)((?:.|\n(?!\s*?\n|bull |blockCode|fences|blockquote|heading|html))+?)\n {0,3}(=+|-+) *(?:\n+|$)/).replace(/bull/g, bullet).replace(/blockCode/g, / {4}/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}/).replace(/html/g, / {0,3}<[^\n>]+>\n/).getRegex();
var _paragraph = /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html|table| +\n)[^\n]+)*)/;
var blockText = /^[^\n]+/;
var _blockLabel = /(?!\s*\])(?:\\.|[^\[\]\\])+/;
var def = edit(/^ {0,3}\[(label)\]: *(?:\n *)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n *)?| *\n *)(title))? *(?:\n+|$)/).replace("label", _blockLabel).replace("title", /(?:"(?:\\"?|[^"\\])*"|'[^'\n]*(?:\n[^'\n]+)*\n?'|\([^()]*\))/).getRegex();
var list = edit(/^( {0,3}bull)([ \t][^\n]+?)?(?:\n|$)/).replace(/bull/g, bullet).getRegex();
var _tag = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul";
var _comment = /<!--(?:-?>|[\s\S]*?(?:-->|$))/;
var html = edit("^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n+|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>\\n*|$)|<![A-Z][\\s\\S]*?(?:>\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>\\n*|$)|</?(tag)(?: +|\\n|/?>)[\\s\\S]*?(?:(?:\\n *)+\\n|$)|<(?!script|pre|style|textarea)([a-z][\\w-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n *)+\\n|$)|</(?!script|pre|style|textarea)[a-z][\\w-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n *)+\\n|$))", "i").replace("comment", _comment).replace("tag", _tag).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex();
var paragraph = edit(_paragraph).replace("hr", hr).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("|table", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").replace("list", " {0,3}(?:[*+-]|1[.)]) ").replace("html", "</?(?:tag)(?: +|\\n|/?>)|<(?:script|pre|style|textarea|!--)").replace("tag", _tag).getRegex();
var blockquote = edit(/^( {0,3}> ?(paragraph|[^\n]*)(?:\n|$))+/).replace("paragraph", paragraph).getRegex();
var blockNormal = {
  blockquote,
  code: blockCode,
  def,
  fences,
  heading,
  hr,
  html,
  lheading,
  list,
  newline,
  paragraph,
  table: noopTest,
  text: blockText
};
var gfmTable = edit("^ *([^\\n ].*)\\n {0,3}((?:\\| *)?:?-+:? *(?:\\| *:?-+:? *)*(?:\\| *)?)(?:\\n((?:(?! *\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)").replace("hr", hr).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("blockquote", " {0,3}>").replace("code", " {4}[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").replace("list", " {0,3}(?:[*+-]|1[.)]) ").replace("html", "</?(?:tag)(?: +|\\n|/?>)|<(?:script|pre|style|textarea|!--)").replace("tag", _tag).getRegex();
var blockGfm = {
  ...blockNormal,
  table: gfmTable,
  paragraph: edit(_paragraph).replace("hr", hr).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("table", gfmTable).replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").replace("list", " {0,3}(?:[*+-]|1[.)]) ").replace("html", "</?(?:tag)(?: +|\\n|/?>)|<(?:script|pre|style|textarea|!--)").replace("tag", _tag).getRegex()
};
var blockPedantic = {
  ...blockNormal,
  html: edit(`^ *(?:comment *(?:\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1> *(?:\\n{2,}|\\s*$)|<tag(?:"[^"]*"|'[^']*'|\\s[^'"/>\\s]*)*?/?> *(?:\\n{2,}|\\s*$))`).replace("comment", _comment).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
  def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
  heading: /^(#{1,6})(.*)(?:\n+|$)/,
  fences: noopTest,
  // fences not supported
  lheading: /^(.+?)\n {0,3}(=+|-+) *(?:\n+|$)/,
  paragraph: edit(_paragraph).replace("hr", hr).replace("heading", " *#{1,6} *[^\n]").replace("lheading", lheading).replace("|table", "").replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").replace("|tag", "").getRegex()
};
var escape = /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/;
var inlineCode = /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/;
var br = /^( {2,}|\\)\n(?!\s*$)/;
var inlineText = /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/;
var _punctuation = "\\p{P}\\p{S}";
var punctuation = edit(/^((?![*_])[\spunctuation])/, "u").replace(/punctuation/g, _punctuation).getRegex();
var blockSkip = /\[[^[\]]*?\]\([^\(\)]*?\)|`[^`]*?`|<[^<>]*?>/g;
var emStrongLDelim = edit(/^(?:\*+(?:((?!\*)[punct])|[^\s*]))|^_+(?:((?!_)[punct])|([^\s_]))/, "u").replace(/punct/g, _punctuation).getRegex();
var emStrongRDelimAst = edit("^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)[punct](\\*+)(?=[\\s]|$)|[^punct\\s](\\*+)(?!\\*)(?=[punct\\s]|$)|(?!\\*)[punct\\s](\\*+)(?=[^punct\\s])|[\\s](\\*+)(?!\\*)(?=[punct])|(?!\\*)[punct](\\*+)(?!\\*)(?=[punct])|[^punct\\s](\\*+)(?=[^punct\\s])", "gu").replace(/punct/g, _punctuation).getRegex();
var emStrongRDelimUnd = edit("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)[punct](_+)(?=[\\s]|$)|[^punct\\s](_+)(?!_)(?=[punct\\s]|$)|(?!_)[punct\\s](_+)(?=[^punct\\s])|[\\s](_+)(?!_)(?=[punct])|(?!_)[punct](_+)(?!_)(?=[punct])", "gu").replace(/punct/g, _punctuation).getRegex();
var anyPunctuation = edit(/\\([punct])/, "gu").replace(/punct/g, _punctuation).getRegex();
var autolink = edit(/^<(scheme:[^\s\x00-\x1f<>]*|email)>/).replace("scheme", /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/).replace("email", /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/).getRegex();
var _inlineComment = edit(_comment).replace("(?:-->|$)", "-->").getRegex();
var tag = edit("^comment|^</[a-zA-Z][\\w:-]*\\s*>|^<[a-zA-Z][\\w-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>").replace("comment", _inlineComment).replace("attribute", /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/).getRegex();
var _inlineLabel = /(?:\[(?:\\.|[^\[\]\\])*\]|\\.|`[^`]*`|[^\[\]\\`])*?/;
var link = edit(/^!?\[(label)\]\(\s*(href)(?:\s+(title))?\s*\)/).replace("label", _inlineLabel).replace("href", /<(?:\\.|[^\n<>\\])+>|[^\s\x00-\x1f]*/).replace("title", /"(?:\\"?|[^"\\])*"|'(?:\\'?|[^'\\])*'|\((?:\\\)?|[^)\\])*\)/).getRegex();
var reflink = edit(/^!?\[(label)\]\[(ref)\]/).replace("label", _inlineLabel).replace("ref", _blockLabel).getRegex();
var nolink = edit(/^!?\[(ref)\](?:\[\])?/).replace("ref", _blockLabel).getRegex();
var reflinkSearch = edit("reflink|nolink(?!\\()", "g").replace("reflink", reflink).replace("nolink", nolink).getRegex();
var inlineNormal = {
  _backpedal: noopTest,
  // only used for GFM url
  anyPunctuation,
  autolink,
  blockSkip,
  br,
  code: inlineCode,
  del: noopTest,
  emStrongLDelim,
  emStrongRDelimAst,
  emStrongRDelimUnd,
  escape,
  link,
  nolink,
  punctuation,
  reflink,
  reflinkSearch,
  tag,
  text: inlineText,
  url: noopTest
};
var inlinePedantic = {
  ...inlineNormal,
  link: edit(/^!?\[(label)\]\((.*?)\)/).replace("label", _inlineLabel).getRegex(),
  reflink: edit(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", _inlineLabel).getRegex()
};
var inlineGfm = {
  ...inlineNormal,
  escape: edit(escape).replace("])", "~|])").getRegex(),
  url: edit(/^((?:ftp|https?):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/, "i").replace("email", /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![-_])/).getRegex(),
  _backpedal: /(?:[^?!.,:;*_'"~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_'"~)]+(?!$))+/,
  del: /^(~~?)(?=[^\s~])([\s\S]*?[^\s~])\1(?=[^~]|$)/,
  text: /^([`~]+|[^`~])(?:(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|https?:\/\/|ftp:\/\/|www\.|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)))/
};
var inlineBreaks = {
  ...inlineGfm,
  br: edit(br).replace("{2,}", "*").getRegex(),
  text: edit(inlineGfm.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
};
var block = {
  normal: blockNormal,
  gfm: blockGfm,
  pedantic: blockPedantic
};
var inline = {
  normal: inlineNormal,
  gfm: inlineGfm,
  breaks: inlineBreaks,
  pedantic: inlinePedantic
};
var _Lexer = class __Lexer {
  tokens;
  options;
  state;
  tokenizer;
  inlineQueue;
  constructor(options2) {
    this.tokens = [];
    this.tokens.links = /* @__PURE__ */ Object.create(null);
    this.options = options2 || _defaults;
    this.options.tokenizer = this.options.tokenizer || new _Tokenizer();
    this.tokenizer = this.options.tokenizer;
    this.tokenizer.options = this.options;
    this.tokenizer.lexer = this;
    this.inlineQueue = [];
    this.state = {
      inLink: false,
      inRawBlock: false,
      top: true
    };
    const rules = {
      block: block.normal,
      inline: inline.normal
    };
    if (this.options.pedantic) {
      rules.block = block.pedantic;
      rules.inline = inline.pedantic;
    } else if (this.options.gfm) {
      rules.block = block.gfm;
      if (this.options.breaks) {
        rules.inline = inline.breaks;
      } else {
        rules.inline = inline.gfm;
      }
    }
    this.tokenizer.rules = rules;
  }
  /**
   * Expose Rules
   */
  static get rules() {
    return {
      block,
      inline
    };
  }
  /**
   * Static Lex Method
   */
  static lex(src, options2) {
    const lexer2 = new __Lexer(options2);
    return lexer2.lex(src);
  }
  /**
   * Static Lex Inline Method
   */
  static lexInline(src, options2) {
    const lexer2 = new __Lexer(options2);
    return lexer2.inlineTokens(src);
  }
  /**
   * Preprocessing
   */
  lex(src) {
    src = src.replace(/\r\n|\r/g, "\n");
    this.blockTokens(src, this.tokens);
    for (let i = 0; i < this.inlineQueue.length; i++) {
      const next = this.inlineQueue[i];
      this.inlineTokens(next.src, next.tokens);
    }
    this.inlineQueue = [];
    return this.tokens;
  }
  blockTokens(src, tokens = []) {
    if (this.options.pedantic) {
      src = src.replace(/\t/g, "    ").replace(/^ +$/gm, "");
    } else {
      src = src.replace(/^( *)(\t+)/gm, (_, leading, tabs) => {
        return leading + "    ".repeat(tabs.length);
      });
    }
    let token;
    let lastToken;
    let cutSrc;
    let lastParagraphClipped;
    while (src) {
      if (this.options.extensions && this.options.extensions.block && this.options.extensions.block.some((extTokenizer) => {
        if (token = extTokenizer.call({ lexer: this }, src, tokens)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          return true;
        }
        return false;
      })) {
        continue;
      }
      if (token = this.tokenizer.space(src)) {
        src = src.substring(token.raw.length);
        if (token.raw.length === 1 && tokens.length > 0) {
          tokens[tokens.length - 1].raw += "\n";
        } else {
          tokens.push(token);
        }
        continue;
      }
      if (token = this.tokenizer.code(src)) {
        src = src.substring(token.raw.length);
        lastToken = tokens[tokens.length - 1];
        if (lastToken && (lastToken.type === "paragraph" || lastToken.type === "text")) {
          lastToken.raw += "\n" + token.raw;
          lastToken.text += "\n" + token.text;
          this.inlineQueue[this.inlineQueue.length - 1].src = lastToken.text;
        } else {
          tokens.push(token);
        }
        continue;
      }
      if (token = this.tokenizer.fences(src)) {
        src = src.substring(token.raw.length);
        tokens.push(token);
        continue;
      }
      if (token = this.tokenizer.heading(src)) {
        src = src.substring(token.raw.length);
        tokens.push(token);
        continue;
      }
      if (token = this.tokenizer.hr(src)) {
        src = src.substring(token.raw.length);
        tokens.push(token);
        continue;
      }
      if (token = this.tokenizer.blockquote(src)) {
        src = src.substring(token.raw.length);
        tokens.push(token);
        continue;
      }
      if (token = this.tokenizer.list(src)) {
        src = src.substring(token.raw.length);
        tokens.push(token);
        continue;
      }
      if (token = this.tokenizer.html(src)) {
        src = src.substring(token.raw.length);
        tokens.push(token);
        continue;
      }
      if (token = this.tokenizer.def(src)) {
        src = src.substring(token.raw.length);
        lastToken = tokens[tokens.length - 1];
        if (lastToken && (lastToken.type === "paragraph" || lastToken.type === "text")) {
          lastToken.raw += "\n" + token.raw;
          lastToken.text += "\n" + token.raw;
          this.inlineQueue[this.inlineQueue.length - 1].src = lastToken.text;
        } else if (!this.tokens.links[token.tag]) {
          this.tokens.links[token.tag] = {
            href: token.href,
            title: token.title
          };
        }
        continue;
      }
      if (token = this.tokenizer.table(src)) {
        src = src.substring(token.raw.length);
        tokens.push(token);
        continue;
      }
      if (token = this.tokenizer.lheading(src)) {
        src = src.substring(token.raw.length);
        tokens.push(token);
        continue;
      }
      cutSrc = src;
      if (this.options.extensions && this.options.extensions.startBlock) {
        let startIndex = Infinity;
        const tempSrc = src.slice(1);
        let tempStart;
        this.options.extensions.startBlock.forEach((getStartIndex) => {
          tempStart = getStartIndex.call({ lexer: this }, tempSrc);
          if (typeof tempStart === "number" && tempStart >= 0) {
            startIndex = Math.min(startIndex, tempStart);
          }
        });
        if (startIndex < Infinity && startIndex >= 0) {
          cutSrc = src.substring(0, startIndex + 1);
        }
      }
      if (this.state.top && (token = this.tokenizer.paragraph(cutSrc))) {
        lastToken = tokens[tokens.length - 1];
        if (lastParagraphClipped && lastToken.type === "paragraph") {
          lastToken.raw += "\n" + token.raw;
          lastToken.text += "\n" + token.text;
          this.inlineQueue.pop();
          this.inlineQueue[this.inlineQueue.length - 1].src = lastToken.text;
        } else {
          tokens.push(token);
        }
        lastParagraphClipped = cutSrc.length !== src.length;
        src = src.substring(token.raw.length);
        continue;
      }
      if (token = this.tokenizer.text(src)) {
        src = src.substring(token.raw.length);
        lastToken = tokens[tokens.length - 1];
        if (lastToken && lastToken.type === "text") {
          lastToken.raw += "\n" + token.raw;
          lastToken.text += "\n" + token.text;
          this.inlineQueue.pop();
          this.inlineQueue[this.inlineQueue.length - 1].src = lastToken.text;
        } else {
          tokens.push(token);
        }
        continue;
      }
      if (src) {
        const errMsg = "Infinite loop on byte: " + src.charCodeAt(0);
        if (this.options.silent) {
          console.error(errMsg);
          break;
        } else {
          throw new Error(errMsg);
        }
      }
    }
    this.state.top = true;
    return tokens;
  }
  inline(src, tokens = []) {
    this.inlineQueue.push({ src, tokens });
    return tokens;
  }
  /**
   * Lexing/Compiling
   */
  inlineTokens(src, tokens = []) {
    let token, lastToken, cutSrc;
    let maskedSrc = src;
    let match;
    let keepPrevChar, prevChar;
    if (this.tokens.links) {
      const links = Object.keys(this.tokens.links);
      if (links.length > 0) {
        while ((match = this.tokenizer.rules.inline.reflinkSearch.exec(maskedSrc)) != null) {
          if (links.includes(match[0].slice(match[0].lastIndexOf("[") + 1, -1))) {
            maskedSrc = maskedSrc.slice(0, match.index) + "[" + "a".repeat(match[0].length - 2) + "]" + maskedSrc.slice(this.tokenizer.rules.inline.reflinkSearch.lastIndex);
          }
        }
      }
    }
    while ((match = this.tokenizer.rules.inline.blockSkip.exec(maskedSrc)) != null) {
      maskedSrc = maskedSrc.slice(0, match.index) + "[" + "a".repeat(match[0].length - 2) + "]" + maskedSrc.slice(this.tokenizer.rules.inline.blockSkip.lastIndex);
    }
    while ((match = this.tokenizer.rules.inline.anyPunctuation.exec(maskedSrc)) != null) {
      maskedSrc = maskedSrc.slice(0, match.index) + "++" + maskedSrc.slice(this.tokenizer.rules.inline.anyPunctuation.lastIndex);
    }
    while (src) {
      if (!keepPrevChar) {
        prevChar = "";
      }
      keepPrevChar = false;
      if (this.options.extensions && this.options.extensions.inline && this.options.extensions.inline.some((extTokenizer) => {
        if (token = extTokenizer.call({ lexer: this }, src, tokens)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          return true;
        }
        return false;
      })) {
        continue;
      }
      if (token = this.tokenizer.escape(src)) {
        src = src.substring(token.raw.length);
        tokens.push(token);
        continue;
      }
      if (token = this.tokenizer.tag(src)) {
        src = src.substring(token.raw.length);
        lastToken = tokens[tokens.length - 1];
        if (lastToken && token.type === "text" && lastToken.type === "text") {
          lastToken.raw += token.raw;
          lastToken.text += token.text;
        } else {
          tokens.push(token);
        }
        continue;
      }
      if (token = this.tokenizer.link(src)) {
        src = src.substring(token.raw.length);
        tokens.push(token);
        continue;
      }
      if (token = this.tokenizer.reflink(src, this.tokens.links)) {
        src = src.substring(token.raw.length);
        lastToken = tokens[tokens.length - 1];
        if (lastToken && token.type === "text" && lastToken.type === "text") {
          lastToken.raw += token.raw;
          lastToken.text += token.text;
        } else {
          tokens.push(token);
        }
        continue;
      }
      if (token = this.tokenizer.emStrong(src, maskedSrc, prevChar)) {
        src = src.substring(token.raw.length);
        tokens.push(token);
        continue;
      }
      if (token = this.tokenizer.codespan(src)) {
        src = src.substring(token.raw.length);
        tokens.push(token);
        continue;
      }
      if (token = this.tokenizer.br(src)) {
        src = src.substring(token.raw.length);
        tokens.push(token);
        continue;
      }
      if (token = this.tokenizer.del(src)) {
        src = src.substring(token.raw.length);
        tokens.push(token);
        continue;
      }
      if (token = this.tokenizer.autolink(src)) {
        src = src.substring(token.raw.length);
        tokens.push(token);
        continue;
      }
      if (!this.state.inLink && (token = this.tokenizer.url(src))) {
        src = src.substring(token.raw.length);
        tokens.push(token);
        continue;
      }
      cutSrc = src;
      if (this.options.extensions && this.options.extensions.startInline) {
        let startIndex = Infinity;
        const tempSrc = src.slice(1);
        let tempStart;
        this.options.extensions.startInline.forEach((getStartIndex) => {
          tempStart = getStartIndex.call({ lexer: this }, tempSrc);
          if (typeof tempStart === "number" && tempStart >= 0) {
            startIndex = Math.min(startIndex, tempStart);
          }
        });
        if (startIndex < Infinity && startIndex >= 0) {
          cutSrc = src.substring(0, startIndex + 1);
        }
      }
      if (token = this.tokenizer.inlineText(cutSrc)) {
        src = src.substring(token.raw.length);
        if (token.raw.slice(-1) !== "_") {
          prevChar = token.raw.slice(-1);
        }
        keepPrevChar = true;
        lastToken = tokens[tokens.length - 1];
        if (lastToken && lastToken.type === "text") {
          lastToken.raw += token.raw;
          lastToken.text += token.text;
        } else {
          tokens.push(token);
        }
        continue;
      }
      if (src) {
        const errMsg = "Infinite loop on byte: " + src.charCodeAt(0);
        if (this.options.silent) {
          console.error(errMsg);
          break;
        } else {
          throw new Error(errMsg);
        }
      }
    }
    return tokens;
  }
};
var _Renderer = class {
  options;
  constructor(options2) {
    this.options = options2 || _defaults;
  }
  code(code, infostring, escaped) {
    const lang = (infostring || "").match(/^\S*/)?.[0];
    code = code.replace(/\n$/, "") + "\n";
    if (!lang) {
      return "<pre><code>" + (escaped ? code : escape$1(code, true)) + "</code></pre>\n";
    }
    return '<pre><code class="language-' + escape$1(lang) + '">' + (escaped ? code : escape$1(code, true)) + "</code></pre>\n";
  }
  blockquote(quote) {
    return `<blockquote>
${quote}</blockquote>
`;
  }
  html(html2, block2) {
    return html2;
  }
  heading(text, level, raw) {
    return `<h${level}>${text}</h${level}>
`;
  }
  hr() {
    return "<hr>\n";
  }
  list(body, ordered, start) {
    const type = ordered ? "ol" : "ul";
    const startatt = ordered && start !== 1 ? ' start="' + start + '"' : "";
    return "<" + type + startatt + ">\n" + body + "</" + type + ">\n";
  }
  listitem(text, task, checked) {
    return `<li>${text}</li>
`;
  }
  checkbox(checked) {
    return "<input " + (checked ? 'checked="" ' : "") + 'disabled="" type="checkbox">';
  }
  paragraph(text) {
    return `<p>${text}</p>
`;
  }
  table(header, body) {
    if (body)
      body = `<tbody>${body}</tbody>`;
    return "<table>\n<thead>\n" + header + "</thead>\n" + body + "</table>\n";
  }
  tablerow(content) {
    return `<tr>
${content}</tr>
`;
  }
  tablecell(content, flags) {
    const type = flags.header ? "th" : "td";
    const tag2 = flags.align ? `<${type} align="${flags.align}">` : `<${type}>`;
    return tag2 + content + `</${type}>
`;
  }
  /**
   * span level renderer
   */
  strong(text) {
    return `<strong>${text}</strong>`;
  }
  em(text) {
    return `<em>${text}</em>`;
  }
  codespan(text) {
    return `<code>${text}</code>`;
  }
  br() {
    return "<br>";
  }
  del(text) {
    return `<del>${text}</del>`;
  }
  link(href, title, text) {
    const cleanHref = cleanUrl(href);
    if (cleanHref === null) {
      return text;
    }
    href = cleanHref;
    let out = '<a href="' + href + '"';
    if (title) {
      out += ' title="' + title + '"';
    }
    out += ">" + text + "</a>";
    return out;
  }
  image(href, title, text) {
    const cleanHref = cleanUrl(href);
    if (cleanHref === null) {
      return text;
    }
    href = cleanHref;
    let out = `<img src="${href}" alt="${text}"`;
    if (title) {
      out += ` title="${title}"`;
    }
    out += ">";
    return out;
  }
  text(text) {
    return text;
  }
};
var _TextRenderer = class {
  // no need for block level renderers
  strong(text) {
    return text;
  }
  em(text) {
    return text;
  }
  codespan(text) {
    return text;
  }
  del(text) {
    return text;
  }
  html(text) {
    return text;
  }
  text(text) {
    return text;
  }
  link(href, title, text) {
    return "" + text;
  }
  image(href, title, text) {
    return "" + text;
  }
  br() {
    return "";
  }
};
var _Parser = class __Parser {
  options;
  renderer;
  textRenderer;
  constructor(options2) {
    this.options = options2 || _defaults;
    this.options.renderer = this.options.renderer || new _Renderer();
    this.renderer = this.options.renderer;
    this.renderer.options = this.options;
    this.textRenderer = new _TextRenderer();
  }
  /**
   * Static Parse Method
   */
  static parse(tokens, options2) {
    const parser2 = new __Parser(options2);
    return parser2.parse(tokens);
  }
  /**
   * Static Parse Inline Method
   */
  static parseInline(tokens, options2) {
    const parser2 = new __Parser(options2);
    return parser2.parseInline(tokens);
  }
  /**
   * Parse Loop
   */
  parse(tokens, top = true) {
    let out = "";
    for (let i = 0; i < tokens.length; i++) {
      const token = tokens[i];
      if (this.options.extensions && this.options.extensions.renderers && this.options.extensions.renderers[token.type]) {
        const genericToken = token;
        const ret = this.options.extensions.renderers[genericToken.type].call({ parser: this }, genericToken);
        if (ret !== false || !["space", "hr", "heading", "code", "table", "blockquote", "list", "html", "paragraph", "text"].includes(genericToken.type)) {
          out += ret || "";
          continue;
        }
      }
      switch (token.type) {
        case "space": {
          continue;
        }
        case "hr": {
          out += this.renderer.hr();
          continue;
        }
        case "heading": {
          const headingToken = token;
          out += this.renderer.heading(this.parseInline(headingToken.tokens), headingToken.depth, unescape(this.parseInline(headingToken.tokens, this.textRenderer)));
          continue;
        }
        case "code": {
          const codeToken = token;
          out += this.renderer.code(codeToken.text, codeToken.lang, !!codeToken.escaped);
          continue;
        }
        case "table": {
          const tableToken = token;
          let header = "";
          let cell = "";
          for (let j = 0; j < tableToken.header.length; j++) {
            cell += this.renderer.tablecell(this.parseInline(tableToken.header[j].tokens), { header: true, align: tableToken.align[j] });
          }
          header += this.renderer.tablerow(cell);
          let body = "";
          for (let j = 0; j < tableToken.rows.length; j++) {
            const row = tableToken.rows[j];
            cell = "";
            for (let k = 0; k < row.length; k++) {
              cell += this.renderer.tablecell(this.parseInline(row[k].tokens), { header: false, align: tableToken.align[k] });
            }
            body += this.renderer.tablerow(cell);
          }
          out += this.renderer.table(header, body);
          continue;
        }
        case "blockquote": {
          const blockquoteToken = token;
          const body = this.parse(blockquoteToken.tokens);
          out += this.renderer.blockquote(body);
          continue;
        }
        case "list": {
          const listToken = token;
          const ordered = listToken.ordered;
          const start = listToken.start;
          const loose = listToken.loose;
          let body = "";
          for (let j = 0; j < listToken.items.length; j++) {
            const item = listToken.items[j];
            const checked = item.checked;
            const task = item.task;
            let itemBody = "";
            if (item.task) {
              const checkbox = this.renderer.checkbox(!!checked);
              if (loose) {
                if (item.tokens.length > 0 && item.tokens[0].type === "paragraph") {
                  item.tokens[0].text = checkbox + " " + item.tokens[0].text;
                  if (item.tokens[0].tokens && item.tokens[0].tokens.length > 0 && item.tokens[0].tokens[0].type === "text") {
                    item.tokens[0].tokens[0].text = checkbox + " " + item.tokens[0].tokens[0].text;
                  }
                } else {
                  item.tokens.unshift({
                    type: "text",
                    text: checkbox + " "
                  });
                }
              } else {
                itemBody += checkbox + " ";
              }
            }
            itemBody += this.parse(item.tokens, loose);
            body += this.renderer.listitem(itemBody, task, !!checked);
          }
          out += this.renderer.list(body, ordered, start);
          continue;
        }
        case "html": {
          const htmlToken = token;
          out += this.renderer.html(htmlToken.text, htmlToken.block);
          continue;
        }
        case "paragraph": {
          const paragraphToken = token;
          out += this.renderer.paragraph(this.parseInline(paragraphToken.tokens));
          continue;
        }
        case "text": {
          let textToken = token;
          let body = textToken.tokens ? this.parseInline(textToken.tokens) : textToken.text;
          while (i + 1 < tokens.length && tokens[i + 1].type === "text") {
            textToken = tokens[++i];
            body += "\n" + (textToken.tokens ? this.parseInline(textToken.tokens) : textToken.text);
          }
          out += top ? this.renderer.paragraph(body) : body;
          continue;
        }
        default: {
          const errMsg = 'Token with "' + token.type + '" type was not found.';
          if (this.options.silent) {
            console.error(errMsg);
            return "";
          } else {
            throw new Error(errMsg);
          }
        }
      }
    }
    return out;
  }
  /**
   * Parse Inline Tokens
   */
  parseInline(tokens, renderer) {
    renderer = renderer || this.renderer;
    let out = "";
    for (let i = 0; i < tokens.length; i++) {
      const token = tokens[i];
      if (this.options.extensions && this.options.extensions.renderers && this.options.extensions.renderers[token.type]) {
        const ret = this.options.extensions.renderers[token.type].call({ parser: this }, token);
        if (ret !== false || !["escape", "html", "link", "image", "strong", "em", "codespan", "br", "del", "text"].includes(token.type)) {
          out += ret || "";
          continue;
        }
      }
      switch (token.type) {
        case "escape": {
          const escapeToken = token;
          out += renderer.text(escapeToken.text);
          break;
        }
        case "html": {
          const tagToken = token;
          out += renderer.html(tagToken.text);
          break;
        }
        case "link": {
          const linkToken = token;
          out += renderer.link(linkToken.href, linkToken.title, this.parseInline(linkToken.tokens, renderer));
          break;
        }
        case "image": {
          const imageToken = token;
          out += renderer.image(imageToken.href, imageToken.title, imageToken.text);
          break;
        }
        case "strong": {
          const strongToken = token;
          out += renderer.strong(this.parseInline(strongToken.tokens, renderer));
          break;
        }
        case "em": {
          const emToken = token;
          out += renderer.em(this.parseInline(emToken.tokens, renderer));
          break;
        }
        case "codespan": {
          const codespanToken = token;
          out += renderer.codespan(codespanToken.text);
          break;
        }
        case "br": {
          out += renderer.br();
          break;
        }
        case "del": {
          const delToken = token;
          out += renderer.del(this.parseInline(delToken.tokens, renderer));
          break;
        }
        case "text": {
          const textToken = token;
          out += renderer.text(textToken.text);
          break;
        }
        default: {
          const errMsg = 'Token with "' + token.type + '" type was not found.';
          if (this.options.silent) {
            console.error(errMsg);
            return "";
          } else {
            throw new Error(errMsg);
          }
        }
      }
    }
    return out;
  }
};
var _Hooks = class {
  options;
  constructor(options2) {
    this.options = options2 || _defaults;
  }
  static passThroughHooks = /* @__PURE__ */ new Set([
    "preprocess",
    "postprocess",
    "processAllTokens"
  ]);
  /**
   * Process markdown before marked
   */
  preprocess(markdown) {
    return markdown;
  }
  /**
   * Process HTML after marked is finished
   */
  postprocess(html2) {
    return html2;
  }
  /**
   * Process all tokens before walk tokens
   */
  processAllTokens(tokens) {
    return tokens;
  }
};
var Marked = class {
  defaults = _getDefaults();
  options = this.setOptions;
  parse = this.#parseMarkdown(_Lexer.lex, _Parser.parse);
  parseInline = this.#parseMarkdown(_Lexer.lexInline, _Parser.parseInline);
  Parser = _Parser;
  Renderer = _Renderer;
  TextRenderer = _TextRenderer;
  Lexer = _Lexer;
  Tokenizer = _Tokenizer;
  Hooks = _Hooks;
  constructor(...args) {
    this.use(...args);
  }
  /**
   * Run callback for every token
   */
  walkTokens(tokens, callback) {
    let values = [];
    for (const token of tokens) {
      values = values.concat(callback.call(this, token));
      switch (token.type) {
        case "table": {
          const tableToken = token;
          for (const cell of tableToken.header) {
            values = values.concat(this.walkTokens(cell.tokens, callback));
          }
          for (const row of tableToken.rows) {
            for (const cell of row) {
              values = values.concat(this.walkTokens(cell.tokens, callback));
            }
          }
          break;
        }
        case "list": {
          const listToken = token;
          values = values.concat(this.walkTokens(listToken.items, callback));
          break;
        }
        default: {
          const genericToken = token;
          if (this.defaults.extensions?.childTokens?.[genericToken.type]) {
            this.defaults.extensions.childTokens[genericToken.type].forEach((childTokens) => {
              const tokens2 = genericToken[childTokens].flat(Infinity);
              values = values.concat(this.walkTokens(tokens2, callback));
            });
          } else if (genericToken.tokens) {
            values = values.concat(this.walkTokens(genericToken.tokens, callback));
          }
        }
      }
    }
    return values;
  }
  use(...args) {
    const extensions = this.defaults.extensions || { renderers: {}, childTokens: {} };
    args.forEach((pack) => {
      const opts = { ...pack };
      opts.async = this.defaults.async || opts.async || false;
      if (pack.extensions) {
        pack.extensions.forEach((ext) => {
          if (!ext.name) {
            throw new Error("extension name required");
          }
          if ("renderer" in ext) {
            const prevRenderer = extensions.renderers[ext.name];
            if (prevRenderer) {
              extensions.renderers[ext.name] = function(...args2) {
                let ret = ext.renderer.apply(this, args2);
                if (ret === false) {
                  ret = prevRenderer.apply(this, args2);
                }
                return ret;
              };
            } else {
              extensions.renderers[ext.name] = ext.renderer;
            }
          }
          if ("tokenizer" in ext) {
            if (!ext.level || ext.level !== "block" && ext.level !== "inline") {
              throw new Error("extension level must be 'block' or 'inline'");
            }
            const extLevel = extensions[ext.level];
            if (extLevel) {
              extLevel.unshift(ext.tokenizer);
            } else {
              extensions[ext.level] = [ext.tokenizer];
            }
            if (ext.start) {
              if (ext.level === "block") {
                if (extensions.startBlock) {
                  extensions.startBlock.push(ext.start);
                } else {
                  extensions.startBlock = [ext.start];
                }
              } else if (ext.level === "inline") {
                if (extensions.startInline) {
                  extensions.startInline.push(ext.start);
                } else {
                  extensions.startInline = [ext.start];
                }
              }
            }
          }
          if ("childTokens" in ext && ext.childTokens) {
            extensions.childTokens[ext.name] = ext.childTokens;
          }
        });
        opts.extensions = extensions;
      }
      if (pack.renderer) {
        const renderer = this.defaults.renderer || new _Renderer(this.defaults);
        for (const prop in pack.renderer) {
          if (!(prop in renderer)) {
            throw new Error(`renderer '${prop}' does not exist`);
          }
          if (prop === "options") {
            continue;
          }
          const rendererProp = prop;
          const rendererFunc = pack.renderer[rendererProp];
          const prevRenderer = renderer[rendererProp];
          renderer[rendererProp] = (...args2) => {
            let ret = rendererFunc.apply(renderer, args2);
            if (ret === false) {
              ret = prevRenderer.apply(renderer, args2);
            }
            return ret || "";
          };
        }
        opts.renderer = renderer;
      }
      if (pack.tokenizer) {
        const tokenizer = this.defaults.tokenizer || new _Tokenizer(this.defaults);
        for (const prop in pack.tokenizer) {
          if (!(prop in tokenizer)) {
            throw new Error(`tokenizer '${prop}' does not exist`);
          }
          if (["options", "rules", "lexer"].includes(prop)) {
            continue;
          }
          const tokenizerProp = prop;
          const tokenizerFunc = pack.tokenizer[tokenizerProp];
          const prevTokenizer = tokenizer[tokenizerProp];
          tokenizer[tokenizerProp] = (...args2) => {
            let ret = tokenizerFunc.apply(tokenizer, args2);
            if (ret === false) {
              ret = prevTokenizer.apply(tokenizer, args2);
            }
            return ret;
          };
        }
        opts.tokenizer = tokenizer;
      }
      if (pack.hooks) {
        const hooks = this.defaults.hooks || new _Hooks();
        for (const prop in pack.hooks) {
          if (!(prop in hooks)) {
            throw new Error(`hook '${prop}' does not exist`);
          }
          if (prop === "options") {
            continue;
          }
          const hooksProp = prop;
          const hooksFunc = pack.hooks[hooksProp];
          const prevHook = hooks[hooksProp];
          if (_Hooks.passThroughHooks.has(prop)) {
            hooks[hooksProp] = (arg) => {
              if (this.defaults.async) {
                return Promise.resolve(hooksFunc.call(hooks, arg)).then((ret2) => {
                  return prevHook.call(hooks, ret2);
                });
              }
              const ret = hooksFunc.call(hooks, arg);
              return prevHook.call(hooks, ret);
            };
          } else {
            hooks[hooksProp] = (...args2) => {
              let ret = hooksFunc.apply(hooks, args2);
              if (ret === false) {
                ret = prevHook.apply(hooks, args2);
              }
              return ret;
            };
          }
        }
        opts.hooks = hooks;
      }
      if (pack.walkTokens) {
        const walkTokens2 = this.defaults.walkTokens;
        const packWalktokens = pack.walkTokens;
        opts.walkTokens = function(token) {
          let values = [];
          values.push(packWalktokens.call(this, token));
          if (walkTokens2) {
            values = values.concat(walkTokens2.call(this, token));
          }
          return values;
        };
      }
      this.defaults = { ...this.defaults, ...opts };
    });
    return this;
  }
  setOptions(opt) {
    this.defaults = { ...this.defaults, ...opt };
    return this;
  }
  lexer(src, options2) {
    return _Lexer.lex(src, options2 ?? this.defaults);
  }
  parser(tokens, options2) {
    return _Parser.parse(tokens, options2 ?? this.defaults);
  }
  #parseMarkdown(lexer2, parser2) {
    return (src, options2) => {
      const origOpt = { ...options2 };
      const opt = { ...this.defaults, ...origOpt };
      if (this.defaults.async === true && origOpt.async === false) {
        if (!opt.silent) {
          console.warn("marked(): The async option was set to true by an extension. The async: false option sent to parse will be ignored.");
        }
        opt.async = true;
      }
      const throwError = this.#onError(!!opt.silent, !!opt.async);
      if (typeof src === "undefined" || src === null) {
        return throwError(new Error("marked(): input parameter is undefined or null"));
      }
      if (typeof src !== "string") {
        return throwError(new Error("marked(): input parameter is of type " + Object.prototype.toString.call(src) + ", string expected"));
      }
      if (opt.hooks) {
        opt.hooks.options = opt;
      }
      if (opt.async) {
        return Promise.resolve(opt.hooks ? opt.hooks.preprocess(src) : src).then((src2) => lexer2(src2, opt)).then((tokens) => opt.hooks ? opt.hooks.processAllTokens(tokens) : tokens).then((tokens) => opt.walkTokens ? Promise.all(this.walkTokens(tokens, opt.walkTokens)).then(() => tokens) : tokens).then((tokens) => parser2(tokens, opt)).then((html2) => opt.hooks ? opt.hooks.postprocess(html2) : html2).catch(throwError);
      }
      try {
        if (opt.hooks) {
          src = opt.hooks.preprocess(src);
        }
        let tokens = lexer2(src, opt);
        if (opt.hooks) {
          tokens = opt.hooks.processAllTokens(tokens);
        }
        if (opt.walkTokens) {
          this.walkTokens(tokens, opt.walkTokens);
        }
        let html2 = parser2(tokens, opt);
        if (opt.hooks) {
          html2 = opt.hooks.postprocess(html2);
        }
        return html2;
      } catch (e) {
        return throwError(e);
      }
    };
  }
  #onError(silent, async) {
    return (e) => {
      e.message += "\nPlease report this to https://github.com/markedjs/marked.";
      if (silent) {
        const msg = "<p>An error occurred:</p><pre>" + escape$1(e.message + "", true) + "</pre>";
        if (async) {
          return Promise.resolve(msg);
        }
        return msg;
      }
      if (async) {
        return Promise.reject(e);
      }
      throw e;
    };
  }
};
var markedInstance = new Marked();
function marked(src, opt) {
  return markedInstance.parse(src, opt);
}
marked.options = marked.setOptions = function(options2) {
  markedInstance.setOptions(options2);
  marked.defaults = markedInstance.defaults;
  changeDefaults(marked.defaults);
  return marked;
};
marked.getDefaults = _getDefaults;
marked.defaults = _defaults;
marked.use = function(...args) {
  markedInstance.use(...args);
  marked.defaults = markedInstance.defaults;
  changeDefaults(marked.defaults);
  return marked;
};
marked.walkTokens = function(tokens, callback) {
  return markedInstance.walkTokens(tokens, callback);
};
marked.parseInline = markedInstance.parseInline;
marked.Parser = _Parser;
marked.parser = _Parser.parse;
marked.Renderer = _Renderer;
marked.TextRenderer = _TextRenderer;
marked.Lexer = _Lexer;
marked.lexer = _Lexer.lex;
marked.Tokenizer = _Tokenizer;
marked.Hooks = _Hooks;
marked.parse = marked;
var options = marked.options;
var setOptions = marked.setOptions;
var use = marked.use;
var walkTokens = marked.walkTokens;
var parseInline = marked.parseInline;
var parser = _Parser.parse;
var lexer = _Lexer.lex;

// src/app.js
var $ = (id) => document.getElementById(id);
var esc2 = (s) => String(s ?? "").replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
var state = {
  ctx: null,
  session: null,
  sessionId: null,
  pumping: false,
  turnText: "",
  turnEl: null,
  tools: /* @__PURE__ */ new Map(),
  pendingPermission: null
};
function setStatus(text, ok) {
  const el = $("status");
  el.textContent = text;
  el.className = "pill " + (ok ? "ok" : ok === false ? "bad" : "");
}
function logLine(text) {
  const el = $("events");
  el.textContent += text + "\n";
  el.scrollTop = el.scrollHeight;
}
function addMsg(role, html2) {
  const wrap = document.createElement("div");
  wrap.className = "msg " + role;
  wrap.innerHTML = html2;
  $("msgs").appendChild(wrap);
  wrap.scrollIntoView({ block: "end" });
  return wrap;
}
function textContentOf(block2) {
  if (!block2) return "";
  if (typeof block2 === "string") return block2;
  if (block2.type === "text" && typeof block2.text === "string") return block2.text;
  return "";
}
function updateTextOf(update) {
  const c = update.content;
  if (Array.isArray(c)) return c.map(textContentOf).join("");
  return textContentOf(c);
}
function endStreamingTurn() {
  if (state.turnEl && state.turnText.trim() !== "") {
    state.turnEl.querySelector(".body").innerHTML = marked.parse(state.turnText);
  }
  state.turnEl = null;
  state.turnText = "";
}
function renderToolContent(container, contentArray) {
  for (const item of contentArray ?? []) {
    if (item.type === "content") {
      const t = textContentOf(item.content);
      if (t) {
        const pre = document.createElement("pre");
        pre.textContent = t.slice(0, 4e3);
        container.appendChild(pre);
      }
    } else if (item.type === "diff") {
      const details = document.createElement("details");
      details.open = true;
      const path = item.path ?? item.filePath ?? "diff";
      details.innerHTML = `<summary>diff: ${esc2(path)}</summary>`;
      const pre = document.createElement("pre");
      pre.textContent = `--- before
${item.oldText ?? ""}
+++ after
${item.newText ?? item.new_text ?? ""}`.slice(0, 6e3);
      details.appendChild(pre);
      container.appendChild(details);
    } else if (item.type === "terminal") {
      const div = document.createElement("div");
      div.className = "meta";
      div.textContent = `terminal: ${item.terminalId ?? ""}`;
      container.appendChild(div);
    } else {
      const pre = document.createElement("pre");
      pre.textContent = JSON.stringify(item).slice(0, 2e3);
      container.appendChild(pre);
    }
  }
}
function upsertToolCard(tc) {
  let card = state.tools.get(tc.toolCallId);
  if (!card) {
    card = document.createElement("div");
    card.className = "msg tool";
    card.innerHTML = `<div class="meta"></div><div class="title"></div><div class="content"></div>`;
    $("msgs").appendChild(card);
    card.scrollIntoView({ block: "end" });
    state.tools.set(tc.toolCallId, card);
  }
  if (tc.title != null) card.querySelector(".title").textContent = tc.title;
  card.querySelector(".meta").textContent = `\u{1F527} ${tc.name ?? "tool"} \xB7 ${tc.kind ?? "?"} \xB7 ${tc.status ?? "pending"}`;
  card.dataset.status = tc.status ?? "pending";
  if (tc.content) {
    const box = card.querySelector(".content");
    box.innerHTML = "";
    renderToolContent(box, tc.content);
  }
}
function renderUpdate(update) {
  switch (update.sessionUpdate) {
    case "agent_message_chunk": {
      const t = updateTextOf(update);
      if (!state.turnEl) {
        state.turnEl = addMsg("agent", `<div class="meta">star</div><div class="body"></div>`);
      }
      state.turnText += t;
      state.turnEl.querySelector(".body").textContent = state.turnText;
      break;
    }
    case "agent_thought_chunk": {
      const t = updateTextOf(update);
      if (t.trim()) logLine(`thought: ${t.slice(0, 200)}`);
      break;
    }
    case "tool_call":
    case "tool_call_update":
      upsertToolCard(update);
      break;
    case "plan": {
      const entries = update.plan?.entries ?? update.entries ?? [];
      $("plan").innerHTML = entries.map((e) => `<li><b>${esc2(e.status ?? "")}</b> ${esc2(e.content ?? e.title ?? "")}</li>`).join("") || "<li class='dim'>empty plan</li>";
      break;
    }
    case "available_commands_update": {
      const cmds = update.availableCommands ?? [];
      $("cmds").innerHTML = "";
      for (const c of cmds) {
        const b = document.createElement("button");
        b.className = "chip";
        b.textContent = "/" + c.name;
        b.title = c.description ?? "";
        b.onclick = () => {
          $("prompt").value = `/${c.name} `;
          $("prompt").focus();
        };
        $("cmds").appendChild(b);
      }
      break;
    }
    case "current_mode_update":
      if (update.currentModeId) $("modeLabel").textContent = `mode: ${update.currentModeId}`;
      break;
    default:
      logLine(`update ${update.sessionUpdate}: ` + JSON.stringify(update).slice(0, 300));
  }
}
function askPermission(toolCall, options2) {
  return new Promise((resolve) => {
    const box = $("perms");
    box.querySelector(".empty")?.remove();
    const div = document.createElement("div");
    div.className = "card perm";
    div.innerHTML = `<div><b>Permission:</b> ${esc2(toolCall.title ?? toolCall.name ?? toolCall.toolCallId)}</div><div class="dim small">${esc2(toolCall.kind ?? "")} \xB7 ${esc2(toolCall.status ?? "")}</div>`;
    const row = document.createElement("div");
    row.className = "row";
    for (const opt of options2) {
      const b = document.createElement("button");
      b.textContent = opt.name ?? opt.optionId;
      b.title = opt.optionId;
      b.onclick = () => {
        div.remove();
        resolve(opt.optionId);
      };
      row.appendChild(b);
    }
    const cancel = document.createElement("button");
    cancel.textContent = "cancel turn";
    cancel.onclick = () => {
      div.remove();
      resolve(null);
    };
    row.appendChild(cancel);
    div.appendChild(row);
    box.appendChild(div);
  });
}
async function connect() {
  if (state.ctx) return;
  setStatus("connecting\u2026");
  const stream = createWebSocketStream(`ws://${location.host}/acp`);
  const app = client({ name: "star-acp-web", title: "Star ACP Web UI" }).onRequest(methods.client.session.requestPermission, async (req) => {
    const { toolCall, options: options2 } = req.params;
    logLine(`permission: ${toolCall.title ?? toolCall.name} (${options2.map((o) => o.optionId).join("/")})`);
    const optionId = await askPermission(toolCall, options2);
    if (!optionId) return { outcome: { outcome: "cancelled" } };
    return { outcome: { outcome: "selected", optionId } };
  });
  await app.connectWith(stream, async (ctx) => {
    state.ctx = ctx;
    const initialized = await ctx.request(methods.agent.initialize, {
      protocolVersion: PROTOCOL_VERSION,
      clientCapabilities: {},
      clientInfo: { name: "star-acp-web", version: "0.1.0" }
    });
    const info = initialized.agentInfo ?? {};
    $("agent").textContent = `${info.title ?? info.name ?? "agent"} ${info.version ?? ""}`.trim();
    setStatus(`connected \xB7 ${info.name ?? ""}`, true);
    logLine(`initialize: protocol ${initialized.protocolVersion}, loadSession=${initialized.agentCapabilities?.loadSession}`);
    try {
      await ctx.request(methods.agent.authenticate, { methodId: "local" });
      logLine("authenticated (local)");
    } catch (e) {
      logLine(`authenticate skipped/failed: ${e?.message ?? e}`);
    }
    await newSession();
    await new Promise((resolve) => ctx.signal.addEventListener("abort", resolve));
    state.ctx = null;
    state.session = null;
    setStatus("disconnected", false);
  }).catch((e) => {
    state.ctx = null;
    setStatus("connection failed", false);
    alert(`ACP connection failed: ${e?.message ?? e}`);
  });
}
async function newSession() {
  if (!state.ctx) {
    alert("Connect first.");
    return;
  }
  if (state.session) {
    try {
      state.session.dispose();
    } catch {
    }
    state.tools.clear();
  }
  endStreamingTurn();
  const cwd = $("cwd").value.trim();
  if (!cwd.startsWith("/")) {
    alert("cwd must be an absolute path.");
    return;
  }
  state.session = await state.ctx.buildSession({ cwd, mcpServers: [] }).start();
  state.sessionId = state.session.sessionId;
  $("sess").textContent = state.sessionId;
  logLine(`session/new \u2192 ${state.sessionId}`);
  const modes = state.session.modes;
  renderModes(modes);
  if (!state.pumping) {
    state.pumping = true;
    pump();
  }
}
function renderModes(modes) {
  const sel = $("modeSel");
  sel.innerHTML = "";
  const list2 = modes?.availableModes ?? [];
  if (!list2.length) {
    $("modeLabel").textContent = modes?.currentModeId ? `mode: ${modes.currentModeId}` : "mode: n/a";
    sel.style.display = "none";
    return;
  }
  sel.style.display = "";
  for (const m of list2) {
    const o = document.createElement("option");
    o.value = m.id;
    o.textContent = m.name ?? m.id;
    o.title = m.description ?? "";
    sel.appendChild(o);
  }
  if (modes.currentModeId) sel.value = modes.currentModeId;
  $("modeLabel").textContent = `mode: ${sel.value}`;
}
async function pump() {
  while (state.session) {
    let msg;
    try {
      msg = await state.session.nextUpdate();
    } catch {
      break;
    }
    if (msg.kind === "stop") {
      endStreamingTurn();
      logLine(`turn stopped: ${msg.stopReason}`);
      $("send").disabled = false;
    } else {
      renderUpdate(msg.update);
    }
  }
  state.pumping = false;
}
async function sendPrompt() {
  if (!state.session) {
    alert("Create a session first.");
    return;
  }
  const text = $("prompt").value.trim();
  if (!text) return;
  $("prompt").value = "";
  addMsg("user", `<div class="meta">you</div><div class="body">${marked.parse(text)}</div>`);
  $("send").disabled = true;
  try {
    await state.session.prompt(text);
  } catch (e) {
    endStreamingTurn();
    addMsg("sys", `<div class="meta">error</div><div class="body">${esc2(e?.message ?? e)}</div>`);
    $("send").disabled = false;
  }
}
$("connect").onclick = () => connect().catch((e) => alert(e.message));
$("newSession").onclick = () => newSession().catch((e) => alert(e.message));
$("send").onclick = () => sendPrompt().catch((e) => alert(e.message));
$("prompt").addEventListener("keydown", (e) => {
  if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) sendPrompt();
});
$("cancel").onclick = async () => {
  if (!state.ctx || !state.sessionId) return;
  try {
    await state.ctx.request(methods.agent.session.cancel, { sessionId: state.sessionId });
    logLine("cancel requested");
  } catch (e) {
    alert(e.message);
  }
};
$("modeSel").onchange = async (e) => {
  if (!state.ctx || !state.sessionId) return;
  await state.ctx.request(methods.agent.session.setMode, {
    sessionId: state.sessionId,
    modeId: e.target.value
  });
  $("modeLabel").textContent = `mode: ${e.target.value}`;
};
fetch("/config.json").then((r) => r.json()).then((c) => {
  if (c.cwd) $("cwd").value = c.cwd;
}).catch(() => {
});
