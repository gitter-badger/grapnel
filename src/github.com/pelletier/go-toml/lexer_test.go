package toml

import "testing"

func testFlow(t *testing.T, input string, expectedFlow []token) {
	ch := lexToml(input)
	for _, expected := range expectedFlow {
		token := <-ch
		if token != expected {
			t.Log("While testing: ", input)
			t.Log("compared", token, "to", expected)
			t.Log(token.val, "<->", expected.val)
			t.Log(token.typ, "<->", expected.typ)
			t.Log(token.Line, "<->", expected.Line)
			t.Log(token.Col, "<->", expected.Col)
			t.FailNow()
		}
	}

	tok, ok := <-ch
	if ok {
		t.Log("channel is not closed!")
		t.Log(len(ch)+1, "tokens remaining:")

		t.Log("token ->", tok)
		for token := range ch {
			t.Log("token ->", token)
		}
		t.FailNow()
	}
}

func TestValidKeyGroup(t *testing.T) {
	testFlow(t, "[hello world]", []token{
		token{Position{1, 1}, tokenLeftBracket, "["},
		token{Position{1, 2}, tokenKeyGroup, "hello world"},
		token{Position{1, 13}, tokenRightBracket, "]"},
		token{Position{1, 14}, tokenEOF, ""},
	})
}

func TestUnclosedKeyGroup(t *testing.T) {
	testFlow(t, "[hello world", []token{
		token{Position{1, 1}, tokenLeftBracket, "["},
		token{Position{1, 2}, tokenError, "unclosed key group"},
	})
}

func TestComment(t *testing.T) {
	testFlow(t, "# blahblah", []token{
		token{Position{1, 11}, tokenEOF, ""},
	})
}

func TestKeyGroupComment(t *testing.T) {
	testFlow(t, "[hello world] # blahblah", []token{
		token{Position{1, 1}, tokenLeftBracket, "["},
		token{Position{1, 2}, tokenKeyGroup, "hello world"},
		token{Position{1, 13}, tokenRightBracket, "]"},
		token{Position{1, 25}, tokenEOF, ""},
	})
}

func TestMultipleKeyGroupsComment(t *testing.T) {
	testFlow(t, "[hello world] # blahblah\n[test]", []token{
		token{Position{1, 1}, tokenLeftBracket, "["},
		token{Position{1, 2}, tokenKeyGroup, "hello world"},
		token{Position{1, 13}, tokenRightBracket, "]"},
		token{Position{2, 1}, tokenLeftBracket, "["},
		token{Position{2, 2}, tokenKeyGroup, "test"},
		token{Position{2, 6}, tokenRightBracket, "]"},
		token{Position{2, 7}, tokenEOF, ""},
	})
}

func TestBasicKey(t *testing.T) {
	testFlow(t, "hello", []token{
		token{Position{1, 1}, tokenKey, "hello"},
		token{Position{1, 6}, tokenEOF, ""},
	})
}

func TestBasicKeyWithUnderscore(t *testing.T) {
	testFlow(t, "hello_hello", []token{
		token{Position{1, 1}, tokenKey, "hello_hello"},
		token{Position{1, 12}, tokenEOF, ""},
	})
}

func TestBasicKeyWithDash(t *testing.T) {
	testFlow(t, "hello-world", []token{
		token{Position{1, 1}, tokenKey, "hello-world"},
		token{Position{1, 12}, tokenEOF, ""},
	})
}

func TestBasicKeyWithUppercaseMix(t *testing.T) {
	testFlow(t, "helloHELLOHello", []token{
		token{Position{1, 1}, tokenKey, "helloHELLOHello"},
		token{Position{1, 16}, tokenEOF, ""},
	})
}

func TestBasicKeyWithInternationalCharacters(t *testing.T) {
	testFlow(t, "héllÖ", []token{
		token{Position{1, 1}, tokenKey, "héllÖ"},
		token{Position{1, 6}, tokenEOF, ""},
	})
}

func TestBasicKeyAndEqual(t *testing.T) {
	testFlow(t, "hello =", []token{
		token{Position{1, 1}, tokenKey, "hello"},
		token{Position{1, 7}, tokenEqual, "="},
		token{Position{1, 8}, tokenEOF, ""},
	})
}

func TestKeyWithSharpAndEqual(t *testing.T) {
	testFlow(t, "key#name = 5", []token{
		token{Position{1, 1}, tokenError, "keys cannot contain # character"},
	})
}

func TestKeyWithSymbolsAndEqual(t *testing.T) {
	testFlow(t, "~!@$^&*()_+-`1234567890[]\\|/?><.,;:' = 5", []token{
		token{Position{1, 1}, tokenKey, "~!@$^&*()_+-`1234567890[]\\|/?><.,;:'"},
		token{Position{1, 38}, tokenEqual, "="},
		token{Position{1, 40}, tokenInteger, "5"},
		token{Position{1, 41}, tokenEOF, ""},
	})
}

