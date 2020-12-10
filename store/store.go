package github.com/jamesBan/sensitive/store

type Store interface {
	Write(word string) error
	Remove(word string) error
	ReadAll() (<-chan string)
	Version() uint64
}