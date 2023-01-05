package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const tmpl string = `
<!DOCTYPE html>
  <html lang="en">
  <style>
  table, th, td {
    border:1px solid black;
    border-collapse: collapse;
  }
  </style>
  <head>
    <meta charset="utf-8">
    <title>lsblk</title>
  </head>
  <body>
    <!-- page content -->
    <table>
    <tr>
          <th>Raw</th>
          <th>Gzip</th>
          <th>Path</th>
          <th>Majmin</th>
          <th>Size</th>
          <th>Type</th>
          <th>Tran</th>
          <th>Model</th>
          <th>Serial</th>
          <th>Pttype</th>
          <th>Ptuuid</th>
          <th>Fstype</th>
          <th>Fsver</th>
          <th>Label</th>
          <th>Uuid</th>
          <th>Fsavail</th>
          <th>Fsuse</th>
          <th>Mountpoints</th>
    </tr>
    {{range .}}
      <tr>
            <td><a href="/download?file={{.Path}}">Download</a></td>
            <td><a href="/download?file={{.Path}}&compress=gzip">Download</a></td>
            <td>{{.Path}}</td>
            <td>{{.Majmin}}</td>
            <td>{{.Size}}</td>
            <td>{{.Type}}</td>
            <td>{{.Tran}}</td>
            <td>{{.Model}}</td>
            <td>{{.Serial}}</td>
            <td>{{.Pttype}}</td>
            <td>{{.Ptuuid}}</td>
            <td>{{.Fstype}}</td>
            <td>{{.Fsver}}</td>
            <td>{{.Label}}</td>
            <td>{{.Uuid}}</td>
            <td>{{.Fsavail}}</td>
            <td>{{.Fsuse}}</td>
            <td>{{.Mountpoints}}</td>
      </tr>
    {{end}}
    </table>
  </body>
</html>
`

// root is the Json root struct
type root struct {
	Blockdevices []blockdevice
}

// blockdevice is a recursive struct
type blockdevice struct {
	Children    []blockdevice
	Fsavail     string
	Fstype      string
	Fsuse       string `json:"fsuse%"`
	Fsver       string
	Label       string
	Majmin      string `json:"maj:min"`
	Model       string
	Mountpoints []string
	Path        string
	Pttype      string
	Ptuuid      string
	Serial      string
	Size        string
	Tran        string
	Type        string
	Uuid        string
}

// flatten is used to get a flat list from the tree of blockdevices
func flatten(a *[]blockdevice, b *[]blockdevice) {
	for _, new_a := range *a {
		*b = append(*b, new_a)
		flatten(&new_a.Children, b)
	}
}

// table presents a html table of lsblk's output
func table(w http.ResponseWriter, req *http.Request) {
	out, err := exec.Command("lsblk", "-O", "-J").Output()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "lsblk failed: %v", err)
		return
	}
	var blks root
	err = json.Unmarshal(out, &blks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, string(out))
		fmt.Fprintf(w, "Bad json: %v", err)
		return
	}
	var flat []blockdevice
	flatten(&blks.Blockdevices, &flat)
	tmpl, err := template.New("lsblk").Parse(tmpl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Template parse failed: %v", err)
		return
	}
	err = tmpl.Execute(w, flat)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Template execute failed: %v", err)
		return
	}
}

// download sends block device as a file
func download(w http.ResponseWriter, req *http.Request) {

	// Get path and filename
	values := req.URL.Query()
	path := values.Get("file")
	if path == "" {
		fmt.Fprintf(w, "Expected file parameter.")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	filename := strings.Trim(path, "/")
	filename = strings.ReplaceAll(filename, "/", "-")

	// Inject compression
	var dest io.Writer
	var ext string
	if values.Get("compress") == "" {
		length, err := exec.Command("blockdev", "--getsize64", path).Output()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Failed to check size of %s: %v", path, err)
			return
		}
		w.Header().Set("Content-Length", strings.TrimSpace(string(length)))
		dest = w
	} else {
		gzipW := gzip.NewWriter(w)
		gzipW.Name = filename
		defer gzipW.Close()
		dest = gzipW
		ext = ".gz"
	}

	// Set remaining headers
	cd := mime.FormatMediaType("attachment", map[string]string{"filename": filename + ext})
	w.Header().Set("Content-Disposition", cd)
	w.Header().Set("Content-Type", "application/octet-stream")

	// Open block device
	f, err := os.Open(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to open %s: %v", path, err)
		return
	}
	defer f.Close()

	// Write data to response
	_, err = io.Copy(dest, f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to copy data to response: %v", err)
		return
	}
}

func main() {
	http.HandleFunc("/", table)
	http.HandleFunc("/download", download)
	addr := ":8090"
	fmt.Printf("Listening on %s\n", addr)
	http.ListenAndServe(addr, nil)
}
