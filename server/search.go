package server

import (
	"net/http"
	"strings"
)

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusOK, []List{})
		return
	}

	store := s.storage.GetStore()
	var results []List

	for _, l := range store.Lists {
		if l.Archived && !store.Settings.ShowArchived {
			continue
		}

		if l.Type == "clipboard" {
			matchedList := l
			matchedList.Items = []Item{}

			for _, itm := range l.Items {
				// Match keys only and not value text
				if strings.Contains(strings.ToLower(itm.Key), q) {
					if l.Features.Sensitive {
						itm.Value = "••••••••"
					}
					matchedList.Items = append(matchedList.Items, itm)
				}
			}

			if len(matchedList.Items) > 0 {
				results = append(results, matchedList)
			}
			continue
		}

		listMatches := strings.Contains(strings.ToLower(l.Title), q)
		matchedList := l
		matchedList.Items = []Item{}

		for _, itm := range l.Items {
			if strings.Contains(strings.ToLower(itm.Title), q) || listMatches {
				matchedList.Items = append(matchedList.Items, itm)
			}
		}

		if len(matchedList.Items) > 0 || listMatches {
			results = append(results, matchedList)
		}
	}

	if results == nil {
		results = []List{}
	}
	writeJSON(w, http.StatusOK, results)
}
