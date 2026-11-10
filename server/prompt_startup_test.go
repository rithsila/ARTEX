package server

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// This database-free guard follows the startup seeding call graph: only initial
// seeding may write bodies. Never test this against an operator's database.
func TestPromptStartupNeverReplacesActiveBodies(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "assembly.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	funcs := map[string]*ast.FuncDecl{}
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok {
			funcs[fn.Name.Name] = fn
		}
	}
	visited := map[string]bool{}
	seeded := false
	var visit func(string)
	visit = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true
		fn := funcs[name]
		if fn == nil {
			return
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			switch target := call.Fun.(type) {
			case *ast.Ident:
				if target.Name == "refreshEnglishPrompts" {
					t.Error("startup invokes language-based prompt replacement")
				}
				visit(target.Name)
			case *ast.SelectorExpr:
				switch target.Sel.Name {
				case "SeedPromptIfEmpty":
					seeded = true
				case "SavePrompt":
					t.Errorf("startup %s replaces active prompt bodies", name)
				}
			}
			return true
		})
	}
	visit("seedPrompts")
	if !seeded {
		t.Fatal("initial-only prompt seeding was removed")
	}
	if funcs["refreshEnglishPrompts"] != nil {
		t.Error("obsolete body replacement migration remains")
	}
	if !visited["refreshEnglishToolCopy"] || !visited["refreshEnglishPromptNotes"] {
		t.Error("unrelated copy/note migrations were removed")
	}
}
