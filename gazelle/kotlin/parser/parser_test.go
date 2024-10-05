package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var testCases = []struct {
	desc, kt string
	filename string
	want     *ParseResult
}{
	{
		desc:     "empty",
		kt:       "",
		filename: "empty.kt",
		want: &ParseResult{
			File:    "empty.kt",
			Package: "",
			Imports: []string{},
		},
	},
	{
		desc: "simple",
		kt: `
import a.B
import c.D as E
	`,
		filename: "simple.kt",
		want: &ParseResult{
			File:    "simple.kt",
			Package: "",
			Imports: []string{"a", "c"},
		},
	},
	{
		desc: "stars",
		kt: `package a.b.c

import  d.y.* 
		`,
		filename: "stars.kt",
		want: &ParseResult{
			File:    "stars.kt",
			Package: "a.b.c",
			Imports: []string{"d.y"},
		},
	},
	{
		desc: "comments",
		kt: `
/*dlfkj*/package /*dlfkj*/ x // x
//z
import a.B // y
//z

/* asdf */ import /* asdf */ c./* asdf */D // w
import /* fdsa */ d/* asdf */.* // w
				`,
		filename: "comments.kt",
		want: &ParseResult{
			File:    "comments.kt",
			Package: "x",
			Imports: []string{"a", "c", "d"},
		},
	},
	{
		desc: "value class",
		kt: `
import a.b.C
import c.d.E as EEE

// Maybe xyz.numbers should be an import?
@JvmInline
value class Energy(val kwh: xyz.numbers.Double) { fun thing(): Unit {}}
	`,
		filename: "simple.kt",
		want: &ParseResult{
			File:    "simple.kt",
			Package: "",
			Imports: []string{"a.b", "c.d"},
		},
	},
	{
		desc: "companion object main",
		kt: `
package foo

class MyClass {
	companion object {
		fun main() = TODO("write me")
	}
}
	`,
		filename: "simple.kt",
		want: &ParseResult{
			File:    "simple.kt",
			Package: "foo",
			Imports: []string{},
			HasMain: true,
		},
	},
}

func TestTreesitterParser(t *testing.T) {

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			res, _ := NewParser().Parse(tc.filename, tc.kt)

			if diff := cmp.Diff(tc.want, res, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}

	t.Run("main detection", func(t *testing.T) {
		res, _ := NewParser().Parse("main.kt", "fun main() {}")
		if !res.HasMain {
			t.Errorf("main method should be detected")
		}

		res, _ = NewParser().Parse("x.kt", `
package my.demo
fun main() {}
		`)
		if !res.HasMain {
			t.Errorf("main method should be detected with package")
		}

		res, _ = NewParser().Parse("x.kt", `
package my.demo
import kotlin.text.*
fun main() {}
		`)
		if !res.HasMain {
			t.Errorf("main method should be detected with imports")
		}
	})
}

func equal[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
