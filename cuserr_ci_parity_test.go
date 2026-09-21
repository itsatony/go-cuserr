// {project_root}/cuserr_ci_parity_test.go
package cuserr

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Hosted GitHub Actions is retired (operator decision, 2026-09-19; issue #10).
// The workflows are kept, inert, in .github/workflows-disabled/ because they are
// the SPECIFICATION the local gate is compared against — a deleted file cannot
// be compared against anything.
const (
	CI_PARITY_DISABLED_WORKFLOW_DIR = ".github/workflows-disabled"
	CI_PARITY_ACTIVE_WORKFLOW_DIR   = ".github/workflows"
	CI_PARITY_WORKFLOW_CI           = "ci.yml"
	CI_PARITY_WORKFLOW_RELEASE      = "release.yml"
	CI_PARITY_LOCAL_GATE            = "ci-local"
	CI_PARITY_RELEASE_GATE          = "release-check"
	CI_PARITY_MAKEFILE              = "Makefile"
	CI_PARITY_README                = "README.md"

	// A gate that compares nothing passes. These floors make a parser that
	// silently stops matching a FAILURE rather than a green run.
	CI_PARITY_MIN_JOBS     = 9
	CI_PARITY_MIN_COMMANDS = 20
)

// ciParityArm says how one workflow job is answered locally. Exactly one of
// makeTarget / exception is set: an arm is either reproduced or consciously not.
type ciParityArm struct {
	// makeTarget is the make target that reproduces this job.
	makeTarget string
	// exception is the STATED REASON the job has no local counterpart. An empty
	// reason on an exception is itself a failure: "we decided not to" is only a
	// decision when it says why.
	exception string
}

// ciParityJobs maps every job in the parked workflows to its local answer.
//
// ⚠️ This map is the hand-kept half; the LIST of jobs is derived from the
// workflow YAML, so a job added there fails this test until someone answers it.
// That asymmetry is the whole ratchet: the failure mode being defended against
// is a parked workflow that quietly specifies more than the local gate runs.
func ciParityJobsTable() map[string]ciParityArm {
	return map[string]ciParityArm{
		// --- ci.yml ---
		"quality":  {makeTarget: "ci-lint"},
		"test":     {makeTarget: "ci-test"},
		"coverage": {makeTarget: "ci-coverage"},
		"security": {exception: "the hosted job ran NO scanner — its whole body is an echo saying gosec was unavailable. " +
			"There is a `make security` target, but reproducing a placeholder would be reproducing nothing. " +
			"If a real scanner is ever wanted, add it as an arm and delete this exception."},
		"integration":  {makeTarget: "ci-examples"},
		"performance":  {makeTarget: "ci-perf"},
		"build-matrix": {makeTarget: "build-matrix"},

		// --- release.yml ---
		"validate": {makeTarget: CI_PARITY_RELEASE_GATE},
		"release": {exception: "creating a GitHub Release, uploading its body and pinging proxy.golang.org are GitHub-side " +
			"publishing actions, not checks. They have no local equivalent by construction; releases are now cut by hand " +
			"with `gh release create` after `make release-ready` is green."},
		"notify": {exception: "an echo-only success notification. There is nothing to reproduce."},
	}
}

// ciParityMatrixExceptions records the workflow matrix axes that a single host
// cannot reproduce, with the reason. The GOOS/GOARCH axes are deliberately NOT
// here: they are reproduced by `make build-matrix`.
func ciParityMatrixExceptionsTable() map[string]string {
	return map[string]string{
		"os": "3 operating systems cannot run on one host. Accepted (issue #10): the local gate runs the suite on this " +
			"host's OS only. The compatibility promise is now carried by the cross-compile arm plus the Go version floor " +
			"in go.mod, not by a test matrix.",
		"go-version": "one host has one installed toolchain. Accepted (issue #10): single-version coverage. Note the " +
			"parked matrix names 1.21/1.22/1.23 while go.mod already requires a newer Go — the matrix was stale, which is " +
			"itself an argument that an arm nobody reads is an arm nobody maintains.",
	}
}

// ciParityCommandExceptions covers run-step commands with no local counterpart,
// keyed by the command's leading word.
func ciParityCommandExceptionsTable() map[string]string {
	return map[string]string{
		"echo": "reporting only.",
		"cat":  "prints a benchmark log the step just wrote; the benchmarks themselves are the ci-perf arm.",
		"VERSION": "shell variable assignment deriving the version from the pushed tag; " +
			"the tag guard it feeds is checked below.",
		"if":           "shell control flow around a command that is checked on its own line.",
		"fi":           "shell control flow.",
		"else":         "shell control flow.",
		"exit":         "shell control flow.",
		"COVERAGE":     "shell variable assignment; the coverage floor itself is checked below.",
		"FILE_VERSION": "shell variable assignment; the tag guard itself is checked below.",
		"TAG_VERSION":  "shell variable assignment; the tag guard itself is checked below.",
		"gofmt": "the `gofmt -s -l` check is an arm (ci-fmt); this line is the diagnostic " +
			"re-print inside the same step.",
		"ls":      "an example-artifact listing inside the integration job's success echo.",
		"#":       "a comment inside a run block.",
		"timeout": "covered by ci-examples, which runs the same `timeout 10s go run -tags basic ...` line.",
	}
}

// ciParityCommandCoverage maps a command the hosted runner ran to the substring
// that proves some local arm reproduces it.
func ciParityCommandCoverageTable() map[string]string {
	return map[string]string{
		"go mod download": "go mod download",
		"go mod verify":   "go mod verify",
		"go vet":          "go vet ./...",
		"go build":        "go build",
		"go test":         "go test",
		"go tool":         "go tool cover",
		"go run":          "go run",
	}
}

// TestLocalCIReproducesTheDisabledWorkflows is what makes "hosted CI is off" a
// safe statement rather than a lowered bar.
func TestLocalCIReproducesTheDisabledWorkflows(t *testing.T) {
	root := ciParityRepoRoot(t)
	jobs := ciParityJobsTable()
	matrixExceptions := ciParityMatrixExceptionsTable()
	commandExceptions := ciParityCommandExceptionsTable()
	commandCoverage := ciParityCommandCoverageTable()
	makefile := ciParityReadFile(t, filepath.Join(root, CI_PARITY_MAKEFILE))
	gateText := ciParityGateText(t, root, makefile)

	workflows := map[string]string{}
	for _, name := range []string{CI_PARITY_WORKFLOW_CI, CI_PARITY_WORKFLOW_RELEASE} {
		path := filepath.Join(root, CI_PARITY_DISABLED_WORKFLOW_DIR, name)
		workflows[name] = ciParityReadFile(t, path)
	}

	// 1. Every JOB is an arm or a declared exception with a reason.
	totalJobs := 0
	for name, body := range workflows {
		for _, job := range ciParityWorkflowJobs(body) {
			totalJobs++
			arm, known := jobs[job]
			if !known {
				t.Errorf("%s declares job %q, which the local gate does not answer. Either add an arm to `make %s` "+
					"and map it in ciParityJobs, or declare it an exception there WITH A REASON.", name, job, CI_PARITY_LOCAL_GATE)
				continue
			}
			switch {
			case arm.makeTarget != "" && arm.exception != "":
				t.Errorf("job %q is both an arm and an exception; decide which", job)
			case arm.makeTarget != "":
				if !ciParityMakeTargetExists(makefile, arm.makeTarget) {
					t.Errorf("job %q maps to make target %q, which does not exist", job, arm.makeTarget)
				}
				inLocalGate := ciParityGateDependsOn(makefile, CI_PARITY_LOCAL_GATE, arm.makeTarget)
				if arm.makeTarget != CI_PARITY_RELEASE_GATE && !inLocalGate {
					t.Errorf("make target %q exists but `%s` does not depend on it — an arm nobody runs is not an arm",
						arm.makeTarget, CI_PARITY_LOCAL_GATE)
				}
			case arm.exception == "":
				t.Errorf("job %q is declared an exception with no stated reason", job)
			}
		}
	}
	if totalJobs < CI_PARITY_MIN_JOBS {
		t.Fatalf("only %d jobs parsed out of the parked workflows (expected at least %d) — the parser has gone blind "+
			"and every assertion above passed vacuously", totalJobs, CI_PARITY_MIN_JOBS)
	}

	// 2. Every STRATEGY MATRIX axis is reproduced locally or declared.
	for name, body := range workflows {
		for _, axis := range ciParityMatrixAxes(body) {
			if axis == "goos" || axis == "goarch" {
				if !strings.Contains(gateText, "GOOS=") || !strings.Contains(gateText, "GOARCH=") {
					t.Errorf("%s has a %s matrix and the local gate has no GOOS/GOARCH loop", name, axis)
				}
				continue
			}
			if reason := matrixExceptions[axis]; reason == "" {
				t.Errorf("%s has a %q matrix axis that is neither reproduced locally nor a declared exception "+
					"with a reason (ciParityMatrixExceptions)", name, axis)
			}
		}
	}

	// 3. Every run-step COMMAND is reproduced by some arm, or declared.
	commands := 0
	compared := 0
	for name, body := range workflows {
		for job, jobCommands := range ciParityJobCommands(body) {
			// A job that is a declared exception carries its steps with it: the
			// reason recorded for the job is the reason for its commands.
			if jobs[job].exception != "" {
				continue
			}
			for _, command := range jobCommands {
				commands++
				head := ciParityCommandHead(command)
				if coverage, ok := commandCoverage[head]; ok {
					compared++
					if !strings.Contains(gateText, coverage) {
						t.Errorf("%s job %q runs %q and the local gate contains no %q — the local gate is weaker "+
							"than the workflow it replaced", name, job, command, coverage)
					}
					continue
				}
				if reason := commandExceptions[head]; reason == "" {
					t.Errorf("%s job %q runs %q, which this gate cannot classify. Reproduce it in an arm and add it "+
						"to ciParityCommandCoverage, or record in ciParityCommandExceptions why it needs no local "+
						"counterpart.", name, job, command)
				}
			}
		}
	}
	if commands < CI_PARITY_MIN_COMMANDS {
		t.Fatalf("only %d run commands parsed (expected at least %d) — the parser has gone blind",
			commands, CI_PARITY_MIN_COMMANDS)
	}
	if compared < 5 {
		t.Fatalf("only %d commands were actually compared against the local gate; the coverage map has gone vacuous",
			compared)
	}

	// 4. The two load-bearing release guards must be present in the release gate,
	//    not merely named by it.
	releaseScript := ciParityReadFile(t, filepath.Join(root, "scripts", "release-check.sh"))
	if !strings.Contains(releaseScript, "does not match tag") {
		t.Error("scripts/release-check.sh no longer carries the tag <-> VERSION guard from release.yml")
	}
	coverageScript := ciParityReadFile(t, filepath.Join(root, "scripts", "coverage-gate.sh"))
	if !strings.Contains(coverageScript, "workflows-disabled") {
		t.Error("scripts/coverage-gate.sh must DERIVE the coverage floor from the parked workflows, not retype it")
	}
	if regexp.MustCompile(`FLOOR="?8[0-9]`).MatchString(coverageScript) {
		t.Error("scripts/coverage-gate.sh appears to hardcode the coverage floor; it is derived from the workflow on purpose")
	}
}

// TestHostedCIIsDisabledDeliberately pins the operator decision so it cannot be
// undone by accident, and points at the README that says how to undo it on purpose.
func TestHostedCIIsDisabledDeliberately(t *testing.T) {
	root := ciParityRepoRoot(t)

	active, err := filepath.Glob(filepath.Join(root, CI_PARITY_ACTIVE_WORKFLOW_DIR, "*.y*ml"))
	if err != nil {
		t.Fatalf("globbing %s: %v", CI_PARITY_ACTIVE_WORKFLOW_DIR, err)
	}
	if len(active) > 0 {
		t.Errorf("hosted CI is deliberately disabled (see %s/README.md), but %v is present. Re-enabling is a decision, "+
			"not an accident — make it on purpose and update this test.", CI_PARITY_DISABLED_WORKFLOW_DIR, active)
	}

	readme := filepath.Join(root, CI_PARITY_DISABLED_WORKFLOW_DIR, CI_PARITY_README)
	if _, err := os.Stat(readme); err != nil {
		t.Errorf("%s must exist: a disabled gate with no stated reason gets re-enabled or deleted by the next person", readme)
	}
}

var (
	ciParityJobPattern    = regexp.MustCompile(`(?m)^  ([A-Za-z0-9_-]+):\s*$`)
	ciParityRunPattern    = regexp.MustCompile(`(?m)^\s*run:\s*(.*)$`)
	ciParityMatrixPattern = regexp.MustCompile(`(?m)^\s{6,}([a-z0-9_-]+):\s*\[`)
	ciParityTargetPattern = regexp.MustCompile(`(?m)^([A-Za-z0-9_.-]+):`)
)

// ciParityWorkflowJobs returns the job ids of a workflow: the keys indented by
// exactly two spaces under `jobs:`.
func ciParityWorkflowJobs(workflow string) []string {
	body := workflow
	if idx := strings.Index(workflow, "\njobs:"); idx >= 0 {
		body = workflow[idx:]
	}
	matches := ciParityJobPattern.FindAllStringSubmatch(body, -1)
	jobs := make([]string, 0, len(matches))
	for _, match := range matches {
		jobs = append(jobs, match[1])
	}
	return jobs
}

// ciParityMatrixAxes returns the axis names of every `strategy.matrix` block.
func ciParityMatrixAxes(workflow string) []string {
	var axes []string
	lines := strings.Split(workflow, "\n")
	inMatrix := false
	matrixIndent := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if trimmed == "matrix:" {
			inMatrix = true
			matrixIndent = indent
			continue
		}
		if !inMatrix {
			continue
		}
		if trimmed != "" && indent <= matrixIndent {
			inMatrix = false
			continue
		}
		if match := ciParityMatrixPattern.FindStringSubmatch(line); match != nil && indent == matrixIndent+2 {
			axes = append(axes, match[1])
		}
	}
	return axes
}

