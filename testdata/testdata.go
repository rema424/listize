package testdata

import (
	"listize/testdata/field"
)

type Struct struct {
	Field_1  string
	Field_2  *string
	field_3  int
	field_4  *int
	Field_5  Field
	Field_6  *Field
	Field_7  []Field
	Field_8  []*Field
	Field_9  map[int]Field
	Field_10 map[*Field]int
	Field_11 struct {
		field string
	}
	Field_12 *struct {
		field string
	}
	Field_13 field.Field
	Field_14 *field.Field
}

type Field struct {
	field string
}
