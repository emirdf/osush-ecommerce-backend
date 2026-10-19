package log

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap/zapcore"
)

const (
	DefaultErrorFieldName = "error"
)

type FieldType int

const (
	FieldTypeNil FieldType = iota
	FieldTypeString
	FieldTypeBinary
	FieldTypeBoolean
	FieldTypeSigned
	FieldTypeUnsigned
	FieldTypeFloat
	FieldTypeTime
	FieldTypeDuration
	FieldTypeError
	FieldTypeArray
	FieldTypeAny
	FieldTypeReflect
	FieldTypeByteString
	FieldTypeContext
	FieldTypeLazyCall
	FieldTypeStringer
)

type Field struct {
	key      string
	ftype    FieldType
	string   string
	signed   int64
	unsigned uint64
	float    float64
	iface    any
}

func (f Field) Key() string {
	return f.key
}

func (f Field) Type() FieldType {
	return f.ftype
}

func (f Field) String() string {
	return f.string
}

func (f Field) Binary() []byte {
	if f.iface == nil {
		return nil
	}
	return f.iface.([]byte)
}

func (f Field) Bool() bool {
	return f.Signed() != 0
}

func (f Field) Signed() int64 {
	return f.signed
}

func (f Field) Unsigned() uint64 {
	return f.unsigned
}

func (f Field) Float() float64 {
	return f.float
}

func (f Field) Time() time.Time {
	return time.Unix(0, f.signed)
}

func (f Field) Duration() time.Duration {
	return time.Nanosecond * time.Duration(f.signed)
}

func (f Field) Error() error {
	if f.iface == nil {
		return nil
	}
	return f.iface.(error)
}

func (f Field) Interface() interface{} {
	return f.iface
}

func (f Field) Any() interface{} {
	switch f.Type() {
	case FieldTypeNil:
		return nil
	case FieldTypeString:
		return f.String()
	case FieldTypeBinary:
		return f.Interface()
	case FieldTypeBoolean:
		return f.Bool()
	case FieldTypeSigned:
		return f.Signed()
	case FieldTypeUnsigned:
		return f.Unsigned()
	case FieldTypeFloat:
		return f.Float()
	case FieldTypeTime:
		return f.Time()
	case FieldTypeDuration:
		return f.Duration()
	case FieldTypeError:
		return f.Error()
	case FieldTypeArray:
		return f.Interface()
	case FieldTypeAny:
		return f.Interface()
	case FieldTypeReflect:
		return f.Interface()
	case FieldTypeByteString:
		return f.Interface()
	case FieldTypeContext:
		return f.Interface()
	case FieldTypeLazyCall:
		return f.Interface()
	case FieldTypeStringer:
		return f.Interface()
	default:
		panic(fmt.Sprintf("unknown field type: %d", f.Type()))
	}
}

func Nil(key string) Field {
	return Field{key: key, ftype: FieldTypeNil}
}

func String(key, value string) Field {
	return Field{key: key, ftype: FieldTypeString, string: value}
}

func Stringer(key string, value fmt.Stringer) Field {
	return Field{key: key, ftype: FieldTypeStringer, iface: value}
}

func Sprintf(key, format string, args ...interface{}) Field {
	return Field{key: key, ftype: FieldTypeString, string: fmt.Sprintf(format, args...)}
}

func Strings(key string, value []string) Field {
	return Array(key, value)
}

func Binary(key string, value []byte) Field {
	return Field{key: key, ftype: FieldTypeBinary, iface: value}
}

func Bool(key string, value bool) Field {
	field := Field{key: key, ftype: FieldTypeBoolean}
	if value {
		field.signed = 1
	} else {
		field.signed = 0
	}

	return field
}

func Bools(key string, value []bool) Field {
	return Array(key, value)
}

func Int(key string, value int) Field {
	return Int64(key, int64(value))
}

func Ints(key string, value []int) Field {
	return Array(key, value)
}

func Int8(key string, value int8) Field {
	return Int64(key, int64(value))
}

func Int8s(key string, value []int8) Field {
	return Array(key, value)
}

func Int16(key string, value int16) Field {
	return Int64(key, int64(value))
}

func Int16s(key string, value []int16) Field {
	return Array(key, value)
}

func Int32(key string, value int32) Field {
	return Int64(key, int64(value))
}

func Int32s(key string, value []int32) Field {
	return Array(key, value)
}

func Int64(key string, value int64) Field {
	return Field{key: key, ftype: FieldTypeSigned, signed: value}
}

func Int64s(key string, value []int64) Field {
	return Array(key, value)
}

func UInt(key string, value uint) Field {
	return UInt64(key, uint64(value))
}

