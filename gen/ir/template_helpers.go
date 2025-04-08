package ir

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ogen-go/ogen/internal/naming"
	"github.com/ogen-go/ogen/jsonschema"
)

func (t *Type) EncodeFn() string {
	if t.Is(KindArray) && t.Item.EncodeFn() != "" {
		return t.Item.EncodeFn() + arraySuffix
	}
	switch t.Primitive {
	case ByteSlice:
		return "Bytes"
	case Int, Int8, Int16, Int32, Int64,
		Uint, Uint8, Uint16, Uint32, Uint64,
		Float32, Float64,
		String, Bool:
		return naming.Capitalize(t.Primitive.String())
	case UUID, Time, MAC, IP, Duration, URL:
		return naming.AfterDot(t.Primitive.String())
	default:
		return ""
	}
}

func (t *Type) IsBase64Stream() bool {
	if t == nil || !t.IsStream() {
		return false
	}
	s := t.Schema
	if s == nil || s.Type != jsonschema.String {
		return false
	}
	switch s.Format {
	case "base64", "byte":
		return true
	default:
		return false
	}
}

func (t Type) uriFormat() string {
	if s := t.Schema; s != nil {
		switch f := s.Format; f {
		case "time", "date":
			return naming.Capitalize(f)
		case "date-time":
			return "DateTime"
		case "mac":
			return "MAC"
		case "int8",
			"int16",
			"int32",
			"int64",
			"uint",
			"uint8",
			"uint16",
			"uint32",
			"uint64":
			if s.Type != jsonschema.String {
				break
			}
			return "String" + naming.Capitalize(f)
		case "unix", "unix-seconds":
			return "UnixSeconds"
		case "unix-nano":
			return "UnixNano"
		case "unix-micro":
			return "UnixMicro"
		case "unix-milli":
			return "UnixMilli"
		}
	}
	return t.EncodeFn()
}

func (t Type) ToString() string {
	encodeFn := t.uriFormat()
	if encodeFn == "" {
		panic(fmt.Sprintf("unexpected %+v", t))
	}
	return encodeFn + "ToString"
}

func (t Type) FromString() string {
	encodeFn := t.uriFormat()
	if encodeFn == "" {
		panic(fmt.Sprintf("unexpected %+v", t))
	}
	return "To" + encodeFn
}

func (t *Type) IsInteger() bool {
	switch t.Primitive {
	case Int, Int8, Int16, Int32, Int64,
		Uint, Uint8, Uint16, Uint32, Uint64:
		return true
	default:
		return false
	}
}

func (t *Type) IsFloat() bool {
	switch t.Primitive {
	case Float32, Float64:
		return true
	default:
		return false
	}
}

func (t *Type) IsStringifiedFloat() bool {
	s := t.Schema
	return t.IsFloat() &&
		s != nil &&
		s.Type == jsonschema.String &&
		(s.Format == "float32" || s.Format == "float64")
}

func (t *Type) IsNull() bool {
	return t.Primitive == Null
}

func (t *Type) IsArray() bool     { return t.Is(KindArray) }
func (t *Type) IsMap() bool       { return t.Is(KindMap) }
func (t *Type) IsPrimitive() bool { return t.Is(KindPrimitive) }
func (t *Type) IsStruct() bool    { return t.Is(KindStruct) }
func (t *Type) IsPointer() bool   { return t.Is(KindPointer) }
func (t *Type) IsEnum() bool      { return t.Is(KindEnum) }
func (t *Type) IsGeneric() bool   { return t.Is(KindGeneric) }
func (t *Type) IsAlias() bool     { return t.Is(KindAlias) }
func (t *Type) IsInterface() bool { return t.Is(KindInterface) }
func (t *Type) IsSum() bool       { return t.Is(KindSum) }
func (t *Type) IsAny() bool       { return t.Is(KindAny) }
func (t *Type) IsStream() bool    { return t.Is(KindStream) }
func (t *Type) IsNumeric() bool   { return t.IsInteger() || t.IsFloat() }

func (t *Type) MustField(name string) *Field {
	if t.IsAlias() {
		return t.AliasTo.MustField(name)
	}

	if !t.Is(KindStruct) {
		panic(unreachable(t))
	}

	for _, f := range t.Fields {
		if f.Name == name {
			return f
		}
	}

	panic(fmt.Sprintf("field with name %q not found", name))
}

