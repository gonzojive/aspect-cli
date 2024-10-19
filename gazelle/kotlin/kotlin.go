package gazelle

import (
	"iter"
	"maps"
	"path"
	"strings"

	jvm_java "github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	jvm_types "github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"

	"aspect.build/gazelle/gazelle/kotlin/parser"
)

// IsNativeImport reports if the import literal is a native Kotlin or Java import.
func IsNativeImport(impt string) bool {
	if strings.HasPrefix(impt, "kotlin.") || strings.HasPrefix(impt, "kotlinx.") {
		return true
	}

	// Java native/standard libraries
	if jvm_java.IsStdlib(jvm_types.NewPackageName(impt)) {
		return true
	}

	return false
}

type KotlinTarget struct {
	Imports map[string]*ImportStatement
}

func (t *KotlinTarget) addImport(impt *ImportStatement) {
	t.Imports[impt.ImportHeader.Identifier().Literal()] = impt
}

func (t *KotlinTarget) importsSeq() iter.Seq[*ImportStatement] {
	return maps.Values(t.Imports)
}

/**
 * Information for kotlin library target including:
 * - kotlin files
 * - kotlin import statements from all files
 * - kotlin identifiers defined by the src files of this target.
 */
type KotlinLibTarget struct {
	KotlinTarget

	// Kotlin identifiers defiend by the srcs of this target.
	Identifiers map[string]*parser.Identifier

	// File names of Kotlin src files. File names should be relative to
	// this package.
	Files map[string]struct{}

	// The name of this library target if it already existed in the BUILD file
	// before generation of new rules.
	ExistingName string
}

func (t *KotlinLibTarget) addFile(file string) {
	t.Files[file] = struct{}{}
}

func (t *KotlinLibTarget) addExportedKotlinIdentifier(pkg *parser.Identifier) {
	t.Identifiers[pkg.Literal()] = pkg
}

func NewKotlinLibTarget() *KotlinLibTarget {
	return &KotlinLibTarget{
		KotlinTarget: KotlinTarget{
			Imports: make(map[string]*ImportStatement),
		},
		Identifiers: make(map[string]*parser.Identifier),
		Files:       make(map[string]struct{}),
	}
}

/**
 * Information for kotlin binary (main() method) including:
 * - kotlin import statements from all files
 * - the package
 * - the file
 */
type KotlinBinTarget struct {
	KotlinTarget

	File    string
	Package *parser.Identifier
}

func NewKotlinBinTarget(file string, pkg *parser.Identifier) *KotlinBinTarget {
	return &KotlinBinTarget{
		KotlinTarget: KotlinTarget{
			Imports: make(map[string]*ImportStatement),
		},
		File:    file,
		Package: pkg,
	}
}

// packagesKey is the name of a private attribute set on generated kt_library
// rules. This attribute contains the KotlinTarget for the target.
const packagesKey = "_kotlin_package"

func toBinaryTargetName(mainFile string) string {
	base := strings.ToLower(strings.TrimSuffix(path.Base(mainFile), path.Ext(mainFile)))

	// TODO: move target name template to directive
	return base + "_bin"
}

func toTestTargetName(mainFile string) string {
	base := strings.ToLower(strings.TrimSuffix(path.Base(mainFile), path.Ext(mainFile)))

	// TODO: move target name template to directive
	return base + "_test"
}

/**
 * Information for kotlin test including:
 * - kotlin import statements from all files
 * - the package
 * - the files
 * - the fully-qualified class name of the test
 */
type KotlinTestTarget struct {
	KotlinTarget

	Files     []string
	Package   *parser.Identifier
	TestClass *parser.Identifier
}

func NewKotlinTestTarget(files []string, pkg, testClass *parser.Identifier) *KotlinTestTarget {
	return &KotlinTestTarget{
		KotlinTarget: KotlinTarget{
			Imports: make(map[string]*ImportStatement),
		},
		Files:     files,
		Package:   pkg,
		TestClass: testClass,
	}
}
