package godrop

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
)

// WriteTarball traverses over dir and streams a tar archive
// into writer
func WriteTarball(writer io.Writer, dir string) error {
	tw := tar.NewWriter(writer)

	defer tw.Close()

	//walk path
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())

		if err != nil {
			return err
		}

		header.Name = path

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(path)

		if err != nil {
			return err
		}

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		f.Close()

		return nil
	})
}

// ReadTarball reads from reader and creates the resulting directory at target
func ReadTarball(reader io.Reader, target string) error {

	tr := tar.NewReader(reader)

	for {
		header, err := tr.Next()

		switch {
		case err == io.EOF:
			return nil

		case err != nil:
			return err

		case header == nil:
			continue

		}

		target := filepath.Join(target, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			//it is a directory. create it if it does not exist
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		case tar.TypeReg:
			//regular file. create it
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()
		}
	}
}
