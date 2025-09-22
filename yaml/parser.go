// NOTE: WIP
package yaml

import (
	"strings"

	"github.com/s-hammon/p"
)

type parser struct {
	state       state
	lineNumber  int
	indentStack *p.Stack[int]
}

func newParser() *parser {
	return &parser{
		indentStack: p.NewStack[int](),
	}
}

func (p *parser) transition(line string) {
	trimmed := strings.TrimSpace(line)
	indent := len(line) - len(strings.TrimLeft(line, " "))
	p.checkIndent(indent)

	switch p.state {
	case stateStart:
		if trimmed == "---" {
			p.state = stateDocStart
		} else if trimmed == "" {
			p.state = stateNewLine
		} else if strings.HasPrefix(trimmed, "#") {
			p.state = stateComment
		} else if strings.HasPrefix(trimmed, "- ") {
			p.checkIndent(indent)
			p.state = stateListItem
		} else if strings.Contains(trimmed, ":") {
			p.checkIndent(indent)
			p.state = stateKey
		} else {
			p.state = stateErr
		}
	case stateDocStart, stateNewLine:
		if trimmed == "" {
			p.state = stateNewLine
		} else if strings.HasPrefix(trimmed, "#") {
			p.state = stateComment
		} else if strings.HasPrefix(trimmed, "- ") {
			p.checkIndent(indent)
			p.state = stateListItem
		} else if strings.Contains(trimmed, ":") {
			p.checkIndent(indent)
			p.state = stateKey
		} else {
			p.state = stateErr
		}
	case stateKey:
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			p.state = stateErr
			return
		}

		if val := strings.TrimSpace(parts[1]); val != "" {
			p.state = stateScalar
		} else {
			p.state = stateVal
		}
	case stateVal:
		if trimmed == "" {
			p.state = stateNewLine
		} else {
			p.state = stateScalar
		}
	case stateListItem:
		p.state = stateVal
	case stateScalar, stateComment:
		p.state = stateNewLine
	case stateErr:
	}
}

func (p *parser) checkIndent(indent int) {
	if indent%2 != 0 {
		p.state = stateErr
		return
	}

	top, ok := p.indentStack.Peek()
	if !ok {
		p.indentStack.Push(indent)
		return
	}

	switch {
	case indent > top:
		p.indentStack.Push(indent)
	case indent == top:
	case indent < top:
		for {
			_, ok := p.indentStack.Pop()
			if !ok {
				p.state = stateErr
				return
			}

			peek, ok := p.indentStack.Peek()
			if !ok || peek == indent {
				break
			}
			if peek < indent {
				p.state = stateErr
				return
			}
		}
	}
}

func (p *parser) parse(lines []string) {
	p.state = stateStart
	for i, line := range lines {
		p.lineNumber = i + 1
		p.transition(line)

		if p.state == stateEnd {
			break
		}
	}
}
