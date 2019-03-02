package main

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fname := os.Args[1:] // gimmie myfile.tar.gz.bz2
	identifyFileType(fname[0])
}

/*
   Get the filetype of a given file.

   Check the 1st two bytes of the header for the gz/bz2/tar magic bytes.
   If we find one, call the appropriate decompression function, then
   call again with the new filename. Unless it's tar, then just drop
*/
func identifyFileType(fname string) error {
	// magic bytes for each format
	gz := [2]byte{31, 139}
	bz2 := [2]byte{66, 90}
	tar := [2]byte{116, 101}

	fReader, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer fReader.Close()

	// read the header for the magic bytes
	var header [2]byte
	_, err = io.ReadFull(fReader, header[:])
	if err != nil {
		return err
	}

	// go through the magic bytes and see if we have a match.
	// call the corresponding decompression function and get the new
	// filename. Then recall this function to do it over again.
	// If we get a tar, we know that we will be done (assuming no error)
	// so don't recursive again
	switch header {
	case bz2:
		fmt.Println("[+] Found bz2 compression")
		fname, err := unbz(fname)
		if err != nil {
			return err
		}
		identifyFileType(fname)
	case gz:
		fmt.Println("[+] Found gz compression")
		fname, err := ungzip(fname)
		if err != nil {
			return err
		}
		identifyFileType(fname)
	case tar:
		fmt.Println("[+] Found a tar archive")
		err := untar(fname)
		if err != nil {
			return err
		}
	default:
		fmt.Println("[+] Your file is in: ", fname)
		break
	}
	return nil
}

/*
   decompress a bz2 flie and return with the new filename, or an error
*/
func unbz(source string) (string, error) {
	file, err := os.Open(source)
	if err != nil {
		return "", err
	}

	// if we have a file extension get rid of it for the next decompression
	if strings.HasSuffix(source, ".bz2") {
		source = source[0 : len(source)-len(filepath.Ext(source))]
	}

	dst := bzip2.NewReader(file)
	writer, err := os.Create(source)
	if err != nil {
		return "", err
	}
	defer writer.Close()
	// write our new file out
	_, err = io.Copy(writer, dst)

	return source, nil
}

/*
   decompress a gz file and return the new filename, or error
*/
func ungzip(source string) (string, error) {
	reader, err := os.Open(source)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return "", err
	}
	defer archive.Close()

	if strings.HasSuffix(source, ".gz") {
		source = source[0 : len(source)-len(filepath.Ext(source))]
	}

	target := filepath.Join(source, archive.Name)
	writer, err := os.Create(target)
	if err != nil {
		return "", err
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	return target, nil
}

/*
   decompresses a .tar file.
   loop through the archive and pull out files. If it's a directory, create
   that. If it's a file, copy the contents out
   CREDIT: https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
*/
func untar(dst string) error {
	reader, _ := os.Open(dst)
	tr := tar.NewReader(reader)

	if strings.HasSuffix(dst, ".tar") {
		dst = dst[0 : len(dst)-len(filepath.Ext(dst))]
	}
	// create our destination, decompressed directory if it doesn't exist
	if _, err := os.Stat(dst); err != nil {
		if err := os.MkdirAll(dst, 0755); err != nil {
			return err
		}
	}

	for {
		header, err := tr.Next()
		switch {
		// if no more files are found return
		case err == io.EOF:
			return err

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}
