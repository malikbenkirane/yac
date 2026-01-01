package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/4sp1/yac/internal/agent"
	"github.com/4sp1/yac/internal/commit/config"
	"github.com/4sp1/yac/internal/commit/scope"
	"github.com/4sp1/yac/internal/commit/wip"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type vertexClient struct {
	projectId string
	model     string
	location  string
	token     string

	commitBody string
	userPrompt string

	logger *zap.Logger
}

func (vc *vertexClient) updateToken() error {
	{
		out, err := exec.Command("gcloud", "auth", "print-access-token").Output()
		if err != nil {
			return err
		}
		vc.token = string(out[:len(out)-1])
		vc.logger.Debug("googlcloud aiplatform token retrieved")
	}
	return nil
}

func (vc *vertexClient) preparePrompt(agent agent.Agent) error {
	user, err := agent.UserPrompt()
	if err != nil {
		return fmt.Errorf("agent: user prompt: %w", err)
	}
	vc.userPrompt = user
	return nil
}

func (vc *vertexClient) post(agent agent.Agent) error {
	var url = fmt.Sprintf("https://aiplatform.googleapis.com/v1/projects/%[3]s/locations/%[2]s/publishers/anthropic/models/%[1]s:streamRawPredict", vc.model, vc.location, vc.projectId)

	var err error
	defer func() {
		if err != nil {
			vc.logger.Error("vertexClient operator", zap.Error(err))
		}
	}()

	vc.logger.Debug("vertex ai post", zap.String("url", url))

	if err := vc.preparePrompt(agent); err != nil {
		return err
	}

	payload := struct {
		Version   string      `json:"anthropic_version"`
		Messges   []claudeMsg `json:"messages"`
		System    string      `json:"system"`
		Stream    bool        `json:"stream"`
		MaxTokens int         `json:"max_tokens"`
	}{
		Version: "vertex-2023-10-16",
		Messges: []claudeMsg{
			{
				Role:    "user",
				Content: []claudeMsgContent{newClaudeMsgTxt(vc.userPrompt)},
			},
		},
		MaxTokens: 3072,
		Stream:    false,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return err
	}

	if err = vc.updateToken(); err != nil {
		return fmt.Errorf("updateToken: %w", err)
	}

	vc.logger.Debug("new request for vertexai api",
		zap.String("anthropic_version", payload.Version),
		zap.Int("max_tokens", payload.MaxTokens),
		zap.String("model", vc.model),
		zap.String("location", vc.location),
		zap.String("project_id", vc.projectId))

	req.Header.Add("Authorization", "Bearer "+vc.token)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err = res.Body.Close()
	}()
	vc.logger.Debug("Vertex AI API response returned",
		zap.String("status", res.Status))

	var out bytes.Buffer
	_, err = io.Copy(&out, res.Body)
	if err != nil {
		return fmt.Errorf("io copy body: %w", err)
	}
	valid := strings.ToValidUTF8(out.String(), "?")
	r := strings.NewReader(valid)
	vc.logger.Debug("to valid utf-8", zap.String("valid?", valid))
	{
		var buf bytes.Buffer
		xc := exec.Command("jq", "-r", ".content[0].text")
		xc.Stdin = r
		xc.Stdout = &buf
		xc.Stderr = &buf
		if err = xc.Run(); err != nil {
			return fmt.Errorf("jq extract: %w", err)
		}
		vc.commitBody = buf.String()
	}

	return nil

}

func newScopt() []*bool {
	opts := make([]*bool, scope.UpperBound)
	for i := range opts {
		b := new(bool)
		opts[i] = b
	}
	return opts
}

type runtimeFlagsNew struct {
	path      string
	overwrite bool
	scopt     []*bool
	mode      config.Mode
}