func UInts(key string, value []uint) Field {
	return Array(key, value)
}

func UInt8(key string, value uint8) Field {
	return UInt64(key, uint64(value))
}

func UInt8s(key string, value []uint8) Field {
	return Array(key, value)
}

func UInt16(key string, value uint16) Field {
	return UInt64(key, uint64(value))
}

func UInt16s(key string, value []uint16) Field {
	return Array(key, value)
}

func UInt32(key string, value uint32) Field {
	return UInt64(key, uint64(value))
}

func UInt32s(key string, value []uint32) Field {
	return Array(key, value)
}

func UInt64(key string, value uint64) Field {
	return Field{key: key, ftype: FieldTypeUnsigned, unsigned: value}
}

func UInt64s(key string, value []uint64) Field {
	return Array(key, value)
}

func Float32(key string, value float32) Field {
	return Float64(key, float64(value))
}

func Float32s(key string, value []float32) Field {
	return Array(key, value)
}

func Float64(key string, value float64) Field {
	return Field{key: key, ftype: FieldTypeFloat, float: value}
}

func Float64s(key string, value []float64) Field {
	return Array(key, value)
}

func Time(key string, value time.Time) Field {
	return Field{key: key, ftype: FieldTypeTime, signed: value.UnixNano()}
}

func Times(key string, value []time.Time) Field {
	return Array(key, value)
}

func Duration(key string, value time.Duration) Field {
	return Field{key: key, ftype: FieldTypeDuration, signed: value.Nanoseconds()}
}

func Durations(key string, value []time.Duration) Field {
	return Array(key, value)
}

func NamedError(key string, value error) Field {
	return Field{key: key, ftype: FieldTypeError, iface: value}
}

func Error(value error) Field {
	return NamedError(DefaultErrorFieldName, value)
}

func Errors(key string, value []error) Field {
	return Array(key, value)
}

func Array(key string, value interface{}) Field {
	return Field{key: key, ftype: FieldTypeArray, iface: value}
}

func Reflect(key string, value interface{}) Field {
	return Field{key: key, ftype: FieldTypeReflect, iface: value}
}

func ByteString(key string, value []byte) Field {
	return Field{key: key, ftype: FieldTypeByteString, iface: value}
}

func Context(ctx context.Context) Field {
	return Field{ftype: FieldTypeContext, iface: ctx}
}

type LazyEvaluator interface {
	func() (any, error) | zapcore.ObjectMarshalerFunc
}

func Lazy[T LazyEvaluator](key string, fn T) Field {
	return Field{key: key, ftype: FieldTypeLazyCall, iface: fn}
}

// Any tries to deduce interface{} underlying type and constructs Field from it.
// Use of this function is ok only for the sole purpose of not repeating its entire code
// or parts of it in user's code (when you need to log interface{} types with unknown content).
// Otherwise please use specialized functions.
// nolint: gocyclo
func Any(key string, value any) Field {
	switch val := value.(type) {
	case bool:
		return Bool(key, val)
	case float64:
		return Float64(key, val)
	case float32:
		return Float32(key, val)
	case int:
		return Int(key, val)
	case []int:
		return Ints(key, val)
	case int64:
		return Int64(key, val)
	case []int64:
		return Int64s(key, val)
	case int32:
		return Int32(key, val)
	case []int32:
		return Int32s(key, val)
	case int16:
		return Int16(key, val)
	case []int16:
		return Int16s(key, val)
	case int8:
		return Int8(key, val)
	case []int8:
		return Int8s(key, val)
	case string:
		return String(key, val)
	case []string:
		return Strings(key, val)
	case uint:
		return UInt(key, val)
	case []uint:
		return UInts(key, val)
	case uint64:
		return UInt64(key, val)
	case []uint64:
		return UInt64s(key, val)
	case uint32:
		return UInt32(key, val)
	case []uint32:
		return UInt32s(key, val)
	case uint16:
		return UInt16(key, val)
	case []uint16:
		return UInt16s(key, val)
	case uint8:
		return UInt8(key, val)
	case []byte:
		return Binary(key, val)
	case time.Time:
		return Time(key, val)
	case []time.Time:
		return Times(key, val)
	case time.Duration:
		return Duration(key, val)
	case []time.Duration:
		return Durations(key, val)
	case error:
		return NamedError(key, val)
	case []error:
		return Errors(key, val)
	case context.Context:
		return Context(val)
	case func() (any, error):
		return Lazy(key, val)
	case zapcore.ObjectMarshalerFunc:
		return Lazy(key, val)
	default:
		return Field{key: key, ftype: FieldTypeAny, iface: value}
	}
}