// ciParityJobCommands groups every run-step command by the job that runs it, so
// that a job declared an exception carries its own steps with it.
func ciParityJobCommands(workflow string) map[string][]string {
	body := workflow
	if idx := strings.Index(workflow, "\njobs:"); idx >= 0 {
		body = workflow[idx:]
	}
	result := map[string][]string{}
	lines := strings.Split(body, "\n")

	current := ""
	start := 0
	flush := func(end int) {
		if current == "" {
			return
		}
		result[current] = append(result[current], ciParityRunCommands(strings.Join(lines[start:end], "\n"))...)
	}
	for i, line := range lines {
		if match := ciParityJobPattern.FindStringSubmatch(line); match != nil {
			flush(i)
			current = match[1]
			start = i
		}
	}
	flush(len(lines))
	return result
}

// ciParityRunCommands flattens both the inline (`run: cmd`) and block (`run: |`)
// forms into one command per line.
func ciParityRunCommands(workflow string) []string {
	lines := strings.Split(workflow, "\n")
	var commands []string

	for i, line := range lines {
		match := ciParityRunPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		inline := strings.TrimSpace(match[1])
		if inline != "" && inline != "|" && inline != ">" {
			commands = append(commands, inline)
			continue
		}
		runIndent := len(line) - len(strings.TrimLeft(line, " "))
		for _, body := range lines[i+1:] {
			if strings.TrimSpace(body) == "" {
				continue
			}
			indent := len(body) - len(strings.TrimLeft(body, " "))
			if indent <= runIndent {
				break
			}
			commands = append(commands, strings.TrimSpace(body))
		}
	}
	return commands
}

