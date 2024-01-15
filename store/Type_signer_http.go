package store

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/if-itb/siasn-jf-backend/store/models"
)

func (c *Client) HandleTypeSignerGet(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionSubmit)
	defer cancel()

	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return
	}
	rows, err := mtx.Query("SELECT penandatangan_id, key, nama FROM jenis_penandatangan")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var data []models.TypeSigner
	for rows.Next() {
		var d models.TypeSigner
		if err := rows.Scan(&d.PenandatanganId, &d.Key, &d.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data = append(data, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
