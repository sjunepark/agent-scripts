package sjskills

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const statusCacheVersion = 1
const maxStatusCacheBytes = 1 << 20
const statusRetention = 30 * 24 * time.Hour

// Retry metadata is tied to the same identity even without a successful map.
type statusCacheEntry struct {
	Version    int                 `json:"version"`
	Identity   string              `json:"identity"`
	Expected   map[string]TreeHash `json:"expected,omitempty"`
	ObservedAt time.Time           `json:"observedAt"`
	RetryAt    time.Time           `json:"retryAt"`
}

func (e statusCacheEntry) fresh(now time.Time) bool {
	return !e.ObservedAt.IsZero() && !e.ObservedAt.After(now) && now.Sub(e.ObservedAt) < statusRefreshInterval
}
func (e statusCacheEntry) retryDue(now time.Time) bool {
	// A retry farther ahead than the entire cooldown means the clock moved
	// backwards; permit a new attempt instead of waiting on the old clock.
	return e.RetryAt.IsZero() || !now.Before(e.RetryAt) || e.RetryAt.After(now.Add(statusRetryInterval))
}

func (s StatusService) cacheDirectory() (string, error) {
	root := s.CacheRoot
	var base string
	if root == "" {
		var err error
		base, err = os.UserCacheDir()
		if err != nil {
			return "", err
		}
		// The OS-selected cache directory may intentionally be redirected. Only
		// this platform anchor is canonicalized; our own descendants reject links.
		if err := os.MkdirAll(base, 0700); err != nil {
			return "", err
		}
		base, err = filepath.EvalSymlinks(base)
		if err != nil {
			return "", err
		}
		root = filepath.Join(base, "sjskills", "status")
	} else {
		var err error
		root, err = filepath.Abs(root)
		if err != nil {
			return "", err
		}
		base = filepath.Dir(root)
		if err := os.MkdirAll(base, 0700); err != nil {
			return "", err
		}
		base, err = filepath.EvalSymlinks(base)
		if err != nil {
			return "", err
		}
		root = filepath.Join(base, filepath.Base(root))
	}
	relative, err := filepath.Rel(base, root)
	if err != nil || !filepath.IsLocal(relative) {
		return "", errors.New("invalid status cache directory")
	}
	current := base
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		if err := os.Mkdir(current, 0700); err != nil && !errors.Is(err, os.ErrExist) {
			return "", err
		}
		info, err := os.Lstat(current)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("unsafe status cache directory")
		}
	}
	return root, nil
}

