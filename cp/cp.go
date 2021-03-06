package cp

import (
  "fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

  "github.com/monochromegane/go-gitignore"
)

const (
	// tmpPermissionForDirectory makes the destination directory writable,
	// so that stuff can be copied recursively even if any original directory is NOT writable.
	// See https://github.com/otiai10/copy/pull/9 for more information.
	tmpPermissionForDirectory = os.FileMode(0755)
)

var pathToGitDir string
var pathToGitignore string
var hasGitignore bool
var isGoOnly bool

func SetGitignore(path string) {
  if path != "" {
    pathToGitignore = path
    hasGitignore = true
  }
}

func SetIsGoOnly(val bool) {
  isGoOnly = val
}

// Copy copies src to dest, doesn't matter if src is a directory or a file.
func Copy(src, dest string, opt ...Options) error {
  pathToGitDir = filepath.Join(src, ".git")

	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	return copy(src, dest, info, assure(opt...))
}

func isIgnored(path string, isDir bool) bool {
  result := false
  if hasGitignore {
    gitignore, _ := gitignore.NewGitIgnore(pathToGitignore)
    result = gitignore.Match(path, isDir)
  }
  return result
}

// copy dispatches copy-funcs according to the mode.
// Because this "copy" could be called recursively,
// "info" MUST be given here, NOT nil.
func copy(src, dest string, info os.FileInfo, opt Options) error {
  if src == pathToGitDir {
    fmt.Println("Ignoring", src)
    return nil
  }

	if opt.Skip(src) {
		return nil
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return onsymlink(src, dest, info, opt)
	}

	if info.IsDir() {
		return dcopy(src, dest, info, opt)
	}
	return fcopy(src, dest, info, opt)
}

// fcopy is for just a file,
// with considering existence of parent directory
// and file permission.
func fcopy(src, dest string, info os.FileInfo, opt Options) (err error) {
  if isGoOnly {
    if filepath.Ext(src) != ".go" {
      return nil
    }
  }
  if isIgnored(src, info.IsDir()) {
    fmt.Println("Ignoring", src)
    return nil
  }

  fmt.Println("Cloning", src)
	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer fclose(f, &err)

	if err = os.Chmod(f.Name(), info.Mode()|opt.AddPermission); err != nil {
		return err
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fclose(s, &err)

	_, err = io.Copy(f, s)
	return err
}

// dcopy is for a directory,
// with scanning contents inside the directory
// and pass everything to "copy" recursively.
func dcopy(srcdir, destdir string, info os.FileInfo, opt Options) (err error) {

	originalMode := info.Mode()

	// Make dest dir with 0755 so that everything writable.
	if err := os.MkdirAll(destdir, tmpPermissionForDirectory); err != nil {
		return err
	}
	// Recover dir mode with original one.
	defer chmod(destdir, originalMode|opt.AddPermission, &err)

	contents, err := ioutil.ReadDir(srcdir)
	if err != nil {
		return err
	}

	for _, content := range contents {
		cs, cd := filepath.Join(srcdir, content.Name()), filepath.Join(destdir, content.Name())
		if err := copy(cs, cd, content, opt); err != nil {
			// If any error, exit immediately
			return err
		}
	}

	return nil
}

func onsymlink(src, dest string, info os.FileInfo, opt Options) error {

	switch opt.OnSymlink(src) {
	case Shallow:
		return lcopy(src, dest)
	case Deep:
		orig, err := os.Readlink(src)
		if err != nil {
			return err
		}
		info, err = os.Lstat(orig)
		if err != nil {
			return err
		}
		return copy(orig, dest, info, opt)
	case Skip:
		fallthrough
	default:
		return nil // do nothing
	}
}

// lcopy is for a symlink,
// with just creating a new symlink by replicating src symlink.
func lcopy(src, dest string) error {
	src, err := os.Readlink(src)
	if err != nil {
		return err
	}
	return os.Symlink(src, dest)
}

// fclose ANYHOW closes file,
// with asiging error raised during Close,
// BUT respecting the error already reported.
func fclose(f *os.File, reported *error) {
	if err := f.Close(); *reported == nil {
		*reported = err
	}
}

// chmod ANYHOW changes file mode,
// with asiging error raised during Chmod,
// BUT respecting the error already reported.
func chmod(dir string, mode os.FileMode, reported *error) {
	if err := os.Chmod(dir, mode); *reported == nil {
		*reported = err
	}
}

// assure Options struct, should be called only once.
// All optional values MUST NOT BE nil/zero after assured.
func assure(opts ...Options) Options {
	if len(opts) == 0 {
		return DefaultOptions
	}
	if opts[0].OnSymlink == nil {
		opts[0].OnSymlink = DefaultOptions.OnSymlink
	}
	if opts[0].Skip == nil {
		opts[0].Skip = DefaultOptions.Skip
	}
	return opts[0]
}
