package main

import (
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/leopoldxx/go-utils/middleware"
	"github.com/leopoldxx/go-utils/server"
	"github.com/leopoldxx/go-utils/trace"
)

func main() {
	s := server.New(server.ListenAddr(":8001"), server.APIPrefix("/example"))
	s.Register(new(filesvr))
	s.ListenAndServe()
}

type filesvr struct{}

func (f *filesvr) Register(router *mux.Router) {
	subrouter := router.Path("/files").Subrouter()
	subrouter.Methods("POST").HandlerFunc(
		middleware.RecoverWithTrace("fileupload").
			HandlerFunc(f.upload))
}

func (f *filesvr) upload(w http.ResponseWriter, r *http.Request) {
	tracer := trace.GetTraceFromRequest(r)
	tracer.Info("file uploading...")

	r.ParseMultipartForm(1 << 30)
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		tracer.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}
	defer file.Close()

	newfile, err := os.OpenFile("./files/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		tracer.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	defer newfile.Close()

	sz, err := io.Copy(newfile, file)
	if err != nil {
		tracer.Error(err)
	}
	tracer.Infof("file size is : %d bytes", sz)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

// curl -v -XPOST http://127.0.0.1:8001/example/files -H "Content-Type: multipart/form-data" -F "uploadfile=@main.go"
