package tests

import (
	"fmt"
	"kano/internal/utils/parser"
	"testing"
)

type ArgTest struct {
	Name        string
	Input       string
	Expected    parser.ParseResult
	ShouldError bool
}

func isContentSame(a, b parser.Content) error {
	if a.Data != b.Data {
		return fmt.Errorf("expected content data %q, got %q", b.Data, a.Data)
	}
	if a.Start != b.Start {
		return fmt.Errorf("expected content start pos %d, got %d", b.Start, a.Start)
	}
	if a.End != b.End {
		return fmt.Errorf("expected content end pos %d, got %d", b.End, a.End)
	}
	return nil
}

func isArgSame(a, b parser.Argument) error {
	if err := isContentSame(a.Content, b.Content); err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	if a.Start != b.Start {
		return fmt.Errorf("expected arg start pos %d, got %d", b.Start, a.Start)
	}
	if a.End != b.End {
		return fmt.Errorf("expected arg end pos %d, got %d", b.End, a.End)
	}
	if a.InsideQuote != b.InsideQuote {
		return fmt.Errorf("expected inside quote is %t, got %t", b.InsideQuote, a.InsideQuote)
	}
	if a.InsideQuote {
		if a.UsedQuote != b.UsedQuote {
			return fmt.Errorf("expected used quote is %d, got %d", b.UsedQuote, a.UsedQuote)
		}
	}
	return nil
}

func isCommandSame(a, b parser.Command) error {
	if err := isContentSame(a.Name, b.Name); err != nil {
		return fmt.Errorf("command name: %s", err)
	}
	if a.UsedPrefix != b.UsedPrefix {
		return fmt.Errorf("expected command prefix is %s, got %s", b.UsedPrefix, a.UsedPrefix)
	}
	if err := isContentSame(a.Raw, b.Raw); err != nil {
		return fmt.Errorf("command raw: %s", err)
	}
	return nil
}

func checkParseResult(outResult, expectedParse parser.ParseResult) error {
	if outResult.Text != expectedParse.Text {
		return fmt.Errorf("expected text is %q, got %q", expectedParse.Text, outResult.Text)
	}

	if err := isCommandSame(outResult.Command, expectedParse.Command); err != nil {
		return err
	}

	if err := isArgSame(outResult.RawArg, expectedParse.RawArg); err != nil {
		return fmt.Errorf("rawarg: %s", err.Error())
	}

	if len(outResult.Args) != len(expectedParse.Args) {
		return fmt.Errorf("expected args length is %d, got %d", len(expectedParse.Args), len(outResult.Args))
	}

	for i := range outResult.Args {
		o := outResult.Args[i]
		e := expectedParse.Args[i]

		if err := isArgSame(o, e); err != nil {
			return fmt.Errorf("arg #%d: %s", i, err.Error())
		}
	}

	if len(outResult.NamedArgs) != len(expectedParse.NamedArgs) {
		return fmt.Errorf("expected named args length is %d, got %d", len(expectedParse.NamedArgs), len(outResult.NamedArgs))
	}

	for key, expVals := range expectedParse.NamedArgs {
		outVals, ok := outResult.NamedArgs[key]
		if !ok {
			return fmt.Errorf("named arg with key %s is not found", key)
		}

		if len(outVals) != len(expVals) {
			return fmt.Errorf("named arg key %s: expected args length is %d, got %d", key, len(expectedParse.Args), len(outResult.Args))
		}
		for i := range outVals {
			o := outVals[i]
			e := expVals[i]

			if err := isArgSame(o, e); err != nil {
				return fmt.Errorf("named arg key %s: %s", key, err.Error())
			}
		}
	}

	return nil
}