func prepareFlags(r runtimeFlagsNew) (err error) {
	if r.scopt == nil {
		r.scopt = newScopt()
	}
	if r.path == "" {
		switch r.mode {
		case config.ModeJSON:
			r.path = config.DefaultJsonFile
		case config.ModeYAML:
			r.path = config.DefaultYamlFile
		}
	}
	f, err := config.Prepare(r.path, r.overwrite)
	if err != nil {
		return fmt.Errorf("config prepare: %w", err)
	}
	defer func() {
		err = f.Close()
	}()
	var commitScope = scope.Other
	for i := range r.scopt {
		if *r.scopt[i] {
			commitScope = scope.Scope(i)
			break
		}
	}

	switch r.mode {
	case config.ModeJSON:
		encoder := json.NewEncoder(f)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(
			config.NewJson(commitScope)); err != nil {
			return fmt.Errorf("json encoder: %w", err)
		}
	case config.ModeYAML:
		if err := yaml.NewEncoder(f).Encode(
			config.NewYaml(commitScope)); err != nil {
			return fmt.Errorf("yaml encoder: %w", err)
		}
	}
	fmt.Printf("Edit %q to prepare your next claude commit\n", r.path)
	return nil
}

func checkConfigFlags(json bool) (config.Mode, error) {
	mode := config.ModeYAML
	if json {
		mode = config.ModeJSON
	}
	return mode, nil
}

func newCommandClaudeCommitFlagsNew() *cobra.Command {
	var pathJson, pathYaml *string
	var overwrite, json *bool
	var scopt = make([]*bool, scope.UpperBound)
	cmd := &cobra.Command{
		Args: func(cmd *cobra.Command, args []string) error {
			_, err := checkConfigFlags(*json)
			return err
		},
		Use: "new",
		RunE: func(cmd *cobra.Command, args []string) error {
			mode, _ := checkConfigFlags(*json)
			path := *pathYaml
			if *json {
				path = *pathJson
			}
			return prepareFlags(runtimeFlagsNew{
				path:      path,
				scopt:     scopt,
				overwrite: *overwrite,
				mode:      mode,
			})
		}}

	pathJson = cmd.Flags().String("config-json", config.DefaultJsonFile,
		"specify json config file")
	pathYaml = cmd.Flags().String("config-yaml", config.DefaultYamlFile,
		"specify json config file")

	for i := range scope.UpperBound {
		scopt[i] = cmd.Flags().Bool(i.Flag(), false, "set commit scope")
	}

	overwrite = cmd.Flags().Bool("overwrite", false, "truncate current configuration")

	json = cmd.Flags().Bool("json", false, "read config-json instead of config-yaml")

	return cmd
}

func newCommandClaudeCommitFlags() *cobra.Command {
	cmd := &cobra.Command{
		Use: "flags",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newCommandClaudeCommitFlagsNew())
	return cmd
}