func (t *Type) SetFieldType(name string, newT *Type) {
	if !t.Is(KindStruct) {
		panic(unreachable(t))
	}

	for _, f := range t.Fields {
		if f.Name == name {
			f.Type = newT
			return
		}
	}

	panic(fmt.Sprintf("field with name %q not found", name))
}

// TypeDiscriminatorCase is a helper struct for describing type discriminator case.
type TypeDiscriminatorCase struct {
	// JXTypes is jx.Type values list.
	JXTypes string
	// Type is the type to be used for this case.
	Type *Type
	// IntType is the type to be used for this case when the type discriminator should distinguish
	// between integer and float types.
	IntType *Type
}

func (t *Type) TypeDiscriminator() (r []TypeDiscriminatorCase) {
	if !t.Is(KindSum) || !t.SumSpec.TypeDiscriminator {
		panic(unreachable(t))
	}

	var (
		numberType *Type
		intType    *Type
	)
	for _, v := range t.SumOf {
		typ := v.JSON().Type()
		if typ != "Number" {
			if typ == "" {
				typ = v.JSON().SumTypes()
			} else {
				typ = "jx." + typ
			}
			r = append(r, TypeDiscriminatorCase{
				JXTypes: typ,
				Type:    v,
			})
			continue
		}
		if s := v.Schema; s != nil && s.Type == jsonschema.Integer {
			intType = v
			continue
		}
		numberType = v
	}

	if intType != nil || numberType != nil {
		cse := TypeDiscriminatorCase{
			JXTypes: "jx.Number",
			Type:    numberType,
			IntType: intType,
		}
		if numberType == nil {
			cse.Type = intType
			cse.IntType = nil
		}
		r = append(r, cse)
	}
	slices.SortStableFunc(r, func(a, b TypeDiscriminatorCase) int {
		return strings.Compare(a.JXTypes, b.JXTypes)
	})
	return r
}

// UniqueObjPropertyNamesDescriminatorCases is a helper struct for describing unique object property name discriminator case.
type UniqueObjPropertyNamesDescriminatorCases struct {
	Default *UniqueObjPropertyNamesDescriminatorCase
	Cases   []UniqueObjPropertyNamesDescriminatorCase
}

type UniqueObjPropertyNamesDescriminatorCase struct {
	Type         *Type
	NumPropNames int
	CaseStmtBody string
}

func (t *Type) UniqueObjPropertyNamesDescriminator() (r UniqueObjPropertyNamesDescriminatorCases) {
	if !t.Is(KindSum) || !t.SumSpec.UniqueObjPropertyNamesDescriminator {
		panic(unreachable(t))
	}

	propNamesMap := make(map[string]*int)

	for _, s := range t.SumOf {
		typ := s.JSON().Type()
		if typ != "Object" {
			panic(unreachable(s))
		}

		for _, p := range s.Schema.Properties {
			if n, ok := propNamesMap[p.Name]; ok {
				*n = *n + 1
			} else {
				i := 0
				propNamesMap[p.Name] = &i
			}
		}
	}

	for _, s := range t.SumOf {
		quotedPropNames := make([]string, 0, len(s.Schema.Properties))
		for _, name := range s.Schema.Required {
			if n, ok := propNamesMap[name]; ok && *n == 0 {
				quotedPropNames = append(quotedPropNames, fmt.Sprintf("%q", name))
			}
		}

		if len(quotedPropNames) == 0 {
			if r.Default != nil {
				panic(unreachable(t))
			}
			r.Default = &UniqueObjPropertyNamesDescriminatorCase{
				Type: s,
			}
		} else {
			r.Cases = append(r.Cases, UniqueObjPropertyNamesDescriminatorCase{
				Type:         s,
				NumPropNames: len(quotedPropNames),
				CaseStmtBody: strings.Join(quotedPropNames, ", "),
			})
		}
	}

	slices.SortStableFunc(r.Cases, func(a, b UniqueObjPropertyNamesDescriminatorCase) int {
		return b.NumPropNames - a.NumPropNames
	})
	return r
}

// DoPassByPointer returns true if type should be passed by pointer.
func (t *Type) DoPassByPointer() bool {
	switch t.Kind {
	case KindStruct:
		return true
	case KindAlias:
		return t.AliasTo.DoPassByPointer()
	default:
		return false
	}
}

// ReadOnlyReceiver returns the receiver type for read-only methods.
func (t *Type) ReadOnlyReceiver() string {
	if t.DoPassByPointer() {
		return "*" + t.Name
	}
	return t.Name
}
