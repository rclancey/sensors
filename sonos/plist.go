package sonos

import (
	"bytes"
	"encoding"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	ErrTagMismatch = errors.New("open/close tag mismatch")
	ErrTagNotAllowed = errors.New("tag not allowed")
	ErrNotPointer = errors.New("receiver is not a pointer")
	ErrIncompatibleType = errors.New("incompatible types")
	ErrUnknownTag = errors.New("unknown tag")
)

func MarshalPlist(obj interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteString("\n")
	buf.WriteString(`<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">`)
	buf.WriteString("\n")
	buf.WriteString(`<plist version="1.0">`)
	err := writePlist(buf, obj, 0)
	if err != nil {
		return nil, err
	}
	buf.WriteString("\n</plist>")
	return buf.Bytes(), nil
}

func indent(w io.Writer, depth int) error {
	x := make([]byte, depth)
	for i := 0; i < depth; i++ {
		x[i] = '\t'
	}
	w.Write(x)
	return nil
}

func writePlist(w io.Writer, obj interface{}, depth int) error {
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		return writePlist(w, rv.Elem().Interface(), depth)
	}
	t, ok := obj.(time.Time)
	if ok {
		w.Write([]byte(fmt.Sprintf(`<date>%s</date>`, t.Format(time.RFC3339))))
		return nil
	}
	switch rv.Kind() {
	case reflect.Struct:
		w.Write([]byte{'\n'})
		indent(w, depth)
		w.Write([]byte("<dict>\n"))
		rt := rv.Type()
		n := rt.NumField()
		for i := 0; i < n; i++ {
			rf := rt.Field(i)
			if rf.PkgPath != "" {
				continue
			}
			tag := rf.Tag.Get("plist")
			if tag == "" {
				tag = rf.Name
			}
			indent(w, depth+1)
			w.Write([]byte("<key>"))
			xml.Escape(w, []byte(tag))
			w.Write([]byte("</key>"))
			err := writePlist(w, rv.Field(i).Interface(), depth+1)
			if err != nil {
				return err
			}
			w.Write([]byte{'\n'})
		}
		indent(w, depth)
		w.Write([]byte("</dict>"))
		return nil
	case reflect.Map:
		w.Write([]byte{'\n'})
		indent(w, depth)
		w.Write([]byte("<dict>\n"))
		iter := rv.MapRange()
		for iter.Next() {
			indent(w, depth+1)
			w.Write([]byte("<key>"))
			k := iter.Key()
			switch k.Kind() {
			case reflect.String:
				w.Write([]byte(k.String()))
			case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
				w.Write([]byte(strconv.FormatInt(k.Int(), 10)))
			case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
				w.Write([]byte(strconv.FormatUint(k.Uint(), 10)))
			case reflect.Float64, reflect.Float32:
				w.Write([]byte(strconv.FormatFloat(k.Float(), 'f', -1, 64)))
			default:
				switch xk := k.Interface().(type) {
				case fmt.Stringer:
					w.Write([]byte(xk.String()))
				case encoding.TextMarshaler:
					data, err := xk.MarshalText()
					if err != nil {
						return err
					}
					w.Write(data)
				default:
					return ErrIncompatibleType
				}
			}
			w.Write([]byte("</key>"))
			err := writePlist(w, iter.Value().Interface(), depth+1)
			if err != nil {
				return err
			}
			w.Write([]byte{'\n'})
		}
		indent(w, depth)
		w.Write([]byte("</dict>"))
		return nil
	case reflect.Slice:
		w.Write([]byte{'\n'})
		indent(w, depth)
		w.Write([]byte(fmt.Sprintf("<array>\n")))
		n := rv.Len()
		for i := 0; i < n; i++ {
			err := writePlist(w, rv.Index(i).Interface(), depth+1)
			if err != nil {
				return err
			}
		}
		w.Write([]byte{'\n'})
		indent(w, depth)
		w.Write([]byte(fmt.Sprintf("</array>")))
		return nil
	case reflect.String:
		w.Write([]byte(`<string>`))
		xml.Escape(w, []byte(rv.String()))
		w.Write([]byte(`</string>`))
		return nil
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		w.Write([]byte(`<integer>`))
		w.Write([]byte(strconv.FormatInt(rv.Int(), 10)))
		w.Write([]byte(`</integer>`))
		return nil
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		w.Write([]byte(`<integer>`))
		w.Write([]byte(strconv.FormatUint(rv.Uint(), 10)))
		w.Write([]byte(`</integer>`))
		return nil
	case reflect.Float64, reflect.Float32:
		w.Write([]byte(`<real>`))
		w.Write([]byte(strconv.FormatFloat(rv.Float(), 'f', -1, 64)))
		w.Write([]byte(`</real>`))
		return nil
	case reflect.Bool:
		if rv.Bool() {
			w.Write([]byte(`<true/>`))
		} else {
			w.Write([]byte(`<false/>`))
		}
		return nil
	}
	return ErrIncompatibleType
}

