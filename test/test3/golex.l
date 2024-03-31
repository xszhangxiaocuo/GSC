%{
package main

import (
		"fmt"
    	"io/ioutil"
    	"os"
)

var (
	c    byte
	src  string
	in   []byte
	un   []byte
	mark int
	charCount int
    lineCount int = 1
)

func lex() (s string) {
%}

%yyn next()
%yyc c
%yym mark = len(in)

%%
	in = in[:0]
	mark = -1

\0
	return "EOF"

[_a-zA-Z][_a-zA-Z0-9]*
	return fmt.Sprintf("IDENTIFIER(%q)", in)
[0-9]+
    return fmt.Sprintf("INTEGER(%q)", in)

%%
	if mark >= 0 {
		if len(in) > mark {
			unget(c)
			for i := len(in)-1; i >= mark; i-- {
				unget(in[i])
			}
			next()
		}
		in = in[:mark]
		goto yyAction // Hook: Execute the semantic action of the last matched rule.
	}

	switch n := len(in); n {
	case 0: // [] z
		s = fmt.Sprintf("%q", c)
		next()
	case 1: // [x] z
		s = fmt.Sprintf("%q", in[0])
	default: // [x, y, ...], z
		s = fmt.Sprintf("%q", in[0])
		unget(c) // z
		for i := n - 1; i > 1; i-- {
			unget(in[i]) // ...
		}
		c = in[1] // y
	}
	return s
}

func next() {
	if len(un) != 0 {
		c = un[len(un)-1]
		un = un[:len(un)-1]
		return
	}

	in = append(in, c)
	charCount++
    	if c == '\n' {
    		lineCount++
    	}
	if len(src) == 0 {
		c = 0
		return
	}

	c = src[0]
	src = src[1:]
}

func unget(b byte) {
	un = append(un, b)
}

func main() {
    		file, err := os.Open("test.txt")
    		if err != nil {
    			fmt.Printf("Error opening file: %v\n", err)
    			return
    		}
    		defer file.Close()

    		content, err := ioutil.ReadAll(file)
    		if err != nil {
    			fmt.Printf("Error reading file: %v\n", err)
    			return
    		}

    		src = string(content)

    	next()
    	for {
    		s := lex()
    		if s!="'\\n'"&&s!="EOF"{
    		fmt.Println(s)
    		}
    		if s == "EOF" {
    			break
    		}
    	}
	fmt.Printf("Character count: %d\n", charCount)
    fmt.Printf("Line count: %d\n", lineCount)
}