// ciParityCommandHead returns the key a command is classified by: two words for
// `go <subcommand>`, one otherwise. A head this gate cannot classify is a
// FINDING at the call site, never a pass.
func ciParityCommandHead(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return ""
	}
	if fields[0] == "go" && len(fields) >= 2 {
		if fields[1] == "mod" && len(fields) >= 3 {
			return "go mod " + fields[2]
		}
		return "go " + fields[1]
	}
	if idx := strings.Index(fields[0], "="); idx > 0 {
		return fields[0][:idx] // shell variable assignment
	}
	return fields[0]
}

// ciParityGateText is every recipe reachable from `make ci-local`, plus the
// release scripts, with comments and quoted strings stripped.
//
// ⛔ The stripping is load-bearing. Without it a recipe's COMMENT saying "runs
// go vet" satisfies an assertion about `go vet ./...`, and prose about a command
// is not the command.
func ciParityGateText(t *testing.T, root, makefile string) string {
	t.Helper()

	targets := ciParityGateDeps(makefile, CI_PARITY_LOCAL_GATE)
	targets = append(targets, CI_PARITY_LOCAL_GATE, CI_PARITY_RELEASE_GATE)
	parts := make([]string, 0, len(targets)+2)
	for _, target := range targets {
		parts = append(parts, ciParityRecipeBody(makefile, target))
	}
	for _, script := range []string{"coverage-gate.sh", "release-check.sh"} {
		parts = append(parts, ciParityStripComments(ciParityReadFile(t, filepath.Join(root, "scripts", script))))
	}
	return strings.Join(parts, "\n")
}