func UnmarshalPlist(data []byte, obj interface{}) error {
	dec := xml.NewDecoder(bytes.NewReader(data))
	inPlist := false
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch el := tok.(type) {
		case xml.StartElement:
			if el.Name.Local == "plist" {
				inPlist = true
			} else if inPlist {
				return decodePlistTo(dec, el, obj)
			}
		}
	}
}

func readSimpleValue(dec *xml.Decoder, elem xml.StartElement) ([]byte, error) {
	valBytes := []byte{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch el := tok.(type) {
		case xml.CharData:
			valBytes = append(valBytes, []byte(el)...)
		case xml.EndElement:
			if el.Name.Space != elem.Name.Space || el.Name.Local != elem.Name.Local {
				return nil, ErrTagMismatch
			}
			return valBytes, nil
		case xml.StartElement:
			return nil, ErrTagNotAllowed
		}
	}
	return valBytes, nil
}

func findField(rv reflect.Value, key string) interface{} {
	rt := rv.Type()
	n := rt.NumField()
	for i := 0; i < n; i++ {
		rf := rt.Field(i)
		if rf.PkgPath != "" {
			continue
		}
		tag := rf.Tag.Get("plist")
		if tag == "-" {
			continue
		}
		if tag == key || strings.ToLower(rf.Name) == strings.ToLower(key) {
			f := rv.Field(i)
			return f.Addr().Interface()
		}
	}
	var obj interface{}
	return &obj
}

