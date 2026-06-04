package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/AntTheLimey/pgecloudctl/internal/auth"
	"github.com/AntTheLimey/pgecloudctl/internal/config"
	"github.com/AntTheLimey/pgecloudctl/internal/output"
	"github.com/spf13/cobra"
)

// --- report structs ----------------------------------------------------------

type versionInfo struct {
	Pgecloudctl string `json:"pgecloudctl"`
	Go          string `json:"go"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
}

type latestInfo struct {
	Latest   string `json:"latest"`
	Current  string `json:"current"`
	UpToDate bool   `json:"up_to_date"`
}

type authInfo struct {
	Authenticated bool   `json:"authenticated"`
	Source        string `json:"source,omitempty"`
	TokenValid    bool   `json:"token_valid"`
	ExpiresAt     string `json:"expires_at,omitempty"`
}

type apiInfo struct {
	Reachable bool   `json:"reachable"`
	URL       string `json:"url"`
	Status    int    `json:"status,omitempty"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
}

type configInfo struct {
	Dir          string `json:"dir"`
	ConfigExists bool   `json:"config_exists"`
	TokenExists  bool   `json:"token_exists"`
}

type environmentInfo struct {
	OS                string `json:"os"`
	Arch              string `json:"arch"`
	PgedgeClientIDSet bool   `json:"pgedge_client_id_set"`
	PgedgeAPIURLSet   bool   `json:"pgedge_api_url_set"`
	NoColorSet        bool   `json:"no_color_set"`
}

type shellInfo struct {
	Shell  string `json:"shell"`
	InPath bool   `json:"in_path"`
}

type skillInfo struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version,omitempty"`
}

type doctorReport struct {
	Version       versionInfo     `json:"version"`
	LatestVersion latestInfo      `json:"latest_version"`
	Auth          authInfo        `json:"auth"`
	API           apiInfo         `json:"api"`
	Config        configInfo      `json:"config"`
	Environment   environmentInfo `json:"environment"`
	Shell         shellInfo       `json:"shell"`
	InstallMethod string          `json:"install_method"`
	Skill         skillInfo       `json:"skill"`
}

// --- table row ---------------------------------------------------------------

type doctorRow struct {
	check   string
	status  string
	details string
}

func (r doctorRow) Columns() []string {
	return []string{r.check, output.ColorStatus(r.status), r.details}
}

// --- command wiring ----------------------------------------------------------

func init() {
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run environment diagnostics",
	Long: `doctor runs a series of checks to verify that pgecloudctl and its
environment are configured correctly. It does not require authentication
and can be run when auth is broken.`,
	RunE: runDoctor,
}

// --- check implementations ---------------------------------------------------

func checkVersion() versionInfo {
	return versionInfo{
		Pgecloudctl: Version,
		Go:          runtime.Version(),
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
	}
}

func checkLatestVersion() latestInfo {
	info := latestInfo{Current: Version}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(
		"https://api.github.com/repos/AntTheLimey/pgecloudctl/releases/latest",
	)
	if err != nil {
		return info
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return info
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return info
	}

	info.Latest = payload.TagName
	info.UpToDate = payload.TagName == Version ||
		payload.TagName == "v"+Version ||
		strings.TrimPrefix(payload.TagName, "v") == Version
	return info
}

func checkAuth() authInfo {
	store := config.DefaultStore()
	a := auth.Auth{Store: store, APIURL: flagAPIURL}

	_, source, err := a.ResolveCredentials(flagClientID, flagSecret)
	if err != nil {
		return authInfo{}
	}

	info := authInfo{Authenticated: true, Source: source}

	tok, err := store.LoadToken()
	if err != nil {
		return info
	}

	info.ExpiresAt = tok.ExpiresAt.Format(time.RFC3339)
	info.TokenValid = time.Until(tok.ExpiresAt) > 0
	return info
}

func checkAPI() apiInfo {
	info := apiInfo{URL: flagAPIURL}
	client := &http.Client{Timeout: 10 * time.Second}

	start := time.Now()
	resp, err := client.Get(flagAPIURL)
	elapsed := time.Since(start)

	if err != nil {
		return info
	}
	defer func() { _ = resp.Body.Close() }()

	info.Reachable = true
	info.Status = resp.StatusCode
	info.LatencyMs = elapsed.Milliseconds()
	return info
}

func checkConfig() configInfo {
	store := config.DefaultStore()
	info := configInfo{Dir: store.Dir}

	if _, err := os.Stat(filepath.Join(store.Dir, "config.json")); err == nil {
		info.ConfigExists = true
	}
	if _, err := os.Stat(filepath.Join(store.Dir, "token.json")); err == nil {
		info.TokenExists = true
	}
	return info
}

func checkEnvironment() environmentInfo {
	return environmentInfo{
		OS:                runtime.GOOS,
		Arch:              runtime.GOARCH,
		PgedgeClientIDSet: os.Getenv("PGEDGE_CLIENT_ID") != "",
		PgedgeAPIURLSet:   os.Getenv("PGEDGE_API_URL") != "",
		NoColorSet:        os.Getenv("NO_COLOR") != "",
	}
}

func checkShell() shellInfo {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "unknown"
	}

	_, err := exec.LookPath("pgecloudctl")
	return shellInfo{
		Shell:  shell,
		InPath: err == nil,
	}
}

func checkInstallMethod() string {
	exe, err := os.Executable()
	if err != nil {
		return "unknown"
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		resolved = exe
	}
	lower := strings.ToLower(resolved)
	switch {
	case strings.Contains(lower, "cellar") || strings.Contains(lower, "homebrew"):
		return "homebrew"
	case strings.Contains(lower, ".local/bin"):
		return "install-script"
	case strings.Contains(lower, "go/bin"):
		return "go-install"
	default:
		return "unknown"
	}
}

