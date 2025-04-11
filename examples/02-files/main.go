package main

import (
	"fmt"
	"net/http"

	"github.com/ClickerMonkey/deps"
	"github.com/ClickerMonkey/rez"
	"github.com/ClickerMonkey/rez/api"
	"github.com/go-chi/chi/v5"
)

var _ rez.FileConstrainer = PNG{}

type PNG struct{}

func (PNG) FileConstraints() rez.FileConstraints {
	return rez.FileConstraints{
		ContentType: api.ContentTypePNG,
	}
}

type UploadedFile = rez.File[rez.AnyFile]
type UploadedFiles = rez.Files[rez.AnyFile]

type FileResponse[FD rez.FileConstrainer] struct {
	Filename string `json:"filename"`
	Size     int    `json:"size"`
	DataSize int    `json:"dataSize"`
	Data     string `json:"data"`
}

func NewFileResponse[FD rez.FileConstrainer](file rez.File[FD], req *http.Request) (FileResponse[FD], error) {
	s, err := file.ParseData(req)
	if err != nil {
		return FileResponse[FD]{}, err
	}
	return FileResponse[FD]{
		Filename: file.Filename,
		Size:     int(file.Size),
		DataSize: len(s),
		Data:     string(s),
	}, nil
}

type EmbeddedFile struct {
	EntityID string       `json:"entityID"`
	File     UploadedFile `json:"file"`
}

type FilesRequest struct {
	Files UploadedFiles `json:"files"`
}

func main() {
	site := rez.New(chi.NewRouter())

	site.Post("/file", func(file UploadedFile) (FileResponse[rez.AnyFile], error) {
		return NewFileResponse(file, nil)
	})

	site.Post("/png-file", func(file rez.File[PNG]) (FileResponse[PNG], error) {
		return NewFileResponse(file, nil)
	})

	site.DefineBody(FilesRequest{})
	site.Post("/files", func(files FilesRequest, req *http.Request) ([]FileResponse[rez.AnyFile], error) {
		err := files.Files.Parse(req)
		if err != nil {
			return nil, err
		}
		out := make([]FileResponse[rez.AnyFile], 0, len(files.Files))
		for i := range files.Files {
			file, err := NewFileResponse(files.Files[i], req)
			if err != nil {
				return nil, err
			}
			out = append(out, file)
		}
		return out, nil
	})

	site.DefineBody(EmbeddedFile{})
	site.Post("/embedded-file", func(ef EmbeddedFile, req *http.Request) (FileResponse[rez.AnyFile], error) {
		fmt.Printf("embedded-file:entityID: %s\n", ef.EntityID)

		return NewFileResponse(ef.File, req)
	})

	site.SetErrorHandler(func(err error, response http.ResponseWriter, request *http.Request, scope *deps.Scope) (bool, error) {
		fmt.Printf("ERROR: %+v\n", err)

		return false, nil
	})

	site.SetInternalErrorHandler(func(err error) {
		fmt.Printf("INTERNAL ERROR: %+v\n", err)
	})

	site.PrintPaths()
	site.ServeSwaggerUI("/doc/swagger", nil)
	site.ServeRedoc("/doc/redoc", nil)
	site.Listen(":3000")
}
