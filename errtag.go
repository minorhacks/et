package et

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// private is an unexported type that is used in interface methods as a
// parameter type to ensure that only types in this package can implement said
// interface.
type private struct{}

var reExtractNamespace = regexp.MustCompile(`^errtag\.Inside\[.+\.(.+)\]$`)

// Namespace is a type that, when embedded into another type, denotes a custom
// error tag namespace.
//
// These custom namespace types can be exported (and other packages can add
// errors to the namespace) or unexported (so other packages cannot extend
// them).
type Namespace struct{}

// isNamespace is an interface/constraint that only Namespace and types that
// embed Namespace implement.
type isNamespace interface {
	embedsNamespace(private)
}

func (n Namespace) embedsNamespace(_ private) {}

type Tagged interface {
	Tag() string
}

// isTagged is an interface/constraint that only Inside[N] and types that embed
// Inside[N] implement.
type isTagged interface {
	error
	Tagged
	embedsInside(private)
	setNamespace(string)
	setName(string)
	getNamespace() string
	getName() string
	setError(error)
}

// Inside[N] is a type that, when embedded into another type, denotes a custom
// error tag within namespace N.
type Inside[N isNamespace] struct {
	namespace string
	name      string
	inner     error
}

// embedsInside implements isTagged for Inside[N].
func (e *Inside[N]) embedsInside(_ private) {}

func (e *Inside[N]) setNamespace(ns string) {
	e.namespace = ns
}

func (e *Inside[N]) setName(name string) {
	e.name = name
}

func (e *Inside[N]) setError(err error) {
	e.inner = err
}

func (e *Inside[N]) getNamespace() string {
	return e.namespace
}

func (e *Inside[N]) getName() string {
	return e.name
}

func (e *Inside[N]) Error() string {
	return fmt.Sprintf("%s::%s: %v", e.namespace, e.name, e.inner)
}

// As will fill a supplied *Inside[N] value if it is found in the error chain of
// e.
//
// *Inside[N] values supplied to `errors.As()` can be constructed with
// AsKind[N]().
func (e *Inside[N]) As(target any) bool {
	if reflect.TypeOf(target).Elem() == reflect.TypeOf(e) {
		targetPtr := target.(**Inside[N])
		*targetPtr = e
		return true
	}
	return errors.As(e.inner, target)
}

// Is() will return true if:
//
//   - comparator returned by OfKind[N] has a namespace N matching the namespace
//     of this error
//   - comparator returned by OfType[E] has an error tag E matching the namespace
//     and tag of this error
//   - error matches namespace, tag, and underlying error of this error
func (e *Inside[N]) Is(target error) bool {
	// Check for a matching namespace-only comparison
	if namespace, ok := target.(*namespaceCompare); ok {
		if e.namespace == namespace.name {
			return true
		}
		return errors.Is(e.inner, target)
	}
	// Check for a matching namespace-and-name comparison
	if errType, ok := target.(*typeCompare); ok {
		if errType.namespace == e.namespace && errType.name == e.name {
			return true
		}
		return errors.Is(e.inner, target)
	}
	return errors.Is(e.inner, target)
}

// Tag returns the fully-qualified tag string for this error.
func (e *Inside[N]) Tag() string {
	return fmt.Sprintf("%s::%s", e.namespace, e.name)
}

// insideValueFromError will introspect an error value for a struct containing
// an embedded Inside[N] field, and return said value if it can successfully
// find and cast it. Otherwise, it returns nil (for both a missing field and a
// failed cast).
func insideValueFromError[N isNamespace](err error) *Inside[N] {
	if err == nil {
		return nil
	}

	v := reflect.ValueOf(err)
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if field.Anonymous && strings.HasPrefix(field.Type.Name(), "Inside[") {
			insideVal, ok := v.Field(i).Interface().(Inside[N])
			if ok {
				return &insideVal
			}
		}
	}
	return nil
}

// Wrap[E] wraps the supplied error with a custom error tag of type E.
func Wrap[E any, PE interface {
	*E
	isTagged
}](err error) error {
	wrapper := PE(new(E))
	wrapperType := reflect.TypeOf(wrapper).Elem()

	wrapper.setNamespace(namespaceFromType(wrapperType))
	wrapper.setName(nameFromType(wrapperType))
	wrapper.setError(err)

	return wrapper
}

// Errorf[E] is a convenience wrapper for fmt.Errorf() wrapped with a custom
// error tag of type E.
func Errorf[E any, PE interface {
	*E
	isTagged
}](f string, args ...any) error {
	return Wrap[E, PE](fmt.Errorf(f, args...))
}

func namespaceFromType(t reflect.Type) string {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.Anonymous {
			continue
		}
		if match := reExtractNamespace.FindStringSubmatch(f.Type.String()); match != nil {
			return match[1]
		}
	}
	panic(fmt.Sprintf("failed to find namespace for error type: %s", t))
}

func nameFromType(t reflect.Type) string {
	pieces := strings.Split(t.String(), ".")
	return pieces[len(pieces)-1]
}

// OfType[E] returns a comparison object that can be used to match tagged errors
// using errors.Is(); the returned object will match errors within an error
// chain also tagged with E.
func OfType[E any, PE interface {
	*E
	isTagged
}]() error {
	example := Wrap[E, PE](nil).(PE)
	return &typeCompare{
		namespace: example.getNamespace(),
		name:      example.getName(),
	}
}

type typeCompare struct {
	namespace string
	name      string
}

// Error implements error for typeCompare using a string that is useful for
// debugging inside a unit testing context.
func (t *typeCompare) Error() string {
	return fmt.Sprintf("<error with tag %s::%s>", t.namespace, t.name)
}

// OfKind[N] returns a comparison objet that can be used to match tagged errors
// using errors.Is; the returned object will match errors within an error chain
// tagged with any tag inside namespace N.
func OfKind[N isNamespace]() error {
	var namespace N
	return &namespaceCompare{name: nameFromType(reflect.TypeOf(namespace))}
}

type namespaceCompare struct {
	name string
}

// Error implements error for namespaceCompare using a string that is useful for
// debugging inside a unit testing context.
func (n *namespaceCompare) Error() string {
	return fmt.Sprintf("<error with tag %s::*>", n.name)
}

// AsKind[N] returns an object that errors can be unpacked into using
// `errors.As`; this will match the first error in the chain tagged with any tag
// in namespace N.
func AsKind[N isNamespace]() *Inside[N] {
	return &Inside[N]{}
}


