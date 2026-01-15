package parser

import (
	"strings"
	"unicode"
)

// Return n-th until m-th argument and join them using single space.
// This function might ignore the quote character
func (r ParseResult) GetJoinedArg(n, m int) string {
	if n >= len(r.Args) {
		return ""
	}
	if m >= len(r.Args) {
		m = len(r.Args) - 1
	}
	if m < n {
		m = n
	}

	allArgsStr := []string{}
	for i := n; i <= m; i++ {
		allArgsStr = append(allArgsStr, r.Args[i].Content.Data)
	}

	return strings.Join(allArgsStr, " ")
}

func (r ParseResult) GetAllJoinedArg() string {
	return r.GetJoinedArg(0, len(r.Args)-1)
}

// Return n-th until m-th argument and join them using original space
func (r ParseResult) GetOriginalArg(n, m int) string {
	if n >= len(r.Args) {
		return ""
	}
	if m >= len(r.Args) {
		m = len(r.Args) - 1
	}
	if m < n {
		m = n
	}

	var res strings.Builder
	for i := n; i <= m; i++ {
		res.WriteString(r.Args[n].Content.Data)
		endIdx := r.Args[n].Position.End

		if endIdx == len(r.Text)-1 {
			continue
		}

		for j, rn := range r.Text[endIdx:] {
			if unicode.IsSpace(rn) {
				continue
			}

			res.WriteString(r.Text[endIdx:j])
			break
		}
	}

	return res.String()
}

func (r ParseResult) GetAllOriginalArg() string {
	return r.GetOriginalArg(0, len(r.Args)-1)
}
