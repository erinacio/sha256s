# sha256s
`sha256sum` compatible CLI tool with SIMD and parallelism.

## Usage

```
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
```
