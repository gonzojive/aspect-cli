package kotlinconfig

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/javaconfig"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// GenericDirective is a version of [Directive] without type arguments.
type GenericDirective interface {
	// ConfigKey returns the string key used for this configuration directive
	// in gazelle comments.
	//
	// For a directive like "# gazelle:blah foo bar", returns "blah".
	ConfigKey() string

	// Parse parses the directive value and updates the config.
	Parse(dir rule.Directive, cfg *KotlinConfig) error
}

type Directive[ParsedType any] struct {
	configKey string
	parseFn   func(d rule.Directive) (ParsedType, error)
	setFn     func(val ParsedType, cfg *KotlinConfig)
}

// ConfigKey returns the string key used for this configuration directive
// in gazelle comments.
//
// For a directive like "# gazelle:blah foo bar", returns "blah".
func (d *Directive[ParsedType]) ConfigKey() string { return d.configKey }

// parse parses the directive value.
func (d *Directive[ParsedType]) parse(dir rule.Directive) (ParsedType, error) {
	return d.parseFn(dir)
}

// Parse parses the directive value and updates the config.
func (d *Directive[ParsedType]) Parse(dir rule.Directive, cfg *KotlinConfig) error {
	val, err := d.parse(dir)
	if err != nil {
		return err
	}
	d.setFn(val, cfg)
	return nil
}

var librarySuffixRegexp = regexp.MustCompile(`^[^\s]*$`)

var (
	// The directive for enable or disabling the gazelle plugin.
	EnabledDirective = &Directive[bool]{
		"kotlin",
		parseEnabledDisableDirective,
		func(val bool, cfg *KotlinConfig) { cfg.SetGenerationEnabled(val) },
	}

	// A directive taking a single enabled/disabled argument that configures whether the
	// plugin should generate new library sources.
	OnlyUseExistingLibraryTargetsDirective = &Directive[bool]{
		"kotlin_only_use_existing_library_targets",
		parseEnabledDisableDirective,
		func(val bool, cfg *KotlinConfig) { cfg.SetOnlyUseExistingLibraryTargets(val) },
	}

	// A directive that configures the suffix used to name kt_jvm_library rules generated
	// by the plugin.
	LibraryRuleNameSuffix = &Directive[string]{
		"kotlin_library_suffix",
		func(d rule.Directive) (string, error) {
			value := strings.TrimSpace(d.Value)
			if !librarySuffixRegexp.MatchString(value) {
				return "", fmt.Errorf("invalid starlark name part %q - doesn't match regex %s", value, librarySuffixRegexp)
			}
			return value, nil
		},
		func(val string, cfg *KotlinConfig) { cfg.SetLibrarySuffix(val) },
	}
)

// AllDirectives returns all directives defined by the kotlin plugin. This list excludes
// directives relevant to the Kotlin plugin but not defined by the plugin such as those
// defined by the rules_jvm plugin.
func AllDirectives() []GenericDirective {
	return []GenericDirective{
		EnabledDirective,
		OnlyUseExistingLibraryTargetsDirective,
		LibraryRuleNameSuffix,
	}
}

func parseEnabledDisableDirective(d rule.Directive) (bool, error) {
	switch strings.TrimSpace(d.Value) {
	case "enabled":
		return true, nil
	case "disabled":
		return false, nil
	default:
		return false, fmt.Errorf("invalid directive value %q for key %q: expected enabled or disabled", d.Key, d.Value)
	}
}

type KotlinConfig struct {
	javaConfig *javaconfig.Config

	parent *KotlinConfig
	rel    string

	librarySuffix    string
	testFileSuffixes []string

	generationEnabled             bool
	onlyUseExistingLibraryTargets bool
}

type Configs = map[string]*KotlinConfig

func New(repoRoot string) *KotlinConfig {
	return &KotlinConfig{
		javaConfig:        javaconfig.New(repoRoot),
		generationEnabled: true,
		parent:            nil,
		testFileSuffixes:  []string{"Test.kt"},
		librarySuffix:     "_lib",
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

// SetOnlyUseExistingLibraryTargets sets the value of the
// only-use-existing-library-targets configuration value.
func (c *KotlinConfig) SetOnlyUseExistingLibraryTargets(enabled bool) {
	c.onlyUseExistingLibraryTargets = enabled
}

// OnlyUseExistingLibraryTargets returns the value of the
// only-use-existing-library-targets configuration value.
func (c *KotlinConfig) OnlyUseExistingLibraryTargets() bool {
	return c.onlyUseExistingLibraryTargets
}

// SetLibrarySuffix sets the suffix to be appended to the names of kt_jvm_library
// targets generated by the plugin.
func (c *KotlinConfig) SetLibrarySuffix(suffix string) {
	c.librarySuffix = suffix
}

// LibrarySuffix returns the suffix of kt_jvm_library targets generated by the
// plugin.
func (c *KotlinConfig) LibrarySuffix() string {
	return c.librarySuffix
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
