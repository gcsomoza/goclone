package main

import (
  "fmt"
  "io/ioutil"
  "os"
	"path/filepath"
  "strings"

  "github.com/gcsomoza/goclone/cp"
)

const Help = `USAGE: goclone -s directory -d directory [-o --go-only]
This tool allows you to copy a go project to another directory
eliminating the hassle to modify your imports manually!

Clone and build your project straight away!

Arguments:
  -s, --source        Source directory.
  -d, --destination   Destination directory.
  -o, --overwrite     Overwrite destination directory.
  --go-only           Clone .go files only.

Example:
  goclone -s github.com/gcsomoza/hello -d github.com/gcsomoza/hello_world -o
`

var OLD string
var NEW string

func main() {

  isError := false
  isOverwrite := false
  isGoOnly := false
  args := os.Args

  n := len(args)
  if n < 5 {
    printHelp("Invalid number of arguments.")
    isError = true
  }

  source := ""
  destination := ""
  if !isError {
    source, destination, isError = parseArgs(args)
    if n > 5 {
      isOverwrite = hasOverwriteArg(args)
      isGoOnly = hasGoOnlyArg(args)
    }
  }

  goPath := os.Getenv("GOPATH")
  if !isError {
		if goPath == "" {
      printHelp("Your GOPATH environment variable is not set.")
      isError = true
		}
  }

  src := ""
  dst := ""
  hasGitignore := false
  pathToGitignore := ""
  if !isError {
    src = filepath.Join(goPath, "src", source)
    dst = filepath.Join(goPath, "src", destination)
    if _, err := os.Stat(src); os.IsNotExist(err) {
      printHelp(source + " does not exist")
      isError = true
    } else {
      pathToGitignore = filepath.Join(src, ".gitignore")
      if _, err := os.Stat(pathToGitignore); !os.IsNotExist(err) {
        hasGitignore = true
      }
    }
    if _, err := os.Stat(dst); !os.IsNotExist(err) {
      // Destination already exist
      if !isOverwrite {
        fmt.Println("ERROR: " + dst + " already exist.")
        isError = true
      }
    }
  }

  if !isError {
    fmt.Println("Cloning", source, "to", destination)

    if hasGitignore {
      fmt.Println("INFO: .gitignore detected. Ignoring files written in .gitignore")
      cp.SetGitignore(pathToGitignore)
      cp.SetIsGoOnly(isGoOnly)
    }

    err := cp.Copy(src, dst)
    if err != nil {
      fmt.Println("ERROR:", err)
      isError = true
    }
  }

  if !isError {
    OLD = source
    NEW = destination
    err := filepath.Walk(dst, visit)
    if err != nil {
      fmt.Println("ERROR:", err)
      isError = true
    }
  }

  if !isError {
    fmt.Println("Clone successful!")
  }
}

func inSlice(needle string, haystack []string) int {
  for i, hay := range haystack {
    if hay == needle {
      return i
    }
  }
  return -1
}

func printHelp(msg string) {
  fmt.Println("ERROR: " + msg)
  fmt.Println("")
  fmt.Println(Help)
}

func parseArgs(args []string) (string, string, bool) {
  params := make(map[string]string)
  params["source"] = ""
  params["destination"] = ""

  key := ""
  for i, arg := range args {
    if i == 1 || i == 3 {
      if arg == "-s" || arg == "--source" {
        key = "source"
      } else if arg == "-d" || arg == "--destination" {
        key = "destination"
      } else {
        printHelp("Invalid arguments.")
        return "", "", true
        break
      }
    } else if i == 2 || i == 4 {
      params[key] = arg
    }
  }
  if params["source"] == "" || params["destination"] == "" {
    return "", "", true
  }
  return params["source"], params["destination"], false
}

func hasOverwriteArg(args []string) bool {
  if inSlice("-o", args) > -1 {
    return true
  }
  if inSlice("--overwrite", args) > -1 {
    return true
  }
  return false
}

func hasGoOnlyArg(args []string) bool {
  if inSlice("--go-only", args) > -1 {
    return true
  }
  return false
}

func visit(path string, fi os.FileInfo, err error) error {

	if err != nil {
		return err
	}

	if !!fi.IsDir() {
		return nil //
	}

	matched, err := filepath.Match("*.go", fi.Name())

	if err != nil {
		panic(err)
		return err
	}

	if matched {
		read, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		//fmt.Println(string(read))
		//fmt.Println(path)

		newContents := strings.Replace(string(read), OLD, NEW, -1)

    //fmt.Println(fi.Mode())
		//fmt.Println(newContents)

		err = ioutil.WriteFile(path, []byte(newContents), fi.Mode())
		if err != nil {
			panic(err)
		}

	}

	return nil
}
