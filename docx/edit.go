package docx

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"sort"
)

// EditableDocument 表示一个从已有 docx 文件打开、可局部修改后再写出的文档。
type EditableDocument struct {
	files map[string][]byte
	order []string
}

// Open 打开一个已有 docx 文件。它保留原包中的其它部件，只修改 word/document.xml。
func Open(path string) (*EditableDocument, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	files := make(map[string][]byte, len(zr.File))
	order := make([]string, 0, len(zr.File))
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		data, readErr := io.ReadAll(rc)
		closeErr := rc.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		files[f.Name] = data
		order = append(order, f.Name)
	}
	if _, ok := files["word/document.xml"]; !ok {
		return nil, errors.New("docx: missing word/document.xml")
	}
	return &EditableDocument{files: files, order: order}, nil
}

// Write 写出修改后的 docx。
func (d *EditableDocument) Write(w io.Writer) error {
	if d == nil {
		return errors.New("docx: nil editable document")
	}
	if w == nil {
		return errors.New("docx: nil writer")
	}
	zw := zip.NewWriter(w)
	written := make(map[string]bool, len(d.files))
	for _, name := range d.order {
		data, ok := d.files[name]
		if !ok || written[name] {
			continue
		}
		if err := writeZipFile(zw, name, data); err != nil {
			return err
		}
		written[name] = true
	}

	extras := make([]string, 0)
	for name := range d.files {
		if !written[name] {
			extras = append(extras, name)
		}
	}
	sort.Strings(extras)
	for _, name := range extras {
		if err := writeZipFile(zw, name, d.files[name]); err != nil {
			return err
		}
	}
	return zw.Close()
}

// Save 把修改后的已有文档写出到本地文件。
func (d *EditableDocument) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return d.Write(f)
}

// writeZipFile 把一个命名条目原样写入 zip writer。
func writeZipFile(zw *zip.Writer, name string, data []byte) error {
	f, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}
