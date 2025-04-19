package rest

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"yadro.com/course/api/core"
)

func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := make(map[string]string)
		for name, pinger := range pingers {
			if err := pinger.Ping(r.Context()); err != nil {
				resp[name] = "unavailable"
			} else {
				resp[name] = "ok"
			}
		}
		wrappedResp := map[string]interface{}{
			"replies": resp,
		}
		err := json.NewEncoder(w).Encode(wrappedResp)
		if err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}

func NewWordsHandler(log *slog.Logger, normalizer core.Normalizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, core.ErrBadArguments.Error())
			return
		}

		words, err := normalizer.Norm(r.Context(), phrase)
		if err != nil {
			currStatus := http.StatusInternalServerError
			if code := status.Code(err); code == codes.ResourceExhausted {
				currStatus = http.StatusBadRequest
				log.Debug("received message larger than 4KB", "phrase", phrase)
			} else {
				log.Error("failed to normalize phrase", "error", err)
			}
			w.WriteHeader(currStatus)
			fmt.Fprintf(w, "error normalizing phrase")
			return
		}

		response := map[string]interface{}{
			"words": words,
			"total": len(words),
		}
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}

func NewUpdateHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := updater.Update(r.Context())
		if err != nil {
			if code := status.Code(err); code == codes.AlreadyExists {
				log.Debug("already updating", "error", err)
				w.WriteHeader(http.StatusAccepted)
				return
			}
			log.Error("failed to update", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error updating")
			return
		}
	}
}

func NewUpdateStatsHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := updater.Stats(r.Context())
		if err != nil {
			log.Error("failed to get stats", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error getting stats")
			return
		}

		response := map[string]interface{}{
			"words_total":    stats.WordsTotal,
			"words_unique":   stats.WordsUnique,
			"comics_fetched": stats.ComicsFetched,
			"comics_total":   stats.ComicsTotal,
		}
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}

func NewUpdateStatusHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := updater.Status(r.Context())
		if err != nil {
			log.Error("failed to get status", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error getting status")
			return
		}
		response := map[string]interface{}{
			"status": status,
		}
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}

func NewDropHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := updater.Drop(r.Context())
		if err != nil {
			log.Error("failed to drop", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error dropping")
			return
		}
	}
}

func NewSearchHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")
		limit := r.URL.Query().Get("limit")
		if phrase == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, core.ErrBadArguments.Error())
			return
		}

		if limit == "" {
			limit = "10"
		}

		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, core.ErrBadArguments.Error())
			return
		}

		if limitInt < 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, core.ErrBadArguments.Error())
			return
		}

		comics, total, err := searcher.Search(r.Context(), phrase, limitInt)
		if err != nil {
			log.Error("failed to search", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error searching")
			return
		}

		response := map[string]interface{}{
			"comics": comics,
			"total":  total,
		}
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}

func NewISearchHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")
		limit := r.URL.Query().Get("limit")
		if phrase == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, core.ErrBadArguments.Error())
			return
		}

		if limit == "" {
			limit = "10"
		}

		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, core.ErrBadArguments.Error())
			return
		}

		if limitInt < 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, core.ErrBadArguments.Error())
			return
		}

		comics, total, err := searcher.ISearch(r.Context(), phrase, limitInt)
		if err != nil {
			log.Error("failed to search", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error searching")
			return
		}

		response := map[string]interface{}{
			"comics": comics,
			"total":  total,
		}
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Error("failed to encode response", "error", err)
		}
	}
}
