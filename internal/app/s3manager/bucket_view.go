package s3manager

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/facette/natsort"
	"github.com/minio/minio-go/v7"
)

// HandleBucketView shows the details page of a bucket.
func HandleBucketView(s3 S3, templates fs.FS, allowDelete bool, allowDeleteBucket bool, listRecursive bool) http.HandlerFunc {
	type objectWithIcon struct {
		Key          string
		Size         int64
		LastModified time.Time
		Owner        string
		Icon         string
		IsFolder     bool
		DisplayName  string
	}

	type pageData struct {
		BucketName        string
		Objects           []objectWithIcon
		AllowDelete       bool
		AllowDeleteBucket bool
		Paths             []string
		CurrentPath       string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		regex := regexp.MustCompile(`\/buckets\/([^\/]*)\/?(.*)`)
		matches := regex.FindStringSubmatch(r.RequestURI)
		bucketName := matches[1]
		path := matches[2]

		var objs []objectWithIcon
		doneCh := make(chan struct{})
		defer close(doneCh)
		opts := minio.ListObjectsOptions{
			Recursive: listRecursive,
			Prefix:    path,
		}
		objectCh := s3.ListObjects(r.Context(), bucketName, opts)
		for object := range objectCh {
			if object.Err != nil {
				handleHTTPError(w, fmt.Errorf("error listing objects: %w", object.Err))
				return
			}

			obj := objectWithIcon{
				Key:          object.Key,
				Size:         object.Size,
				LastModified: object.LastModified,
				Owner:        object.Owner.DisplayName,
				Icon:         icon(object.Key),
				IsFolder:     strings.HasSuffix(object.Key, "/"),
				DisplayName:  strings.TrimSuffix(strings.TrimPrefix(object.Key, path), "/"),
			}
			objs = append(objs, obj)
		}
		sort.Slice(objs, func(i, j int) bool {
			if objs[i].IsFolder != objs[j].IsFolder {
				return objs[i].IsFolder
			}
			return natsort.Compare(objs[i].Key, objs[j].Key)
		})
		data := pageData{
			BucketName:        bucketName,
			Objects:           objs,
			AllowDelete:       allowDelete,
			AllowDeleteBucket: allowDeleteBucket,
			Paths:             removeEmptyStrings(strings.Split(path, "/")),
			CurrentPath:       path,
		}

		t, err := template.ParseFS(templates, "layout.html.tmpl", "bucket.html.tmpl")
		if err != nil {
			handleHTTPError(w, fmt.Errorf("error parsing template files: %w", err))
			return
		}
		err = t.ExecuteTemplate(w, "layout", data)
		if err != nil {
			handleHTTPError(w, fmt.Errorf("error executing template: %w", err))
			return
		}
	}
}

// icon returns an icon for a file type.
func icon(fileName string) string {
	if strings.HasSuffix(fileName, "/") {
		return "folder"
	}

	e := path.Ext(fileName)
	switch e {
	case ".aac", ".aiff", ".au", ".flac", ".m4a", ".mka", ".mid", ".mp3", ".mpa", ".ogg", ".opus", ".ra", ".wav", ".wma":
		return "audio_track"
	case ".apk":
		return "android"
	case ".3g2", ".3gp", ".avi", ".flv", ".h264", ".m4v", ".mkv", ".mov", ".mp4", ".mpeg", ".mpg", ".rm", ".swf", ".ts", ".vob", ".wmv":
		return "ondemand_video"
	case ".c", ".cc", ".cpp", ".cs", ".h", ".htm", ".html", ".java", ".js", ".php", ".pl", ".py", ".sh", ".swift", ".vb", ".xml", ".css", ".bat", ".json", ".hpp":
		return "code"
	case ".dat", ".doc", ".docx", ".ods", ".odp", ".odt", ".pdf", ".ppt", ".pptx", ".rtf", ".tex", ".txt", ".wpd", ".xls", ".xlsx", ".csv":
		return "article"
	case ".dmg", ".toast", ".vcd":
		return "album"
	case ".db", ".dbf", ".mdb", ".pdb", ".sql":
		return "storage"
	case ".fon", ".fnt", ".otf", ".ttf", ".woff", ".woff2":
		return "font_download"
	case ".ai", ".bmp", ".eps", ".gif", ".heif", ".ico", ".indd", ".jpeg", ".jpg", ".png", ".raw", ".svg", ".tiff", ".webp":
		return "photo"
	case ".7z", ".arj", ".bz", ".gz", ".iso", ".jar", ".pkg", ".rar", ".rpm", ".tar", ".tgz", ".zip", ".deb", ".xz":
		return "folder_zip"
	}

	return "insert_drive_file"
}

func removeEmptyStrings(input []string) []string {
	var result []string
	for _, str := range input {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}
