package main

import "strings"

const PathSeparator = "/"

type Path struct {
	Path string
	ID string
}

// NewPath parses the specified path string and
// returns a new instance of the Path type.
// Leading and trailing slashes are trimmed and
// the remaining path is split by the PathSeparator
// constant.
// If there is more than one segment, the last one
// is considered to be the ID. We re-slice the slice
// of string to select the last item for the ID and
// the rest of the items for the remainder of the
// path.
func NewPath(p string) *Path {
	var id string
	p = strings.Trim(p, PathSeparator)
	s := strings.Split(p, PathSeparator)
	if len(s) > 1 {
		id = s[len(s)-1]
		p = strings.Join(s[:len(s)-1], PathSeparator)
	}
	return &Path{Path: p, ID: id}
}

func (p *Path) HasID() bool {
	return len(p.ID) > 0
}
