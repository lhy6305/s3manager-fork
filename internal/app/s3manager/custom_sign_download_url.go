package s3manager

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type Hash [32]byte

func hash_func(algo, str string) string {
	switch algo {
	case "SHA-512":
		sum := sha512.Sum512([]byte(str))
		return hex.EncodeToString(sum[:])
	case "SHA-384":
		sum := sha512.Sum384([]byte(str))
		return hex.EncodeToString(sum[:])
	case "SHA-256":
		sum := sha256.Sum256([]byte(str))
		return hex.EncodeToString(sum[:])
	default:
		return ""
	}
}

func custom_sign_path(str, custom_path_sign_salt string) (string, error) {
	if custom_path_sign_salt == "" {
		return "", errors.New("CUSTOM_PATH_SIGN_SALT not set")
	}
	hash1 := hash_func("SHA-512", str+custom_path_sign_salt+str)
	hash2 := hash_func("SHA-384", hash1+custom_path_sign_salt+str+custom_path_sign_salt)
	hash3 := hash_func("SHA-384", hash2+custom_path_sign_salt+hash1+custom_path_sign_salt)
	hash4 := hash_func("SHA-256", hash2+str+hash3+custom_path_sign_salt+hash1)
	hash5 := hash_func("SHA-256", hash3+str+hash1+custom_path_sign_salt+hash4+str+hash2)
	return hash5[:24], nil
}

func HandleCustomGenerateUrl(endpoint string, custom_path_sign_salt string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		objectName, ok := vars["objectName"]
		if !ok {
			handleHTTPError(w, fmt.Errorf("error getting object: objectName not found"))
			return
		}
		key, err := custom_sign_path(objectName, custom_path_sign_salt)
		if err != nil {
			handleHTTPError(w, fmt.Errorf("sign error: %w", err))
			return
		}
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		err = encoder.Encode(map[string]string{"url": "https://" + endpoint + "/" + key + "/" + objectName})
		if err != nil {
			handleHTTPError(w, fmt.Errorf("error encoding JSON: %w", err))
			return
		}
	}
}
