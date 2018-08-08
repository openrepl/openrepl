package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"unsafe"
)

// Example is an example.
type Example struct {
	// Path is the path to the example.
	Path string `json:"path"`

	// Name is the name of the example.
	Name string `json:"name"`

	// NameSan is the sanitized name of the example.
	NameSan string `json:"-"`

	// Language is the programming language that the example was written in.
	Language string `json:"lang"`

	// Tags is the set of tags on the Example.
	Tags []Tag `json:"tags,omitempty"`

	// Code is the code of the example.
	Code string `json:"code"`
}

// Tag is an example tag.
type Tag struct {
	// Raw is the raw text of the tag.
	// All characters other than the following must be removed: [a-z]|[0-9]|[ -]
	// All uppercase letters [A-Z] will be replaced with lowercase runes.
	// Underscores must be replaced with dashes.
	Raw string

	// Parts is the broken-down parts of the Tag.
	// If the raw text is one part, this should be nil.
	Parts []string

	// IsDash is whether the parts are seperated by a dash.
	// If raw contains both dashes and spaces, this should be set to false.
	// Conditions:
	//  no dash    && no space => false
	//  dash       && no space => true
	//  no dash    && space    => false
	//  dash       && space    => false
	IsDash bool
}

// MarshalJSON marshals a tag as JSON.
// Equivalent to json.Marshal(t.Raw).
func (t Tag) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Raw)
}

// UnmarshalJSON unmarshals a raw tag from a JSON string.
func (t *Tag) UnmarshalJSON(dat []byte) error {
	// get string
	var str string
	err := json.Unmarshal(dat, &str)
	if err != nil {
		return err
	}

	// parse tag
	*t = ParseTag(str)

	return nil
}

// cleanRune sanitizes the rune for the rules defined in Tag.Raw.
// '\000' is returned if the rune must be deleted.
func cleanRune(r rune) rune {
	switch {
	case r >= 'a' && r <= 'z':
		fallthrough
	case r >= '0' && r <= '9':
		fallthrough
	case r == '-' || r == ' ':
		return r
	case r == '_':
		return '-'
	case r >= 'A' && r <= 'Z':
		return r - ('A' - 'a')
	default:
		return '\000'
	}
}

// sanitizeText sanitizes text by the rules defined in Tag.Raw.
func sanitizeText(str string) string {
	// count bytes in output & check whether modification is required
	nlen := 0
	mustmod := false
	for _, r := range []rune(str) {
		outrune := cleanRune(r)
		if outrune != '\000' {
			nlen++
		}
		if outrune != r {
			mustmod = true
		}
	}

	// if no modification is required, use the original string
	if !mustmod {
		return str
	}

	// build new string
	nstr := make([]byte, nlen)
	i := 0
	for _, r := range []rune(str) {
		outrune := cleanRune(r)
		if outrune != '\000' {
			nstr[i] = byte(outrune)
			i++
		}
	}

	// unsafe convert byte slice to string
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: uintptr(unsafe.Pointer(&nstr[0])),
		Len:  len(nstr),
	}))
}

// ParseTag parses a Tag.
func ParseTag(str string) Tag {
	// sanitize text
	str = sanitizeText(str)

	// detect split type
	hasSpace := strings.ContainsRune(str, ' ')
	hasDash := strings.ContainsRune(str, '-')
	var spl []string
	switch {
	case hasSpace && hasDash:
		// this is rare enough that I dont want to bother optimizing it
		spl = strings.Split(strings.Replace(str, " ", "-", -1), "-")
	case hasSpace:
		spl = strings.Split(str, " ")
	case hasDash:
		spl = strings.Split(str, "-")
	}

	// generate tag
	return Tag{
		Raw:    str,
		Parts:  spl,
		IsDash: hasDash && !hasSpace,
	}
}

// ExampleSet is a set of Examples.
type ExampleSet []Example
