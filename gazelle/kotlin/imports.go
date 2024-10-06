package gazelle

import (
	"strings"

	"aspect.build/cli/gazelle/kotlin/parser"
	"github.com/bazelbuild/bazel-gazelle/resolve"
)

// ImportStatement corresponds to a single Kotlin import.
type ImportStatement struct {
	resolve.ImportSpec

	// The path of the file containing the import
	SourcePath string

	// All of the parsed import information.
	ImportHeader *parser.ImportStatement
}

// packageFullyQualifiedName returns a [javaFullyQualifiedName] of package of the import.
//
// for "import foo.Bar as X", returns "foo"
// for "import foo.* as X", returns "foo"
func (is *ImportStatement) packageFullyQualifiedName() *javaFullyQualifiedName {
	return &javaFullyQualifiedName{strings.Split(is.Imp, ".")}
}

// javaFullyQualifiedName represents a fully-qualified name in Java, which is
// a dot-delimited list of [identifiers].
//
// [identifiers]: https://docs.oracle.com/javase/specs/jls/se23/html/jls-3.html#jls-Identifier
type javaFullyQualifiedName struct {
	// Each component of the path is an identifier as specified here:
	// https://kotlinlang.org/spec/syntax-and-grammar.html#grammar-rule-importList.
	//
	// A Java identifier that should mostly correspond to [3.8. Identifiers] from
	// the Java spec.
	//
	// [3.8. Identifiers]: https://docs.oracle.com/javase/specs/jls/se23/html/jls-3.html#jls-JavaLetter
	//
	// [L letter class]: https://stackoverflow.com/questions/5969440/what-is-the-l-unicode-category
	parts []string
}

// String  returns the dot-delimited java package name as it would appear in
// source code.
func (jpn *javaFullyQualifiedName) String() string {
	return strings.Join(jpn.parts, ".")
}

// Parent returns the parent package.
func (jpn *javaFullyQualifiedName) Parent() *javaFullyQualifiedName {
	if jpn == nil || len(jpn.parts) <= 1 {
		return nil
	}
	return &javaFullyQualifiedName{jpn.parts[0 : len(jpn.parts)-1]}
}

func importAsFQN(is *ImportStatement) *javaFullyQualifiedName {
	return &javaFullyQualifiedName{strings.Split(is.Imp, ".")}
}
