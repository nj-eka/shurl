package fsutils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// ResolvePath {"" -> ".", "~..." -> "user.HomeDir..."} -> Abs
func ResolvePath(path string, usr *user.User) (string, error) {
	//	var err error
	if path == "" {
		path = "."
	}
	if strings.HasPrefix(path, "~") && usr != nil {
		//if usr == nil {
		//	usr, err = GetCurrentUser()
		//	if err != nil {
		//		return path, fmt.Errorf("resolving path [%s] failed due to inability to get user info: %w", path, err)
		//	}
		//}
		path = usr.HomeDir + path[1:]
	}
	return filepath.Abs(path)
}

func SafeParentResolvePath(path string, usr *user.User, perm os.FileMode) (string, error) {
	fullPath, err := ResolvePath(path, usr)
	if err != nil {
		return path, err
	}
	err = os.MkdirAll(filepath.Dir(fullPath), perm)
	if err != nil {
		return path, err
	}
	return fullPath, nil
}

func GetCurrentUser() (usr *user.User, err error) {
	if userName := os.Getenv("SUDO_USER"); userName != "" {
		usr, err = user.Lookup(userName)
	} else {
		usr, err = user.Current()
	}
	return usr, err
}

// IsDirectory checks whether path is directory and exists
func IsDirectory(path string) (b bool, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, err
		}
	}
	if !fi.IsDir() {
		return false, fmt.Errorf(`not a directory: %v`, path)
	}
	return true, nil
}
