package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// BackupConfig is persisted to ~/.config/self-ledger/backup.yaml.
// It has both yaml and json tags so it round-trips correctly through
// the YAML file and Wails IPC (which uses JSON).
type BackupConfig struct {
	Provider string `yaml:"provider" json:"provider"` // github | gitlab | forgejo | gitea | custom
	Host     string `yaml:"host"     json:"host"`     // required for forgejo/gitea/custom
	Repo     string `yaml:"repo"     json:"repo"`     // "owner/repo-name"
	Token    string `yaml:"token"    json:"token"`    // personal access token
}

// RemoteURL builds an authenticated HTTPS remote URL with the token embedded
// so git uses it directly without needing a credential helper.
func (c BackupConfig) RemoteURL() string {
	host := c.Host
	if host == "" {
		switch c.Provider {
		case "github":
			host = "github.com"
		case "gitlab":
			host = "gitlab.com"
		}
	}
	u := &url.URL{Scheme: "https", Host: host, Path: "/" + c.Repo + ".git"}
	if c.Provider == "gitlab" {
		u.User = url.UserPassword("oauth2", c.Token)
	} else {
		u.User = url.User(c.Token) // token as username; handles special chars via url encoding
	}
	return u.String()
}

// BackupManager handles config I/O and git operations.
type BackupManager struct {
	configPath string // ~/.config/self-ledger/backup.yaml
	repoPath   string // ~/.config/self-ledger/backup-repo/
}

func NewBackupManager() (*BackupManager, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	appDir := filepath.Join(configDir, "self-ledger")
	return &BackupManager{
		configPath: filepath.Join(appDir, "backup.yaml"),
		repoPath:   filepath.Join(appDir, "backup-repo"),
	}, nil
}

// LoadConfig reads backup.yaml. Returns an empty BackupConfig (no error) if the
// file does not exist yet.
func (bm *BackupManager) LoadConfig() (BackupConfig, error) {
	data, err := os.ReadFile(bm.configPath)
	if os.IsNotExist(err) {
		return BackupConfig{}, nil
	}
	if err != nil {
		return BackupConfig{}, err
	}
	var cfg BackupConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return BackupConfig{}, err
	}
	return cfg, nil
}

// SaveConfig writes cfg to backup.yaml with mode 0600 (owner-only, contains token).
func (bm *BackupManager) SaveConfig(cfg BackupConfig) error {
	if err := os.MkdirAll(filepath.Dir(bm.configPath), 0750); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(bm.configPath, data, 0600)
}

// runGit runs a git command inside repoPath and returns combined output.
// "nothing to commit" is treated as a non-error.
func (bm *BackupManager) runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = bm.repoPath
	out, err := cmd.CombinedOutput()
	outStr := string(out)
	if err != nil {
		if strings.Contains(outStr, "nothing to commit") {
			return outStr, nil
		}
		return outStr, fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, outStr)
	}
	return outStr, nil
}

// ensureRepo initialises the local repo if needed and sets the remote URL.
func (bm *BackupManager) ensureRepo(remoteURL string) error {
	gitDir := filepath.Join(bm.repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		if _, err := bm.runGit("init"); err != nil {
			return err
		}
		if _, err := bm.runGit("remote", "add", "origin", remoteURL); err != nil {
			return err
		}
	} else {
		// Update remote URL in case it changed (e.g. token rotated).
		if _, err := bm.runGit("remote", "set-url", "origin", remoteURL); err != nil {
			return err
		}
	}
	return nil
}

// BackupNow writes jsonData to the local repo and pushes to the remote.
func (bm *BackupManager) BackupNow(jsonData string) error {
	cfg, err := bm.LoadConfig()
	if err != nil {
		return err
	}
	if cfg.Repo == "" {
		return fmt.Errorf("backup repo not configured")
	}

	if err := os.MkdirAll(bm.repoPath, 0750); err != nil {
		return err
	}

	remoteURL := cfg.RemoteURL()
	if err := bm.ensureRepo(remoteURL); err != nil {
		return err
	}

	backupFile := filepath.Join(bm.repoPath, "self_ledger_backup.json")
	if err := os.WriteFile(backupFile, []byte(jsonData), 0600); err != nil {
		return err
	}

	if _, err := bm.runGit("add", "self_ledger_backup.json"); err != nil {
		return err
	}

	msg := fmt.Sprintf("backup: %s", time.Now().UTC().Format(time.RFC3339))
	// Inline identity so backup works even on machines without global git user config.
	if _, err := bm.runGit(
		"-c", "user.email=self-ledger@local",
		"-c", "user.name=Self Ledger",
		"commit", "-m", msg,
	); err != nil {
		return err
	}

	if _, err := bm.runGit("push", "--force", "--set-upstream", "origin", "HEAD"); err != nil {
		return err
	}
	return nil
}

// FetchBackup pulls the latest backup from the remote and returns the JSON content.
func (bm *BackupManager) FetchBackup() (string, error) {
	cfg, err := bm.LoadConfig()
	if err != nil {
		return "", err
	}
	if cfg.Repo == "" {
		return "", fmt.Errorf("backup repo not configured")
	}

	remoteURL := cfg.RemoteURL()
	gitDir := filepath.Join(bm.repoPath, ".git")

	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Fresh machine — clone the repo.
		if err := os.MkdirAll(bm.repoPath, 0750); err != nil {
			return "", err
		}
		if _, err := bm.runGit("clone", remoteURL, "."); err != nil {
			return "", err
		}
	} else {
		// Repo exists — update remote URL and fetch latest file.
		if _, err := bm.runGit("remote", "set-url", "origin", remoteURL); err != nil {
			return "", err
		}
		if _, err := bm.runGit("fetch", "origin"); err != nil {
			return "", err
		}
		if _, err := bm.runGit("checkout", "origin/HEAD", "--", "self_ledger_backup.json"); err != nil {
			return "", err
		}
	}

	data, err := os.ReadFile(filepath.Join(bm.repoPath, "self_ledger_backup.json"))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
