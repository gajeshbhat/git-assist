// Package docgen provides documentation extraction and generation utilities.
//
// This package can extract documentation from Go source code, CLI definitions,
// and configuration structs to automatically generate Diataxis-structured
// documentation for Hugo.
package docgen

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"strings"
)

// PackageDoc represents documentation for a Go package
type PackageDoc struct {
	Name       string    `json:"name"`
	ImportPath string    `json:"import_path"`
	Doc        string    `json:"doc"`
	Functions  []FuncDoc `json:"functions"`
	Types      []TypeDoc `json:"types"`
	Examples   []Example `json:"examples"`
}

// FuncDoc represents documentation for a function
type FuncDoc struct {
	Name string `json:"name"`
	Doc  string `json:"doc"`
	Decl string `json:"decl"`
	Recv string `json:"recv,omitempty"` // receiver type for methods
}

// TypeDoc represents documentation for a type
type TypeDoc struct {
	Name    string     `json:"name"`
	Doc     string     `json:"doc"`
	Decl    string     `json:"decl"`
	Methods []FuncDoc  `json:"methods"`
	Fields  []FieldDoc `json:"fields,omitempty"`
}

// FieldDoc represents documentation for a struct field
type FieldDoc struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Tag  string `json:"tag,omitempty"`
	Doc  string `json:"doc"`
}

// Example represents a code example
type Example struct {
	Name   string `json:"name"`
	Doc    string `json:"doc"`
	Code   string `json:"code"`
	Output string `json:"output,omitempty"`
}

// Extractor extracts documentation from Go source code
type Extractor struct {
	fset *token.FileSet
}

// NewExtractor creates a new documentation extractor
func NewExtractor() *Extractor {
	return &Extractor{
		fset: token.NewFileSet(),
	}
}

// ExtractPackage extracts documentation from a Go package directory
func (e *Extractor) ExtractPackage(pkgPath string) (*PackageDoc, error) {
	// Parse the package
	pkgs, err := parser.ParseDir(e.fset, pkgPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package %s: %w", pkgPath, err)
	}

	// Find the main package (non-test)
	var pkg *ast.Package
	for name, p := range pkgs {
		if !strings.HasSuffix(name, "_test") {
			pkg = p
			break
		}
	}

	if pkg == nil {
		return nil, fmt.Errorf("no non-test package found in %s", pkgPath)
	}

	// Create documentation
	docPkg := doc.New(pkg, pkgPath, doc.AllDecls)

	result := &PackageDoc{
		Name:       docPkg.Name,
		ImportPath: pkgPath,
		Doc:        docPkg.Doc,
	}

	// Extract functions
	for _, f := range docPkg.Funcs {
		funcDoc := FuncDoc{
			Name: f.Name,
			Doc:  f.Doc,
			Decl: e.formatDecl(f.Decl),
		}
		result.Functions = append(result.Functions, funcDoc)
	}

	// Extract types
	for _, t := range docPkg.Types {
		typeDoc := TypeDoc{
			Name: t.Name,
			Doc:  t.Doc,
			Decl: e.formatDecl(t.Decl),
		}

		// Extract methods
		for _, m := range t.Funcs {
			methodDoc := FuncDoc{
				Name: m.Name,
				Doc:  m.Doc,
				Decl: e.formatDecl(m.Decl),
				Recv: t.Name,
			}
			typeDoc.Methods = append(typeDoc.Methods, methodDoc)
		}

		// Extract struct fields if it's a struct
		if structType, ok := t.Decl.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType); ok {
			for _, field := range structType.Fields.List {
				for _, name := range field.Names {
					fieldDoc := FieldDoc{
						Name: name.Name,
						Type: e.formatType(field.Type),
					}
					if field.Tag != nil {
						fieldDoc.Tag = field.Tag.Value
					}
					if field.Doc != nil {
						fieldDoc.Doc = field.Doc.Text()
					}
					typeDoc.Fields = append(typeDoc.Fields, fieldDoc)
				}
			}
		}

		result.Types = append(result.Types, typeDoc)
	}

	return result, nil
}

// formatDecl formats an AST declaration as a string
func (e *Extractor) formatDecl(decl ast.Decl) string {
	// This is a simplified implementation
	// In a real implementation, you'd use go/format
	return "// Declaration formatting not implemented yet"
}

// formatType formats an AST type as a string
func (e *Extractor) formatType(expr ast.Expr) string {
	// This is a simplified implementation
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + e.formatType(t.X)
	case *ast.ArrayType:
		return "[]" + e.formatType(t.Elt)
	default:
		return "unknown"
	}
}

// GenerateHugoContent generates Hugo markdown content from package documentation
func (e *Extractor) GenerateHugoContent(pkg *PackageDoc) string {
	var content strings.Builder

	// Hugo front matter
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("title: \"Package %s\"\n", pkg.Name))
	content.WriteString("weight: 10\n")
	content.WriteString("bookToc: true\n")
	content.WriteString("---\n\n")

	// Package documentation
	content.WriteString(fmt.Sprintf("# Package %s\n\n", pkg.Name))
	if pkg.Doc != "" {
		content.WriteString(pkg.Doc)
		content.WriteString("\n\n")
	}

	// Import path
	content.WriteString("```go\n")
	content.WriteString(fmt.Sprintf("import \"%s\"\n", pkg.ImportPath))
	content.WriteString("```\n\n")

	// Types
	if len(pkg.Types) > 0 {
		content.WriteString("## Types\n\n")
		for _, t := range pkg.Types {
			content.WriteString(fmt.Sprintf("### %s\n\n", t.Name))
			if t.Doc != "" {
				content.WriteString(t.Doc)
				content.WriteString("\n\n")
			}

			// Struct fields
			if len(t.Fields) > 0 {
				content.WriteString("**Fields:**\n\n")
				for _, field := range t.Fields {
					content.WriteString(fmt.Sprintf("- `%s %s`", field.Name, field.Type))
					if field.Doc != "" {
						content.WriteString(" - " + strings.TrimSpace(field.Doc))
					}
					content.WriteString("\n")
				}
				content.WriteString("\n")
			}

			// Methods
			if len(t.Methods) > 0 {
				content.WriteString("**Methods:**\n\n")
				for _, method := range t.Methods {
					content.WriteString(fmt.Sprintf("#### %s\n\n", method.Name))
					if method.Doc != "" {
						content.WriteString(method.Doc)
						content.WriteString("\n\n")
					}
				}
			}
		}
	}

	// Functions
	if len(pkg.Functions) > 0 {
		content.WriteString("## Functions\n\n")
		for _, f := range pkg.Functions {
			content.WriteString(fmt.Sprintf("### %s\n\n", f.Name))
			if f.Doc != "" {
				content.WriteString(f.Doc)
				content.WriteString("\n\n")
			}
		}
	}

	return content.String()
}
