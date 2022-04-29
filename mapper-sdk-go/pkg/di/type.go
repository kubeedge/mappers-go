package di

import "reflect"

// TypeInstanceToName converts an instance of a type to a unique name.
func TypeInstanceToName(v interface{}) string {
	t := reflect.TypeOf(v)

	if name := t.Name(); name != "" {
		// non-interface types
		return t.PkgPath() + "." + name
	}

	// interface types
	e := t.Elem()
	return e.PkgPath() + "." + e.Name()
}
