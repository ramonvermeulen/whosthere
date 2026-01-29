package discovery

import (
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestDiscoveryHasNoRepoInternalDependencies(t *testing.T) {
	modulePath, err := readModulePath()
	if err != nil {
		t.Fatalf("read module path: %v", err)
	}

	allowedPrefix := modulePath + "/pkg/discovery"
	repoPrefix := modulePath + "/"

	var violations []string
	walkErr := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if base == "testdata" || strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fset := token.NewFileSet()
		f, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			violations = append(violations, path+": parse error: "+parseErr.Error())
			return nil
		}

		for _, spec := range f.Imports {
			imp, unqErr := strconv.Unquote(spec.Path.Value)
			if unqErr != nil {
				violations = append(violations, path+": invalid import path: "+spec.Path.Value)
				continue
			}

			if strings.HasPrefix(imp, allowedPrefix) {
				continue
			}

			if strings.HasPrefix(imp, repoPrefix) {
				pos := fset.Position(spec.Pos())
				violations = append(violations, pos.String()+": forbidden import "+imp)
			}
		}

		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk pkg: %v", walkErr)
	}

	if len(violations) > 0 {
		t.Fatalf("discovery must not depend on repo packages outside %q (GOOS=%s):\n%s", allowedPrefix, runtime.GOOS, strings.Join(violations, "\n"))
	}
}

func readModulePath() (string, error) {
	goModPathBytes, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		return "", err
	}
	goModPath := strings.TrimSpace(string(goModPathBytes))
	b, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	return "", os.ErrNotExist
}