func TestKeyEqualStringEscape(t *testing.T) {
	testFlow(t, `foo = "hello\""`, []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 8}, tokenString, "hello\""},
		token{Position{1, 16}, tokenEOF, ""},
	})
}

func TestKeyEqualStringUnfinished(t *testing.T) {
	testFlow(t, `foo = "bar`, []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 8}, tokenError, "unclosed string"},
	})
}

func TestKeyEqualString(t *testing.T) {
	testFlow(t, `foo = "bar"`, []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 8}, tokenString, "bar"},
		token{Position{1, 12}, tokenEOF, ""},
	})
}

func TestKeyEqualTrue(t *testing.T) {
	testFlow(t, "foo = true", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenTrue, "true"},
		token{Position{1, 11}, tokenEOF, ""},
	})
}

func TestKeyEqualFalse(t *testing.T) {
	testFlow(t, "foo = false", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenFalse, "false"},
		token{Position{1, 12}, tokenEOF, ""},
	})
}

func TestArrayNestedString(t *testing.T) {
	testFlow(t, `a = [ ["hello", "world"] ]`, []token{
		token{Position{1, 1}, tokenKey, "a"},
		token{Position{1, 3}, tokenEqual, "="},
		token{Position{1, 5}, tokenLeftBracket, "["},
		token{Position{1, 7}, tokenLeftBracket, "["},
		token{Position{1, 9}, tokenString, "hello"},
		token{Position{1, 15}, tokenComma, ","},
		token{Position{1, 18}, tokenString, "world"},
		token{Position{1, 24}, tokenRightBracket, "]"},
		token{Position{1, 26}, tokenRightBracket, "]"},
		token{Position{1, 27}, tokenEOF, ""},
	})
}

func TestArrayNestedInts(t *testing.T) {
	testFlow(t, "a = [ [42, 21], [10] ]", []token{
		token{Position{1, 1}, tokenKey, "a"},
		token{Position{1, 3}, tokenEqual, "="},
		token{Position{1, 5}, tokenLeftBracket, "["},
		token{Position{1, 7}, tokenLeftBracket, "["},
		token{Position{1, 8}, tokenInteger, "42"},
		token{Position{1, 10}, tokenComma, ","},
		token{Position{1, 12}, tokenInteger, "21"},
		token{Position{1, 14}, tokenRightBracket, "]"},
		token{Position{1, 15}, tokenComma, ","},
		token{Position{1, 17}, tokenLeftBracket, "["},
		token{Position{1, 18}, tokenInteger, "10"},
		token{Position{1, 20}, tokenRightBracket, "]"},
		token{Position{1, 22}, tokenRightBracket, "]"},
		token{Position{1, 23}, tokenEOF, ""},
	})
}

func TestArrayInts(t *testing.T) {
	testFlow(t, "a = [ 42, 21, 10, ]", []token{
		token{Position{1, 1}, tokenKey, "a"},
		token{Position{1, 3}, tokenEqual, "="},
		token{Position{1, 5}, tokenLeftBracket, "["},
		token{Position{1, 7}, tokenInteger, "42"},
		token{Position{1, 9}, tokenComma, ","},
		token{Position{1, 11}, tokenInteger, "21"},
		token{Position{1, 13}, tokenComma, ","},
		token{Position{1, 15}, tokenInteger, "10"},
		token{Position{1, 17}, tokenComma, ","},
		token{Position{1, 19}, tokenRightBracket, "]"},
		token{Position{1, 20}, tokenEOF, ""},
	})
}

func TestMultilineArrayComments(t *testing.T) {
	testFlow(t, "a = [1, # wow\n2, # such items\n3, # so array\n]", []token{
		token{Position{1, 1}, tokenKey, "a"},
		token{Position{1, 3}, tokenEqual, "="},
		token{Position{1, 5}, tokenLeftBracket, "["},
		token{Position{1, 6}, tokenInteger, "1"},
		token{Position{1, 7}, tokenComma, ","},
		token{Position{2, 1}, tokenInteger, "2"},
		token{Position{2, 2}, tokenComma, ","},
		token{Position{3, 1}, tokenInteger, "3"},
		token{Position{3, 2}, tokenComma, ","},
		token{Position{4, 1}, tokenRightBracket, "]"},
		token{Position{4, 2}, tokenEOF, ""},
	})
}

func TestKeyEqualArrayBools(t *testing.T) {
	testFlow(t, "foo = [true, false, true]", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenLeftBracket, "["},
		token{Position{1, 8}, tokenTrue, "true"},
		token{Position{1, 12}, tokenComma, ","},
		token{Position{1, 14}, tokenFalse, "false"},
		token{Position{1, 19}, tokenComma, ","},
		token{Position{1, 21}, tokenTrue, "true"},
		token{Position{1, 25}, tokenRightBracket, "]"},
		token{Position{1, 26}, tokenEOF, ""},
	})
}

