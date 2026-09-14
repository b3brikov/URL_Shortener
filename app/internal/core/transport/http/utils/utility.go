package utilshttp

import (
	"net/http"
)

func SendError(w http.ResponseWriter, message string, code int) {
	http.Error(w, message, code)
}
