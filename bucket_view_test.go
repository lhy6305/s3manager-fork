package s3manager_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	. "github.com/mastertinner/s3manager"
	minio "github.com/minio/minio-go"
	"github.com/stretchr/testify/assert"
)

func TestBucketViewHandler(t *testing.T) {
	cases := map[string]struct {
		s3                   S3
		bucketName           string
		expectedStatusCode   int
		expectedBodyContains string
	}{
		"renders a bucket containing a file": {
			s3: &s3Mock{
				Buckets: []minio.BucketInfo{
					{Name: "testBucket"},
				},
				Objects: []minio.ObjectInfo{
					{Key: "testFile"},
				},
			},
			bucketName:           "testBucket",
			expectedStatusCode:   http.StatusOK,
			expectedBodyContains: "testBucket",
		},
		"renders placeholder for an empty bucket": {
			s3: &s3Mock{
				Buckets: []minio.BucketInfo{
					{Name: "testBucket"},
				},
			},
			bucketName:           "testBucket",
			expectedStatusCode:   http.StatusOK,
			expectedBodyContains: "No objects in",
		},
		"renders a bucket containing an archive": {
			s3: &s3Mock{
				Buckets: []minio.BucketInfo{
					{Name: "testBucket"},
				},
				Objects: []minio.ObjectInfo{
					{Key: "archive.tar.gz"},
				},
			},
			bucketName:           "testBucket",
			expectedStatusCode:   http.StatusOK,
			expectedBodyContains: "archive",
		},
		"renders a bucket containing an image": {
			s3: &s3Mock{
				Buckets: []minio.BucketInfo{
					{Name: "testBucket"},
				},
				Objects: []minio.ObjectInfo{
					{Key: "testImage.png"},
				},
			},
			bucketName:           "testBucket",
			expectedStatusCode:   http.StatusOK,
			expectedBodyContains: "photo",
		},
		"renders a bucket containing a sound file": {
			s3: &s3Mock{
				Buckets: []minio.BucketInfo{
					{Name: "testBucket"},
				},
				Objects: []minio.ObjectInfo{
					{Key: "testSound.mp3"},
				},
			},
			bucketName:           "testBucket",
			expectedStatusCode:   http.StatusOK,
			expectedBodyContains: "music_note",
		},
		"returns error if the bucket doesn't exist": {
			s3:                   &s3Mock{},
			bucketName:           "testBucket",
			expectedStatusCode:   http.StatusNotFound,
			expectedBodyContains: http.StatusText(http.StatusNotFound),
		},
		"returns error if there is an S3 error": {
			s3: &s3Mock{
				Err: errors.New("mocked S3 error"),
			},
			bucketName:           "testBucket",
			expectedStatusCode:   http.StatusInternalServerError,
			expectedBodyContains: http.StatusText(http.StatusInternalServerError),
		},
	}

	for tcID, tc := range cases {
		t.Run(tcID, func(t *testing.T) {
			assert := assert.New(t)

			r := mux.NewRouter()
			r.
				Methods(http.MethodGet).
				Path("/buckets/{bucketName}").
				Handler(BucketViewHandler(tc.s3))

			ts := httptest.NewServer(r)
			defer ts.Close()

			url := fmt.Sprintf("%s/buckets/%s", ts.URL, tc.bucketName)
			resp, err := http.Get(url)
			assert.NoError(err, tcID)
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			assert.NoError(err, tcID)

			assert.Equal(tc.expectedStatusCode, resp.StatusCode, tcID)
			assert.Contains(string(body), tc.expectedBodyContains, tcID)
		})
	}
}
