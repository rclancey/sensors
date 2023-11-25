package api

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type DBLike interface {
	Exec(string, ...interface{}) (sql.Result, error)
}

type IDable interface {
	GetID() int64
	SetID(int64)
	Table() string
}

func strSliceHas(slc []string, s string) bool {
	for _, x := range slc {
		if x == s {
			return true
		}
	}
	return false
}

func InsertStruct(db DBLike, obj IDable) error {
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	rt := rv.Type()
	cols := []string{}
	vals := []interface{}{}
	qms := []string{}
	n := rv.NumField()
	for i := 0; i < n; i++ {
		f := rt.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tag := f.Tag.Get("dbignore")
		if strings.Contains(tag, "insert") {
			continue
		}
		parts := strings.Split(f.Tag.Get("db"), ",")
		tag = parts[0]
		if tag == "" {
			tag = strings.ToLower(f.Name)
		}
		if tag == "-" {
			continue
		}
		if len(parts) > 1 && strSliceHas(parts[1:], "primary_key") {
			continue
		}
		cols = append(cols, tag)
		vals = append(vals, rv.Field(i).Interface())
		qms = append(qms, "?")
	}
	qs := fmt.Sprintf(`INSERT INTO %s (%s) VALUES(%s)`, obj.Table(), strings.Join(cols, ","), strings.Join(qms, ","))
	res, err := db.Exec(qs, vals...)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	obj.SetID(id)
	return nil
}

func UpdateStruct(db DBLike, obj IDable) error {
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	rt := rv.Type()
	cols := []string{}
	vals := []interface{}{}
	n := rv.NumField()
	var pkey string
	for i := 0; i < n; i++ {
		f := rt.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tag := f.Tag.Get("dbignore")
		if strings.Contains(tag, "update") {
			continue
		}
		parts := strings.Split(f.Tag.Get("db"), ",")
		tag = parts[0]
		if tag == "" {
			tag = strings.ToLower(f.Name)
		}
		if tag == "-" {
			continue
		}
		if len(parts) > 1 && strSliceHas(parts[1:], "primary_key") {
			pkey = tag
			continue
		}
		cols = append(cols, fmt.Sprintf("%s = ?", tag))
		vals = append(vals, rv.Field(i).Interface())
	}
	qs := fmt.Sprintf(`UPDATE %s SET %s WHERE %s = ?`, obj.Table(), strings.Join(cols, ", "), pkey)
	vals = append(vals, obj.GetID())
	_, err := db.Exec(qs, vals...)
	return err
}
