package p

type Field struct {
	Name, Type string
}

type Schema struct {
	Fields   []Field
	FieldMap map[string]*Field
}

func NewSchema(fields ...Field) Schema {
	m := make(map[string]*Field)
	for _, field := range fields {
		m[field.Name] = &field
	}
	return Schema{Fields: fields, FieldMap: m}
}
