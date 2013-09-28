package eval

import (
	"os"
	"fmt"
	"strings"
	"syscall"
	"../parse"
)

const (
	// A special impossible fd value. Used for "close fd" in
	// syscall.ProcAttr.Files and various other things internally.
	FD_NIL uintptr = ^uintptr(0)
)

type command struct {
	name string // command name, used in error messages
	// full argument list. args[0] is always some form of command name.
	args []string
}

type externalCommand struct {
	command
	ios [3]uintptr
}

func isExecutable(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return false
	}
	fm := fi.Mode()
	return !fm.IsDir() && (fm & 0111 != 0)
}

// Search for executable `exe`.
func search(exe string) (string, error) {
	for _, p := range []string{"/", "./", "../"} {
		if strings.HasPrefix(exe, p) {
			if isExecutable(exe) {
				return exe, nil
			}
			return "", fmt.Errorf("not executable")
		}
	}
	for _, p := range search_paths {
		full := p + "/" + exe
		if isExecutable(full) {
			return full, nil
		}
	}
	return "", fmt.Errorf("not found")
}

func envAsSlice(env map[string]string) (s []string) {
	s = make([]string, 0, len(env))
	for k, v := range env {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	return
}

func evalTerm(n parse.Node) (string, error) {
	return n.(*parse.StringNode).Text, nil
}

func evalTermList(ln *parse.ListNode) ([]string, error) {
	ss := make([]string, len(ln.Nodes))
	for i, n := range ln.Nodes {
		var e error
		ss[i], e = evalTerm(n)
		if e != nil {
			return nil, e
		}
	}
	return ss, nil
}

// CommandErrors holds multiple errors.
type CommandErrors struct {
	Errors []error
}

func (ce CommandErrors) Error() string {
	return fmt.Sprintf("%v", ce.Errors)
}

// ExecPipeline executes a pipeline.
//
// As many things as possible are done before any command actually gets
// executed, to avoid leaving the pipeline broken - resolving command names,
// opening files, and in future, evaluating shell constructs. If any error is
// encountered, pids is nil and err contains the error.
//
// However, if error is encountered when executing individual commands, the
// rest of the pipeline will still be executed. In that case, the
// corresponding elements in pids is -1 and err is typed *CommandErrors. For
// each pids[i] == -1, err.(*CommandErrors)Errors[i] contains the
// corresponding error.
func ExecPipeline(pl *parse.ListNode) (pids []int, err error) {
	ncmds := len(pl.Nodes)
	if ncmds == 0 {
		return []int{}, nil
	}

	cmds := make([]externalCommand, 0, ncmds)

	nextReadPipe := FD_NIL

	for i, n := range pl.Nodes {
		n := n.(*parse.CommandNode)

		if len(n.Nodes) == 0 {
			return nil, fmt.Errorf("command #%d is emtpy", i)
		}

		cmd := externalCommand{ios: [3]uintptr{1, 2, 3}}
		cmd.args, err = evalTermList(&n.ListNode)

		if err != nil {
			return nil, fmt.Errorf("error evaluating command #%d: %s", err)
		}

		// Save unresolved args[0] as name.
		cmd.name = cmd.args[0]

		cmd.args[0], err = search(n.Nodes[0].(*parse.StringNode).Text)
		if err != nil {
			return nil, fmt.Errorf("can't resolve command #%d: %s", i, err)
		}

		// Create pipes.
		// XXX Check whether output is fd IO
		var readPipe, writePipe uintptr
		readPipe = nextReadPipe
		writePipe = FD_NIL
		if i != ncmds - 1 {
			// os.Pipe sets O_CLOEXEC, which is what we want.
			reader, writer, e := os.Pipe()
			if e != nil {
				return nil, fmt.Errorf("failed to create pipe: %s", e)
			}
			defer reader.Close()
			defer writer.Close()
			nextReadPipe = reader.Fd()
			writePipe = writer.Fd()
		}

		// Check IO redirections, turn all FilenameRedir to FdRedir.
		for _, r := range n.Redirs {
			fd := r.Fd()
			if fd > 2 {
				return nil, fmt.Errorf("redir on fd > 2 not yet supported")
			} else if fd == 0 && readPipe != FD_NIL {
				return nil, fmt.Errorf("input already connected to pipe")
			} else if fd == 1 && writePipe != FD_NIL {
				return nil, fmt.Errorf("output already connected to pipe")
			}
			switch r := r.(type) {
			case *parse.CloseRedir:
				cmd.ios[fd] = FD_NIL
			case *parse.FdRedir:
				if r.OldFd > 2 {
					return nil, fmt.Errorf("fd redir from fd > 2 not yet supported")
				}
				cmd.ios[fd] = r.OldFd
			case *parse.FilenameRedir:
				fname, err := evalTerm(r.Filename)
				if err != nil {
					return nil,
					       fmt.Errorf("failed to evaluate filename: %q: %s",
					                  r.Filename, err)
				}
				// TODO haz hardcoded permbits now
				f, err := os.OpenFile(fname, r.Flag, 0644)
				if err != nil {
					return nil, fmt.Errorf("failed to open file %q: %s",
					                       r.Filename, err)
				}
				oldFd := f.Fd()
				cmd.ios[fd] = oldFd
				defer f.Close()
			}
		}

		// Connect pipes.
		if readPipe != FD_NIL {
			cmd.ios[0] = readPipe
		}
		if writePipe != FD_NIL {
			cmd.ios[1] = writePipe
		}

		cmds = append(cmds, cmd)
	}

	pids = make([]int, ncmds)
	cmderr := CommandErrors{Errors: make([]error, ncmds)}
	haserr := false

	for i, cmd := range cmds {
		pid, err := ExecCommand(&cmd)

		if err != nil {
			pids[i] = -1
			cmderr.Errors[i] = err
			haserr = true
		} else {
			pids[i] = pid
		}
	}

	if haserr {
		return pids, cmderr
	}
	return pids, nil
}

// ExecCommand executes a command.
func ExecCommand(cmd *externalCommand) (pid int, err error) {
	sys := syscall.SysProcAttr{}
	attr := syscall.ProcAttr{Env: envAsSlice(env), Files: cmd.ios[:], Sys: &sys}

	return syscall.ForkExec(cmd.args[0], cmd.args, &attr)
}
