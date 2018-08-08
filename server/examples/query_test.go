package main

import (
	"reflect"
	"testing"
)

func TestSearch(t *testing.T) {
	ex := ExampleSet{
		Example{
			NameSan:  "hello world",
			Language: "go",
			Tags: []Tag{
				Tag{
					Raw:   "hello-world",
					Parts: []string{"hello", "world"},
				},
			},
		},
		Example{
			NameSan:  "counterexample",
			Language: "ruby",
			Tags: []Tag{
				Tag{
					Raw: "wow",
				},
			},
		},
	}
	tbl := []struct {
		query  string
		expect ExampleSet
	}{
		{
			query: "lang:go",
			expect: ExampleSet{
				ex[0],
			},
		},
		{
			query: "hello world &&",
			expect: ExampleSet{
				ex[0],
			},
		},
		{
			query: "hello world AND",
			expect: ExampleSet{
				ex[0],
			},
		},
		{
			query: "hello world",
			expect: ExampleSet{
				ex[0],
			},
		},
		{
			query: "tag:hello-world",
			expect: ExampleSet{
				ex[0],
			},
		},
		{
			query: "tag:hello tag:world",
			expect: ExampleSet{
				ex[0],
			},
		},
		{
			query: "hello",
			expect: ExampleSet{
				ex[0],
			},
		},
		{
			query: "lang:ruby",
			expect: ExampleSet{
				ex[1],
			},
		},
		{
			query:  "lang:go lang:ruby ||",
			expect: ex,
		},
		{
			query: "lang:go ! ruby ||",
			expect: ExampleSet{
				ex[1],
			},
		},
		{
			query: "name:c",
			expect: ExampleSet{
				ex[1],
			},
		},
		{
			query:  "",
			expect: ex,
		},
		{
			query:  "     ",
			expect: ex,
		},
		{
			query:  "nomatch",
			expect: ExampleSet{},
		},
	}
	for _, v := range tbl {
		got := ex.SearchQuery(v.query)
		if !reflect.DeepEqual(v.expect, got) {
			t.Errorf("expected %v but got %v", v.expect, got)
		}
	}
}
