package dispatcher

import "github.com/smarthome-go/homescript/v3/homescript/parser/ast"

func (i *InstanceT) RegisterAnnotations() {
	// TODO: get all homescripts from the user hms table and the driver table.
	// Then, loop over them and analyze the annotations.
	// NOTE: an interesting problem is when to re-run this function.
}

func (i *InstanceT) RegisterAnnotation(annotaion ast.AnnotationItem) {
	panic("TODO: process annotation")
}
