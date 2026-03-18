package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"unicode/utf8"

	"github.com/electro98/gobf/interpreter"

	"github.com/urfave/cli/v3"
)

type inputArguments struct {
	input         string
	inputAsIs     bool
	mode          interpreterModeArg
	bufferedInput string
	outputFile    string
}

type interpreterMode uint8

const (
	interactiveMode interpreterMode = iota
	bufferedMode
	immediateMode
)

type interpreterModeArg struct {
	interactive bool
	buffered    bool
	immediate   bool
}

func (modeArg interpreterModeArg) getMode() interpreterMode {
	switch modeArg {
	case interpreterModeArg{interactive: true}:
		return interactiveMode
	case interpreterModeArg{buffered: true}:
		return bufferedMode
	case interpreterModeArg{immediate: true}:
		return immediateMode
	default:
		return interactiveMode
	}
}

func RunCLI() {
	inputArgs := inputArguments{}
	cmd := &cli.Command{
		Name:  "bf-cli",
		Usage: "small brainfuck interpreter",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:        "input",
				UsageText:   "File containing bf code or code itself",
				Destination: &inputArgs.input,
			},
			&cli.StringArg{
				Name:        "program input",
				UsageText:   "Input for the program in buffered mode",
				Destination: &inputArgs.bufferedInput,
			},
		},
		ArgsUsage: "file/code [buffered program input]",
		MutuallyExclusiveFlags: []cli.MutuallyExclusiveFlags{
			{
				Flags: [][]cli.Flag{
					{
						&cli.BoolFlag{
							Name:        "interactive",
							Aliases:     []string{"i"},
							Usage:       "Provide user input when interpreted program expects it",
							Destination: &inputArgs.mode.interactive,
						},
					},
					{
						&cli.BoolFlag{
							Name:        "buffered",
							Aliases:     []string{"b"},
							Usage:       "Use the second argument as program input (null-terminated)",
							Destination: &inputArgs.mode.buffered,
						},
					},
					{
						&cli.BoolFlag{
							Name:        "immediate",
							Aliases:     []string{"m"},
							Usage:       "When program expects input send single key-presses",
							Destination: &inputArgs.mode.immediate,
						},
					},
				},
				Category: "Mode Selection [Default - interactive]",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{"o"},
				Value:       "",
				Usage:       "Duplicate output in `FILE`",
				Destination: &inputArgs.outputFile,
			},
			&cli.BoolFlag{
				Name:        "as-code",
				Aliases:     []string{"s"},
				Value:       false,
				Usage:       "Use the first argument as code itself",
				Destination: &inputArgs.inputAsIs,
			},
		},
		Action: func(context.Context, *cli.Command) error {
			if err := validateArgs(&inputArgs); err != nil {
				return err
			}
			return interpretAction(&inputArgs)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func validateArgs(inputArgs *inputArguments) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Code panicked, sorry, probably bug: %s", r)
		}
	}()
	if len(inputArgs.input) == 0 {
		if inputArgs.inputAsIs {
			return fmt.Errorf("`as-code` flag set, but no code provided!")
		} else {
			return fmt.Errorf("No input file is provided")
		}
	}
	if len(inputArgs.outputFile) != 0 {
		panic("Output files are not yet implemented")
	}
	return nil

}

func interpretAction(inputArgs *inputArguments) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Code panicked, sorry, probably bug: %s", r)
		}
	}()
	getRune, putRune := getInputOutputMethod(inputArgs.mode.getMode(), &inputArgs.bufferedInput)
	itBuilder := interpreter.NewInterpreter()
	if inputArgs.inputAsIs {
		itBuilder.SetStringSource(inputArgs.input)
	} else {
		runeStream, fileSize, err := getRunesFromFile(inputArgs.input)
		if err != nil {
			return err
		}
		itBuilder.SetRuneStreamSource(runeStream, int(fileSize/2))
	}

	it, err := itBuilder.Finalize()
	if err != nil {
		return err
	}
	if err := it.Execute(getRune, putRune); err != nil {
		fmt.Println()
		return err
	}

	fmt.Println("\n ~ Finished execution ~ ")

	return nil
}

func getInputOutputMethod(mode interpreterMode, bufferedInput *string) (getRune func() rune, putRune func(rune)) {
	putRune = func(r rune) {
		fmt.Printf("%c", r)
	}
	switch mode {
	case interactiveMode:
		var userInput string
		getRune = func() rune {
			for len(userInput) == 0 {
				fmt.Scanln(&userInput)
			}
			r, size := utf8.DecodeRune([]byte(userInput))
			userInput = userInput[size:]
			return r
		}
	case immediateMode:
		panic("Not implemented")
	case bufferedMode:
		i := 0
		getRune = func() rune {
			if i <= len(*bufferedInput) {
				return 0
			}
			r, size := utf8.DecodeRune([]byte(*bufferedInput)[i:])
			i += size
			return r
		}
	}
	return
}
