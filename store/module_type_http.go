package store

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/if-itb/siasn-jf-backend/store/models"
	"github.com/if-itb/siasn-libs-backend/httputil"
)

func (c *Client) HandleModuleTypeSubmit(writer http.ResponseWriter, request *http.Request) {
	ar := &models.ModuleType{}
	err := c.decodeRequestJson(writer, request, ar)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionSubmit)
	defer cancel()

	modulId, err := c.InsertModuleTypeCtx(ctx, ar)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, map[string]string{
		"modul_id": modulId,
	})
}

func (c *Client) HandleModuleTypeGet(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutActivityAdmissionSubmit)
	defer cancel()

	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return
	}
	rows, err := mtx.Query("SELECT modul_id, nama FROM jenis_modul")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var data []models.ModuleType
	for rows.Next() {
		var d models.ModuleType
		if err := rows.Scan(&d.ModuleId, &d.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data = append(data, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