func assignToReceiver(obj interface{}, val interface{}) error {
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Ptr {
		return ErrNotPointer
	}
	rv = rv.Elem()
	xv := reflect.ValueOf(val)
	if rv.Type() == xv.Type() {
		rv.Set(xv)
		return nil
	}
	t, ok := val.(time.Time)
	if ok {
		switch rv.Kind() {
		case reflect.Interface:
			rv.Set(xv)
			return nil
		case reflect.String:
			rv.SetString(t.Format(time.RFC3339))
			return nil
		case reflect.Int, reflect.Int64:
			rv.SetInt(t.UnixMilli())
			return nil
		case reflect.Int32:
			rv.SetInt(t.Unix())
			return nil
		case reflect.Uint, reflect.Uint64:
			rv.SetUint(uint64(t.UnixMilli()))
			return nil
		case reflect.Uint32:
			rv.SetUint(uint64(t.Unix()))
			return nil
		case reflect.Float64:
			rv.SetFloat(float64(t.Unix()) + float64(t.Nanosecond()) / 1e9)
			return nil
		}
	}
	switch rv.Kind() {
	case reflect.Interface:
		rv.Set(xv)
		return nil
	case reflect.String:
		switch xv.Kind() {
		case reflect.String:
			rv.SetString(xv.String())
			return nil
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			rv.SetString(strconv.FormatInt(xv.Int(), 10))
			return nil
		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			rv.SetString(strconv.FormatUint(xv.Uint(), 10))
			return nil
		case reflect.Float64, reflect.Float32:
			rv.SetString(strconv.FormatFloat(xv.Float(), 'f', -1, 64))
			return nil
		case reflect.Bool:
			rv.SetString(strconv.FormatBool(xv.Bool()))
			return nil
		}
		switch tv := val.(type) {
		case fmt.Stringer:
			rv.SetString(tv.String())
			return nil
		case encoding.TextMarshaler:
			data, err := tv.MarshalText()
			if err != nil {
				return err
			}
			rv.SetString(string(data))
			return nil
		}
		return ErrIncompatibleType
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		switch xv.Kind() {
		case reflect.String:
			i, err := strconv.ParseInt(xv.String(), 10, 64)
			if err != nil {
				return err
			}
			rv.SetInt(i)
			return nil
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			rv.SetInt(xv.Int())
			return nil
		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			rv.SetInt(int64(xv.Uint()))
			return nil
		case reflect.Float64, reflect.Float32:
			rv.SetInt(int64(xv.Float()))
			return nil
		case reflect.Bool:
			if xv.Bool() {
				rv.SetInt(1)
			} else {
				rv.SetInt(0)
			}
			return nil
		}
		return ErrIncompatibleType
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		switch xv.Kind() {
		case reflect.String:
			u, err := strconv.ParseUint(xv.String(), 10, 64)
			if err != nil {
				return err
			}
			rv.SetUint(u)
			return nil
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			rv.SetUint(uint64(xv.Int()))
			return nil
		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			rv.SetUint(xv.Uint())
			return nil
		case reflect.Float64, reflect.Float32:
			rv.SetUint(uint64(xv.Float()))
			return nil
		case reflect.Bool:
			if xv.Bool() {
				rv.SetUint(1)
			} else {
				rv.SetUint(0)
			}
			return nil
		}
		return ErrIncompatibleType
	case reflect.Float64, reflect.Float32:
		switch xv.Kind() {
		case reflect.String:
			f, err := strconv.ParseFloat(xv.String(), 64)
			if err != nil {
				return err
			}
			rv.SetFloat(f)
			return nil
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			rv.SetFloat(float64(xv.Int()))
			return nil
		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			rv.SetFloat(float64(xv.Uint()))
			return nil
		case reflect.Float64, reflect.Float32:
			rv.SetFloat(xv.Float())
			return nil
		case reflect.Bool:
			if xv.Bool() {
				rv.SetFloat(1)
			} else {
				rv.SetFloat(0)
			}
			return nil
		}
	case reflect.Bool:
		switch xv.Kind() {
		case reflect.String:
			b, err := strconv.ParseBool(xv.String())
			if err != nil {
				return err
			}
			rv.SetBool(b)
			return nil
		case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			rv.SetBool(xv.Int() > 0)
			return nil
		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			rv.SetBool(xv.Uint() > 0)
			return nil
		case reflect.Float64, reflect.Float32:
			rv.SetBool(xv.Float() > 0)
			return nil
		case reflect.Bool:
			rv.SetBool(xv.Bool())
			return nil
		}
		return ErrIncompatibleType
	}
	return ErrIncompatibleType
}

