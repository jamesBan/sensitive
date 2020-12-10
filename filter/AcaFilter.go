package filter

import (
	"bytes"
	"github.com/eachain/aca"
	"github.com/jamesBan/sensitive/store"
	"sort"
	"unicode/utf8"
)

type AcaFilter struct {
	aca *aca.ACA
}

func NewAcaFilter() Filter {
	filter := &AcaFilter{}
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

func (f *AcaFilter) Replace(content string, replace string) string {
	return replaceAll(f.aca, content, []rune(replace)[0])
}

type byPos []aca.Block

func (bs byPos) Len() int { return len(bs) }

func (bs byPos) Swap(i, j int) { bs[i], bs[j] = bs[j], bs[i] }

func (bs byPos) Less(i, j int) bool {
	if bs[i].Start < bs[j].Start {
		return true
	}
	if bs[i].Start == bs[j].Start {
		return bs[i].End < bs[j].End
	}
	return false
}

func replaceAll(a *aca.ACA, s string, new rune) string {
	tmp := make([]rune, utf8.RuneCountInString(s))
	for i := range tmp {
		tmp[i] = new
	}

	now := 0
	buf := &bytes.Buffer{}
	for _, b := range unionBlocks(a.Blocks(s)) {
		buf.WriteString(s[now:b.Start])
		cnt := utf8.RuneCountInString(s[b.Start:b.End])
		buf.WriteString(string(tmp[:cnt]))
		now = b.End
	}
	if now < len(s) {
		buf.WriteString(s[now:])
	}
	return buf.String()
}

func unionBlocks(blocks []aca.Block) []aca.Block {
	if len(blocks) == 0 {
		return blocks
	}

	sort.Sort(byPos(blocks))
	n := 0
	for i := 1; i < len(blocks); i++ {
		if blocks[i].Start <= blocks[n].End {
			if blocks[i].End > blocks[n].End {
				blocks[n].End = blocks[i].End
			}
		} else {
			n++
			blocks[n] = blocks[i]
		}
	}
	return blocks[:n+1]
}
