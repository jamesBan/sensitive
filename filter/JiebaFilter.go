package filter

import (
	"fmt"
	"os"
	"log"
	"github.com/yanyiwu/gojieba"
	"strings"
	"unicode/utf8"
	"sensitive/store"
)

const WORD_SPPECH = "mg"

type Jieba struct {
	dictPath      string
	hmmPath       string
	userDictPath  string
	idfPath       string
	stopWordsPath string
	jieba         *gojieba.Jieba
}

func NewJiebaFilter(dictPath, hmmPath, userDictPath, idfPath, stopWordsPath string) (Filter) {
	j := &Jieba{
		dictPath:      dictPath,
		hmmPath:       hmmPath,
		userDictPath:  userDictPath,
		idfPath:       idfPath,
		stopWordsPath: stopWordsPath,
	}

	j.updateGojieba()

	return j
}

func (j *Jieba) updateGojieba() {
	j.jieba = gojieba.NewJieba(j.dictPath, j.hmmPath, j.userDictPath, j.idfPath, j.stopWordsPath)
}

//更新用户词典
func (j *Jieba) updateDict(wordChannel <-chan string) error {
	err := os.Truncate(j.userDictPath, 0)
	if err != nil {
		log.Printf("truncate file err:%s", err.Error())
		return err
	}

	handle, err := os.OpenFile(j.userDictPath, os.O_APPEND|os.O_WRONLY, 077)
	if err != nil {
		log.Printf("read file err:%s", err.Error())
		return err
	}
	defer handle.Close()

	for word := range wordChannel {
		fmt.Fprintln(handle, j.formatCustomWord(word))
	}

	return nil
}

func (c *Jieba) Find(content string) ([]string) {
	words, _ := c.checkWord(content, false, "*")
	return words
}

func (c *Jieba) Replace(content string, replace string) (string) {
	_, content = c.checkWord(content, true, replace)
	return content
}

func (c *Jieba) UpdateAll(s store.Store) () {
	c.updateDict(s.ReadAll())
	c.updateGojieba()
}

//检查敏感词
func (j *Jieba) checkWord(content string, isReplace bool, replace string) ([]string, string) {
	badWordList := make([]string, 0)
	words := j.jieba.Tag(content)

	for _, word := range words {
		if j.isBadWord(word) {
			realWord := j.getRealWord(word)
			badWordList = append(badWordList, realWord)
			if isReplace {
				wordLen := utf8.RuneCountInString(realWord)
				content = strings.Replace(content, realWord, strings.Repeat(replace, wordLen), -1)
			}
		}
	}

	return badWordList, content
}

//格式化单词
func (j *Jieba) formatCustomWord(word string) (string) {
	return fmt.Sprintf("%s 10 %s", word, WORD_SPPECH)
}

func (j *Jieba) isBadWord(word string) bool {
	return strings.HasSuffix(word, "/"+WORD_SPPECH);
}

func (j *Jieba) getRealWord(word string) string {
	return strings.TrimRight(word, "/"+WORD_SPPECH)
}
