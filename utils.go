package teak

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/urfave/cli.v1"
)

//DumpJSON - dumps JSON representation of given data to stdout
func DumpJSON(o interface{}) {
	b, err := json.MarshalIndent(o, "", "    ")
	if err == nil {
		fmt.Println(string(b))
	} else {
		LogErrorX("t.utils", "Failed to marshal data to JSON", err)
	}
}

//GetAsJSON - converts given data to JSON and returns as pretty printed
func GetAsJSON(o interface{}) (jstr string, err error) {
	b, err := json.MarshalIndent(o, "", "    ")
	if err == nil {
		jstr = string(b)
	}
	return jstr, LogErrorX("t.utils", "Failed to marshal data to JSON", err)
}

//GetExecDir - gives absolute path of the directory in which the executable
//for the current application is present
func GetExecDir() (dirPath string) {
	execPath, err := os.Executable()
	if err == nil {
		dirPath = filepath.Dir(execPath)
	} else {
		LogErrorX("t.utils", "Failed to get the executable path", err)
	}

	return dirPath
}

//ExistsAsFile - checks if a regular file exists at given path. If a error
//occurs while stating whatever exists at given location, false is returned
func ExistsAsFile(path string) (yes bool) {
	stat, err := os.Stat(path)
	if err == nil && !stat.IsDir() {
		yes = true
	}
	return yes
}

//ExistsAsDir - checks if a directory exists at given path. If a error
//occurs while stating whatever exists at given location, false is returned
func ExistsAsDir(path string) (yes bool) {
	stat, err := os.Stat(path)
	if err == nil && stat.IsDir() {
		yes = true
	}
	return yes
}

//ErrString - returns the error string if the given error is not nil
func ErrString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

//FirstValid - returns the first error that is not nil
func FirstValid(errs ...error) (err error) {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

//GetFirstValidStr - gets first string that is not empty
func GetFirstValidStr(strs ...string) (str string) {
	for _, s := range strs {
		if len(s) == 0 {
			str = s
			break
		}
	}
	return str
}

//MergeEnpoints - merge endpoints from different arrays to single array
func MergeEnpoints(epss ...[]*Endpoint) []*Endpoint {
	out := make([]*Endpoint, 0, 100)
	for _, eps := range epss {
		out = append(out, eps...)
	}
	return out
}

//MergeCommands - merge commands from different arrays to single array
func MergeCommands(cmdss ...[]*cli.Command) []*cli.Command {
	out := make([]*cli.Command, 0, 100)
	for _, cmds := range cmdss {
		out = append(out, cmds...)
	}
	return out
}
