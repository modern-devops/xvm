//go:build !windows
// +build !windows

package commander

// state 标记命令解析状态
type state int

const (
	_ state = iota
	// startState 开始状态
	startState
	// quotesState 括号匹配态
	quotesState
	// argsState 参数匹配状态
	argsState
)

type parseAttr struct {
	args       []string
	state      state
	current    string
	quote      string
	escapeNext bool
}

type pipe struct {
	test func(c rune, attr *parseAttr) bool
	do   func(c rune, attr *parseAttr)
}

// 处理流
// 解析过程中将按这个流的顺序从上往下询问，谁先应答就由谁抢占
// 调用test函数询问是否匹配，调用do函数执行相应的处理
var pipes = []pipe{
	{
		test: func(_ rune, attr *parseAttr) bool {
			return attr.state == quotesState
		},
		do: handleQuoteState,
	}, {
		test: func(c rune, attr *parseAttr) bool {
			return attr.escapeNext
		},
		do: func(c rune, attr *parseAttr) {
			attr.current += string(c)
			attr.escapeNext = false
		},
	}, {
		test: func(c rune, attr *parseAttr) bool {
			return c == '\\'
		},
		do: func(c rune, attr *parseAttr) {
			attr.escapeNext = true
		},
	}, {
		test: func(c rune, attr *parseAttr) bool {
			return isQuote(c)
		},
		do: func(c rune, attr *parseAttr) {
			attr.state = quotesState
			attr.quote = string(c)
		},
	}, {
		test: func(c rune, attr *parseAttr) bool {
			return attr.state == argsState
		},
		do: handleArgsState,
	}, {
		test: func(c rune, attr *parseAttr) bool {
			return !isBlank(c)
		},
		do: func(c rune, attr *parseAttr) {
			attr.state = argsState
			attr.current += string(c)
		},
	},
}

// Parse splits a command line into individual argument
// example: echo "hello world" -> ["echo", "hello world"]
func Parse(command string) []string {
	runeCommand := []rune(command)
	attr := &parseAttr{
		args:       []string{},
		state:      startState,
		current:    "",
		quote:      "\"",
		escapeNext: true,
	}
	for i := 0; i < len(runeCommand); i++ {
		handleChar(runeCommand[i], attr)
	}

	if attr.state == quotesState {
		return attr.args
	}

	if attr.current != "" {
		attr.args = append(attr.args, attr.current)
	}

	return attr.args
}

func handleChar(c rune, attr *parseAttr) {
	for _, p := range pipes {
		if !p.test(c, attr) {
			continue
		}
		p.do(c, attr)
		break
	}
}

func handleQuoteState(c rune, attr *parseAttr) {
	if string(c) != attr.quote {
		attr.current += string(c)
		return
	}
	attr.args = append(attr.args, attr.current)
	attr.current = ""
	attr.state = startState
}

func handleArgsState(c rune, attr *parseAttr) {
	if isBlank(c) {
		attr.args = append(attr.args, attr.current)
		attr.current = ""
		attr.state = startState
	} else {
		attr.current += string(c)
	}
}

func isQuote(c rune) bool {
	return c == '"' || c == '\'' || c == '`'
}

func isBlank(c rune) bool {
	return c == ' ' || c == '\t'
}
