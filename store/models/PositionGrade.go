package models

// PositionGrade represents single entry of position.
type PositionGrade struct {
	PositionGradeId   string `json:"jabatan_id"`
	PositionGradeName string `json:"nama_jabatan"`

	// AssignedEmployeeCount is the number of employee, the same as bezetting.
	AssignedEmployeeCount int `json:"jumlah_pegawai,omitempty"`
}
