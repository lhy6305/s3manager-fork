package s3manager

import (
	"fmt"
	//"log"
	"io"
	//"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/encrypt"
)

// HandleCreateObject uploads a new object.
func HandleCreateObject(s3 S3, sseInfo SSEType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bucketName := mux.Vars(r)["bucketName"]

		/*
			//err := r.ParseMultipartForm(500 * (1 << 20)) // 500M Bytes
			err := r.ParseMultipartForm(0) // Write all files(>10M) to disk
			if err != nil {
				handleHTTPError(w, fmt.Errorf("error parsing multipart form: %w", err))
				return
			}
			file, _, err := r.FormFile("file")
			path := r.FormValue("path")
			if err != nil {
				handleHTTPError(w, fmt.Errorf("error getting file from form: %w", err))
				return
			}
			defer func(file multipart.File) {
				if err = file.Close(); err != nil {
					log.Fatal(fmt.Errorf("file cannot be closed: %w", err))
				}
			}(file)
		*/

		mr, err := r.MultipartReader()
		if err != nil {
			handleHTTPError(w, fmt.Errorf("failed to get a multipart reader: %w", err))
			return
		}

		part_index := 0
		var file_size int64 = -1
		var file_name string = ""

		for {
			part, err := mr.NextPart()
			if err != nil {
				if err == io.EOF { //end of multipart stream
					break
				}
				handleHTTPError(w, fmt.Errorf("error getting part #"+strconv.Itoa(part_index)+": %w", err))
				return
			}

			if part.FormName() == "path" {
				fnb, err := io.ReadAll(part)
				if err != nil {
					handleHTTPError(w, fmt.Errorf("error getting part #"+strconv.Itoa(part_index)+" as file_path: %w", err))
					return
				}
				file_name = string(fnb)
				continue
			}

			if part.FormName() == "size" {
				fsb, err := io.ReadAll(part)
				file_size, err = strconv.ParseInt(string(fsb), 10, 64)
				if err != nil {
					handleHTTPError(w, fmt.Errorf("error getting part #"+strconv.Itoa(part_index)+" as file_size: %w", err))
					return
				}
				continue
			}

			if part.FormName() == "file" {
				if part == nil {
					handleHTTPError(w, fmt.Errorf("error parsing form data: file entry not found"))
					return
				}

				//if file_size < 0 {
				//	handleHTTPError(w, fmt.Errorf("error parsing form data: file size not set or is negative"))
				//	return
				//}

				if len(file_name) <= 0 {
					handleHTTPError(w, fmt.Errorf("error parsing form data: file name not set or empty"))
					return
				}

				//opts := minio.PutObjectOptions{ContentType: "application/octet-stream"}
				opts := minio.PutObjectOptions{ContentType: "b2/x-auto"}

				if sseInfo.Type == "KMS" {
					opts.ServerSideEncryption, _ = encrypt.NewSSEKMS(sseInfo.Key, nil)
				}

				if sseInfo.Type == "SSE" {
					opts.ServerSideEncryption = encrypt.NewSSE()
				}

				if sseInfo.Type == "SSE-C" {
					opts.ServerSideEncryption, err = encrypt.NewSSEC([]byte(sseInfo.Key))
					if err != nil {
						handleHTTPError(w, fmt.Errorf("error setting SSE-C key: %w", err))
						return
					}
				}

				_, err = s3.PutObject(r.Context(), bucketName, file_name, part, file_size, opts)
				if err != nil {
					handleHTTPError(w, fmt.Errorf("error putting object: %w", err))
					return
				}
				continue
			}

			part_index++
		}

		w.WriteHeader(http.StatusCreated)
	}
}
