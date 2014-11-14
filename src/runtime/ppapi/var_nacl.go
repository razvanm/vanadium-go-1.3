// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppapi

import (
	"errors"
	"fmt"
	"unsafe"
)

// Var represents a NaCl value of some kind.  The values include numbers,
// string, objects, arrays, dictionaries, objects, and other resources.
type Var pp_Var // Keep the raw pp_Var representation.

var (
	errVarNotBool           = errors.New("Var is not a bool")
	errVarNotInt            = errors.New("Var is not an int")
	errVarNotDouble         = errors.New("Var is not a double")
	errVarNotString         = errors.New("Var is not a string")
	errVarNotArrayBuffer    = errors.New("Var is not an array buffer")
	errVarNotResource       = errors.New("Var is not a resource")
	errFailedArrayBufferLen = errors.New("Failed to retrieve array buffer length")
)

var (
	VarUndefined = varFromTypeAndValue(PP_VARTYPE_UNDEFINED, 0)
	VarNull      = varFromTypeAndValue(PP_VARTYPE_NULL, 0)
)

func makeVar(vin pp_Var) Var {
	return Var(vin)
}

func varFromTypeAndValue(t VarType, v int32) Var {
	var pv pp_Var
	*(*int32)(unsafe.Pointer(&pv[0])) = int32(t)
	*(*int32)(unsafe.Pointer(&pv[8])) = v
	return Var(pv)
}

// VarFromInt creates a variable from an int32.
func VarFromInt(i int32) Var {
	return varFromTypeAndValue(PP_VARTYPE_INT32, i)
}

// VarFromString creates a variable containing a string.
func VarFromString(s string) Var {
	var v pp_Var
	b := []byte(s)
	// TODO(bprosnitz) This doesn't work with the empty string.
	ppb_var_from_utf8(&v, &b[0], uint32(len(s)))
	return Var(v)
}

// VarFromByteSlice creates an arraybuffer variable corresponding to a byte slice.
func VarFromByteSlice(b []byte) Var {
	var pv pp_Var
	ppb_vararraybuffer_create(&pv, uint32(len(b)))
	rawBytes := ppb_vararraybuffer_map(pv)
	ptrval := uintptr(unsafe.Pointer(rawBytes))
	for i, val := range b {
		*(*byte)(unsafe.Pointer(ptrval + uintptr(i))) = val
	}
	ppb_vararraybuffer_unmap(pv)
	return Var(pv)
}

func NewDictVar() Var {
	var pv pp_Var
	ppb_vardictionary_create(&pv)
	return Var(pv)
}

func (v *Var) fromPP(in pp_Var) {
	*v = Var(in)
}

func (v Var) toPP(out *pp_Var) {
	*out = pp_Var(v)
	return
}

func (v Var) toPPVar() (out pp_Var) {
	v.toPP(&out)
	return
}

// AddRef increments the Var's reference count.
func (v Var) AddRef() {
	ppb_var_add_ref(pp_Var(v))
}

// Release decrements the Var's reference count.
func (v Var) Release() {
	ppb_var_release(pp_Var(v))
}

// Type returns the value's type.
func (v Var) Type() VarType {
	return VarType(*(*int32)(unsafe.Pointer(&v[0])))
}

// IsNull returns true iff the Var is NULL.
func (v Var) IsNull() bool {
	return v.Type() == PP_VARTYPE_NULL
}

// IsBool returns true iff the Var is a Boolean value.
func (v Var) IsBool() bool {
	return v.Type() == PP_VARTYPE_BOOL
}

// IsInt returns true iff the Var is of type int.
func (v Var) IsInt() bool {
	return v.Type() == PP_VARTYPE_INT32
}

// IsDouble returns true iff the Var is of type float64.
func (v Var) IsDouble() bool {
	return v.Type() == PP_VARTYPE_DOUBLE
}

// IsString returns true iff the Var is a string.
func (v Var) IsString() bool {
	return v.Type() == PP_VARTYPE_STRING
}

// IsObject returns true iff the Var is an object reference.
func (v Var) IsObject() bool {
	return v.Type() == PP_VARTYPE_OBJECT
}

// IsArray returns true iff the var is an array.
func (v Var) IsArray() bool {
	return v.Type() == PP_VARTYPE_ARRAY
}

// IsDictionary returns true iff the var is a dictionary.
func (v Var) IsDictionary() bool {
	return v.Type() == PP_VARTYPE_DICTIONARY
}

// IsResource returns true iff the var is a Resource.
func (v Var) IsResource() bool {
	return v.Type() == PP_VARTYPE_RESOURCE
}

