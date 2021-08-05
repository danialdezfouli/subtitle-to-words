package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/asticode/go-astisub"
	"github.com/bbalet/stopwords"
)

type Word struct {
	body string
	freq int
}

func IsFile(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !fileInfo.IsDir()
}
func GetFiles(root string) []string {
	var files []string

	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if IsFile(path) {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	return files

}
func getOutputDir(file string) string {
	dirpath := filepath.Join("./result", strings.TrimLeft(filepath.Dir(file), "srt/"))
	os.MkdirAll(dirpath, os.ModePerm)

	return dirpath
}

func get_words(content string) []string {
	var words []string
	stopwords.DontStripDigits()

	raw_words := stopwords.CleanString(content, "en", true)

	re_quotes := regexp.MustCompile(`\w+[']\w+`)
	raw_words = re_quotes.ReplaceAllString(raw_words, "")

	re_chars := regexp.MustCompile(`\s*\W+\s*`)
	raw_words = re_chars.ReplaceAllString(raw_words, " ")

	for _, word := range strings.Split(raw_words, " ") {
		if len(word) <= 2 {
			continue
		}

		re_num := regexp.MustCompile(`^\d+\w*`)
		if re_num.MatchString(word) {
			continue
		}

		words = append(words, word)
	}
	return words
}

func getSubtitleContent(path string) string {
	subtitle, _ := astisub.OpenFile(path)
	lines := make([]string, 0)

	for _, item := range subtitle.Items {
		line := item.String()
		if len(line) > 1 {
			lines = append(lines, line)
		}
	}
	content := strings.Join(lines, "\r\n")
	return content
}

func words_to_strings(words []Word) []string {
	var result []string

	for _, word := range words {
		result = append(result, fmt.Sprintf("%v, %v", word.body, word.freq))
	}

	return result
}

func unique_words_with_counter(_words []string) []Word {
	word_freq := make(map[string]int)
	var words []Word

	for _, word := range _words {
		word_freq[word]++
	}

	for keyword, freq := range word_freq {
		word := Word{
			body: keyword,
			freq: freq,
		}
		words = append(words, word)
	}

	sort.Slice(words, func(i, j int) bool {
		return words[j].freq < words[i].freq
	})

	return words
}

func SaveQuotes(text_output, content string) {
	l, err := os.Create(text_output)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	l.WriteString(content)
}

func SaveWords(path, content string) []string {

	w, err := os.Create(path)

	if err != nil {
		panic(err)
	}

	words := get_words(content)
	sorted_words := unique_words_with_counter(words)
	words = words_to_strings(sorted_words)

	w.WriteString(strings.Join(words, "\r\n"))

	return words
}

func main() {
	root := "./srt"
	files := GetFiles(root)

	for _, path := range files {
		if strings.HasSuffix(path, ".gitignore") {
			continue
		}

		dirpath := getOutputDir(path)

		filename := strings.TrimRight(filepath.Base(path), filepath.Ext(path))

		words_path := filepath.Join(dirpath, filename+"-words.txt")
		quotes_path := filepath.Join(dirpath, filename+".txt")

		content := getSubtitleContent(path)

		SaveQuotes(quotes_path, content)
		words := SaveWords(words_path, content)

		fmt.Printf("%v words: %v \n", len(words), path)
	}

}
