package interpreter_test

import (
	"bytes"
	"errors"
	"testing"

	it "github.com/electro98/gobf/interpreter"
)

func prepareInterpreter(t *testing.T, memorySize uint, code string) *it.Interpreter {
	interpreter, err := it.NewInterpreter().
		SetMemorySizeSource(memorySize).
		SetStringSource(code).
		Finalize()
	if err != nil {
		t.Fatal("Failed to prepare interpreter -", err)
	}
	return interpreter
}

func toBufferPutRune() (*bytes.Buffer, func(rune)) {
	arr := bytes.Buffer{}
	return &arr, func(r rune) {
		arr.WriteRune(r)
	}
}

func noopGetRune() rune {
	return rune(0)
}

var boundaryTests = []struct {
	name       string
	code       string
	memorySize uint
	out        string
}{
	{"Right Boundary Check", "+[>+++++++++++++++++++++++++++++++++.]", 8, "!!!!!!!"},
	{"Left Boundary Check", "+[<+++++++++++++++++++++++++++++++++.]", 8, ""},
}

func TestInterpreterMemoryRange(t *testing.T) {
	for _, tt := range boundaryTests {
		t.Run(tt.name, func(t *testing.T) {
			i := prepareInterpreter(t, tt.memorySize, tt.code)
			buff, putRune := toBufferPutRune()
			err := i.Execute(noopGetRune, putRune)
			if err != nil {
				if !errors.Is(err, it.MemoryOutOfBound) {
					t.Errorf("Got unexpected error - %s", err)
				}
			} else {
				t.Error("Execution should have failed! Memory boundary breached!")
			}
			if out := buff.String(); out != tt.out {
				t.Errorf("got '%s', expected '%s'", out, tt.out)
			}
		})
	}
}

func TestInterpreterObscure(t *testing.T) {
	const code = "[]++++++++++[>>+>+>++++++[<<+<+++>>>-]<<<<-]\"A*$\";?@![#>>+<<]>[>>]<<<<[>++<[-]]>.>."
	const expOut = "H\n"
	i := prepareInterpreter(t, 3000, code)
	buff, putRune := toBufferPutRune()
	err := i.Execute(noopGetRune, putRune)
	if err != nil {
		t.Errorf("Got unexpected error - %s", err)
	}
	if out := buff.String(); out != expOut {
		t.Errorf("got '%s', expected '%s'", out, expOut)
		t.Errorf("buffer: %d", buff.Bytes())
	}
}