func newCommandClaudeCommit() *cobra.Command {
	var jj, noCommitOpt, noPost, debugDev, debugPrompt *bool
	var isJsonConfig *bool

	var wipt = make(map[wip.Context]*[]string)
	var logs *[]string
	var scopt = make([]*bool, scope.UpperBound)

	var configPath *string
	var prepare, noPrepare *bool

	cmd := &cobra.Command{
		Use:          "commit",
		Short:        "ask claude for a good commit message (vertexai)",
		SilenceUsage: true,
		Args: func(cmd *cobra.Command, args []string) error {
			var set bool
			for i := range scope.UpperBound {
				if *scopt[i] && set {
					return fmt.Errorf("commit scope was specified twice")
				}
				if *scopt[i] {
					set = true
				}
			}
			_, err := checkConfigFlags(*isJsonConfig)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			var debug *zap.Logger
			{
				config := zap.NewProductionConfig()
				if *debugDev {
					config = zap.NewDevelopmentConfig()
				}
				debug, err = config.Build()
				if err != nil {
					return fmt.Errorf("zap logger: %w", err)
				}
			}

			debug = debug.Named("claude_commit")

			configMode, _ := checkConfigFlags(*isJsonConfig)
			debug.Debug("config mode", zap.String("mode", configMode.String()))
			if configMode == config.ModeJSON {
				prefix, found := strings.CutSuffix(*configPath, ".yaml")
				if found {
					*configPath = prefix + ".json"
				}
			}

			if *prepare {
				f, err := config.Prepare(*configPath, true)
				if err := f.Close(); err != nil {
					return fmt.Errorf("close %q: %w", *configPath, err)
				}
				if err != nil {
					return fmt.Errorf("config prepare: %w", err)
				}
				return prepareFlags(runtimeFlagsNew{
					path:      *configPath,
					overwrite: true,
					mode:      configMode,
				})
			}

			debug.Debug("--no-commit flag valued", zap.Bool("no-commit", *noCommitOpt))

			vc := vertexClient{
				projectId: os.Getenv("GCP_VERTEXAI_PROJECT"),
				model:     "claude-sonnet-4-5@20250929",
				location:  "global",
				logger:    debug,
			}

			var hasConfig bool
			var rawConfig bytes.Buffer
			var v config.Flags
			if *noPrepare {
				configMode = config.ModeNone
			}
			if err := func(m config.Mode) (err error) {
				f, err := config.Prepare(*configPath, false)
				if err != nil {
					return fmt.Errorf("config prepare: %w", err)
				}
				defer func() {
					err = f.Close()
				}()
				t := io.TeeReader(f, &rawConfig)
				var w config.Config
				switch m {
				case config.ModeYAML:
					if err := yaml.NewDecoder(t).Decode(&w); err != nil {
						return fmt.Errorf("decode yaml config flags: %w", err)
					}
					v = config.FromYAML(w)
				case config.ModeJSON:
					if err := json.NewDecoder(t).Decode(&w); err != nil {
						return fmt.Errorf("decode json config flags: %w", err)
					}
					v = config.FromJSON(w)
				case config.ModeNone:
					return nil
				}
				hasConfig = true
				return nil
			}(configMode); err != nil {
				return fmt.Errorf("parse flags from json: %w", err)
			}

			var finalScope scope.Scope

			opts := []agent.Option{}

			if *jj {
				opts = append(opts, agent.WithJujutsuGitDiff())
			} else {
				opts = append(opts, agent.WithGitDiff())
			}

			// parse flags from command flags

			for _, hash := range *logs {
				if *jj {
					opts = append(opts, agent.WithJujutsuLog(hash))
				} else {
					opts = append(opts, agent.WithGitLog(hash))
				}
			}
			for i := range wip.UpperBound {
				for _, note := range *wipt[i] {
					opts = append(opts, agent.WithNote(note, i))
				}
			}
			for i := range scope.UpperBound {
				if *scopt[i] {
					opts = append(opts, agent.WithScope(i))
					finalScope = i
				}
			}
			opts = append(opts, agent.WithLogger(debug))

			// parse flags from config

			if v != nil {
				for _, hash := range v.FlagsLogs() {
					opts = append(opts, agent.WithGitLog(hash))
				}
				for section, notes := range v.FlagsWip() {
					for _, note := range notes {
						opts = append(opts, agent.WithNote(note, section))
					}
				}
			}

			debug.Debug("final scope setting", zap.String("scope", finalScope.String()))

			{
				agent, err := agent.New(opts...)
				if err != nil {
					return fmt.Errorf("new agent: %w", err)
				}
				if !*noPost {
					if err = vc.post(agent); err != nil {
						return fmt.Errorf("vx client post: %w", err)
					}
				} else {
					if err = vc.preparePrompt(agent); err != nil {
						return fmt.Errorf("prepare prompt: %w", err)
					}
				}
			}

			var commitTag string
			{
				debug.Debug("yag timestamp")
				ts, err := exec.Command("yag", "timestamp").CombinedOutput()
				if err != nil {
					return err
				}
				commitTag = string(ts)
			}

			// debugPrompt
			if err := func(save bool) error {
				if !save {
					return nil
				}
				_, err := os.Stat(".prompt")
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						if err := os.Mkdir(".prompt", 0700); err != nil {
							return fmt.Errorf("os mkdir: %w", err)
						}
					} else {
						return fmt.Errorf("os stat .prompt: %w", err)
					}
				}
				stat, err := os.Stat(".prompt")
				if err != nil && !stat.IsDir() {
					debug.Warn(".prompt already exist and is not a folder, will not save user prompt!")
					return nil
				}
				commitTag = strings.TrimSpace(commitTag)

				// backup user prompt
				{
					path := path.Join(".prompt", fmt.Sprintf("%s.md", commitTag))
					f, err := os.Create(path)
					if err != nil {
						return fmt.Errorf("os create %q: %w", path, err)
					}
					defer func() {
						if err := f.Close(); err != nil {
							debug.Warn("unable to close user prompt backup", zap.String("path", path), zap.Error(err))
						}
					}()
					debug.Debug("user prompt", zap.String("prompt", vc.userPrompt))
					_, err = io.Copy(f, strings.NewReader(vc.userPrompt))
					if err != nil {
						return fmt.Errorf("io copy user prompt: %w", err)
					}
					fmt.Println("Claude user prompt persists at", path)
					if !hasConfig {
						fmt.Println()
					}
				}

				// backup user configured flags via jsonFlagsPath option
				if hasConfig {
					path := path.Join(".prompt", configMode.File(commitTag))
					f, err := os.Create(path)
					if err != nil {
						return fmt.Errorf("os create %q: %w", path, err)
					}
					defer func() {
						if err := f.Close(); err != nil {
							debug.Debug("unable to close user flags backup",
								zap.String("path", path), zap.Error(err))
						}
					}()
					if _, err = io.Copy(f, &rawConfig); err != nil {
						return fmt.Errorf("io copy json flags backup: %w", err)
					}
					fmt.Println("Configured flags persist at", path)
					fmt.Println()
				}

				return nil
			}(*debugPrompt); err != nil {
				debug.Error("unable to save user prompt", zap.Error(err))
			}

			var finalCommit bytes.Buffer
			if !*noPost {
				var commitMsgBody string

				debug.Debug("create commit-stash")
				f, err := os.Create(".commit-stash")
				if err != nil {
					return err
				}

				debug.Debug(".commit-stash opened")
				defer func() {
					err = f.Close()
					if err != nil {
						debug.Error("unable to close .commit-stash", zap.Error(err))
					}
				}()

				w := io.MultiWriter(f, &finalCommit)
				if _, err = fmt.Fprintln(w, commitTag); err != nil {
					return fmt.Errorf("write finalCommit: %w", err)
				}

				if _, err = fmt.Fprintln(w, vc.commitBody); err != nil {
					return fmt.Errorf("write commitMsgBody: %w", err)
				}

				debug.Debug("write final commit", zap.String("body", commitMsgBody), zap.String("tag", commitTag))
			}

			if *noCommitOpt {
				red("\n\nnothing to commit\n")
				xc := exec.Command("pbcopy")
				xc.Stdout = os.Stdout
				xc.Stdin = strings.NewReader(finalCommit.String())
				xc.Stderr = os.Stderr
				debug.Debug("copy to pastebin", zap.String("final_commit_msg", finalCommit.String()))
				if err = xc.Run(); err != nil {
					return fmt.Errorf("unable to pbcopy: %w", err)
				}
				fmt.Println("To commit your changes, edit .commit-stash and")
				fmt.Println("git commit --file .commit-stash")
				return nil
			}

			{
				cmd := exec.Command("git", "commit", "--file", ".commit-stash")
				cmd.Stdout = os.Stdout
				cmd.Stdin = os.Stdin
				cmd.Stderr = os.Stderr
				if err = cmd.Run(); err != nil {
					return err
				}
				fmt.Println("To amend commit, edit .commit-stash and")
				fmt.Println("git commit --amend --file .commit-stash")
				return nil
			}

		},
	}

	jj = cmd.Flags().Bool("jj", true, "use jj instead of git")

	prepare = cmd.Flags().Bool("prepare", false, "prepare flags file and exit")
	noPrepare = cmd.Flags().Bool("no-prepare", true, "don't read flags from prepare file")
	isJsonConfig = cmd.Flags().Bool("json", false, "read config-json instead of config-yaml")

	debugDev = cmd.Flags().Bool("dev", false,
		"enable zap dev logger")

	debugPrompt = cmd.Flags().Bool("debug-prompt", false, "save prompts and config flags to .prompt")

	noPost = cmd.Flags().Bool("no-post", false, "do not post to claude")

	noCommitOpt = cmd.Flags().Bool("no-commit", true,
		"do not commit (review .commit_stash instead)")

	for i := range wip.UpperBound {
		wipt[i] = cmd.Flags().StringArray(i.Flag(),
			[]string{}, fmt.Sprintf("WIP %q Notes", i.Header()))
	}

	logs = cmd.Flags().StringArray("log",
		[]string{}, "related commit hash (can be repeated)")

	for i := range scope.UpperBound {
		scopt[i] = cmd.Flags().Bool(i.Flag(), false, i.Label())
	}

	configPath = cmd.Flags().String("config", config.DefaultYamlFile, "configure flags from file (--json or --yaml")

	cmd.AddCommand(newCommandClaudeCommitFlags())

	return cmd
}