// IsArrayBuffer returns true iff the var is an array buffer.
func (v Var) IsArrayBuffer() bool {
	return v.Type() == PP_VARTYPE_ARRAY_BUFFER
}

// AsBool returns the Boolean value stored in the Var.  Fails if the Var is not
// a Boolean value.
func (v Var) AsBool() (bool, error) {
	if v.IsBool() {
		i := *(*int32)(unsafe.Pointer(&v[8]))
		return i != 0, nil
	}
	return false, errVarNotBool
}

// AsInt returns the int stored in the Var, or an error if the Var is not an int.
func (v Var) AsInt() (int32, error) {
	if v.IsInt() {
		i := *(*int32)(unsafe.Pointer(&v[8]))
		return i, nil
	}
	return 0, errVarNotInt
}

// AsDouble returns the double stored in the Var, or an error if the Var is not a double.
func (v Var) AsDouble() (float64, error) {
	if v.IsDouble() {
		x := *(*float64)(unsafe.Pointer(&v[8]))
		return x, nil
	}
	return 0, errVarNotDouble
}

// AsString returns the string stored in the Var, or an error if the Var is not a string.
func (v Var) AsString() (string, error) {
	if v.IsString() {
		var len uint32
		b := ppb_var_to_utf8(pp_Var(v), &len)
		if b == nil {
			return "", errVarNotString
		}
		s := gostringn(b, int(len))
		return s, nil
	}
	return "", errVarNotString
}

// AsByteSlice returns the bytes stored in the Var, or an error if the Var is not an array buffer.
func (v Var) AsByteSlice() ([]byte, error) {
	if v.IsArrayBuffer() {
		var length uint32
		if r := ppb_vararraybuffer_byte_length(pp_Var(v), &length); r != ppTrue {
			return nil, errFailedArrayBufferLen
		}
		ret := make([]byte, length)
		rawBytes := ppb_vararraybuffer_map(pp_Var(v))
		ptrval := uintptr(unsafe.Pointer(rawBytes))
		for i, _ := range ret {
			ret[i] = *(*byte)(unsafe.Pointer(ptrval + uintptr(i)))
		}
		ppb_vararraybuffer_unmap(pp_Var(v))
		return ret, nil
	}
	return nil, errVarNotArrayBuffer
}

// AsResource returns the Resource stored in the Var, or an error if the Var is not a Resource.
func (v Var) AsResource() (Resource, error) {
	if v.IsResource() {
		r := makeResource(pp_Resource(*(*int32)(unsafe.Pointer(&v[8]))))
		return r, nil
	}
	return Resource{}, errVarNotResource
}

// GetKeys returns a slice of the dictionary keys.
func (v Var) GetKeys() ([]string, error) {
	var array pp_Var
	ppb_vardictionary_get_keys(&array, pp_Var(v))
	length := ppb_vararray_get_length(array)
	ret := make([]string, length)
	for i := uint32(0); i < length; i++ {
		var val pp_Var
		ppb_vararray_get(&val, array, i)
		var err error
		ret[i], err = Var(val).AsString()
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (v Var) DictionarySet(key string, val Var) error {
	keyVar := VarFromString(key)
	defer keyVar.Release()
	result := ppb_vardictionary_set(pp_Var(v), pp_Var(keyVar), pp_Var(val))
	if result != ppTrue {
		return fmt.Errorf("Error setting dictionary value")
	}
	return nil
}

// LookupKey looks up a key if var is a dictionary.
func (v Var) LookupKey(key string) (Var, error) {
	keyVar := VarFromString(key)
	defer keyVar.Release()
	hasKey := ppb_vardictionary_has_key(pp_Var(v), pp_Var(keyVar))
	if hasKey != ppTrue {
		return Var{}, fmt.Errorf("Key %q not found", key)
	}
	var pval pp_Var
	ppb_vardictionary_get(&pval, pp_Var(v), pp_Var(keyVar))
	return Var(pval), nil
}

// LookupKey looks up a key if var is a dictionary and converts it to a string.
func (v Var) LookupStringValuedKey(key string) (string, error) {
	val, err := v.LookupKey(key)
	if err != nil {
		return "", err
	}
	str, err := val.AsString()
	if err != nil {
		val.Release()
		return "", err
	}
	val.Release()
	return str, nil
}

// LookupKey looks up a key if var is a dictionary and converts it to an int.
func (v Var) LookupIntValuedKey(key string) (int, error) {
	val, err := v.LookupKey(key)
	if err != nil {
		return 0, err
	}
	i, err := val.AsInt()
	if err != nil {
		val.Release()
		return 0, err
	}
	val.Release()
	return int(i), nil
}
