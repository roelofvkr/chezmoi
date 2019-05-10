package chezmoi

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"

	vfs "github.com/twpayne/go-vfs"
)

// Suffixes and prefixes.
const (
	dotPrefix        = "dot_"
	emptyPrefix      = "empty_"
	encryptedPrefix  = "encrypted_"
	exactPrefix      = "exact_"
	executablePrefix = "executable_"
	oncePrefix       = "once_"
	privatePrefix    = "private_"
	runPrefix        = "run_"
	symlinkPrefix    = "symlink_"
	TemplateSuffix   = ".tmpl"
)

// A templateFuncError is an error encountered while executing a template
// function.
type templateFuncError struct {
	err error
}

// An ApplyOptions is a big ball of mud for things that affect Entry.Apply.
type ApplyOptions struct {
	DestDir     string
	DryRun      bool
	Ignore      func(string) bool
	ScriptState interface {
		GetScriptRanState([]byte) (bool, error)
		SetScriptRanState([]byte) error
	}
	Umask   os.FileMode
	Verbose bool
}

// An Entry is either a Dir, a File, or a Symlink.
type Entry interface {
	Apply(fs vfs.FS, mutator Mutator, applyOptions *ApplyOptions) error
	ConcreteValue(destDir string, ignore func(string) bool, sourceDir string, umask os.FileMode, recursive bool) (interface{}, error)
	Evaluate(ignore func(string) bool) error
	SourceName() string
	TargetName() string
	archive(w *tar.Writer, ignore func(string) bool, headerTemplate *tar.Header, umask os.FileMode) error
}

type parsedSourceFilePath struct {
	dirAttributes    []DirAttributes
	fileAttributes   *FileAttributes
	scriptAttributes *ScriptAttributes
}

// ReturnTemplateFuncError causes template execution to return an error.
func ReturnTemplateFuncError(err error) {
	panic(templateFuncError{
		err: err,
	})
}

// dirNames returns the dir names from dirAttributes.
func dirNames(dirAttributes []DirAttributes) []string {
	dns := make([]string, len(dirAttributes))
	for i, da := range dirAttributes {
		dns[i] = da.Name
	}
	return dns
}

// isEmpty returns true if b should be considered empty.
func isEmpty(b []byte) bool {
	return len(bytes.TrimSpace(b)) == 0
}

// parseDirNameComponents parses multiple directory name components.
func parseDirNameComponents(components []string) []DirAttributes {
	das := []DirAttributes{}
	for _, component := range components {
		da := ParseDirAttributes(component)
		das = append(das, da)
	}
	return das
}

// parseSourceFilePath parses a single source file path.
func parseSourceFilePath(path string) parsedSourceFilePath {
	components := splitPathList(path)
	das := parseDirNameComponents(components[0 : len(components)-1])
	sourceName := components[len(components)-1]
	if strings.HasPrefix(sourceName, runPrefix) {
		sa := ParseScriptAttributes(sourceName)
		return parsedSourceFilePath{
			dirAttributes:    das,
			scriptAttributes: &sa,
		}
	}
	fa := ParseFileAttributes(components[len(components)-1])
	return parsedSourceFilePath{
		dirAttributes:  das,
		fileAttributes: &fa,
	}
}

// sortedEntryNames returns a sorted slice of all entry names.
func sortedEntryNames(entries map[string]Entry) []string {
	entryNames := []string{}
	for entryName := range entries {
		entryNames = append(entryNames, entryName)
	}
	sort.Strings(entryNames)
	return entryNames
}

func splitPathList(path string) []string {
	if strings.HasPrefix(path, string(filepath.Separator)) {
		path = strings.TrimPrefix(path, string(filepath.Separator))
	}
	return strings.Split(path, string(filepath.Separator))
}
