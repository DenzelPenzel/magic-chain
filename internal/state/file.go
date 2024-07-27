package state

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/denzelpenzel/magic-chain/internal/core"
)

type OutFiles struct {
	FTxs       *os.File
	FSourcelog *os.File
}

const (
	notFoundError = "could not find state store value for key %s"
)

type FileStore struct {
	uid       core.UUID
	dirname   string
	filesLock *sync.RWMutex
	files     map[int64]*OutFiles

	knownTxs     map[string]time.Time
	knownTxsLock sync.RWMutex
}

func NewFileStore(dirname string) *FileStore {
	return &FileStore{
		dirname:   dirname,
		filesLock: &sync.RWMutex{},
		files:     make(map[int64]*OutFiles),
		knownTxs:  make(map[string]time.Time),
	}
}

func (f *FileStore) GetCSVFile(timestamp int64) (*OutFiles, error) {
	sec := int64(core.BucketMinutes * 60)
	bucketTS := timestamp / sec * sec

	t := time.Unix(bucketTS, 0).UTC()

	f.filesLock.RLock()
	files, ok := f.files[bucketTS]
	f.filesLock.RUnlock()

	if ok {
		return files, nil
	}

	dir := filepath.Join(f.dirname, t.Format(time.DateOnly), "transactions")
	err := os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return nil, err
	}

	p := filepath.Join(dir, f.getFilename("txs", bucketTS))
	ftxs, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}

	dir = filepath.Join(f.dirname, t.Format(time.DateOnly), "sourcelog")
	err = os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return nil, err
	}

	p = filepath.Join(dir, f.getFilename("src", bucketTS))
	fsourcelog, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}

	outFiles := &OutFiles{
		FTxs:       ftxs,
		FSourcelog: fsourcelog,
	}

	f.filesLock.Lock()
	f.files[bucketTS] = outFiles
	f.filesLock.Unlock()

	return outFiles, err
}

func (f *FileStore) GetTx(key string) (time.Time, error) {
	defer f.knownTxsLock.RUnlock()

	f.knownTxsLock.RLock()

	val, exists := f.knownTxs[key]
	if !exists {
		return time.Time{}, fmt.Errorf(notFoundError, key)
	}

	return val, nil
}

func (f *FileStore) SetTx(key string, value time.Time) (time.Time, error) {
	f.knownTxsLock.Lock()
	defer f.knownTxsLock.Unlock()
	f.knownTxs[key] = value
	return value, nil
}

func (f *FileStore) getFilename(prefix string, timestamp int64) string {
	t := time.Unix(timestamp, 0).UTC()
	if prefix != "" {
		prefix += "_"
	}
	return fmt.Sprintf("%s%s_%s.csv", prefix, t.Format("2020-01-01_12-01"), f.uid)
}

func (f *FileStore) Cleaner() {
	for {
		time.Sleep(time.Minute)

		f.knownTxsLock.Lock()

		for k, v := range f.knownTxs {
			if time.Since(v) > core.TXCacheTime {
				delete(f.knownTxs, k)
			}
		}

		f.knownTxsLock.Unlock()

		f.filesLock.Lock()
		for ts, files := range f.files {
			usageSec := core.BucketMinutes * 60 * 2
			if time.Now().UTC().Unix()-ts > int64(usageSec) {
				delete(f.files, ts)
				_ = files.FTxs.Close()
				_ = files.FSourcelog.Close()
			}
		}
		f.filesLock.Unlock()

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
	}
}

type Option = func(*FileStore)

func WithID(uid core.UUID) Option {
	return func(f *FileStore) {
		f.uid = uid
	}
}

func WithDirName(dirname string) Option {
	return func(f *FileStore) {
		f.dirname = dirname
	}
}