func readStatusFile(path string) ([]byte, error) {
	before, err := lstatIdentity(path)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() || before.Size() > maxStatusCacheBytes {
		return nil, errors.New("unsafe status cache file")
	}
	file, err := openStatusFile(path, false)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	after, err := file.Stat()
	if err != nil || !os.SameFile(before, after) {
		return nil, errors.New("status cache changed")
	}
	data, err := io.ReadAll(io.LimitReader(file, maxStatusCacheBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxStatusCacheBytes {
		return nil, errors.New("oversized status cache")
	}
	return data, nil
}
func (s StatusService) read(scope StatusScope) (statusCacheEntry, error) {
	directory, err := s.cacheDirectory()
	if err != nil {
		return statusCacheEntry{}, err
	}
	data, err := readStatusFile(filepath.Join(directory, scope.cacheKey()+".json"))
	if err != nil {
		return statusCacheEntry{}, err
	}
	var entry statusCacheEntry
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&entry); err != nil {
		return statusCacheEntry{}, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return statusCacheEntry{}, errors.New("trailing status cache data")
	}
	if entry.Version != statusCacheVersion || entry.Identity != scope.identity() {
		return statusCacheEntry{}, errors.New("incompatible status cache")
	}
	// A failed cold refresh has no snapshot but retains its bounded cooldown.
	if !entry.ObservedAt.IsZero() && !validStatusExpected(scope.Plan.Desired, entry.Expected) {
		return statusCacheEntry{}, errors.New("incomplete status cache")
	}
	if entry.ObservedAt.IsZero() && len(entry.Expected) != 0 {
		return statusCacheEntry{}, errors.New("unobserved status cache")
	}
	return entry, nil
}
func (s StatusService) lock(scope StatusScope) (func(), error) {
	directory, err := s.cacheDirectory()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(directory, scope.cacheKey()+".lock")
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return nil, errors.New("unsafe cache lock")
	}
	file, err := createApplyLockFile(path)
	if errors.Is(err, os.ErrExist) {
		file, err = openStatusFile(path, true)
	}
	if err != nil {
		return nil, err
	}
	if err := lockApplyFile(file); err != nil {
		file.Close()
		return nil, err
	}
	info, statErr := file.Stat()
	current, pathErr := lstatIdentity(path)
	if statErr != nil || pathErr != nil || !info.Mode().IsRegular() || !current.Mode().IsRegular() || !os.SameFile(info, current) {
		unlockApplyFile(file)
		file.Close()
		return nil, errors.New("cache lock changed")
	}
	return func() {
		// Remove while held. A waiter that opened the old inode must verify it is
		// still linked after obtaining the lock before touching the cache.
		current, err := lstatIdentity(path)
		if err == nil && os.SameFile(info, current) {
			_ = os.Remove(path)
		}
		_ = unlockApplyFile(file)
		_ = file.Close()
	}, nil
}
func (s StatusService) publish(scope StatusScope, entry statusCacheEntry) error {
	unlock, err := s.lock(scope)
	if err != nil {
		return err
	}
	defer unlock()
	old, err := s.read(scope)
	if err == nil && old.ObservedAt.After(entry.ObservedAt) && !old.ObservedAt.After(s.now()) {
		return nil
	}
	if err := s.write(scope, entry); err != nil {
		return err
	}
	s.prune(scope)
	return nil
}
func (s StatusService) write(scope StatusScope, entry statusCacheEntry) error {
	directory, err := s.cacheDirectory()
	if err != nil {
		return err
	}
	path := filepath.Join(directory, scope.cacheKey()+".json")
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return errors.New("unsafe cache destination")
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if len(data) > maxStatusCacheBytes {
		return errors.New("oversized cache snapshot")
	}
	file, err := os.CreateTemp(directory, scope.cacheKey()+".tmp-")
	if err != nil {
		return err
	}
	temp := file.Name()
	defer os.Remove(temp)
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(temp, path)
}
func (s StatusService) prune(scope StatusScope) {
	directory, err := s.cacheDirectory()
	if err != nil {
		return
	}
	dir, err := os.Open(directory)
	if err != nil {
		return
	}
	defer dir.Close()
	// Bound directory work as well as deletion work. A fresh scope replaces its
	// previous identity in place, so ordinary use cannot accumulate versions.
	entries, _ := dir.ReadDir(256)
	for _, entry := range entries {
		name := entry.Name()
		if len(name) < 69 || !lowercaseDigestPattern.MatchString(name[:64]) {
			continue
		}
		if !strings.HasSuffix(name, ".json") && !strings.Contains(name, ".tmp-") && !strings.HasSuffix(name, ".lock") {
			continue
		}
		info, err := lstatIdentity(filepath.Join(directory, name))
		if err != nil || !info.Mode().IsRegular() || s.now().Sub(info.ModTime()) < statusRetention {
			continue
		}
		if strings.HasPrefix(name, scope.cacheKey()) {
			if strings.Contains(name, ".tmp-") {
				_ = os.Remove(filepath.Join(directory, name))
			}
			continue
		}
		// Reuse the per-scope lock so pruning cannot race an active writer. The
		// key alone suffices here; no cache payload or foreign path is followed.
		path := filepath.Join(directory, name[:64]+".lock")
		file, err := createApplyLockFile(path)
		if errors.Is(err, os.ErrExist) {
			file, err = openStatusFile(path, true)
		}
		if err != nil {
			continue
		}
		if lockApplyFile(file) != nil {
			file.Close()
			continue
		}
		linked, e := lstatIdentity(path)
		opened, se := file.Stat()
		if e == nil && se == nil && linked.Mode().IsRegular() && os.SameFile(linked, opened) {
			// Candidate discovery precedes locking. A peer may have published a
			// replacement in between; only remove the same, still-expired file.
			s.removeExpiredStatusFile(filepath.Join(directory, name), info)
			_ = os.Remove(path)
		}
		_ = unlockApplyFile(file)
		_ = file.Close()
	}
}

// The caller holds this entry's scope lock.
func (s StatusService) removeExpiredStatusFile(path string, candidate os.FileInfo) {
	current, err := lstatIdentity(path)
	if err == nil && current.Mode().IsRegular() && os.SameFile(candidate, current) && s.now().Sub(current.ModTime()) >= statusRetention {
		_ = os.Remove(path)
	}
}
