//Package goval exposes functions to return strings
//containing a roundabout type declaration for passed interfaces.
package goval

import (
	"reflect"
	"strconv"
	"strings"
)

func Val(i interface{}) (decl string) {
	return ValType(reflect.TypeOf(i))
}

func indent(s string) (o string) {
	bits := strings.Split(s, "\n")
	for i, v := range bits {
		bits[i] = "\t" + v
	}
	return strings.Join(bits, "\n")
}

func ValType(t reflect.Type) (decl string) {
	switch k := t.Kind(); k {
	case reflect.Struct:
		decl = "struct {\n"
		for i, ed := 0, t.NumField(); i < ed; i++ {
			ft := t.Field(i)
			if ft.Tag != "-" || ft.Tag.Get("goval") == "-" {
				s := ft.Name + " " + ValType(ft.Type)
				if ft.Tag != "" {
					s += " `" + strings.Replace("`", "\\`", string(ft.Tag), -1) + "`"
				}
				decl += indent(s) + "\n"
			}
		}
		decl += "}"
	case reflect.Array:
		decl = "[" + strconv.Itoa(t.Len()) + "]" + Val(t.Elem())
	case reflect.Slice:
		decl = "[]" + Val(t.Elem())
	case reflect.Chan:
		switch t.ChanDir() {
		case reflect.RecvDir:
			decl = "<-chan "
		case reflect.SendDir:
			decl = "chan<- "
		case reflect.BothDir:
			decl = "chan "
		default:
			panic("Didn't expect a dir other than send, recieve or both.")
		}
		decl += Val(t.Elem())
	case reflect.Map:
		decl = "map[" + ValType(t.Key()) + "]" + ValType(t.Elem())
	case reflect.Ptr:
		decl = "*" + ValType(t.Elem())
	case reflect.Interface:
		decl = "interface {\n"
		for i, ed := 0, t.NumMethod(); i < ed; i++ {
			ft := t.Method(i)
			s := ft.Name + FormatFuncArguments(ft.Type)
			decl += indent(s) + "\n"
		}
		decl += "}"
	case reflect.Func:
		decl = "func" + FormatFuncArguments(t)
	default:
		return k.String()
	}

	return
}

func FormatFuncArguments(t reflect.Type) (decl string) {
	decl = "("
	in := make([]string, t.NumIn())

	for i := range in {
		in[i] = ValType(t.In(i))
	}

	decl += strings.Join(in, ",") + ")"

	out := make([]string, t.NumOut())
	if len(out) > 0 {

		for i := range out {
			out[i] = ValType(t.Out(i))
		}

		s := strings.Join(out, ",")

		if len(out) != 1 {
			s = "(" + s + ")"
		}

		decl += " " + s
	}
	return
}