func TestKeyEqualArrayBoolsWithComments(t *testing.T) {
	testFlow(t, "foo = [true, false, true] # YEAH", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenLeftBracket, "["},
		token{Position{1, 8}, tokenTrue, "true"},
		token{Position{1, 12}, tokenComma, ","},
		token{Position{1, 14}, tokenFalse, "false"},
		token{Position{1, 19}, tokenComma, ","},
		token{Position{1, 21}, tokenTrue, "true"},
		token{Position{1, 25}, tokenRightBracket, "]"},
		token{Position{1, 33}, tokenEOF, ""},
	})
}

func TestDateRegexp(t *testing.T) {
	if dateRegexp.FindString("1979-05-27T07:32:00Z") == "" {
		t.Error("basic lexing")
	}
	if dateRegexp.FindString("1979-05-27T00:32:00-07:00") == "" {
		t.Error("offset lexing")
	}
	if dateRegexp.FindString("1979-05-27T00:32:00.999999-07:00") == "" {
		t.Error("nano precision lexing")
	}
}

func TestKeyEqualDate(t *testing.T) {
	testFlow(t, "foo = 1979-05-27T07:32:00Z", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenDate, "1979-05-27T07:32:00Z"},
		token{Position{1, 27}, tokenEOF, ""},
	})
	testFlow(t, "foo = 1979-05-27T00:32:00-07:00", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenDate, "1979-05-27T00:32:00-07:00"},
		token{Position{1, 32}, tokenEOF, ""},
	})
	testFlow(t, "foo = 1979-05-27T00:32:00.999999-07:00", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenDate, "1979-05-27T00:32:00.999999-07:00"},
		token{Position{1, 39}, tokenEOF, ""},
	})
}

func TestFloatEndingWithDot(t *testing.T) {
	testFlow(t, "foo = 42.", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenError, "float cannot end with a dot"},
	})
}

func TestFloatWithTwoDots(t *testing.T) {
	testFlow(t, "foo = 4.2.", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenError, "cannot have two dots in one float"},
	})
}

func TestFloatWithExponent1(t *testing.T) {
	testFlow(t, "a = 5e+22", []token{
		token{Position{1, 1}, tokenKey, "a"},
		token{Position{1, 3}, tokenEqual, "="},
		token{Position{1, 5}, tokenFloat, "5e+22"},
		token{Position{1, 10}, tokenEOF, ""},
	})
}

func TestFloatWithExponent2(t *testing.T) {
	testFlow(t, "a = 5E+22", []token{
		token{Position{1, 1}, tokenKey, "a"},
		token{Position{1, 3}, tokenEqual, "="},
		token{Position{1, 5}, tokenFloat, "5E+22"},
		token{Position{1, 10}, tokenEOF, ""},
	})
}

func TestFloatWithExponent3(t *testing.T) {
	testFlow(t, "a = -5e+22", []token{
		token{Position{1, 1}, tokenKey, "a"},
		token{Position{1, 3}, tokenEqual, "="},
		token{Position{1, 5}, tokenFloat, "-5e+22"},
		token{Position{1, 11}, tokenEOF, ""},
	})
}

func TestFloatWithExponent4(t *testing.T) {
	testFlow(t, "a = -5e-22", []token{
		token{Position{1, 1}, tokenKey, "a"},
		token{Position{1, 3}, tokenEqual, "="},
		token{Position{1, 5}, tokenFloat, "-5e-22"},
		token{Position{1, 11}, tokenEOF, ""},
	})
}

func TestFloatWithExponent5(t *testing.T) {
	testFlow(t, "a = 6.626e-34", []token{
		token{Position{1, 1}, tokenKey, "a"},
		token{Position{1, 3}, tokenEqual, "="},
		token{Position{1, 5}, tokenFloat, "6.626e-34"},
		token{Position{1, 14}, tokenEOF, ""},
	})
}

func TestDoubleEqualKey(t *testing.T) {
	testFlow(t, "foo= = 2", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 4}, tokenEqual, "="},
		token{Position{1, 5}, tokenError, "cannot have multiple equals for the same key"},
	})
}

func TestInvalidEsquapeSequence(t *testing.T) {
	testFlow(t, `foo = "\x"`, []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 8}, tokenError, "invalid escape sequence: \\x"},
	})
}

func TestNestedArrays(t *testing.T) {
	testFlow(t, "foo = [[[]]]", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenLeftBracket, "["},
		token{Position{1, 8}, tokenLeftBracket, "["},
		token{Position{1, 9}, tokenLeftBracket, "["},
		token{Position{1, 10}, tokenRightBracket, "]"},
		token{Position{1, 11}, tokenRightBracket, "]"},
		token{Position{1, 12}, tokenRightBracket, "]"},
		token{Position{1, 13}, tokenEOF, ""},
	})
}

