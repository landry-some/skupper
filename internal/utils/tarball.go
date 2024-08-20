package utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

type Tarball struct {
	buf       *bytes.Buffer
	gz        *gzip.Writer
	tw        *tar.Writer
	lastAdded time.Time
	basePath  string
	mutex     *sync.Mutex
}

// NewTarball returns an initialized Tarball
func NewTarball() *Tarball {
	tb := &Tarball{}
	tb.buf = new(bytes.Buffer)
	tb.gz = gzip.NewWriter(tb.buf)
	tb.tw = tar.NewWriter(tb.gz)
	tb.mutex = &sync.Mutex{}
	return tb
}

// Save saves tarball based on added directories.
// The provided filename will be created or truncated
// if it already exists.
// Once Save method is executed, this instance cannot
// be used anymore.
func (t *Tarball) Save(filename string) error {
	var err error
	err = t.close()
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, t.buf.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}

// SaveData saves written files and returns the tarball data.
// Once SaveData method is executed, this instance cannot
// be used anymore.
func (t *Tarball) SaveData() ([]byte, error) {
	err := t.close()
	if err != nil {
		return nil, err
	}
	return t.buf.Bytes(), nil
}

func (t *Tarball) close() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	var err error
	err = t.tw.Flush()
	if err != nil {
		return err
	}
	err = t.tw.Close()
	if err != nil {
		return err
	}
	err = t.gz.Close()
	if err != nil {
		return err
	}
	return nil
}

// AddFiles adds all files (recursively) based on the
// provided directory.
func (t *Tarball) AddFiles(dir string, includesOnly ...string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.lastAdded = time.Now()
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}
	t.basePath = dir
	var err error
	if len(includesOnly) > 0 {
		for _, inc := range includesOnly {
			err = t.addFiles(path.Join(dir, inc))
		}
	} else {
		err = t.addFiles(dir)
	}
	t.lastAdded = time.Time{}
	t.basePath = ""
	return err
}

func (t *Tarball) addFiles(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	innerDir := dir[len(t.basePath):]
	for _, entry := range entries {
		fileStat, err := os.Stat(path.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
		if entry.IsDir() {
			t.tw.WriteHeader(&tar.Header{
				Name:     path.Join(innerDir, entry.Name()) + "/",
				Mode:     int64(fileStat.Mode()),
				Typeflag: tar.TypeDir,
				ModTime:  fileStat.ModTime(),
			})
			err = t.addFiles(path.Join(dir, entry.Name()))
			if err != nil {
				return err
			}
		} else {
			fileName := path.Join(innerDir, entry.Name())
			err = t.tw.WriteHeader(&tar.Header{
				Name:    fileName,
				Mode:    int64(fileStat.Mode()),
				Size:    fileStat.Size(),
				ModTime: fileStat.ModTime(),
			})
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path.Join(dir, entry.Name()))
			if err != nil {
				return err
			}
			_, err = t.tw.Write(data)
			if err != nil {
				return err
			}
			err = t.tw.Flush()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Tarball) AddFileData(fileName string, mode int64, data []byte) error {
	var err error
	err = t.tw.WriteHeader(&tar.Header{
		Name:    fileName,
		Mode:    mode,
		Size:    int64(len(data)),
		ModTime: time.Now(),
	})
	if err != nil {
		return err
	}
	_, err = t.tw.Write(data)
	if err != nil {
		return err
	}
	err = t.tw.Flush()
	if err != nil {
		return err
	}
	return nil
}

// Extract extracts the given tar.gz filename into the provided outputPath
func (t *Tarball) Extract(fileName, outputPath string) error {
	// Validating outputPath
	if outputPath == "" {
		return fmt.Errorf("outputPath is empty")
	}
	outputPathStat, err := os.Stat(outputPath)
	if err != nil {
		err = os.MkdirAll(outputPath, 0755)
		if err != nil {
			return err
		}
	} else if !outputPathStat.Mode().IsDir() {
		return fmt.Errorf("outputPath is not a directory")
	}

	tgzFile, err := os.Open(fileName)
	if err != nil {
		return err
	}
	gzipReader, err := gzip.NewReader(tgzFile)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}
		targetFilePath := path.Join(outputPath, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(targetFilePath); err != nil {
				if err := os.MkdirAll(targetFilePath, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			file, err := os.OpenFile(targetFilePath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(file, tarReader); err != nil {
				return err
			}
			err = file.Close()
			if err != nil {
				return err
			}
		}
	}
}
