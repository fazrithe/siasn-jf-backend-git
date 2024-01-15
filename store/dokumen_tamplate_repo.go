package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/if-itb/siasn-jf-backend/store/models"
)

func (c *Client) InsertDokumenTemplateCtx(ctx context.Context, request *models.Dokumen_tamplate) (dokumenId string, err error) {
	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return "", err
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()

	activityIdBytes := uuid.New()

	_, err = mtx.ExecContext(
		ctx,
		"insert into dokumen_template(id, modul, filename, penandatangan) values($1, $2, $3, $4)",
		activityIdBytes,
		request.Modul,
		request.Filename,
		request.Penandatangan,
	)

	if err != nil {
		return "", fmt.Errorf("cannot insert entry to kegiatan: %w", err)
	}

	return activityIdBytes.String(), nil
}
