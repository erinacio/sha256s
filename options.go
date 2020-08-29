package main

import (
	"errors"
	"io"
	"runtime"
	"strconv"

	"github.com/spf13/pflag"
)

type Options struct {
	Binary      bool
	Check       bool
	Jobs        int
	Dereference bool
	NativePath  bool
	Recursive   bool
	Tag         bool
	Zero        bool

	CrLf          bool
	IgnoreMissing bool
	Quiet         bool
	Status        bool
	Strict        bool
	Warn          bool

	Paths []string
}

func (o *Options) Parse(args []string) (err error) {
	*o = Options{}
	fs := pflag.NewFlagSet("sha256s", pflag.ContinueOnError)
	fs.BoolVarP(&o.Binary, "binary", "b", false, "")
	fs.BoolVarP(&o.Check, "check", "c", false, "")
	fs.IntVarP(&o.Jobs, "jobs", "j", 1, "")
	fs.BoolVarP(&o.Dereference, "dereference", "L", false, "")
	fs.BoolVar(&o.NativePath, "native-path", false, "")
	fs.VarPF((*NegateBoolValue)(&o.Dereference), "no-dereference", "P", "").NoOptDefVal = "true"
	fs.BoolVarP(&o.Recursive, "recursive", "r", false, "")
	fs.BoolVar(&o.Tag, "tag", false, "")
	fs.VarPF((*NegateBoolValue)(&o.Binary), "text", "t", "").NoOptDefVal = "true"
	fs.BoolVarP(&o.Zero, "zero", "z", false, "")
	fs.BoolVar(&o.CrLf, "crlf", false, "")
	fs.BoolVar(&o.IgnoreMissing, "ignore-missing", false, "")
	fs.BoolVarP(&o.Quiet, "quiet", "q", false, "")
	fs.BoolVar(&o.Status, "status", false, "")
	fs.BoolVar(&o.Strict, "strict", false, "")
	fs.BoolVarP(&o.Warn, "warn", "w", false, "")
	fs.VarPF(HelpRequestedError{}, "help", "h", "").NoOptDefVal = "x"
	fs.VarPF(VersionRequestedError{}, "version", "v", "").NoOptDefVal = "x"
	fs.Lookup("jobs").NoOptDefVal = strconv.Itoa(runtime.NumCPU())
	if err = fs.Parse(args); err != nil {
		return err
	}
	o.Paths = fs.Args()

	if o.CrLf && !o.Check {
		return errors.New("the --crlf option is meaningful only when verifying checksums")
	}
	if o.IgnoreMissing && !o.Check {
		return errors.New("the --ignore-missing option is meaningful only when verifying checksums")
	}
	if o.Quiet && !o.Check {
		return errors.New("the --quiet option is meaningful only when verifying checksums")
	}
	if o.Status && !o.Check {
		return errors.New("the --status option is meaningful only when verifying checksums")
	}
	if o.Strict && !o.Check {
		return errors.New("the --strict option is meaningful only when verifying checksums")
	}
	if o.Warn && !o.Check {
		return errors.New("the --warn option is meaningful only when verifying checksums")
	}
	if o.Recursive && o.Check {
		return errors.New("the --recursive option is meaningless when verifying checksums")
	}
	if o.Dereference && !o.Recursive {
		return errors.New("the --dereference option is meaningful only with --recursive")
	}

	if o.Jobs <= 0 {
		return errors.New("the --jobs option requires a positive integer argument")
	}
	if len(o.Paths) == 0 {
		o.Paths = []string{"-"}
	}

	if o.Check && !o.CrLf && runtime.GOOS == "windows" {
		o.CrLf = true
	}

	return nil
}

func (*Options) PrintHelp(out io.Writer) {
	_, _ = io.WriteString(out, Help[1:])
}

const Help = `
Usage: sha256s [OPTION]... [PATH]...
Print or check SHA256 (256-bit) checksums, using SIMD instructions for
acceleration if possible.

With no PATH, or when PATH is -, read standard input.

  -b, --binary          read in binary mode
  -c, --check           read SHA256 sums from the PATHs and check them
  -j [N], --jobs[=N]    allow N jobs at once, cpu number with no arg
  -L, --dereference     always follow symbolic links in PATHs
      --native-path     use backslash as path separator on Windows
  -P, --no-dereference  never follow symbolic links in PATHs (default)
  -r, --recursive       traverse directories in PATHs
      --tag             create or read a BSD-style checksum
  -t, --text            read in text mode (default)
  -z, --zero            end each output line with NUL, not newline,
                        and disable file name escaping

The following six options are useful only when verifying checksums:
      --crlf            allow checksum lines ending with CRLF, always true on
                          Windows system because Windows file names can't
                          contain "\r"
      --ignore-missing  don't fail or report status for missing files
  -q, --quiet           don't print OK for each successfully verified file
      --status          don't output anything, status code shows success
      --strict          exit non-zero for improperly formatted checksum lines
  -w, --warn            warn about improperly formatted checksum lines

  -h, --help     display this help and exit
  -v, --version  output version information and exit

Note: There is no difference between binary mode and text mode in this
      implementation.  These flags only affects output format, which will add
      '*' before file names in binary mode.  Command-line symbolic links in
      PATHs are always dereferenced, regardless of --no-dereference, so
      --dereference option is meaningful only with --recursive.
`

type HelpRequestedError struct{}

func (e HelpRequestedError) Error() string      { return "help requested" }
func (e HelpRequestedError) String() string     { return "" }
func (e HelpRequestedError) Set(s string) error { return e }
func (e HelpRequestedError) Type() string       { return "" }

type VersionRequestedError struct{}

func (e VersionRequestedError) Error() string      { return "version requested" }
func (e VersionRequestedError) String() string     { return "" }
func (e VersionRequestedError) Set(s string) error { return e }
func (e VersionRequestedError) Type() string       { return "" }

var (
	ErrHelpRequested    HelpRequestedError
	ErrVersionRequested VersionRequestedError
)

type NegateBoolValue bool

func (b NegateBoolValue) String() string { return strconv.FormatBool(bool(b)) }
func (b NegateBoolValue) Type() string   { return "bool" }
func (b *NegateBoolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	*b = NegateBoolValue(!v)
	return err
}
