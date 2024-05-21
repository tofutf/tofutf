package html

import (
	"fmt"
	"html/template"
	"net/url"
	"reflect"

	types "github.com/hashicorp/go-tfe"
	"github.com/tofutf/tofutf/internal"
)

// mergeQuery merges the query string into the given url, replacing any existing
// query parameters with the same name.
func mergeQuery(u string, q string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	mergeQuery, err := url.ParseQuery(q)
	if err != nil {
		return "", err
	}
	existingQuery := parsedURL.Query()
	for k, v := range mergeQuery {
		existingQuery.Set(k, v[0])
	}
	parsedURL.RawQuery = existingQuery.Encode()
	return parsedURL.String(), nil
}

func prevPageQuery(p types.Pagination) *string {
	if p.PreviousPage == 0 {
		return nil
	}
	return internal.String(fmt.Sprintf("page[number]=%d", p.PreviousPage))
}

func nextPageQuery(p types.Pagination) *string {
	if p.NextPage == 0 {
		return nil
	}
	return internal.String(fmt.Sprintf("page[number]=%d", p.NextPage))
}

func disabled(arg any, args ...any) (template.HTMLAttr, error) {
	return attrIf("disabled", arg, args...)
}

func selected(arg any, args ...any) (template.HTMLAttr, error) {
	return attrIf("selected", arg, args...)
}

func checked(arg any, args ...any) (template.HTMLAttr, error) {
	return attrIf("checked", arg, args...)
}

// attrIf returns string as an html attribute, if:
// (a) single arg provided, it is a boolean, and it is true.
// (b) multiple args provided, they are all strings, and they are all equal.
// otherwise it outputs an empty attribute
// This is useful for printing strings in templates or not.
func attrIf(s string, arg any, args ...any) (template.HTMLAttr, error) {
	if len(args) == 0 {
		if reflect.ValueOf(arg).Kind() == reflect.Bool {
			if reflect.ValueOf(arg).Bool() {
				return template.HTMLAttr(s), nil
			}
		}
		return "", nil
	}
	if reflect.ValueOf(arg).Kind() != reflect.String {
		return "", nil
	}
	lastarg := reflect.ValueOf(arg).String()
	for _, a := range args {
		if reflect.ValueOf(a).Kind() != reflect.String {
			return "", nil
		}
		if reflect.ValueOf(a).String() != lastarg {
			return "", nil
		}
	}
	return template.HTMLAttr(s), nil
}
