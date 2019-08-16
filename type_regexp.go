package otto

import (
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/robertkrimen/otto/parser"
	"github.com/dlclark/regexp2"
)

type _regExpObject struct {
	regularExpression *regexp.Regexp
	regularExpression2 *regexp2.Regexp
	global            bool
	ignoreCase        bool
	multiline         bool
	source            string
	flags             string
}

func (runtime *_runtime) newRegExpObject(pattern string, flags string) *_object {
	self := runtime.newObject()
	self.class = "RegExp"

	global := false
	ignoreCase := false
	multiline := false
	re2flags := ""

	// TODO Maybe clean up the panicking here... TypeError, SyntaxError, ?

	for _, chr := range flags {
		switch chr {
		case 'g':
			if global {
				panic(runtime.panicSyntaxError("newRegExpObject: %s %s", pattern, flags))
			}
			global = true
		case 'm':
			if multiline {
				panic(runtime.panicSyntaxError("newRegExpObject: %s %s", pattern, flags))
			}
			multiline = true
			re2flags += "m"
		case 'i':
			if ignoreCase {
				panic(runtime.panicSyntaxError("newRegExpObject: %s %s", pattern, flags))
			}
			ignoreCase = true
			re2flags += "i"
		}
	}

	re2pattern, err := parser.TransformRegExp(pattern)
	if err != nil {
		// panic(runtime.panicTypeError("Invalid regular expression: %s", err.Error()))
	}
	if len(re2flags) > 0 {
		re2pattern = fmt.Sprintf("(?%s:%s)", re2flags, re2pattern)
	}

	// fmt.Printf("--------\n")
	// fmt.Printf("%s\n", re2pattern)
	// fmt.Printf("%s\n", re2flags)
	// fmt.Printf("%s\n", pattern)
	// fmt.Printf("%s\n", flags)

	regularExpression, err := regexp.Compile(pattern)
	if err != nil {
		// panic(runtime.panicTypeError("[1] Invalid regular expression: %s", err.Error()))
	}
	if len(re2flags) > 0 {
		re2pattern = fmt.Sprintf("(?%s:%s)", re2flags, re2pattern)
	}
	
	regularExpression2, err := regexp2.Compile(pattern, 0x0100)
	if err != nil {
		panic(runtime.panicSyntaxError("[2] Invalid regular expression: %s", err.Error()[22:]))
	}

	self.value = _regExpObject{
		regularExpression: regularExpression,
		regularExpression2: regularExpression2,
		global:            global,
		ignoreCase:        ignoreCase,
		multiline:         multiline,
		source:            pattern,
		flags:             flags,
	}
	self.defineProperty("global", toValue_bool(global), 0, false)
	self.defineProperty("ignoreCase", toValue_bool(ignoreCase), 0, false)
	self.defineProperty("multiline", toValue_bool(multiline), 0, false)
	self.defineProperty("lastIndex", toValue_int(0), 0100, false)
	self.defineProperty("source", toValue_string(pattern), 0, false)
	return self
}

func (self *_object) regExpValue() _regExpObject {
	value, _ := self.value.(_regExpObject)
	return value
}

func execRegExp(this *_object, target string) (match bool, result []int) {
	if this.class != "RegExp" {
		panic(this.runtime.panicTypeError("Calling RegExp.exec on a non-RegExp object"))
	}
	lastIndex := this.get("lastIndex").number().int64
	index := lastIndex
	global := this.get("global").bool()
	if !global {
		index = 0
	}
	if 0 > index || index > int64(len(target)) {
	} else {
		match_, err := this.regExpValue().regularExpression2.FindStringMatch(target[index:])
		if err != nil {
			fmt.Printf("bad news")
		}
		groups := match_.Groups()
		var result2 []int
		leng := len(groups)
		result2 = make([]int, 2*leng)
		for i := 0; i < leng; i++ {
			result2[2*i] = groups[i].Index
			result2[2*i+1] = groups[i].Index + groups[i].Length
		}

		result = this.regExpValue().regularExpression.FindStringSubmatchIndex(target[index:])
		fmt.Printf("compare regex results\n")
		fmt.Printf("result: %s\n", result)
		fmt.Printf("result2: %s\n", result2)
	}
	if result == nil {
		//this.defineProperty("lastIndex", toValue_(0), 0111, true)
		this.put("lastIndex", toValue_int(0), true)
		return // !match
	}
	match = true
	startIndex := index
	endIndex := int(lastIndex) + result[1]
	// We do this shift here because the .FindStringSubmatchIndex above
	// was done on a local subordinate slice of the string, not the whole string
	for index, _ := range result {
		result[index] += int(startIndex)
	}
	if global {
		//this.defineProperty("lastIndex", toValue_(endIndex), 0111, true)
		this.put("lastIndex", toValue_int(endIndex), true)
	}
	return // match
}

func execResultToArray(runtime *_runtime, target string, result []int) *_object {
	captureCount := len(result) / 2
	valueArray := make([]Value, captureCount)
	for index := 0; index < captureCount; index++ {
		offset := 2 * index
		if result[offset] != -1 {
			valueArray[index] = toValue_string(target[result[offset]:result[offset+1]])
		} else {
			valueArray[index] = Value{}
		}
	}
	matchIndex := result[0]
	if matchIndex != 0 {
		matchIndex = 0
		// Find the rune index in the string, not the byte index
		for index := 0; index < result[0]; {
			_, size := utf8.DecodeRuneInString(target[index:])
			matchIndex += 1
			index += size
		}
	}
	match := runtime.newArrayOf(valueArray)
	match.defineProperty("input", toValue_string(target), 0111, false)
	match.defineProperty("index", toValue_int(matchIndex), 0111, false)
	return match
}
