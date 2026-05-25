//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"unsafe"
)

const realCodexEnv = "CODEX_REMOTE_CONTROL_REAL_CODEX"

func main() {
	realCodex, err := resolveRealCodex()
	if err != nil {
		fmt.Fprintf(os.Stderr, "codex remote-control shim: %v\n", err)
		os.Exit(127)
	}

	args := append([]string{}, os.Args[1:]...)
	if shouldEnableRemoteControl(args) {
		args = append(args, "--remote-control")
	}

	cmd := exec.Command(realCodex, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = childEnv()

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "codex remote-control shim: failed to run %q: %v\n", realCodex, err)
		os.Exit(126)
	}
	job := assignKillOnCloseJob(cmd.Process.Pid)
	if job != 0 {
		defer closeHandle(job)
	}

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "codex remote-control shim: %q failed: %v\n", realCodex, err)
		os.Exit(126)
	}
}

func shouldEnableRemoteControl(args []string) bool {
	if len(args) < 1 || args[0] != "app-server" {
		return false
	}
	for i := 1; i < len(args); i++ {
		if args[i] == "--enable" && i+1 < len(args) && args[i+1] == "remote_control" {
			return false
		}
		if args[i] == "--experimental-enable-remote-control" || args[i] == "--remote-control" {
			return false
		}
	}
	return true
}

func resolveRealCodex() (string, error) {
	if configured := strings.TrimSpace(os.Getenv(realCodexEnv)); configured != "" {
		return configured, nil
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return "", fmt.Errorf("LOCALAPPDATA is not set and %s was not provided", realCodexEnv)
	}

	root := filepath.Join(localAppData, "OpenAI", "Codex", "bin")
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", fmt.Errorf("cannot read Codex bin directory %q: %w", root, err)
	}

	type candidate struct {
		path    string
		modTime int64
	}
	var candidates []candidate
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(root, entry.Name(), "codex.exe")
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		candidates = append(candidates, candidate{path: path, modTime: info.ModTime().UnixNano()})
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("no codex.exe found under %q", root)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].modTime > candidates[j].modTime
	})
	return candidates[0].path, nil
}

func childEnv() []string {
	env := os.Environ()
	out := env[:0]
	for _, item := range env {
		name, _, found := strings.Cut(item, "=")
		if found && strings.EqualFold(name, "CODEX_CLI_PATH") {
			continue
		}
		out = append(out, item)
	}
	return out
}

func assignKillOnCloseJob(pid int) syscall.Handle {
	job := createJobObject()
	if job == 0 {
		return 0
	}
	if !setJobKillOnClose(job) {
		closeHandle(job)
		return 0
	}
	process := openProcessForJob(pid)
	if process == 0 {
		closeHandle(job)
		return 0
	}
	defer closeHandle(process)
	if !assignProcessToJobObject(job, process) {
		closeHandle(job)
		return 0
	}
	return job
}

const (
	jobObjectExtendedLimitInformationClass = 9
	jobObjectLimitKillOnJobClose           = 0x00002000
	processSetQuota                        = 0x0100
	processTerminate                       = 0x0001
)

type ioCounters struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

type jobObjectBasicLimitInformation struct {
	PerProcessUserTimeLimit int64
	PerJobUserTimeLimit     int64
	LimitFlags              uint32
	MinimumWorkingSetSize   uintptr
	MaximumWorkingSetSize   uintptr
	ActiveProcessLimit      uint32
	Affinity                uintptr
	PriorityClass           uint32
	SchedulingClass         uint32
}

type jobObjectExtendedLimitInformation struct {
	BasicLimitInformation jobObjectBasicLimitInformation
	IoInfo                ioCounters
	ProcessMemoryLimit    uintptr
	JobMemoryLimit        uintptr
	PeakProcessMemoryUsed uintptr
	PeakJobMemoryUsed     uintptr
}

var (
	kernel32                    = syscall.NewLazyDLL("kernel32.dll")
	procCreateJobObjectW        = kernel32.NewProc("CreateJobObjectW")
	procSetInformationJobObject = kernel32.NewProc("SetInformationJobObject")
	procAssignProcessToJob      = kernel32.NewProc("AssignProcessToJobObject")
	procOpenProcess             = kernel32.NewProc("OpenProcess")
	procCloseHandle             = kernel32.NewProc("CloseHandle")
)

func createJobObject() syscall.Handle {
	handle, _, _ := procCreateJobObjectW.Call(0, 0)
	return syscall.Handle(handle)
}

func setJobKillOnClose(job syscall.Handle) bool {
	info := jobObjectExtendedLimitInformation{}
	info.BasicLimitInformation.LimitFlags = jobObjectLimitKillOnJobClose
	ok, _, _ := procSetInformationJobObject.Call(
		uintptr(job),
		uintptr(jobObjectExtendedLimitInformationClass),
		uintptr(unsafe.Pointer(&info)),
		unsafe.Sizeof(info),
	)
	return ok != 0
}

func openProcessForJob(pid int) syscall.Handle {
	handle, _, _ := procOpenProcess.Call(
		uintptr(processSetQuota|processTerminate),
		0,
		uintptr(uint32(pid)),
	)
	return syscall.Handle(handle)
}

func assignProcessToJobObject(job syscall.Handle, process syscall.Handle) bool {
	ok, _, _ := procAssignProcessToJob.Call(uintptr(job), uintptr(process))
	return ok != 0
}

func closeHandle(handle syscall.Handle) {
	if handle != 0 {
		procCloseHandle.Call(uintptr(handle))
	}
}
