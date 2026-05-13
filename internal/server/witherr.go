package server

import "net/http"

func handlerFuncWithErr(logger *Logger, hf func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := hf(w, r); err != nil {
			logger.L.Error("Big ooopsies. Will report where in a hot second")
			http.Error(w, "OH SHIT", http.StatusInternalServerError)
		}

	}

}