func TestKeyEqualNumber(t *testing.T) {
	testFlow(t, "foo = 42", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenInteger, "42"},
		token{Position{1, 9}, tokenEOF, ""},
	})

	testFlow(t, "foo = +42", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenInteger, "+42"},
		token{Position{1, 10}, tokenEOF, ""},
	})

	testFlow(t, "foo = -42", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenInteger, "-42"},
		token{Position{1, 10}, tokenEOF, ""},
	})

	testFlow(t, "foo = 4.2", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenFloat, "4.2"},
		token{Position{1, 10}, tokenEOF, ""},
	})

	testFlow(t, "foo = +4.2", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenFloat, "+4.2"},
		token{Position{1, 11}, tokenEOF, ""},
	})

	testFlow(t, "foo = -4.2", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenFloat, "-4.2"},
		token{Position{1, 11}, tokenEOF, ""},
	})
}

func TestMultiline(t *testing.T) {
	testFlow(t, "foo = 42\nbar=21", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 7}, tokenInteger, "42"},
		token{Position{2, 1}, tokenKey, "bar"},
		token{Position{2, 4}, tokenEqual, "="},
		token{Position{2, 5}, tokenInteger, "21"},
		token{Position{2, 7}, tokenEOF, ""},
	})
}

func TestKeyEqualStringUnicodeEscape(t *testing.T) {
	testFlow(t, `foo = "hello \u2665"`, []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 8}, tokenString, "hello ♥"},
		token{Position{1, 21}, tokenEOF, ""},
	})
	testFlow(t, `foo = "hello \U000003B4"`, []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 8}, tokenString, "hello δ"},
		token{Position{1, 25}, tokenEOF, ""},
	})
}

func TestLiteralString(t *testing.T) {
	testFlow(t, `foo = 'C:\Users\nodejs\templates'`, []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 8}, tokenString, `C:\Users\nodejs\templates`},
		token{Position{1, 34}, tokenEOF, ""},
	})
	testFlow(t, `foo = '\\ServerX\admin$\system32\'`, []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 8}, tokenString, `\\ServerX\admin$\system32\`},
		token{Position{1, 35}, tokenEOF, ""},
	})
	testFlow(t, `foo = 'Tom "Dubs" Preston-Werner'`, []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 8}, tokenString, `Tom "Dubs" Preston-Werner`},
		token{Position{1, 34}, tokenEOF, ""},
	})
	testFlow(t, `foo = '<\i\c*\s*>'`, []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 8}, tokenString, `<\i\c*\s*>`},
		token{Position{1, 19}, tokenEOF, ""},
	})
}

func TestMultilineLiteralString(t *testing.T) {
	testFlow(t, `foo = '''hello 'literal' world'''`, []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 10}, tokenString, `hello 'literal' world`},
		token{Position{1, 34}, tokenEOF, ""},
	})

	testFlow(t, "foo = '''\nhello\n'literal'\nworld'''", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{2, 1}, tokenString, "hello\n'literal'\nworld"},
		token{Position{4, 9}, tokenEOF, ""},
	})
}

func TestMultilineString(t *testing.T) {
	testFlow(t, `foo = """hello "literal" world"""`, []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 10}, tokenString, `hello "literal" world`},
		token{Position{1, 34}, tokenEOF, ""},
	})

	testFlow(t, "foo = \"\"\"\nhello\n\"literal\"\nworld\"\"\"", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{2, 1}, tokenString, "hello\n\"literal\"\nworld"},
		token{Position{4, 9}, tokenEOF, ""},
	})

	testFlow(t, "foo = \"\"\"\\\n    \\\n    \\\n    hello\nmultiline\nworld\"\"\"", []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 10}, tokenString, "hello\nmultiline\nworld"},
		token{Position{6, 9}, tokenEOF, ""},
	})
}

func TestUnicodeString(t *testing.T) {
	testFlow(t, `foo = "hello ♥ world"`, []token{
		token{Position{1, 1}, tokenKey, "foo"},
		token{Position{1, 5}, tokenEqual, "="},
		token{Position{1, 8}, tokenString, "hello ♥ world"},
		token{Position{1, 22}, tokenEOF, ""},
	})
}

func TestKeyGroupArray(t *testing.T) {
	testFlow(t, "[[foo]]", []token{
		token{Position{1, 1}, tokenDoubleLeftBracket, "[["},
		token{Position{1, 3}, tokenKeyGroupArray, "foo"},
		token{Position{1, 6}, tokenDoubleRightBracket, "]]"},
		token{Position{1, 8}, tokenEOF, ""},
	})
}