func ciParityGateDeps(makefile, target string) []string {
	line := ciParityTargetLine(makefile, target)
	if line == "" {
		return nil
	}
	rhs := strings.TrimPrefix(line, target+":")
	if idx := strings.Index(rhs, "##"); idx >= 0 {
		rhs = rhs[:idx]
	}
	return strings.Fields(rhs)
}

func ciParityGateDependsOn(makefile, gate, target string) bool {
	for _, dep := range ciParityGateDeps(makefile, gate) {
		if dep == target {
			return true
		}
	}
	return false
}

func ciParityMakeTargetExists(makefile, target string) bool {
	return ciParityTargetLine(makefile, target) != ""
}

func ciParityTargetLine(makefile, target string) string {
	for _, line := range strings.Split(makefile, "\n") {
		if strings.HasPrefix(line, target+":") {
			return line
		}
	}
	return ""
}

func ciParityRecipeBody(makefile, target string) string {
	lines := strings.Split(makefile, "\n")
	start := -1
	for i, line := range lines {
		if strings.HasPrefix(line, target+":") {
			start = i + 1
			break
		}
	}
	if start == -1 {
		return ""
	}
	var body []string
	for _, line := range lines[start:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, "\t") {
			if ciParityTargetPattern.MatchString(line) || !strings.HasPrefix(line, "#") {
				break
			}
			continue
		}
		body = append(body, line)
	}
	return ciParityStripComments(strings.Join(body, "\n"))
}

// ciParityStripComments removes shell/make comments and quoted spans, so that
// only executable text can satisfy an assertion.
func ciParityStripComments(text string) string {
	var out strings.Builder
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		out.WriteString(ciParityStripQuoted(line))
		out.WriteString("\n")
	}
	return out.String()
}

func ciParityStripQuoted(text string) string {
	var out strings.Builder
	var quote rune
	for _, r := range text {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case r == '"' || r == '\'':
			quote = r
		default:
			out.WriteRune(r)
		}
	}
	return out.String()
}

func ciParityReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path) //nolint:gosec // a fixed, repo-relative path
	if err != nil {
		t.Fatalf("%s must exist: it is part of the gate this test compares. %v", path, err)
	}
	return string(data)
}

// ciParityRepoRoot is the directory holding go.mod — the tests run from the
// package directory, which for this single-module repo is the root.
func ciParityRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate the repository root (no go.mod found)")
		}
		dir = parent
	}
}
