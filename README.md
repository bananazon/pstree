# pstree
## Introduction
`pstree` is a small program that shows the process listing (`ps`) as a tree (as the name implies...). It has several options to make selection criteria and to change the output style. I ported the code from FredHucht's [repository](https://github.com/FredHucht/pstree) and added some more functionality.

It is targeted towards macOS as other Un*x variants already have `pstree` built-in. However, this version provides additional features and customization options.

It uses [gopsutil](https://github.com/shirou/gopsutil) for gathering process information. I may use `ps` later so I can support more Un*x distirbutions, but for now, this is what you get.

## Features

### Display Options
- Show process IDs (`--show-pids`)
- Show process group IDs (`--show-pgids`)
- Show parent process IDs (`--show-ppids`)
- Show command line arguments (`--arguments`)
- Show process owner information (`--show-owner`)
- Show process age in dd:hh:mm:ss format (`--age`)
- Show CPU utilization percentage (`--cpu`)
- Show memory usage in MiB (`--memory`)
- Show thread count for each process (`--threads`)

### Filtering and Selection
- Filter by process ID (`--pid`)
- Filter by username (`--user`)
- Filter by command line pattern (`--contains`)
- Exclude processes owned by root (`--exclude-root`)
- Limit tree depth (`--level`)

### Visualization
- Multiple line drawing character sets:
  - ASCII (default)
  - UTF-8 Unicode (`--utf-8`)
  - IBM-850 (`--ibm-850`)
  - VT-100 (`--vt-100`)
- Colorization options:
  - Standard colorization (`--color`)
  - Color by attribute (`--color-attr`):
    - Age: red (<1 min), orange (1 min-1 hr), yellow (1 hr-1 day), green (>1 day)
    - CPU: green (<5%), yellow (5-15%), red (>15%)
    - Memory: green (<10%), orange (10-20%), red (>20%)
  - Rainbow mode (`--rainbow`) for the adventurous
  - Custom color schemes (`--color-scheme`):
    - darwin (macOS optimized)
    - linux (Linux optimized)
    - powershell (PowerShell optimized)
    - windows10 (Windows optimized)
    - xterm (generic terminal)
- Process group leader indicators (`--show-pgls`)
- Wide output mode to prevent truncation (`--wide`)

### Security and Privilege Tracking
- Highlight user ID transitions (`--uid-transitions`)
- Highlight username transitions (`--user-transitions`)

### Output Control
- Non-compact mode to show all processes individually (`--compact-not`)
- Sort processes by various attributes (`--order-by`): age, cpu, mem, pid, threads, user
- All-inclusive mode to enable multiple options at once (`--all`)

## Compiling
* Clone this repository
* `cd` to the repository root
* Type `make build` and the binary will live under `bin` in the repository root
* You will need to manually copy `share/man/man1/pstree.1` to your `$MANPATH`
* You can also use homebrew
    * `brew tap bananazon/homebrew`
    * `brew update`
    * `brew install bananazon/homebrew/pstree`

## Usage
```
$ pstree --help
pstree $Revision: 0.9.3 $ by Cursed Bananazon (C) 2025, 2026

Usage: pstree [OPTIONS]

Display a tree of processes.

Application Options:
  -G, --age                   show the age of the process using the format (dd:hh:mm:ss)
  -A, --all                   equivalent to -acDGmOpSt
  -a, --arguments             show command line arguments
  -C, --color                 add some beautiful color to the pstree output; cannot be used with --color-attr or --rainbow
  -k, --color-attr string     color the process name by given attribute; implies --compact-not; valid options are: age, cpu, mem;
                              cannot be used with --color or --rainbow
  -q, --color-scheme string   override the default color scheme; valid options are: darwin, linux, powershell, windows10, xterm
  -n, --compact-not           do not compact identical subtrees in output
  -s, --contains string       show only branches containing processes with <pattern> in the command line; implies --compact-not
  -c, --cpu                   show CPU utilization percentage with each process, e.g., (c:0.00%); implies --compact-not
  -d, --debug count           Increase debugging level (-d, -dd, -ddd)
  -X, --exclude-root          don't show branches containing only root processes; cannot be used with --user
  -h, --help                  help for pstree
  -i, --ibm-850               use IBM-850 line drawing characters; only supported on DOS/Windows
  -l, --level int             print tree to <level> level deep
  -m, --memory                show the memory usage with each process, e.g., (m:x.y MiB); implies --compact-not
  -o, --order-by string       sort the results by <field>; valid options are: age, cpu, mem, pid, threads, user
  -P, --pid int32             show only branches containing process <pid>
  -r, --rainbow               for the adventurous; cannot be used with --color-attr or --color
  -O, --show-owner            show the owner of the process
  -g, --show-pgids            show process group IDs
  -S, --show-pgls             show process group leader indicators
  -p, --show-pids             show process IDs
  -D, --show-ppids            show parent process IDs
  -t, --threads               show the number of threads with each process, e.g., (t:xx)
  -I, --uid-transitions       show processes where the user ID changes from the parent process, e.g., (uid→uid); cannot be used with --user-transitions
      --user strings          show only branches containing processes of <user>; this option can be used more than and cannot be used with --exclude-root
  -U, --user-transitions      show processes where the user changes from the parent process, e.g., (user→user); cannot be used with --uid-transitions
  -u, --utf-8                 use UTF-8 (Unicode) line drawing characters
  -V, --version               display version information
  -v, --vt-100                use VT-100 line drawing characters
  -w, --wide                  wide output, not truncated to window width

Process group leaders are marked with '=' for ASCII, '¤' for IBM-850, '◆' for VT-100, and '●' for UTF-8.
```

## Testing

The pstree project includes various test suites to ensure code quality and functionality. Here are all the available test options:

### Running Basic Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests in a specific package
go test ./pkg/pstree
go test ./cmd
go test ./util
```

### Running Integration Tests

```bash
# Run integration tests in the cmd package
go test -v ./cmd -run TestMain

# Run specific integration test functions
go test -v ./cmd -run TestRealPstreeOutput
go test -v ./cmd -run TestShowPPIDsRealOutput
go test -v ./cmd -run TestOrderByRealOutput
go test -v ./cmd -run TestFlagCombinationsRealOutput
```

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run benchmarks in a specific package
go test -bench=. ./pkg/pstree

# Run specific benchmark functions
go test -bench=BenchmarkBuildTree ./pkg/pstree
go test -bench=BenchmarkOriginalBuildTree ./pkg/pstree
go test -bench=BenchmarkOptimizedBuildTree ./pkg/pstree

# Run benchmarks with memory allocation statistics
go test -bench=. -benchmem ./pkg/pstree
```

### Test Coverage

```bash
# Generate test coverage report
go test -cover ./...

# Generate detailed HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Test with Race Detection

```bash
# Run tests with race condition detection
go test -race ./...
```

### Short Mode Tests

```bash
# Run only short tests (skips integration tests)
go test -short ./...
```

### Continuous Integration

The project uses GitHub Actions for CI/CD. You can find the workflow configuration in the `.github/workflows` directory.

To run the same tests locally that run in CI:

```bash
make test
```

## Notes
* To view the man page for accuracy, use the command `groff -man -Tascii ./share/man/man1/pstree.1` or `groff -man -Tutf8 ./share/man/man1/pstree.1` if you've enabled UTF-8
* To generate the HTML man page, use the command `groff -Thtml -mandoc ./share/man/man1/pstree.1 > doc/pstree.1.html`
