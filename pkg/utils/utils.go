package utils

import(
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"archive/tar"
	"compress/gzip"
	"net/http"
)

func ParseBody(r *http.Request, x interface{}) error {
	reqBody, _ := io.ReadAll(r.Body)
	r.Body.Close()

	err := json.Unmarshal(reqBody, x)
	if err != nil {
		log.Printf("❌ Error decoding body: %v", err.Error())
		return fmt.Errorf("Error decoding JSON")
	}

	return nil
}


func CreateArchive(name string, directory string, files []string) error {
	// Create output file
	out, err := os.Create(name)
	if err != nil {
		log.Printf("❌ Error writing archive: ", err)
	}
	defer out.Close()
	
	// Create new Writers for gzip and tar
	// These writers are chained. Writing to the tar writer will
	// write to the gzip writer which in turn will write to
	// the "buf" writer
	gw := gzip.NewWriter(out)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Iterate over files and add them to the tar archive
	for _, file := range files {
		err := AddToArchive(tw, directory + file)
		if err != nil {
			return err
		}
	}

	return nil
}

func AddToArchive(tw *tar.Writer, filename string) error {
	// Open the file which will be written into the archive
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get FileInfo about our file providing file size, mode, etc.
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create a tar Header from the FileInfo data
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	// Use full path as name (FileInfoHeader only takes the basename)
	// If we don't do this the directory strucuture would
	// not be preserved
	// https://golang.org/src/archive/tar/common.go?#L626
	header.Name = filename

	// Write file header to the tar archive
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to tar archive
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}
