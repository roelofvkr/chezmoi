package chezmoi

import (
	"archive/tar"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	vfs "github.com/twpayne/go-vfs"
)

// FIXME allow encrypted scripts
// FIXME add pre- and post- attributes

// A ScriptAttributes holds attributes parsed from a source script name.
type ScriptAttributes struct {
	Name     string
	Once     bool
	Template bool
}

// A Script represents a script to run.
type Script struct {
	sourceName       string
	targetName       string
	Once             bool
	Template         bool
	contents         []byte
	contentsErr      error
	evaluateContents func() ([]byte, error)
}

type scriptConcreteValue struct {
	Type       string `json:"type" yaml:"type"`
	SourcePath string `json:"sourcePath" yaml:"sourcePath"`
	TargetPath string `json:"targetPath" yaml:"targetPath"`
	Once       bool   `json:"once" yaml:"once"`
	Template   bool   `json:"template" yaml:"template"`
	Contents   string `json:"contents" yaml:"contents"`
}

// ParseScriptAttributes parses a source script file name.
func ParseScriptAttributes(sourceName string) ScriptAttributes {
	name := strings.TrimPrefix(sourceName, runPrefix)
	once := false
	template := false
	if strings.HasPrefix(name, oncePrefix) {
		once = true
		name = strings.TrimPrefix(name, oncePrefix)
	}
	if strings.HasSuffix(name, TemplateSuffix) {
		template = true
		name = strings.TrimSuffix(name, TemplateSuffix)
	}
	return ScriptAttributes{
		Name:     name,
		Once:     once,
		Template: template,
	}
}

// SourceName returns sa's source name.
func (sa ScriptAttributes) SourceName() string {
	sourceName := runPrefix
	if sa.Once {
		sourceName += oncePrefix
	}
	sourceName += sa.Name
	if sa.Template {
		sourceName += TemplateSuffix
	}
	return sourceName
}

// Apply runs s.
func (s *Script) Apply(fs vfs.FS, mutator Mutator, applyOptions *ApplyOptions) error {
	if applyOptions.Ignore(s.targetName) {
		return nil
	}
	contents, err := s.Contents()
	if err != nil {
		return err
	}

	if s.Once {
		ran, err := applyOptions.ScriptState.GetScriptRanState(contents)
		if err != nil {
			return err
		}
		if ran {
			return nil
		}
	}

	// FIXME verbose
	if applyOptions.DryRun {
		return nil
	}

	// Write the temporary script file.
	f, err := ioutil.TempFile("", filepath.Base(s.targetName))
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(f.Name())
	}()
	if err := os.Chmod(f.Name(), 0700); err != nil {
		return err
	}
	if _, err := f.Write(contents); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	// Run the temporary script file.
	c := exec.Command(f.Name())
	c.Dir = applyOptions.DestDir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	if err := c.Run(); err != nil {
		return err
	}

	if s.Once {
		if err := applyOptions.ScriptState.SetScriptRanState(contents); err != nil {
			return err
		}
	}

	return err
}

// ConcreteValue implements Entry.ConcreteValue.
func (s *Script) ConcreteValue(destDir string, ignore func(string) bool, sourceDir string, umask os.FileMode, recursive bool) (interface{}, error) {
	if ignore(s.targetName) {
		return nil, nil
	}
	contents, err := s.Contents()
	if err != nil {
		return nil, err
	}
	return &scriptConcreteValue{
		Type:       "script",
		SourcePath: filepath.Join(sourceDir, s.SourceName()),
		TargetPath: filepath.Join(destDir, s.TargetName()),
		Once:       s.Once,
		Template:   s.Template,
		Contents:   string(contents),
	}, nil
}

// Contents returns s's contents.
func (s *Script) Contents() ([]byte, error) {
	if s.evaluateContents != nil {
		s.contents, s.contentsErr = s.evaluateContents()
		s.evaluateContents = nil
	}
	return s.contents, s.contentsErr
}

// Evaluate evaluates s's contents.
func (s *Script) Evaluate(ignore func(string) bool) error {
	if ignore(s.targetName) {
		return nil
	}
	_, err := s.Contents()
	return err
}

// SourceName implements Entry.SourceName.
func (s *Script) SourceName() string {
	return s.sourceName
}

// TargetName implements Entry.TargetName.
func (s *Script) TargetName() string {
	return s.targetName
}

// archive writes s to w.
func (s *Script) archive(w *tar.Writer, ignore func(string) bool, headerTemplate *tar.Header, umask os.FileMode) error {
	if ignore(s.targetName) {
		return nil
	}
	contents, err := s.Contents()
	if err != nil {
		return err
	}
	header := *headerTemplate
	header.Typeflag = tar.TypeReg
	header.Name = s.targetName
	header.Size = int64(len(contents))
	header.Mode = int64(0777 &^ umask)
	if err := w.WriteHeader(&header); err != nil {
		return nil
	}
	_, err = w.Write(contents)
	return err
}
