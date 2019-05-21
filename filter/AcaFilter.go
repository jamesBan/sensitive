package filter

import (
	"sensitive/store"
	"github.com/eachain/aca"
	"strings"
)

type AcaFilter struct {
	aca *aca.ACA
}


func NewAcaFilter(s store.Store) Filter {
	filter := &AcaFilter{}
	a := aca.New()
	for word := range s.ReadAll() {
		a.Add(word)
	}

	a.Build()

	filter.aca = a
	return filter
}

func (f *AcaFilter) UpdateAll(s store.Store) {
	a := aca.New()
	for word := range s.ReadAll() {
		a.Add(word)
	}

	a.Build()

	f.aca = a
}

func (f *AcaFilter) Find(content string) (words []string) {
	return f.aca.Find(content)
}

func (f *AcaFilter) Replace(content string, replace string) (string) {
	words := f.Find(content)
	if len(words) < 1 {
		return content
	}

	for i, l := 0, len(words); i < l; i++ {
		content = strings.Replace(content, words[i], strings.Repeat(replace, len(words)), 1)
	}

	return content
}
