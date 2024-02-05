package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type orderedMap map[string]string

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Need exactly two arguments, got %d: '%s'\n", len(os.Args)-1, strings.Join(os.Args[1:], " "))
		fmt.Printf("%s <base lang> <new lang>\n", os.Args[0])
		os.Exit(-1)
	}

	base_file, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error on 'base lang' read: %v\n", err)
		os.Exit(-1)
	}
	new_file, err := os.ReadFile(os.Args[2])
	if err != nil {
		fmt.Printf("Error on 'new lang' read: %v\n", err)
		os.Exit(-1)
	}

	base_lang := make(orderedMap)
	new_lang := make(map[string]string)
	err = json.Unmarshal(base_file, &base_lang)
	if err != nil {
		fmt.Printf("Error on 'base lang' unmarshal: %v\n", err)
		os.Exit(-1)
	}
	err = json.Unmarshal(new_file, &new_lang)
	if err != nil {
		fmt.Printf("Error on 'new lang' unmarshal: %v\n", err)
		os.Exit(-1)
	}

	for k, v := range new_lang {
		base_lang[k] = v
	}

	base_file, err = json.MarshalIndent(base_lang, "", "  ")
	if err != nil {
		fmt.Printf("Error on json marshal: %v\n", err)
		os.Exit(-1)
	}

	base_file = bytes.ReplaceAll(base_file, []byte("\\u0026"), []byte{'\u0026'})
	base_file = bytes.ReplaceAll(base_file, []byte("\\u003c"), []byte{'\u003c'})
	base_file = bytes.ReplaceAll(base_file, []byte("\\u003e"), []byte{'\u003e'})
	base_file = append(base_file, '\n')

	err = os.WriteFile(os.Args[1], base_file, 0644)
	if err != nil {
		fmt.Printf("Error on save file: %v\n", err)
		os.Exit(-1)
	}

	fmt.Println("Done!")
}

func (om orderedMap) MarshalJSON() ([]byte, error) {
	keys := make([]string, 0, len(om))
	for k := range om {
		keys = append(keys, k)
	}

	numRegex := regexp.MustCompile("(\\d+)")
	slices.SortFunc(keys, func(a, b string) int {
		a = strings.ToLower(a)
		b = strings.ToLower(b)
		aFind := numRegex.FindStringIndex(a)
		bFind := numRegex.FindStringIndex(b)
		if aFind == nil || bFind == nil {
			return strings.Compare(a, b)
		}

		var (
			aBefore = a[:aFind[0]]
			bBefore = b[:bFind[0]]
			aInt, _ = strconv.Atoi(a[aFind[0]:aFind[1]])
			bInt, _ = strconv.Atoi(b[bFind[0]:bFind[1]])
			aAfter  = a[aFind[1]:]
			bAfter  = b[bFind[1]:]
		)

		if aBefore != bBefore {
			return strings.Compare(aBefore, bBefore)
		}
		if aInt < bInt {
			return -1
		}
		if aInt > bInt {
			return 1
		}
		return strings.Compare(aAfter, bAfter)
	})

	buf := bytes.NewBuffer([]byte{})
	buf.WriteByte('{')
	for i, k := range keys {
		if i != 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(fmt.Sprintf("\"%s\":", k))
		v, err := json.Marshal(om[k])
		if err != nil {
			return nil, err
		}
		buf.Write(v)
	}
	buf.WriteByte('}')

	return buf.Bytes(), nil
}
