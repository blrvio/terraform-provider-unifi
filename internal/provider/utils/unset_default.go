package utils

// OrDefault reports an unset controller value (the zero value of an omitted key) as
// the schema default, so a configuration that omits the attribute matches it.
func OrDefault[T comparable](v, def T) T {
	var zero T
	if v == zero {
		return def
	}
	return v
}

// KeepUnset returns the zero value when the controller had the field unset and the
// configuration still asks for the schema default, so a read-modify-write update
// never writes a default the controller did not have.
func KeepUnset[T comparable](current, planned, def T) T {
	var zero T
	if current == zero && planned == def {
		return zero
	}
	return planned
}
