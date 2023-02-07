package diskache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/GitbookIO/syncgroup"
	"io"
	"log"
	"os"
	"path"
	"time"
)

const expiredTimestamp = 0

const expiredTableName = "expired-table"

type Diskache struct {
	directory             string
	expiredTableDirectory string
	items                 int
	lock                  *syncgroup.MutexGroup
	expiredTime           int64
}

type Opts struct {
	Directory string // dir name
}

type Stats struct {
	Directory string
	Items     int
}

func New(opts *Opts) (*Diskache, error) {
	// Create Diskache directory
	if err := os.MkdirAll(opts.Directory, os.ModePerm); err != nil {
		return nil, err
	}
	// Create Diskache instance
	dc := &Diskache{
		directory:             opts.Directory,
		expiredTableDirectory: path.Join(opts.Directory, "expired-table.json"),
		lock:                  syncgroup.NewMutexGroup(),
	}
	return dc, nil
}

func (dc *Diskache) SetJson(key string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return dc.Set(key, b)
}

func (dc *Diskache) SetStr(key string, value string) error {
	if len(value) == 0 {
		return nil
	}
	return dc.Set(key, []byte(value))
}

func (dc *Diskache) Set(key string, data []byte) error {
	// Get encoded key
	filename := dc.buildFilename(key)

	// Lock for writing
	dc.lock.Lock(filename)
	defer dc.lock.Unlock(filename)

	// Open file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write data
	if _, err = file.Write(data); err == nil {
		// Increment items
		dc.items += 1
	}
	return err
}

func (dc *Diskache) SetExpired(key string, data []byte, expired int64) error {
	err := dc.Set(key, data)
	if err != nil {
		return err
	}
	err = dc.setExpiredTime(key, getTimestamp()+expired)
	if err != nil {
		log.Printf("set expired time error, key = %s, err = %s", key, err)
		return err
	}
	return nil
}

func (dc *Diskache) Get(key string) ([]byte, bool) {
	b, exists := dc.getKey(key)
	if !exists {
		return b, false
	}
	// lazy check expired time
	if expired, _ := dc.IsExpired(key); expired {
		_ = dc.setExpiredTime(key, expiredTimestamp)
		return nil, false
	} else {
		return b, true
	}
}

func (dc *Diskache) GetStr(key string) (string, bool) {
	if b, exists := dc.Get(key); exists {
		return string(b), true
	}
	return "", false
}

func (dc *Diskache) GetJson(key string) (string, bool) {
	if b, exists := dc.Get(key); exists {
		return string(b), true
	}
	return "", false
}

func (dc *Diskache) Delete(key string) bool {
	// Get encoded key
	filename := dc.buildFilename(key)
	// Lock for deleting
	dc.lock.RLock(filename)
	defer dc.lock.RUnlock(filename)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return true
	}
	err := os.Remove(filename)
	if err != nil {
		log.Println("[error] fail to remove cache file:", err.Error())
		return false
	}
	return true
}

func (dc *Diskache) Clean() error {
	// Delete directory
	if err := os.RemoveAll(dc.directory); err != nil {
		return err
	}
	// Recreate directory
	return os.MkdirAll(dc.directory, os.ModePerm)
}

func (dc *Diskache) Stats() Stats {
	return Stats{
		Directory: dc.directory,
		Items:     dc.items,
	}
}

func (dc *Diskache) IsExpired(key string) (bool, error) {
	expired, err := dc.getExpiredTime(key)
	if err != nil {
		return true, err
	}
	if expired == expiredTimestamp {
		return false, nil
	}
	return getTimestamp() >= expired, nil
}

func (dc *Diskache) getKey(key string) ([]byte, bool) {
	// Get encoded key
	filename := dc.buildFilename(key)

	// Lock for reading
	dc.lock.RLock(filename)
	defer dc.lock.RUnlock(filename)

	// Open file
	file, err := os.Open(filename)
	if err != nil {
		return nil, false
	}
	defer file.Close()

	// Read file
	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Diskache: Error reading from file %s\n", key)
		return nil, false
	}
	return data, true
}

func (dc *Diskache) getExpiredTime(k string) (int64, error) {
	var table map[string]int64
	b, exists := dc.getKey(expiredTableName)
	if !exists {
		return 0, nil
	}
	err := json.Unmarshal(b, &table)
	if err != nil {
		return -1, nil
	}
	if t, ok := table[k]; !ok {
		return 0, nil
	} else {
		return t, nil
	}
}

func (dc *Diskache) setExpiredTime(k string, t int64) error {
	var table map[string]int64
	b, exists := dc.getKey(expiredTableName)
	if !exists {
		table = make(map[string]int64)
	} else {
		err := json.Unmarshal(b, &table)
		if err != nil {
			return err
		}
	}
	table[k] = t
	// write table
	return dc.SetJson(expiredTableName, table)
}

func (dc *Diskache) buildFilename(key string) string {
	hasher := sha256.New()
	hasher.Write([]byte(key))
	return path.Join(dc.directory, hex.EncodeToString(hasher.Sum(nil)))
}

func getTimestamp() int64 {
	return time.Now().UnixMilli()
}
