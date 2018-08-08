package main

import (
	"strings"
)

// Query is a search query.
type Query func(Example) bool

// QueryAnd returns a Query that is matched if both input Queries are matched.
func QueryAnd(q1, q2 Query) Query {
	return Query(func(e Example) bool {
		return q1(e) && q2(e)
	})
}

// QueryOr returns a Query that is matched if at least one input Query is matched.
func QueryOr(q1, q2 Query) Query {
	return Query(func(e Example) bool {
		return q1(e) || q2(e)
	})
}

// QueryInvert returns a Query that is matched if the input Query is not matched.
func QueryInvert(q Query) Query {
	return Query(func(e Example) bool {
		return !q(e)
	})
}

// QueryLanguage is a query by language.
func QueryLanguage(lang string) Query {
	return Query(func(e Example) bool {
		return e.Language == lang
	})
}

// QueryTag queries by tag.
func QueryTag(tag string) Query {
	return Query(func(e Example) bool {
		if e.Tags == nil {
			return false
		}
		for _, v := range e.Tags {
			if v.Raw == tag {
				return true
			}
		}
		return false
	})
}

// QueryName is a query that matches when the given word is in the name of the Example.
func QueryName(word string) Query {
	word = sanitizeText(word)
	return Query(func(e Example) bool {
		return strings.Contains(e.NameSan, word)
	})
}

// QueryWord runs a full query for the given word.
// Checks:
//  if the word is in the name
//  if the word is in a tag
//  if the word is the language name
func QueryWord(word string) Query {
	word = sanitizeText(word)
	return QueryOr(
		QueryName(word),
		QueryOr(
			QueryTag(word),
			QueryLanguage(word),
		),
	)
}

// QueryWildcard is a query that matches anything.
var QueryWildcard = Query(func(Example) bool {
	return true
})

type queryStack []Query

func (qs *queryStack) push(q Query) {
	*qs = append(*qs, q)
}

// pop pops a query off the stack or returns nil if the stack is empty.
func (qs *queryStack) pop() Query {
	if len(*qs) == 0 {
		return nil
	}
	v := (*qs)[len(*qs)-1]
	*qs = (*qs)[:len(*qs)-1]
	return v
}

// ParseQuery parses a Query.
func ParseQuery(str string) Query {
	qs := queryStack{}

	// parse query into stack
	for _, v := range strings.Split(str, " ") {
		// query operators
		switch v {
		case "AND", "&&":
			if len(qs) >= 2 {
				qs.push(QueryAnd(qs.pop(), qs.pop()))
				continue
			}
		case "OR", "||":
			if len(qs) >= 2 {
				qs.push(QueryOr(qs.pop(), qs.pop()))
				continue
			}
		case "NOT", "!":
			if len(qs) >= 1 {
				qs.push(QueryInvert(qs.pop()))
				continue
			}
		}

		// specialized matching
		if strings.Count(v, ":") == 1 {
			spl := strings.Split(v, ":")
			switch spl[0] {
			case "language", "lang":
				qs.push(QueryLanguage(spl[1]))
				continue
			case "tag":
				qs.push(QueryTag(spl[1]))
				continue
			case "name":
				qs.push(QueryName(spl[1]))
				continue
			}
		}

		// basic matching
		qs.push(QueryWord(v))
	}

	// handle wildcard query
	if len(qs) == 0 {
		return QueryWildcard
	}

	// merge into a single query
	for len(qs) > 1 {
		qs.push(QueryAnd(qs.pop(), qs.pop()))
	}

	// pass out final query
	return qs.pop()
}
