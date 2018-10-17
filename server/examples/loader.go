package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// exampleSetLoader is a loader for an example set directory structure.
type exampleSetLoader struct {
	ex   ExampleSet
	base string
}

// loadFileText loads a file as text relative to the base
func (esl *exampleSetLoader) loadFileText(path string) (string, error) {
	dat, err := ioutil.ReadFile(filepath.Join(esl.base, path))
	if err != nil {
		return "", err
	}
	return string(dat), nil
}

// loadTags loads a tags file (if one exists)
func (esl *exampleSetLoader) loadTags(path string) ([]Tag, error) {
	txt, err := esl.loadFileText(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	tags := make([]Tag, 0, strings.Count(txt, "\n"))
	for _, v := range strings.Split(txt, "\n") {
		if v == "" {
			continue
		}
		tags = append(tags, ParseTag(v))
	}

	return tags, nil
}

var langtbl = map[string]string{
	".sh":   "bash",
	".bash": "bash",
	".cpp":  "cpp",
	".c":    "cpp",
	".fs":   "forth",
	".fth":  "forth",
	".go":   "golang",
	".hs":   "haskell",
	".js":   "javascript",
	".lua":  "lua",
	".php":  "php",
	".py":   "python",
	".ts":   "typescript",
}

func (esl *exampleSetLoader) loadExample(path string) error {
	code, err := esl.loadFileText(path)
	if err != nil {
		return err
	}

	ext := filepath.Ext(path)
	lang, ok := langtbl[ext]
	if !ok {
		return fmt.Errorf("unrecognized lang with extension %q", ext)
	}

	name := strings.TrimSuffix(filepath.Base(path), ext)
	saname := sanitizeText(name)

	tags, err := esl.loadTags(path + ".tags")
	if err != nil {
		return err
	}

	esl.ex = append(esl.ex, Example{
		Path:     path,
		Name:     name,
		NameSan:  saname,
		Language: lang,
		Tags:     tags,
		Code:     code,
	})

	return nil
}

func (esl *exampleSetLoader) walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() || strings.HasSuffix(path, ".tags") {
		return nil
	}
	path, err = filepath.Rel(esl.base, path)
	if err != nil {
		return err
	}
	return esl.loadExample(path)
}

// LoadExampleSet loads an ExampleSet from a dir.
func LoadExampleSet(dir string) (ExampleSet, error) {
	esl := &exampleSetLoader{
		ex:   ExampleSet{},
		base: dir,
	}

	err := filepath.Walk(dir, esl.walk)
	if err != nil {
		return nil, err
	}

	return esl.ex, nil
}
