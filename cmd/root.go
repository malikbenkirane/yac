package cmd

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var srcDir = os.Getenv("YAG_SRCDIR")

func NewCLI() *cobra.Command {
	return newCommandClaudeCommit()
}

func NewLegacyCLI() *cobra.Command {
	if srcDir == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			panic(err)

		}
		srcDir = filepath.Join(h, "i", "fx", "yag")
	}
	out := os.Stdout
	git := gitCli{
		infoOut: os.Stdout,
	}.run
	rootCmd := &cobra.Command{
		Use:   "yag -- [file]*",
		Short: "Yet Another [Git]",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) >= 1 {
				// Check if args are existing files names.
				for _, fname := range args {
					info, err := os.Stat(fname)
					if os.IsNotExist(err) {
						return fmt.Errorf("%q does not exist", fname)
					}
					if info.IsDir() {
						entries, err := os.ReadDir(fname)
						if err != nil {
							return err
						}
						_, err = fmt.Fprintf(out, "%q is a directory, please cherry-pick or use (-d or --directory)\n\n", fname)
						if err != nil {
							return fmt.Errorf("fprintf out: %w", err)
						}
						// Loop throug the directory entries for further review
						for _, e := range entries {
							var note = "  "
							if e.IsDir() {
								note = "D "
							}
							_, err = fmt.Fprintf(out, "\033[33m%s\033[0m \033[32m%q\033[0m\n", note, filepath.Join(fname, e.Name()))
							if err != nil {
								return fmt.Errorf("fprintf out: %w", err)
							}

						}
						_, err = fmt.Fprintln(out)
						if err != nil {
							return fmt.Errorf("fprintln out: %w", err)
						}
						return git("status", fname)
					}
				}
				// Prepare git command args.
				gitArgs := append([]string{"add"}, args...)
				return git(gitArgs...)
			}
			stats, err := stat()
			if err != nil {
				return fmt.Errorf("git stat(): %w", err)
			}
			{
				u := make([]fstat, 0, len(stats))
				m := make([]fstat, 0, len(stats))
				for _, f := range stats {
					switch {
					case f.untrackedNewFile():
						u = append(u, f)
					case f.modified():
						m = append(m, f)
					}
				}
				if len(u) > 0 {
					fmt.Println()
					fmt.Print("💣 ")
					printUtil{
						out:       out,
						cut:       "staging untracked",
						noNewLine: true,
					}.yellowOnBlack()
					fmt.Println(" files:")
					for _, f := range u {
						fmt.Println(f.path)
					}
				}
				if len(m) > 0 {
					fmt.Println()
					fmt.Print("🧨 unstaged ")
					printUtil{
						out:       out,
						cut:       "modified",
						noNewLine: true,
					}.greenOnBlack()
					fmt.Println(" files:")
					fmt.Println()
					for _, f := range m {
						printUtil{out: out, cut: f.path}.greenOnBlack()
					}

					fmt.Println("\n💥💥💥💥💥")
					xc := exec.Command("git", "status", "-uno")
					xc.Stdout = os.Stdout
					xc.Stderr = os.Stderr
					if err = xc.Run(); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
	unstageCmd := &cobra.Command{
		Use:   "unstage [file]...",
		Short: "git restore --staged <file>...",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return git(append([]string{"restore", "--staged"}, args...)...)
		},
	}
	uCmd := &cobra.Command{
		Use:   "u",
		Short: "only 🎶 untracked files",
		RunE: func(cmd *cobra.Command, args []string) error {
			var b bytes.Buffer
			gitRun := gitCli{
				infoOut: out,
				cmdOut:  &b,
			}.run
			if err := gitRun("status", "--short", "--show-stash"); err != nil {
				return err
			}
			scanner := bufio.NewScanner(&b)
			var untracked []string
			for scanner.Scan() {
				line := scanner.Text()
				if cut, found := strings.CutPrefix(line, "?? "); found {
					untracked = append(untracked, cut)
				}
			}
			var err error
			if err = scanner.Err(); err != nil {
				return fmt.Errorf("scanner error: %w", err)
			}
			isEven := false
			sort.Strings(untracked)
			for _, cut := range untracked {
				root, cd, err := gitRoot()
				if err != nil {
					return err
				}
				cut = filepath.Join(cd, cut)
				cut, found := strings.CutPrefix(cut, root)
				if !found {
					return fmt.Errorf("%q not found prefix=%q", cut, root)
				}
				if isEven {
					printUtil{out: out, cut: cut}.seq(printYellow, printYellow, cut, printReset)
				} else {
					printUtil{out: out, cut: cut}.greenOnBlack()
				}
				isEven = !isEven
			}
			return nil
		},
	}
	tsCmd := &cobra.Command{
		Use:     "timestamp",
		Aliases: []string{"ts"},
		Short:   "make a timestamped tag for current location",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tstampFormat{}.print()
		},
	}
	tsLittCmd := &cobra.Command{
		Use:   "litt",
		Short: "litterate version of timestamp command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tstampFormat{
				litt: true,
			}.print()
		},
	}
	installCmd := &cobra.Command{
		Use:     "install",
		Aliases: []string{"i"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			{
				if err = os.Chdir(srcDir); err != nil {
					return err
				}
				c := exec.Command("go", "install", ".")
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				err = c.Run()
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	var tagRemoteOpt *string
	tagCmd := &cobra.Command{ //
		Use:   "tag",
		Short: "tag and push with last commit tag title",
		RunE: func(cmd *cobra.Command, args []string) error {
			var logOut bytes.Buffer
			if err := (gitCli{
				infoOut: out,
				cmdOut:  io.MultiWriter(&logOut),
			}.run("log", "-1", "--oneline")); err != nil {
				return err
			}
			logParts := strings.Split(logOut.String(), " ")
			tag := strings.TrimSpace(logParts[len(logParts)-1])
			if err := git("tag", tag); err != nil {
				return err
			}
			if err := git("push", *tagRemoteOpt, tag); err != nil {
				return err
			}
			return nil
		},
	}
	tagRemoteOpt = tagCmd.Flags().String("remote", "github", "git remote parameter")
	unoCmd := &cobra.Command{
		Use:   "uno",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return git("status", "-uno", ".")
		},
	}
	var listUntrackedOpt *bool
	skCmd := &cobra.Command{
		Use:   "sk",
		Short: "skim through git status",
		RunE: func(cmd *cobra.Command, args []string) error {
			var buf bytes.Buffer
		INSTR_LOOP:
			for {
				buf.Reset()
				status, err := stat()
				if err != nil {
					return fmt.Errorf("yag_stat: %w", err)
				}
				for _, fs := range status {
					if fs.modified() || (*listUntrackedOpt && fs.untracked()) {
						buf.WriteString(fs.path + "\n")
					}
				}
				buf.WriteString("help\ndone\ntag-last-commit\nclaude-commit\nclaude-commit-llamax\n")
				c := exec.Command("sk")
				c.Stdin = &buf
				var outBuf bytes.Buffer
				c.Stdout = &outBuf
				c.Stderr = os.Stderr
				err = c.Run()
				if err != nil {
					if exiterr, ok := err.(*exec.ExitError); ok {
						switch exiterr.ExitCode() {
						case 1:
							// no match
							fmt.Println("exit no match")
							return nil
						case 2:
							// error
							return err
						case 130:
							// interrupted (C-C or ESC)
							fmt.Println("exit interrupted")
							return nil
						}
					}
				}

				switch skInstr := strings.TrimSpace(outBuf.String()); skInstr {
				case "claude-commit-llamax":
					x := exec.Command("yag", "claude", "commit")
					x.Stdout = os.Stdout
					x.Stderr = os.Stderr
					x.Stdin = os.Stdin
					if err = x.Run(); err != nil {
						return fmt.Errorf("yag claude commit: %w", err)
					}
				case "claude-commit":
					x := exec.Command("yag", "claude", "commit", "--no-llama")
					x.Stdout = os.Stdout
					x.Stderr = os.Stderr
					x.Stdin = os.Stdin
					if err = x.Run(); err != nil {
						return fmt.Errorf("yag claude commit: %w", err)
					}
				case "help":
					fmt.Println("Nothing yet to help you here")
					scanner := bufio.NewScanner(os.Stdin)
					_ = scanner.Scan()
					if scanner.Err() != nil {
						return fmt.Errorf("scanner: %w", err)
					}
				case "done":
					break INSTR_LOOP
				case "tag-last-commit":
					x := exec.Command("yag", "tag")
					x.Stdout = os.Stdout
					x.Stderr = os.Stderr
					x.Stdin = os.Stdin
					if err = x.Run(); err != nil {
						return fmt.Errorf("yag claude commit: %w", err)
					}

				default:
					c = exec.Command("git", "add", skInstr)
					c.Stdout = os.Stdout
					c.Stderr = os.Stderr
					err = c.Run()
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
	listUntrackedOpt = skCmd.Flags().Bool("list-untracked",
		false, "include untracked files in skim list")
	claudeCmd := &cobra.Command{
		Use:   "claude",
		Short: "invoke claude",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.AddCommand(
		unstageCmd,
		newCommandOllamaCommit(),
		tagCmd,
		unoCmd,
		tsCmd,
		installCmd,
		skCmd,
		claudeCmd,
		uCmd)

	claudeCmd.AddCommand(newCommandClaudeCommit())

	tsCmd.AddCommand(tsLittCmd)
	yagRootCmd := &cobra.Command{
		Use:   "root",
		Short: "git root command",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, cd, err := gitRoot()
			if err != nil {
				return err
			}
			fmt.Println("root:", root)
			fmt.Println("cdir:", cd)
			return nil
		},
	}
	rootCmd.AddCommand(yagRootCmd)
	return rootCmd
}

type fstat struct {
	staged   byte
	unstaged byte
	path     string
}

func (fs fstat) untrackedNewFile() bool {
	return fs.staged == 'A' && fs.unstaged == ' '
}

func (fs fstat) modified() bool {
	return fs.unstaged == 'M'
}

func (fs fstat) untracked() bool {
	return fs.unstaged == '?' && fs.staged == '?'
}

func (fs fstat) isStaged() bool {
	return fs.staged != ' '
}

type gitCli struct {
	infoOut io.Writer
	cmdOut  io.Writer
	in      io.Reader
}

func (gc gitCli) run(args ...string) error {
	if len(args) == 0 {
		if _, err := fmt.Fprintln(gc.infoOut, "We are looking for changes in current directory:"); err != nil {
			return fmt.Errorf("git cli info out: %w", err)
		}
		if _, err := fmt.Fprintln(gc.infoOut); err != nil {
			return fmt.Errorf("git cli info out: %w", err)
		}
		args = []string{"status", "-uno", "."}
	}
	if gc.in == nil {
		gc.in = os.Stdin
	}
	if gc.cmdOut == nil {
		gc.cmdOut = os.Stdout
	}
	cmd := exec.Command("git", args...)
	cmd.Stdout = gc.cmdOut
	cmd.Stdin = gc.in
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func stat() ([]fstat, error) {
	c := exec.Command("git", "status", "-s")
	c.Stderr = os.Stderr
	var buf bytes.Buffer
	c.Stdout = &buf
	err := c.Run()
	if err != nil {
		return nil, err
	}
	var status []fstat
	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		t := scanner.Text()
		status = append(status, fstat{
			staged:   t[0],
			unstaged: t[1],
			path:     t[3:],
		})
	}
	return status, nil
}

var gitRoot = func() (gitroot, cd string, err error) {
	cd, err = os.Getwd()
	if err != nil {
		return
	}
	gitroot = cd
	{
		for {
			_, err = os.Stat(filepath.Join(gitroot, ".git"))
			if os.IsNotExist(err) {
				gitroot = filepath.Join(gitroot, "..")
				continue
			}
			break
		}
	}
	return
}

func timestamp(litt bool) error {
	tsfmt := "200601021504.05"
	if litt {
		tsfmt = "Mon.Jan.2.34PM"
	}
	tstr := time.Now().Format(tsfmt)
	var (
		gitroot, cd string
		err         error
	)

	gitroot, cd, err = gitRoot()
	if err != nil {
		return fmt.Errorf("gitroot: %w", err)
	}
	{
		var part1, part2 string
		cdpath := strings.Split(cd, string(os.PathSeparator))
		rootpath := strings.Split(gitroot, string(os.PathSeparator))
		delta := len(cdpath) - len(rootpath)
		if delta == 0 { // we are in git root directory
			part1 = "root"
		}
		if delta >= 1 { // only one depth: root of sub directory of the root
			part1 = cdpath[len(rootpath)]
		}
		if delta >= 2 {
			l := 2
			if delta == 2 {
				l = 1
			}
			part2 = strings.Join(cdpath[len(cdpath)-l:], ".")
		}
		tag := fmt.Sprintf("%s.dev-%s.%s", part1, tstr, part2)

		// Remove extra dot character
		if tag[len(tag)-1] == '.' {
			tag = tag[:len(tag)-1]
		}

		fmt.Println(tag)
	}
	return nil
}

type tstampFormat struct{ litt bool }

func (tsf tstampFormat) print() error {
	if err := timestamp(tsf.litt); err != nil {
		return err
	}
	return nil
}

type claudeMsg struct {
	Role    string             `json:"role"`
	Content []claudeMsgContent `json:"content"`
}

type claudeMsgContent interface {
	fmt.Stringer
	json.Marshaler
}

func newClaudeMsgTxt(text string) claudeMsgContent {
	return claudeMsgTxt{
		Type: "text",
		Text: text,
	}
}

func (c claudeMsgTxt) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}{
		Type: c.Type,
		Text: c.Text,
	})
}

func (cmc claudeMsgTxt) String() string {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(cmc); err != nil {
		panic(err)
	}
	return b.String()
}

type claudeMsgTxt struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

var red = func(s string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", s)
}

type printUtil struct {
	out       io.Writer
	cut       string
	noNewLine bool
}

const (
	printBlack  = "\033[40m"
	printYellow = "\033[33m"
	printGreen  = "\033[32m"
	printReset  = "\033[0m"
)

func (u printUtil) seq(steps ...string) {
	for _, lex := range steps {
		if _, err := fmt.Fprint(u.out, lex); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	nl := "\n"
	if u.noNewLine {
		nl = ""
	}
	if _, err := fmt.Fprint(u.out, nl); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (u printUtil) yellowOnBlack() {
	u.seq(
		printBlack,  // background
		printYellow, // foreground
		u.cut,       // cut
		printReset,  // reset colors
	)
}

func (u printUtil) greenOnBlack() {
	u.seq(
		printBlack,
		printGreen,
		u.cut,
		printReset,
	)
}
