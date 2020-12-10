package filter

import "github.com/jamesBan/sensitive/store"

type Filter interface {
	Find(content string) (words []string)
	Replace(content string, replace string) string
	UpdateAll(s store.Store)
}
