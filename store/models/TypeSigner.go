package models

type TypeSigner struct {
	PenandatanganId string `json:"penandatangan_id"`
	Key             string `json:"key" validate:"required"`
	Name            string `json:"name" validate:"required"`
}
