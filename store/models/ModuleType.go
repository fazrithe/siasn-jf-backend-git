package models

type ModuleType struct {
	ModuleId string `json:"modul_id"`
	Name     string `json:"name" validate:"required"`
}
