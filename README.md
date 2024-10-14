A Kotlin plugin for [bazel's gazelle
tool](https://github.com/bazelbuild/bazel-gazelle/tree/master) for automatically
generating BUILD.bazel files from Kotlin source code.

### Fork status
This code was extracted from https://github.com/aspect-build/aspect-cli. I would
like to merge it upstream depending on how the conversation goes with the
original authors. See https://github.com/aspect-build/aspect-cli/issues/750 for
discussion.


# Usage


# Dependency resolution

Gazelle must resolve what Bazel labels (like `@foo//bar/baz`) a Kotlin file depends on.
The rules followed by this plugin are as follows:

1. The dependencies of a Kotlin file are determined from the explicit imports in
   the file. Fully-qualified identifiers present elsewhere in the file are not
   considered when determining dependencies.

2. The resolution algorithm (as implemented in `gazelle/kotlin/resolver.go`)
   
3. An identifier matches a [Maven
   artifact](https://maven.apache.org/repositories/artifacts.html) if one of the
   packages declared as an export of that artifact is a package prefix of an
   import.

4. If the import to be resolved is in the library index, the import will be resolved to that library. If `-index=true`, Gazelle builds an index of library rules in the current repository before starting dependency resolution, and this is how most dependencies are resolved.

   1. For Kotlin, the match is based on the importpath attribute.

   2. For proto, the match is based on the srcs attribute.

## Target Assignment of Source Files

The plugin will use the following algorithm to decide what targets to create and
associate with source files within a Bazel package.

1. A default `kt_jvm_library` target is created for each package in the
   repository for which the plugin is enabled. Initially, the `srcs` associated
   with this target is the empty set. The name of this target is configurable
   using the `# gazelle:kotlin library_suffix foo` directive. By default, the
   name is the basename of the package directory + `"_lib"`.

2. The plugin will inspect each source file in a package and assign it to
   zero or more targets within the same package.

   1. If a source file is already present in the `srcs` of a target known
      to the plugin, the plugin will not create an additional target for
      the source file.

   2. **Binary Source Files:** Files containing a `main` function are assigned
      to a `kt_jvm_binary` target. This target will depend on all the
      `kt_jvm_library` targets in the same package as well as all of the
      labels that resolve based 


   3. **Test Source Files:** Files ending in `Test.kt` are assigned to a
      `kt_jvm_test` target.
   
   4. Otherwise, a `.kt` file will be assigned to the default `kt_jvm_library`.



### Note on target granularity

It is recommended to group all source files within a package into a single
`kt_jvm_library` target.

When there are multiple `kt_jvm_library` targets for a single Bazel package,
resolving dependencies between these targets is more challenging because import
statements are not used to express intra-package dependencies, and the plugin
does not attempt to perform the semantic analysis required to resolve all
identifiers in a source file. Because of this limitation, when choosing to use
multiple `library` targets per Bazel package, explicit `deps` entries with
`#keep` comments must be added by the author.


**8\. `only-use-existing-library-targets` Directive:**

-   `#gazelle:kotlin only-use-existing-library-targets enable` prevents the plugin from generating new library targets.
-   It requires users to explicitly define dependencies between existing targets using `deps` attributes and `#keep` comments to prevent dependency pruning.


# Terminology

**identifier** is used to refer to what the Kotlin spec calls an
[identifier](https://kotlinlang.org/spec/syntax-and-grammar.html#grammar-rule-identifier).
This is a fully-qualified or partially-qualified name. Fully-qualified
identifiers are used in `package` and `import` statements in Kotlin files.

**parent identifier**: In this document and in the code base, "parent identifier" means
an identifier with the last dot-delimited component removed. Sometimes it may be loosely
used to refer to all the ancestor identifiers as well (parent, parent of parent, etc.).

# Directives

Many [gazelle directives](https://github.com/bazelbuild/bazel-gazelle#directives) are generic
and will apply the the behavior of the Kotlin plugin.

The Kotlin plugin has additional directives for configuring behavior:

## gazelle:kotlin *\{enabledStatus: String\}*
Enables or disables the gazelle plugin in this bazel package and all child packages,
unless overridden.

Arguments:

- enabledStatus - String: One of "enabled" or "disabled"

## gazelle:kotlin_only_use_existing_library_targets *\{enabledStatus: String\}*
If enabled, only existing `kt_jvm_library` targets will be used. No new targets
will be generated.

If this mode is enabled, users must explicitly define intra-package dependencies
using `deps` attributes and `#keep` comments to prevent dependency pruning.

Arguments:

- enabledStatus - String: One of "enabled" or "disabled"

## gazelle:java_maven_install_file *\{path: String\}*

Specifies where the `maven_install.json` file is located.

This directive is defined by the [rules_jvm gazelle plugin](), and the Kotlin
plugins hares logic with the Java plugin to parse it.