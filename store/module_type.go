package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/if-itb/siasn-jf-backend/store/models"
)

func (c *Client) InsertModuleTypeCtx(ctx context.Context, request *models.ModuleType) (activityId string, err error) {
	mtx, err := c.createMtxDb(ctx, c.Db)
	if err != nil {
		return "", err
	}

	defer func() {
		c.completeMtx(mtx, err)
	}()

	moduleIdBytes := uuid.New()
	query := "insert into jenis_modul(modul_id, nama) values($1, $2)"
	_, err = mtx.ExecContext(
		ctx,
		query,
		moduleIdBytes,
		request.Name,
	)

	if err != nil {
		return "", fmt.Errorf("cannot insert entry to Module Type: %w", err)
	}

	return moduleIdBytes.String(), nil
}
