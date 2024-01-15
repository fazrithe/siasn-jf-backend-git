package models

const (
	StaffRoleSupervisor = "pejabat_pembina"
)

type OrganizationUnit struct {
	OrganizationUnitId   string `json:"unit_organisasi_id"`
	OrganizationUnitName string `json:"unit_organisasi"`
}
