package p

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

type Scanner func(stderr bool, text string)

type Session struct {
	Subprocess *Subprocess
	Alias      map[string]string
	Env        map[string]string
	mu         sync.Mutex
}

func NewSession() *Session {
	return &Session{
		Alias: make(map[string]string),
		Env:   make(map[string]string),
	}
}

func (s *Session) AddAlias(k, v string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Alias[k] = v
}

type Subprocess struct {
	WorkingDir                   string
	Bin                          string
	Args                         []string
	Env                          map[string]string
	Err                          error
	Cmd                          *exec.Cmd
	mu                           sync.Mutex
	Done                         chan struct{}
	Context                      context.Context
	Stderr, Stdout, Combined     bytes.Buffer
	StderrReader, StdoutReader   io.ReadCloser
	stderrScanner, stdoutScanner *bufio.Scanner
	StdInWriter                  io.Writer
	Nice, Pid                    int
	Label                        Label
}

func NewSubprocess(bin string, args ...string) (*Subprocess, error) {
	subp := &Subprocess{
		Bin:  bin,
		Args: args,
		Env:  make(map[string]string),
		Done: make(chan struct{}),
	}

	if !subp.ExecInPath() {
		return nil, fmt.Errorf("could not find executable '%s'", subp.Bin)
	}
	return subp, nil
}

func (sp *Subprocess) SetArgs(args ...string) {
	sp.Args = args
}

func (sp *Subprocess) ExecInPath() bool {
	if _, err := exec.LookPath(sp.Bin); err != nil {
		return false
	}
	return true
}

func (sp *Subprocess) Start(args ...string) error {
	if len(args) > 0 {
		sp.SetArgs(args...)
	}
	if sp.Context == nil {
		sp.Context = context.Background() // TODO: make func in helper to incorporate max concurrency
	}

	sp.Done = make(chan struct{})
	sp.Cmd = exec.Command(sp.Bin, sp.Args...)
	sp.Cmd.Dir = sp.WorkingDir
	if sp.Env != nil {
		sp.Cmd.Env = SerializeMap(sp.Env)
	}

	// TODO: add stdin/out/err readers
	// TODO: trace, too?
	err := sp.Cmd.Start()
	if err != nil {
		return err
	}

	go sp.scanAndWait()

	if runtime.GOOS != "windows" && sp.Nice != 0 {
		niceCmd := exec.Command("renice", "-n", Format("%d", sp.Nice), "-p", Format("%d", sp.Pid))
		niceCmd.Run()
	}
	return nil
}

func (sp *Subprocess) String() string {
	parts := []string{sp.Bin}
	for _, a := range sp.Args {
		if strings.Contains(a, `"`) {
			a = `"` + strings.ReplaceAll(a, `"`, `""`) + `"`
		}
		parts = append(parts, a)
	}
	return strings.Join(parts, " ")
}

// func (sp *Subprocess) scanAndWait() {
// 	exitCh := make(chan bool)
// 	label := sp.Label.Render()
//
// 	go func() {
// 		for p
// 	}()
// }

type Label struct {
	Value      string
	Len, Color int
}

func (l Label) Render() string {
	if l.Value == "" {
		return ""
	}
	label := l.Value
	if l.Color > 0 {
		label = Colorize(l.Color, l.Value)
	}
	if l.Len > len(l.Value) {
		label = label + strings.Repeat(" ", l.Len-len(l.Value))
	}
	return label
}
