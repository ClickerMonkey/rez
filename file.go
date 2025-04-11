package rez

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/ClickerMonkey/deps"
	"github.com/ClickerMonkey/rez/api"
)

var ErrTooManyFiles = errors.New("expected at most one file")
var ErrNotMultipartForm = errors.New("file parsing only possible with multi-part form data")
var ErrInvalidFileFormKey = errors.New("invalid file form key format")

type FileConstrainer interface {
	FileConstraints() FileConstraints
}

type FileConstraints struct {
	MinSize     int64
	MaxSize     int64
	ContentType api.ContentType
	MaxFiles    int
}

type AnyFile struct{}

var _ FileConstrainer = AnyFile{}

func (AnyFile) FileConstraints() FileConstraints {
	return FileConstraints{
		ContentType: api.ContentTypeStream,
	}
}

// File as body:
//
//	rez.File[rez.AnyFile]
//
// File in body
//
//	type FormData struct {
//	  File rez.File[rez.AnyFile] `json:"file"`
//	}
//
// Files in body
//
//	type FormData struct {
//	  File []rez.File[rez.AnyFile] `json:"files"`
//	}
//
// Files embedded deeply in body
//
//			type DeepFormData struct {
//			  FileWithInfo []struct{
//			    File    rez.File[rez.AnyFile] `json:"file"`
//		     Details string                `json:"details"`
//			  } `json:"files"`
//			}
//	 // files[0][file]
//	 // files[0][details]
type File[FD FileConstrainer] struct {
	multipart.FileHeader
	io.ReadCloser
	formKey   string
	fileIndex int
	parsed    bool
}

var _ api.HasFullSchema = File[AnyFile]{}
var _ CanValidateFull = File[AnyFile]{}
var _ Injectable = &File[AnyFile]{}
var _ json.Unmarshaler = &File[AnyFile]{}

func (File[FD]) Constraints() FileConstraints {
	var fd FD
	return fd.FileConstraints()
}

func (f File[FD]) APIFullSchema() *api.Schema {
	return &api.Schema{
		Type:     api.DataTypeString,
		Format:   "binary",
		FileType: f.Constraints().ContentType,
	}
}

func (f File[FD]) APIRequestTypes() RequestTypes {
	return RequestTypes{Body: deps.TypeOf[File[FD]]()}
}

func (f File[FD]) FullValidate(v *Validator) {
	if !f.IsParsed() {
		return
	}

	fc := f.Constraints()
	if fc.MaxSize > 0 && f.Size > fc.MaxSize {
		v.Add(Validation{Message: fmt.Sprintf("size exceeds max of %d", fc.MaxSize)})
	}
	if fc.MinSize > 0 && f.Size < fc.MinSize {
		v.Add(Validation{Message: fmt.Sprintf("size does not meet min of %d", fc.MinSize)})
	}
}

func (f File[FD]) APIValidate(op *api.Operation, v *Validator) {
	if !f.IsParsed() {
		return
	}

	fc := f.Constraints()
	if fc.MaxSize > 0 && f.Size > fc.MaxSize {
		v.Add(Validation{Message: fmt.Sprintf("size exceeds max of %d", fc.MaxSize)})
	}
	if fc.MinSize > 0 && f.Size < fc.MinSize {
		v.Add(Validation{Message: fmt.Sprintf("size does not meet min of %d", fc.MinSize)})
	}
}

// When this is injected, the file is the entire body
func (f *File[FD]) ProvideDynamic(scope *deps.Scope) error {
	var err error

	request, _ := deps.GetScoped[http.Request](scope)
	f.ReadCloser = request.Body

	if cd := request.Header.Get("Content-Disposition"); cd != "" {
		var params map[string]string
		_, params, err = mime.ParseMediaType(cd)
		f.Filename = params["filename"]
	}

	if cl := request.Header.Get("Content-Length"); cl != "" {
		f.Size, err = strconv.ParseInt(cl, 10, 60)
	}

	f.parsed = true

	return err
}

// When this is injected, the file is the entire body
func (f *File[FD]) UnmarshalJSON(b []byte) error {
	formKey, fileCount, err := getFormFileInfo(b)
	if err != nil {
		return err
	}
	if fileCount > 1 {
		return ErrTooManyFiles
	}

	f.formKey = formKey

	return nil
}

