package tauri

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// RunApplicationImpl implements Tauri.Native.runApplicationImpl: launch the
// configured sibling application with literal arguments in its own process
// group, deliver the optional source on stdin, inherit output and wait. The
// result is the ProcessResult fields code, stderr and success.
func RunApplicationImpl(arguments []string, source *string) (int, string, bool) {
	return runApplication(arguments, source, false)
}

// RunApplicationTaskImpl implements Tauri.Native.runApplicationTaskImpl: like
// RunApplicationImpl, but the task owns its process group, which is
// terminated once the leader exits.
func RunApplicationTaskImpl(arguments []string, source *string) (int, string, bool) {
	return runApplication(arguments, source, true)
}

func runApplication(arguments []string, source *string, task bool) (int, string, bool) {
	executable := applicationPath()
	command := exec.Command(executable, arguments...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if source != nil {
		command.Stdin = strings.NewReader(*source)
	}
	if err := command.Start(); err != nil {
		throw(executable + ": " + rustIOMessage(err))
	}
	err := command.Wait()
	if task {
		_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		throw("waiting for " + executable + ": " + rustIOMessage(err))
	}
	state := command.ProcessState
	code := int(state.ExitCode())
	return code, "", state.Success()
}

func applicationPath() string {
	if filepath.IsAbs(config.Application) {
		return config.Application
	}
	executable, err := os.Executable()
	if err == nil {
		if resolved, err := filepath.EvalSymlinks(executable); err == nil {
			executable = resolved
		}
	}
	name := config.Application
	if name == "" {
		name = filepath.Base(executable) + "-editor"
	}
	return filepath.Join(filepath.Dir(executable), name)
}

func rustIOMessage(err error) string {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return rustErrno(errno)
	}
	return err.Error()
}