func decodePlistTo(dec *xml.Decoder, elem xml.StartElement, obj interface{}) error {
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Ptr {
		return ErrNotPointer
	}
	rv = rv.Elem()
	if rv.Kind() == reflect.Ptr {
		xrv := reflect.New(rv.Type().Elem())
		err := decodePlistTo(dec, elem, xrv.Interface())
		if err != nil {
			return err
		}
		rv.Set(xrv)
		return nil
	}
	switch elem.Name.Local {
	case "dict":
		if rv.Kind() == reflect.Interface {
			xobj := map[string]interface{}{}
			err := decodePlistTo(dec, elem, &xobj)
			if err != nil {
				return err
			}
			rv.Set(reflect.ValueOf(xobj))
			return nil
		}
		if rv.Kind() == reflect.Struct {
			var key string
			for {
				tok, err := dec.Token()
				if err != nil {
					return err
				}
				switch el := tok.(type) {
				case xml.StartElement:
					if el.Name.Local == "key" {
						keyBytes, err := readSimpleValue(dec, el)
						if err != nil {
							return err
						}
						key = strings.TrimSpace(string(keyBytes))
					} else {
						field := findField(rv, key)
						err = decodePlistTo(dec, el, field)
						if err != nil {
							return err
						}
					}
				case xml.EndElement:
					if el.Name.Space != elem.Name.Space || el.Name.Local != elem.Name.Local {
						return ErrTagMismatch
					}
					return nil
				}
			}
		}
		if rv.Kind() == reflect.Map {
			rm := reflect.MakeMap(rv.Type())
			var key string
			for {
				tok, err := dec.Token()
				if err != nil {
					return err
				}
				switch el := tok.(type) {
				case xml.StartElement:
					if el.Name.Local == "key" {
						keyBytes, err := readSimpleValue(dec, el)
						if err != nil {
							return err
						}
						key = strings.TrimSpace(string(keyBytes))
					} else {
						field := reflect.New(rm.Type().Elem())
						err = decodePlistTo(dec, el, field.Interface())
						if err != nil {
							return err
						}
						kv := reflect.New(rm.Type().Key())
						err = assignToReceiver(kv.Interface(), key)
						if err != nil {
							return err
						}
						rm.SetMapIndex(kv.Elem(), field.Elem())
					}
				case xml.EndElement:
					if el.Name.Space != elem.Name.Space || el.Name.Local != elem.Name.Local {
						return ErrTagMismatch
					}
					rv.Set(rm)
					return nil
				}
			}
		}
		return ErrIncompatibleType
	case "array":
		if rv.Kind() == reflect.Interface {
			xobj := []interface{}{}
			err := decodePlistTo(dec, elem, &xobj)
			if err != nil {
				return err
			}
			rv.Set(reflect.ValueOf(xobj))
			return nil
		}
		if rv.Kind() != reflect.Slice {
			return ErrIncompatibleType
		}
		s := reflect.MakeSlice(rv.Type(), 0, 0)
		for {
			tok, err := dec.Token()
			if err != nil {
				return err
			}
			switch el := tok.(type) {
			case xml.StartElement:
				item := reflect.New(rv.Type().Elem())
				err = decodePlistTo(dec, el, item.Interface())
				if err != nil {
					return err
				}
				s = reflect.Append(s, item.Elem())
			case xml.EndElement:
				if el.Name.Space != elem.Name.Space || el.Name.Local != elem.Name.Local {
					return ErrTagMismatch
				}
				rv.Set(s)
				return nil
			}
		}
	case "string":
		valBytes, err := readSimpleValue(dec, elem)
		if err != nil {
			return err
		}
		s := strings.TrimSpace(string(valBytes))
		return assignToReceiver(obj, s)
	case "date":
		valBytes, err := readSimpleValue(dec, elem)
		if err != nil {
			return err
		}
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(string(valBytes)))
		if err != nil {
			return err
		}
		return assignToReceiver(obj, t)
	case "integer":
		valBytes, err := readSimpleValue(dec, elem)
		if err != nil {
			return err
		}
		i, err := strconv.ParseInt(strings.TrimSpace(string(valBytes)), 10, 64)
		if err != nil {
			return err
		}
		return assignToReceiver(obj, i)
	case "true":
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		end, ok := tok.(xml.EndElement)
		if !ok {
			return ErrTagNotAllowed
		}
		if end.Name.Local != "true" {
			return ErrTagMismatch
		}
		return assignToReceiver(obj, true)
	case "false":
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		end, ok := tok.(xml.EndElement)
		if !ok {
			return ErrTagNotAllowed
		}
		if end.Name.Local != "false" {
			return ErrTagMismatch
		}
		return assignToReceiver(obj, false)
	}
	return ErrUnknownTag
}
