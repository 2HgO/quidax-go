package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/2HgO/quidax-go/errors"
	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
)

var Validator = NewStructValidator()
var queryBinder = schema.NewDecoder()

func init() {
	queryBinder.SetAliasTag("query")
	queryBinder.IgnoreUnknownKeys(true)
}

type structValidator struct {
	validator *validator.Validate
}

func (s *structValidator) Validate(v any) error {
	return s.validator.Struct(v)
}

func NewStructValidator() *structValidator {
	v := &structValidator{validator: validator.New()}

	v.validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		var name string
		if tag, ok := fld.Tag.Lookup("query"); ok {
			name = strings.SplitN(tag, ",", 2)[0]
		} else if tag, ok := fld.Tag.Lookup("uri"); ok {
			name = strings.SplitN(tag, ",", 2)[0]
		} else {
			name = strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		}
		return name
	})

	return v
}

func bindUri(r *http.Request, data any) error {
	t := reflect.TypeOf(data)
	switch {
	case t.Kind() != reflect.Pointer,
		t.Elem().Kind() != reflect.Struct:
		return errors.NewValidationError("invalid data type")
	}
	fields := reflect.VisibleFields(t.Elem())
	for _, field := range fields {
		if !field.IsExported() || field.Type.Kind() != reflect.String {
			continue
		}
		if key, ok := field.Tag.Lookup("uri"); ok {
			reflect.Indirect(reflect.ValueOf(data)).FieldByName(field.Name).SetString(r.PathValue(key))
		}
	}
	return nil
}

func Bind[T any](r *http.Request) *T {
	if reflect.TypeFor[T]().Kind() != reflect.Struct {
		panic(errors.NewValidationError("invalid request type"))
	}
	data := new(T)
	err := defaults.Set(data)
	if err != nil {
		panic(errors.HandleBindError(err))
	}
	if err = bindUri(r, data); err != nil {
		panic(errors.HandleBindError(err))
	}
	if err = r.ParseForm(); err != nil {
		panic(errors.HandleBindError(err))
	}
	err = queryBinder.Decode(data, r.Form)
	if err != nil {
		panic(errors.HandleBindError(err))
	}
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		panic(errors.HandleBindError(err))
	}
	if len(bodyData) > 0 {
		err = json.Unmarshal(bodyData, data)
		if err != nil {
			panic(errors.HandleBindError(err))
		}
	}

	err = Validator.Validate(data)
	if err != nil {
		panic(errors.HandleBindError(err))
	}

	return data
}
