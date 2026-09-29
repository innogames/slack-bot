package builder

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/innogames/slack-bot/v2/bot/config"
)

type sourceKind string

const (
	localSource  sourceKind = "local"
	gitSource    sourceKind = "git"
	moduleSource sourceKind = "module"
)

// plugin is a resolved entry of the "plugins" config section
type plugin struct {
	name   string
	cfg    config.PluginConfig
	kind   sourceKind
	commit string // checked out commit of a git source

	// module containing the plugin package
	modulePath string
	// local directory of the module (local and git sources), the module is replaced by this directory
	moduleDir string
	// package which gets imported in the generated main package
	importPath string
}

// e.g. git@gitlab.example.com:team/slack-bot-plugins.git
var scpLikeGitURL = regexp.MustCompile(`^[\w.-]+@[\w.-]+:`)

func getSourceKind(source string) sourceKind {
	switch {
	case isLocalPath(source):
		return localSource
	case strings.Contains(source, "://") || scpLikeGitURL.MatchString(source):
		return gitSource
	default:
		return moduleSource
	}
}

func isLocalPath(source string) bool {
	return source == "." || source == ".." ||
		strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../") ||
		strings.HasPrefix(source, `.\`) || strings.HasPrefix(source, `..\`) ||
		strings.HasPrefix(source, "/") || filepath.IsAbs(source)
}

// resolvePlugin detects the module and the import path of the plugin. Git sources are cloned into the work dir.
func (b *builder) resolvePlugin(ctx context.Context, name string, cfg config.PluginConfig) (*plugin, error) {
	p := &plugin{
		name: name,
		cfg:  cfg,
		kind: getSourceKind(cfg.Source),
	}

	var err error
	switch {
	case cfg.Source == "":
		return nil, errors.New(`no "source" defined`)
	case strings.HasPrefix(cfg.Source, "-") || strings.HasPrefix(cfg.Version, "-"):
		return nil, errors.New(`"source" and "version" must not start with "-"`)
	case cfg.Subdir != "" && !filepath.IsLocal(cfg.Subdir):
		return nil, fmt.Errorf(`"subdir" %q must be a relative directory within the source`, cfg.Subdir)
	case p.kind == localSource:
		dir := cfg.Source
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(b.baseDir, dir)
		}
		err = p.resolveLocal(filepath.Join(dir, cfg.Subdir), "")
	case p.kind == gitSource:
		var cloneDir string
		cloneDir, p.commit, err = b.cloneGit(ctx, cfg.Source, cfg.Version)
		if err == nil {
			err = p.resolveLocal(filepath.Join(cloneDir, cfg.Subdir), cloneDir)
		}
	default:
		err = p.resolveModule()
	}
	if err != nil {
		return nil, err
	}

	if cfg.Package != "" {
		p.importPath = cfg.Package
	}

	return p, nil
}

// resolveLocal detects the module of the plugin directory. The go.mod might be in a parent directory, up to the given root.
func (p *plugin) resolveLocal(dir string, root string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
		return fmt.Errorf("plugin directory %s does not exist", dir)
	}

	p.moduleDir, p.modulePath, err = findModule(dir, root)
	if err != nil {
		return err
	}

	relDir, err := filepath.Rel(p.moduleDir, dir)
	if err != nil {
		return err
	}
	p.importPath = path.Join(p.modulePath, filepath.ToSlash(relDir))

	return nil
}

// resolveModule for a Go module path like "github.com/innogames/slack-bot/plugins/aws"
func (p *plugin) resolveModule() error {
	p.modulePath = strings.TrimSuffix(p.cfg.Source, "/")

	// a Go module path always starts with a domain, like "github.com"
	firstElement, _, _ := strings.Cut(p.modulePath, "/")
	if !strings.Contains(firstElement, ".") {
		return fmt.Errorf("invalid source %q: use a Go module path (like github.com/org/plugin), a git url or a local directory (like ./%s)", p.cfg.Source, p.cfg.Source)
	}

	p.importPath = path.Join(p.modulePath, p.cfg.Subdir)

	return nil
}

// findModule searches the go.mod in the given directory and its parents (up to root, if given)
func findModule(dir string, root string) (moduleDir string, modulePath string, err error) {
	for current := dir; ; {
		goMod, err := os.ReadFile(filepath.Join(current, "go.mod")) // #nosec G304 -- reading the go.mod of the configured plugin
		if err == nil {
			modulePath = parseModulePath(goMod)
			if modulePath == "" {
				return "", "", fmt.Errorf("no module path defined in %s", filepath.Join(current, "go.mod"))
			}

			return current, modulePath, nil
		}

		parent := filepath.Dir(current)
		if current == root || parent == current {
			return "", "", fmt.Errorf("no go.mod found in %s or its parent directories", dir)
		}
		current = parent
	}
}

// parseModulePath extracts the module path of a go.mod file
func parseModulePath(goMod []byte) string {
	for line := range strings.Lines(string(goMod)) {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			return strings.Trim(fields[1], "\"`")
		}
	}

	return ""
}

// cloneGit clones the given ref (tag, branch or commit) of the git repository into the work dir
func (b *builder) cloneGit(ctx context.Context, url string, ref string) (dir string, commit string, err error) {
	url = strings.TrimPrefix(url, "git+")

	hash := sha256.Sum256([]byte(url + "@" + ref))
	dir = filepath.Join(b.workDir, "sources", hex.EncodeToString(hash[:8]))

	// the same repository is used by multiple plugins
	if commit, ok := b.clones[dir]; ok {
		return dir, commit, nil
	}

	if err = os.RemoveAll(dir); err != nil {
		return "", "", err
	}

	if ref == "" {
		err = b.run(ctx, b.workDir, "git", "clone", "--quiet", "--depth", "1", "--", url, dir)
	} else {
		// tags and branches can be cloned directly, other refs (like commits) need the whole history
		err = b.runSilent(ctx, b.workDir, "git", "clone", "--quiet", "--depth", "1", "--branch", ref, "--", url, dir)
		if err != nil {
			_ = os.RemoveAll(dir)
			err = b.run(ctx, b.workDir, "git", "clone", "--quiet", "--", url, dir)
			if err == nil {
				err = b.run(ctx, dir, "git", "-c", "advice.detachedHead=false", "checkout", "--quiet", ref, "--")
			}
		}
	}
	if err != nil {
		return "", "", fmt.Errorf("failed to clone %s: %w", url, err)
	}

	commit, err = b.output(ctx, dir, "git", "rev-parse", "HEAD")
	if err != nil {
		return "", "", err
	}

	b.clones[dir] = commit

	return dir, commit, nil
}