// When this is injected, the file is the entire body
func (f *File[FD]) Parse(r *http.Request) error {
	if f.parsed {
		return nil
	}

	var err error

	if f.formKey != "" {
		if r.MultipartForm == nil {
			return ErrNotMultipartForm
		}
		if r.MultipartForm.File != nil {
			if header := r.MultipartForm.File[f.formKey]; len(header) >= f.fileIndex {
				if fh := header[f.fileIndex]; fh != nil {
					f.FileHeader = *fh
					f.ReadCloser, err = fh.Open()
				}
			}
		}
	}

	f.parsed = true

	if f.ReadCloser == nil && err != nil {
		err = io.EOF
	}

	return err
}

// If the file is parsed (has a ReadCloser)
func (f *File[FD]) IsParsed() bool {
	return f.parsed
}

// If the file is parsed (has a ReadCloser)
func (f *File[FD]) HasData() bool {
	return f.ReadCloser != nil
}

// Returns the data in the file, assuming it's been parsed already.
func (f *File[FD]) Data() ([]byte, error) {
	if f.ReadCloser == nil {
		return nil, io.EOF
	}
	return io.ReadAll(f)
}

// Returns the data in the file, parsing it if necessary.
func (f *File[FD]) ParseData(r *http.Request) ([]byte, error) {
	err := f.Parse(r)
	if err != nil {
		return nil, err
	}
	return f.Data()
}

type Files[FD FileConstrainer] []File[FD]

var _ api.HasFullSchema = Files[AnyFile]{}
var _ CanValidateFull = Files[AnyFile]{}
var _ Injectable = &Files[AnyFile]{}
var _ json.Unmarshaler = &Files[AnyFile]{}

func (Files[FD]) Constraints() FileConstraints {
	var fd FD
	return fd.FileConstraints()
}

func (f Files[FD]) APIFullSchema() *api.Schema {
	return &api.Schema{
		Type:     api.DataTypeArray,
		MaxItems: f.Constraints().MaxFiles,
		Items: &api.Schema{
			Type:   api.DataTypeString,
			Format: "binary",
		},
		FileType: api.ContentTypeFormData,
	}
}

func (f Files[FD]) APIRequestTypes() RequestTypes {
	return RequestTypes{Body: deps.TypeOf[[]File[FD]]()}
}

func (f Files[FD]) FullValidate(v *Validator) {
	for i := range f {
		f[i].FullValidate(v.Next(fmt.Sprintf("%d", i)))
	}
}

func (f Files[FD]) APIValidate(op *api.Operation, v *Validator) {
	for i := range f {
		f[i].APIValidate(op, v.Next(fmt.Sprintf("%d", i)))
	}
}

// When this is injected, the file is the entire body
func (f *Files[FD]) UnmarshalJSON(b []byte) error {
	formKey, fileCount, err := getFormFileInfo(b)
	if err != nil {
		return err
	}

	for i := 0; i < fileCount; i++ {
		*f = append(*f, File[FD]{
			formKey:   formKey,
			fileIndex: i,
		})
	}

	return nil
}

// When this is injected, the file is the entire body
func (f Files[FD]) Parse(r *http.Request) error {
	for i := range f {
		err := f[i].Parse(r)
		if err != nil {
			return err
		}
	}

	return nil
}

// If the file is parsed (has a ReadCloser)
func (f Files[FD]) IsParsed() bool {
	return len(f) > 0 && f[0].IsParsed()
}

// When this is injected, the file is the entire body
func (f *Files[FD]) ProvideDynamic(scope *deps.Scope) error {
	var err error

	request, _ := deps.GetScoped[http.Request](scope)
	router, _ := deps.GetScoped[Router](scope)

	err = request.ParseMultipartForm((*router).GetMemoryLimit())
	if err != nil {
		return nil
	}

	for formKey, fileHeaders := range request.MultipartForm.File {
		for fileIndex := range fileHeaders {
			file := File[FD]{
				formKey:   formKey,
				fileIndex: fileIndex,
			}
			err = file.Parse(request)
			if err != nil {
				return err
			}
			*f = append(*f, file)
		}
	}

	return err
}

func getFormFileInfo(data []byte) (formKey string, fileCount int, err error) {
	// formKey::fileCount
	formKeyLength := ""
	err = json.Unmarshal(data, &formKeyLength)
	if err != nil {
		return
	}

	formKeyParts := strings.SplitN(formKeyLength, "::", 2)
	if len(formKeyParts) != 2 {
		err = ErrInvalidFileFormKey
		return
	}

	formKey = formKeyParts[0]
	fileCount, err = strconv.Atoi(formKeyParts[1])

	return
}