func checkSkill() skillInfo {
	home, err := os.UserHomeDir()
	if err != nil {
		return skillInfo{}
	}

	candidates := []string{
		filepath.Join(home, ".claude", "plugins", "pgecloudctl"),
		filepath.Join(home, ".claude", "plugins", "cache", "pgecloudctl"),
	}

	for _, dir := range candidates {
		skillPaths := []string{
			filepath.Join(dir, "skills", "pgecloudctl", "SKILL.md"),
			filepath.Join(dir, "SKILL.md"),
		}
		found := false
		for _, sp := range skillPaths {
			if _, err := os.Stat(sp); err == nil {
				found = true
				break
			}
		}
		if !found {
			continue
		}

		info := skillInfo{Installed: true}

		pluginPaths := []string{
			filepath.Join(dir, ".claude-plugin", "plugin.json"),
			filepath.Join(dir, "plugin.json"),
		}
		for _, pp := range pluginPaths {
			if data, err := os.ReadFile(pp); err == nil {
				var meta struct {
					Version string `json:"version"`
				}
				if err := json.Unmarshal(data, &meta); err == nil && meta.Version != "" {
					info.Version = meta.Version
				}
				break
			}
		}

		return info
	}

	return skillInfo{}
}

// --- runner ------------------------------------------------------------------

func runDoctor(cmd *cobra.Command, _ []string) error {
	report := doctorReport{
		Version:       checkVersion(),
		LatestVersion: checkLatestVersion(),
		Auth:          checkAuth(),
		API:           checkAPI(),
		Config:        checkConfig(),
		Environment:   checkEnvironment(),
		Shell:         checkShell(),
		InstallMethod: checkInstallMethod(),
		Skill:         checkSkill(),
	}

	if flagOutput != "table" {
		return output.Print(cmd.OutOrStdout(), flagOutput, report, nil)
	}

	v := report.Version
	env := report.Environment

	// Latest version status
	latestStatus := "warning"
	latestDetail := "could not check"
	if report.LatestVersion.Latest != "" {
		if report.LatestVersion.UpToDate {
			latestStatus = "ok"
			latestDetail = fmt.Sprintf("%s (up to date)", report.LatestVersion.Current)
		} else {
			latestDetail = fmt.Sprintf("%s (latest: %s)", report.LatestVersion.Current, report.LatestVersion.Latest)
		}
	}

	// Auth status
	authStatus := "error"
	authDetail := "not authenticated"
	if report.Auth.Authenticated && report.Auth.TokenValid {
		authStatus = "ok"
		authDetail = fmt.Sprintf("authenticated via %s (expires %s)", report.Auth.Source, report.Auth.ExpiresAt)
	} else if report.Auth.Authenticated {
		authStatus = "warning"
		authDetail = fmt.Sprintf("token expired (source: %s)", report.Auth.Source)
	}

	// API status
	apiStatus := "error"
	apiDetail := fmt.Sprintf("%s (unreachable)", report.API.URL)
	if report.API.Reachable {
		apiStatus = "ok"
		apiDetail = fmt.Sprintf("%s (%d, %dms)", report.API.URL, report.API.Status, report.API.LatencyMs)
	}

	// Config status
	configStatus := "ok"
	if !report.Config.ConfigExists {
		configStatus = "warning"
	}

	// Shell status
	shellSt := "ok"
	shellDet := report.Shell.Shell + ", in PATH"
	if !report.Shell.InPath {
		shellSt = "warning"
		shellDet = report.Shell.Shell + ", not in PATH"
	}

	// Env details
	envParts := []string{env.OS + "/" + env.Arch}
	if env.PgedgeClientIDSet {
		envParts = append(envParts, "PGEDGE_CLIENT_ID: set")
	} else {
		envParts = append(envParts, "PGEDGE_CLIENT_ID: not set")
	}

	// Skill status
	skillSt := "warning"
	skillDet := "not installed"
	if report.Skill.Installed {
		skillSt = "ok"
		skillDet = "installed"
		if report.Skill.Version != "" {
			skillDet += " (v" + report.Skill.Version + ")"
		}
	}

	rows := []output.Row{
		doctorRow{
			check:  "Version",
			status: "ok",
			details: fmt.Sprintf("%s (%s, %s/%s)",
				v.Pgecloudctl, v.Go, v.OS, v.Arch),
		},
		doctorRow{check: "Latest version", status: latestStatus, details: latestDetail},
		doctorRow{check: "Auth", status: authStatus, details: authDetail},
		doctorRow{check: "API connectivity", status: apiStatus, details: apiDetail},
		doctorRow{
			check:  "Config",
			status: configStatus,
			details: fmt.Sprintf("%s (config.json: %s, token.json: %s)",
				report.Config.Dir,
				boolYesNo(report.Config.ConfigExists),
				boolYesNo(report.Config.TokenExists)),
		},
		doctorRow{check: "Environment", status: "ok", details: strings.Join(envParts, ", ")},
		doctorRow{check: "Shell", status: shellSt, details: shellDet},
		doctorRow{check: "Install method", status: "ok", details: report.InstallMethod},
		doctorRow{check: "Skill", status: skillSt, details: skillDet},
	}

	headers := []string{"CHECK", "STATUS", "DETAILS"}
	return output.Print(cmd.OutOrStdout(), "table", rows, headers)
}

func boolYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
