//go:build !darwin && !windows

package sjskills

import "errors"

func publishNoReplace(source, destination string) error {
	return errors.New("unsupported project placement platform")
}

func replaceFileAtomic(source, destination string) error {
	return errors.New("unsupported project state platform")
}

func syncApplyDirectory(path string) error { return nil }
