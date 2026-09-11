package handlers

import (
	"errors"
	"net/http"
	"receipt-wrangler/api/internal/constants"
	"receipt-wrangler/api/internal/services"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
)

// searchResultLimit caps how many receipts a REST search returns (unchanged from
// the previous inline limit).
const searchResultLimit = 100

func Search(w http.ResponseWriter, r *http.Request) {
	handler := structs.Handler{
		ErrorMessage: "Error searching",
		Writer:       w,
		Request:      r,
		ResponseType: constants.ApplicationJson,
		// Enforcement (app.receipts.search, group scope narrowed to group.receipts.read,
		// paid-by visibility) lives in ReceiptService.SearchReceiptsForUser — the single
		// shared path also used by the MCP search_receipts tool — so the declarative gate
		// is intentionally omitted here to avoid two sources of truth.
		HandlerFunction: func(w http.ResponseWriter, r *http.Request) (int, error) {
			token := structs.GetClaims(r)
			searchTerm := r.URL.Query().Get("searchTerm")

			results, err := services.NewReceiptService(nil).SearchReceiptsForUser(token.UserId, searchTerm, searchResultLimit)
			if err != nil {
				if errors.Is(err, services.ErrSearchForbidden) {
					return http.StatusForbidden, err
				}
				return http.StatusInternalServerError, err
			}

			bytes, err := utils.MarshalResponseData(results)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			w.WriteHeader(200)
			w.Write(bytes)

			return 0, nil
		},
	}

	HandleRequest(handler)
}
