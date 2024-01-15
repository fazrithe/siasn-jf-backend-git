package models

type Dokumen_tamplate struct {
	ID            string `json:"id"`
	Name          string `json:"name" validate:"required"`
	Modul         string `json:"modul" validate:"required"`
	Filename      string `json:"filename" validate:"required"`
	Filpath       string `json:"filepath" validate:"required"`
	Penandatangan string `json:"penandatangan" validate:"required"`
}

type Dokumen_tamplate_item struct {
	ID            string `json:"id"`
	Name          string `json:"name" validate:"required"`
	Modul         string `json:"modul" validate:"required"`
	Filename      string `json:"filename" validate:"required"`
	Penandatangan string `json:"penandatangan" validate:"required"`
}
