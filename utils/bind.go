package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

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
		} else {
			name = strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		}
		return name
	})

	return v
}

func BindBody[T any](body io.Reader, data T) (T, error) {
	err := defaults.Set(&data)
	if err != nil {
		return data, err
	}

	err = json.NewDecoder(body).Decode(&data)
	if err != nil {
		return data, err
	}

	err = Validator.Validate(data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func BindQuery[T any](form url.Values, data T) (T, error) {
	err := defaults.Set(&data)
	if err != nil {
		return data, err
	}

	err = queryBinder.Decode(&data, form)
	if err != nil {
		return data, err
	}

	err = Validator.Validate(data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func BindUri[T any](r *http.Request) (T, error) {
	var data T
	return data, nil
}

func Bind(r *http.Request, data any) error {
	err := defaults.Set(data)
	if err != nil {
		return err
	}
	if err = r.ParseForm(); err != nil {
		return err
	}
	err = queryBinder.Decode(data, r.Form)
	if err != nil {
		return err
	}
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if len(bodyData) > 0 {
		err = json.Unmarshal(bodyData, data)
		if err != nil {
			return err
		}
	}

	err = Validator.Validate(data)
	if err != nil {
		return err
	}

	return nil
}
