package interpreter

import (
	"errors"
	"fmt"
	"iter"
)

type Interpreter struct {
	code      []Instruction
	codePtr   uint
	memory    []uint8
	memoryPtr int
}

type Instruction uint8

const (
	_                = iota
	Plus Instruction = iota
	Minus
	MemPlus
	MemMinus
	StartLoop
	EndLoop
	Output
	Input
)

const DefaultMemorySize = 3000

var MemoryOutOfBound = errors.New("Not in allocated memory range")
var LoopNotClosed = errors.New("Loop is not properly closed!")

type interpreterBuilder struct {
	source struct {
		code          string
		runes         iter.Seq[rune]
		estimatedSize int
		set           bool
	}
	memorySize int
	err        error
}

var SourceAlreadySet = errors.New("Source already set!")
var UnfinishedBuilder = errors.New("Builder could not finish")

func NewInterpreter() *interpreterBuilder {
	return &interpreterBuilder{}
}

func (builder *interpreterBuilder) SetStringSource(code string) *interpreterBuilder {
	if builder.source.set {
		builder.err = SourceAlreadySet
		return builder
	}
	builder.source.code = code
	builder.source.set = true
	return builder
}

func (builder *interpreterBuilder) SetRuneStreamSource(runes iter.Seq[rune], estimatedSize int) *interpreterBuilder {
	if builder.source.set {
		builder.err = SourceAlreadySet
		return builder
	}
	builder.source.runes = runes
	builder.source.estimatedSize = estimatedSize
	builder.source.set = true
	return builder
}

func (builder *interpreterBuilder) SetMemorySizeSource(size uint) *interpreterBuilder {
	builder.memorySize = int(size)
	return builder
}

func (builder *interpreterBuilder) Finalize() (*Interpreter, error) {
	if builder.err != nil {
		return nil, fmt.Errorf("Builder failed: %w", builder.err)
	}
	if !builder.source.set {
		return nil, fmt.Errorf("Source is not set: %w", UnfinishedBuilder)
	}
	var code []Instruction
	if builder.source.runes != nil {
		code = runeStreamToInstructions(builder.source.runes, builder.source.estimatedSize)
	} else {
		code = codeToInstructions(builder.source.code)
	}
	if builder.memorySize <= 0 {
		builder.memorySize = DefaultMemorySize
	}
	return &Interpreter{
		code:   code,
		memory: make([]uint8, builder.memorySize),
	}, nil
}

func (it *Interpreter) Execute(getRune func() rune, putRune func(rune)) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Tried to access 0x%x: %w", it.memoryPtr, MemoryOutOfBound)
		}
	}()
	loopStack := make([]uint, 0, 16)
	for it.codePtr < uint(len(it.code)) {
		instruction := it.code[it.codePtr]
		switch instruction {
		case Plus:
			it.memory[it.memoryPtr]++
		case Minus:
			it.memory[it.memoryPtr]--
		case MemPlus:
			it.memoryPtr++
		case MemMinus:
			it.memoryPtr--
		case StartLoop:
			if it.memory[it.memoryPtr] > 0 {
				loopStack = append(loopStack, it.codePtr)
			} else {
				loopEnd, err := it.findLoopEnd()
				if err != nil {
					return fmt.Errorf("Missing complementary ']' char: %w", err)
				}
				it.codePtr = loopEnd
			}
		case EndLoop:
			if len(loopStack) == 0 {
				return fmt.Errorf("Missing complementary '[' char: %w", LoopNotClosed)
			}
			if it.memory[it.memoryPtr] > 0 {
				it.codePtr = loopStack[len(loopStack)-1]
			} else {
				loopStack[len(loopStack)-1] = 0 // Maybe unnecessary
				loopStack = loopStack[:len(loopStack)-1]
			}
		case Output:
			putRune(rune(it.memory[it.memoryPtr]))
		case Input:
			it.memory[it.memoryPtr] = uint8(getRune())
		}
		it.codePtr++
	}
	return nil
}

func (it *Interpreter) findLoopEnd() (uint, error) {
	loopDepth := 0
	for codePtr := it.codePtr + 1; codePtr < uint(len(it.code)); codePtr++ {
		instruction := it.code[codePtr]
		switch instruction {
		case StartLoop:
			loopDepth++
		case EndLoop:
			if loopDepth == 0 {
				return codePtr, nil
			}
			loopDepth--
		}
	}
	return 0, LoopNotClosed
}

func codeToInstructions(code string) []Instruction {
	charToInst := map[rune]Instruction{
		'+': Plus,
		'-': Minus,
		'>': MemPlus,
		'<': MemMinus,
		'[': StartLoop,
		']': EndLoop,
		'.': Output,
		',': Input,
	}

	instructions := make([]Instruction, 0, len(code))
	for _, char := range code {
		instruction, ok := charToInst[char]
		if ok {
			instructions = append(instructions, instruction)
		}
	}
	trimmedInstructions := make([]Instruction, len(instructions))
	copy(trimmedInstructions, instructions)
	return trimmedInstructions
}

func runeStreamToInstructions(code iter.Seq[rune], estimatedSize int) []Instruction {
	charToInst := map[rune]Instruction{
		'+': Plus,
		'-': Minus,
		'>': MemPlus,
		'<': MemMinus,
		'[': StartLoop,
		']': EndLoop,
		'.': Output,
		',': Input,
	}

	instructions := make([]Instruction, 0, estimatedSize)
	for char := range code {
		instruction, ok := charToInst[char]
		if ok {
			instructions = append(instructions, instruction)
		}
	}
	trimmedInstructions := make([]Instruction, len(instructions))
	copy(trimmedInstructions, instructions)
	return trimmedInstructions
}

func (i Instruction) String() string {
	instToChar := map[Instruction]string{
		Plus:      "+",
		Minus:     "-",
		MemPlus:   ">",
		MemMinus:  "<",
		StartLoop: "[",
		EndLoop:   "]",
		Output:    ".",
		Input:     ",",
	}
	return instToChar[i]
}
