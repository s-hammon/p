package p

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
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
	Scanner                      Scanner
	StdInWriter                  io.Writer
	Nice, Pid                    int
	Capture, Print               bool
	Label                        Label
}

func NewSubprocess(bin string, args ...string) (*Subprocess, error) {
	subp := &Subprocess{
		Bin:  bin,
		Args: args,
		Env:  make(map[string]string),
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
	_, err := exec.LookPath(sp.Bin)
	return err == nil
}

func (sp *Subprocess) Start(args ...string) (err error) {
	if len(args) > 0 {
		sp.SetArgs(args...)
	}
	if sp.Context == nil {
		sp.Context = context.Background() // TODO: make func in helper to incorporate max concurrency
	}

	sp.Done = make(chan struct{})

	sp.Cmd = exec.CommandContext(sp.Context, sp.Bin, sp.Args...)
	sp.Cmd.Dir = sp.WorkingDir

	if sp.Env != nil {
		sp.Cmd.Env = SerializeMap(sp.Env)
	} else {
		sp.Cmd.Env = os.Environ()
	}

	sp.Stdout.Reset()
	sp.Stderr.Reset()
	sp.Combined.Reset()

	sp.StdoutReader, err = sp.Cmd.StdoutPipe()
	if err != nil {
		return err
	}
	sp.StderrReader, err = sp.Cmd.StderrPipe()
	if err != nil {
		return err
	}

	sp.stderrScanner = bufio.NewScanner(sp.StderrReader)
	sp.stderrScanner.Split(bufio.ScanLines)

	sp.stdoutScanner = bufio.NewScanner(sp.StdoutReader)
	sp.stdoutScanner.Split(bufio.ScanLines)

	sp.StdInWriter, err = sp.Cmd.StdinPipe()
	if err != nil {
		return err
	}

	err = sp.Cmd.Start()
	if err != nil {
		sp.StdoutReader.Close()
		sp.StderrReader.Close()
		return err
	}

	sp.Pid = sp.Cmd.Process.Pid

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

func (sp *Subprocess) scanAndWait() {
	exitCh := make(chan bool)
	label := sp.Label.Render()

	// stderr
	go func() {
		for sp.stderrScanner.Scan() {
			line := sp.stderrScanner.Text()
			sp.mu.Lock()
			if sp.Capture {
				sp.Stderr.WriteString(line + "\n")
				sp.Combined.WriteString(line + "\n")
			}
			if sp.Scanner != nil {
				sp.Scanner(true, line)
			}
			if sp.Print {
				if label != "" {
					line = Format("%s | %s", Colorize(90, label), line)
				}
				fmt.Fprintf(os.Stderr, "%s", line+"\n")
			}
			sp.mu.Unlock()
		}
		exitCh <- true
	}()

	// stdout
	go func() {
		for sp.stdoutScanner.Scan() {
			line := sp.stdoutScanner.Text()
			sp.mu.Lock()
			if sp.Capture {
				sp.Stdout.WriteString(line + "\n")
				sp.Combined.WriteString(line + "\n")
			}
			if sp.Scanner != nil {
				sp.Scanner(false, line)
			}
			if sp.Print {
				if label != "" {
					line = Format("%s | %s", Colorize(90, label), line)
				}
				fmt.Fprintf(os.Stdout, "%s", line+"\n")
			}
			sp.mu.Unlock()
		}
		exitCh <- true
	}()

	err := sp.Cmd.Wait()
	if err != nil {
		sp.Err = err
	}

	<-exitCh
	<-exitCh

	sp.Done <- struct{}{}
}

func (sp *Subprocess) Run(args ...string) error {
	err := sp.Start(args...)
	if err != nil {
		return err
	}
	err = sp.Wait()
	if err != nil {
		return err
	}
	return nil
}

func (sp *Subprocess) Wait() error {
	select {
	case <-sp.Done:
	case <-sp.Context.Done():
		sp.Cmd.Process.Signal(syscall.SIGINT)
		t := time.NewTimer(5 * time.Second)
		select {
		case <-sp.Done:
		case <-t.C:
			log.Fatal(sp.Cmd.Process.Kill())
		}
	}
	if sp.Err != nil {
		return sp.Err
	}
	if sp.Cmd != nil && sp.Cmd.ProcessState != nil {
		if code := sp.Cmd.ProcessState.ExitCode(); code != 0 {
			return fmt.Errorf("exit code: %d", code)
		}
	}
	return nil
}

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
