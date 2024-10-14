package kotlinconfig

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/javaconfig"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

type Directive[ParsedType any] struct {
	configKey string
	parseFn   func(d rule.Directive) (ParsedType, error)
}

// ConfigKey returns the string key used for this configuration directive
// in gazelle comments.
//
// For a directive like "# gazelle:blah foo bar", returns "blah".
func (d *Directive[ParsedType]) ConfigKey() string { return d.configKey }

// Parse parses the directive value.
func (d *Directive[ParsedType]) Parse(dir rule.Directive) (ParsedType, error) {
	return d.parseFn(dir)
}

// The directive for enable or disabling the gazelle plugin.
var EnabledDirective = &Directive[bool]{
	"kotlin",
	func(d rule.Directive) (bool, error) {
		switch strings.TrimSpace(d.Value) {
		case "enabled":
			return true, nil
		case "disabled":
			return false, nil
		default:
			return false, fmt.Errorf("invalid directive value %q for key %q: expected enabled or disabled", d.Key, d.Value)
		}
	},
}

type KotlinConfig struct {
	javaConfig *javaconfig.Config

	parent *KotlinConfig
	rel    string

	testFileSuffixes []string

	generationEnabled bool
}

type Configs = map[string]*KotlinConfig

func New(repoRoot string) *KotlinConfig {
	return &KotlinConfig{
		javaConfig:        javaconfig.New(repoRoot),
		generationEnabled: true,
		parent:            nil,
		testFileSuffixes:  []string{"Test.kt"},
	}
}

// String returns a debug string for the config.
func (c *KotlinConfig) String() string {
	return fmt.Sprintf("(KotlinConfig %q: enabled=%v; parent=\n  %s)", c.path(), c.generationEnabled, c.parent)
}

func (c *KotlinConfig) path() string {
	if c.parent == nil {
		return c.rel
	}
	return c.rel
	// return c.parent.path() + "/" + c.rel
}

// NewChild creates a new child Config. It inherits desired values from the
// current Config and sets itself as the parent to the child.
func (c *KotlinConfig) NewChild(childPath string) *KotlinConfig {
	cCopy := *c
	cCopy.javaConfig = c.javaConfig.NewChild()
	cCopy.rel = childPath
	cCopy.parent = c
	cCopy.testFileSuffixes = append([]string(nil), c.testFileSuffixes...)
	return &cCopy
}

// SetGenerationEnabled sets whether the extension is enabled or not.
func (c *KotlinConfig) SetGenerationEnabled(enabled bool) {
	c.generationEnabled = enabled
}

// GenerationEnabled returns whether the extension is enabled or not.
func (c *KotlinConfig) GenerationEnabled() bool {
	return c.generationEnabled
}

// JavaConfig returns the [javaconfig.Config] used as part of the Kotlin config.
func (c *KotlinConfig) JavaConfig() *javaconfig.Config {
	return c.javaConfig
}

// IsTestBaseName reports if the given basename within the same bazel package
// as the config should be considered a test.
func (c *KotlinConfig) IsTestBaseName(baseName string) bool {
	for _, suffix := range c.testFileSuffixes {
		if strings.HasSuffix(baseName, suffix) {
			return true
		}
	}
	return false
}

// ParentForPackage returns the parent Config for the given Bazel package.
func ParentForPackage(c Configs, pkg string) *KotlinConfig {
	dir := filepath.Dir(pkg)
	if dir == "." {
		dir = ""
	}
	parent := (map[string]*KotlinConfig)(c)[dir]
	return parent
}