var tests []ArgTest = []ArgTest{
	{Name: "no_text", Input: "", Expected: parser.ParseResult{
		Text:      "",
		Command:   parser.Command{},
		RawArg:    parser.Argument{},
		Args:      []parser.Argument{},
		NamedArgs: parser.NamedArgument{},
	}},

	{Name: "no_command_or_no_prefix", Input: "/just a normal text", Expected: parser.ParseResult{
		Text:      "/just a normal text",
		Command:   parser.Command{},
		RawArg:    parser.Argument{},
		Args:      []parser.Argument{},
		NamedArgs: parser.NamedArgument{},
	}},

	{Name: "command_no_args", Input: ".test", Expected: parser.ParseResult{
		Text: ".test",
		Command: parser.Command{
			Name:       parser.Content{Data: "test", Position: parser.Position{Start: 1, End: 4}},
			UsedPrefix: ".",
			Raw:        parser.Content{Data: ".test", Position: parser.Position{Start: 0, End: 4}},
		},
		RawArg:    parser.Argument{},
		Args:      []parser.Argument{},
		NamedArgs: parser.NamedArgument{},
	}},

	{Name: "command_space_no_args", Input: "   .\n   test", Expected: parser.ParseResult{
		Text: ".\n   test",
		Command: parser.Command{
			Name:       parser.Content{Data: "test", Position: parser.Position{Start: 5, End: 8}},
			UsedPrefix: ".",
			Raw:        parser.Content{Data: ".\n   test", Position: parser.Position{Start: 0, End: 8}},
		},
		RawArg:    parser.Argument{},
		Args:      []parser.Argument{},
		NamedArgs: parser.NamedArgument{},
	}},

	{Name: "normal_args", Input: ".test arg_1 arg_2\narg_newline", Expected: parser.ParseResult{
		Text: ".test arg_1 arg_2\narg_newline",
		Command: parser.Command{
			Name:       parser.Content{Data: "test", Position: parser.Position{Start: 1, End: 4}},
			UsedPrefix: ".",
			Raw:        parser.Content{Data: ".test", Position: parser.Position{Start: 0, End: 4}},
		},
		RawArg: parser.Argument{
			Position: parser.Position{Start: 6, End: 28},
			Content: parser.Content{
				Data:     "arg_1 arg_2\narg_newline",
				Position: parser.Position{Start: 6, End: 28},
			},
			InsideQuote: false, UsedQuote: 0,
		},
		Args: []parser.Argument{
			{
				Position: parser.Position{Start: 6, End: 10},
				Content: parser.Content{
					Data:     "arg_1",
					Position: parser.Position{Start: 6, End: 10},
				},
				InsideQuote: false, UsedQuote: 0,
			},
			{
				Position: parser.Position{Start: 12, End: 16},
				Content: parser.Content{
					Data:     "arg_2",
					Position: parser.Position{Start: 12, End: 16},
				},
				InsideQuote: false, UsedQuote: 0,
			},
			{
				Position: parser.Position{Start: 18, End: 28},
				Content: parser.Content{
					Data:     "arg_newline",
					Position: parser.Position{Start: 18, End: 28},
				},
				InsideQuote: false, UsedQuote: 0,
			},
		},
		NamedArgs: parser.NamedArgument{},
	}},

	{Name: "crazy_space", Input: ".  test          arg_1        'arg     \n\n\n\n\n            2'      'arg_3'          arg_\n4", Expected: parser.ParseResult{
		Text: ".  test          arg_1        'arg     \n\n\n\n\n            2'      'arg_3'          arg_\n4",
		Command: parser.Command{
			Name:       parser.Content{Data: "test", Position: parser.Position{Start: 3, End: 6}},
			UsedPrefix: ".",
			Raw:        parser.Content{Data: ".  test", Position: parser.Position{Start: 0, End: 6}},
		},
		RawArg: parser.Argument{
			Position: parser.Position{Start: 17, End: 86},
			Content: parser.Content{
				Data:     "arg_1        'arg     \n\n\n\n\n            2'      'arg_3'          arg_\n4",
				Position: parser.Position{Start: 17, End: 86},
			},
			InsideQuote: false, UsedQuote: 0,
		},
		Args: []parser.Argument{
			{
				Position: parser.Position{Start: 17, End: 21},
				Content: parser.Content{
					Data:     "arg_1",
					Position: parser.Position{Start: 17, End: 21},
				},
				InsideQuote: false, UsedQuote: 0,
			},
			{
				Position: parser.Position{Start: 30, End: 57},
				Content: parser.Content{
					Data:     "arg     \n\n\n\n\n            2",
					Position: parser.Position{Start: 31, End: 56},
				},
				InsideQuote: true, UsedQuote: '\'',
			},
			{
				Position: parser.Position{Start: 64, End: 70},
				Content: parser.Content{
					Data:     "arg_3",
					Position: parser.Position{Start: 65, End: 69},
				},
				InsideQuote: true, UsedQuote: '\'',
			},
			{
				Position: parser.Position{Start: 81, End: 84},
				Content: parser.Content{
					Data:     "arg_",
					Position: parser.Position{Start: 81, End: 84},
				},
				InsideQuote: false, UsedQuote: 0,
			},
			{
				Position: parser.Position{Start: 86, End: 86},
				Content: parser.Content{
					Data:     "4",
					Position: parser.Position{Start: 86, End: 86},
				},
				InsideQuote: false, UsedQuote: 0,
			},
		},
		NamedArgs: parser.NamedArgument{},
	}},

	{Name: "quoted_args", Input: ".test 'quote 1' \"quote\n2\" `quote \n 3` no quote", Expected: parser.ParseResult{
		Text: ".test 'quote 1' \"quote\n2\" `quote \n 3` no quote",
		Command: parser.Command{
			Name:       parser.Content{Data: "test", Position: parser.Position{Start: 1, End: 4}},
			UsedPrefix: ".",
			Raw:        parser.Content{Data: ".test", Position: parser.Position{Start: 0, End: 4}},
		},
		RawArg: parser.Argument{
			Position: parser.Position{Start: 6, End: 45},
			Content: parser.Content{
				Data:     "'quote 1' \"quote\n2\" `quote \n 3` no quote",
				Position: parser.Position{Start: 6, End: 45},
			},
			InsideQuote: false, UsedQuote: 0,
		},
		Args: []parser.Argument{
			{
				Position: parser.Position{Start: 6, End: 14},
				Content: parser.Content{
					Data:     "quote 1",
					Position: parser.Position{Start: 7, End: 13},
				},
				InsideQuote: true, UsedQuote: '\'',
			},
			{
				Position: parser.Position{Start: 16, End: 24},
				Content: parser.Content{
					Data:     "quote\n2",
					Position: parser.Position{Start: 17, End: 23},
				},
				InsideQuote: true, UsedQuote: '"',
			},
			{
				Position: parser.Position{Start: 26, End: 36},
				Content: parser.Content{
					Data:     "quote \n 3",
					Position: parser.Position{Start: 27, End: 35},
				},
				InsideQuote: true, UsedQuote: '`',
			},
			{
				Position: parser.Position{Start: 38, End: 39},
				Content: parser.Content{
					Data:     "no",
					Position: parser.Position{Start: 38, End: 39},
				},
				InsideQuote: false, UsedQuote: 0,
			},
			{
				Position: parser.Position{Start: 41, End: 45},
				Content: parser.Content{
					Data:     "quote",
					Position: parser.Position{Start: 41, End: 45},
				},
				InsideQuote: false, UsedQuote: 0,
			},
		},
		NamedArgs: parser.NamedArgument{},
	}},

	{Name: "named_args", Input: ".test filter=normal other='spaced value' other=other empty= normal_arg 'normal arg'", Expected: parser.ParseResult{
		Text: ".test filter=normal other='spaced value' other=other empty= normal_arg 'normal arg'",
		Command: parser.Command{
			Name:       parser.Content{Data: "test", Position: parser.Position{Start: 1, End: 4}},
			UsedPrefix: ".",
			Raw:        parser.Content{Data: ".test", Position: parser.Position{Start: 0, End: 4}},
		},
		RawArg: parser.Argument{
			Position: parser.Position{Start: 6, End: 82},
			Content: parser.Content{
				Data:     "filter=normal other='spaced value' other=other empty= normal_arg 'normal arg'",
				Position: parser.Position{Start: 6, End: 82},
			},
			InsideQuote: false, UsedQuote: 0,
		},
		Args: []parser.Argument{
			{
				Position: parser.Position{Start: 60, End: 69},
				Content: parser.Content{
					Data:     "normal_arg",
					Position: parser.Position{Start: 60, End: 69},
				},
				InsideQuote: false, UsedQuote: 0,
			},
			{
				Position: parser.Position{Start: 71, End: 82},
				Content: parser.Content{
					Data:     "normal arg",
					Position: parser.Position{Start: 72, End: 81},
				},
				InsideQuote: true, UsedQuote: '\'',
			},
		},
		NamedArgs: parser.NamedArgument{
			"filter": []parser.Argument{{
				Position: parser.Position{Start: 6, End: 18},
				Content: parser.Content{
					Data:     "normal",
					Position: parser.Position{Start: 13, End: 18},
				},
				InsideQuote: false, UsedQuote: 0,
			}},
			"other": []parser.Argument{
				{
					Position: parser.Position{Start: 20, End: 39},
					Content: parser.Content{
						Data:     "spaced value",
						Position: parser.Position{Start: 27, End: 38},
					},
					InsideQuote: true, UsedQuote: '\'',
				},
				{
					Position: parser.Position{Start: 41, End: 51},
					Content: parser.Content{
						Data:     "other",
						Position: parser.Position{Start: 47, End: 51},
					},
					InsideQuote: false, UsedQuote: 0,
				},
			},
			"empty": []parser.Argument{{
				Position: parser.Position{Start: 53, End: 58},
				Content: parser.Content{
					Data:     "",
					Position: parser.Position{Start: 0, End: 0},
				},
				InsideQuote: false, UsedQuote: 0,
			}},
		},
	}},

	{Name: "empty_quoted_arg", Input: ".test \"\"", Expected: parser.ParseResult{
		Text: ".test \"\"",
		Command: parser.Command{
			Name:       parser.Content{Data: "test", Position: parser.Position{Start: 1, End: 4}},
			UsedPrefix: ".",
			Raw:        parser.Content{Data: ".test", Position: parser.Position{Start: 0, End: 4}},
		},
		RawArg: parser.Argument{
			Content: parser.Content{
				Data:     "\"\"",
				Position: parser.Position{Start: 6, End: 7},
			},
			Position:    parser.Position{Start: 6, End: 7},
			InsideQuote: false, UsedQuote: 0,
		},
		Args: []parser.Argument{
			{
				Content: parser.Content{
					Data:     "",
					Position: parser.Position{Start: 0, End: 0},
				},
				Position:    parser.Position{Start: 6, End: 7},
				InsideQuote: true, UsedQuote: 34,
			},
		},
		NamedArgs: parser.NamedArgument{},
	}},

	{Name: "error_no_quote_close", Input: ".test 'hehe", ShouldError: true},
	{Name: "named_after_normal_arg", Input: ".test hehe named=", ShouldError: true},
}

func TestParseResult(t *testing.T) {
	prefix := []string{"."}
	theParser := parser.Init(prefix)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			parseResult, err := theParser.Parse(test.Input)
			if test.ShouldError {
				if err == nil {
					t.Errorf("this should throw error")
					t.SkipNow()
				} else {
					t.Skip()
				}
			}
			if !test.ShouldError {
				if err != nil {
					t.Errorf("this should not throw error: %s", err.Error())
					t.SkipNow()
				} else {
					t.Skip()
				}
			}

			if err := checkParseResult(parseResult, test.Expected); err != nil {
				t.Errorf("%s", err.Error())
			}
		})
	}
}
