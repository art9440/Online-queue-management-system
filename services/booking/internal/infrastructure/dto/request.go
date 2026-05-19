package dto

type GetAppointmentsRequest struct {
	EmployeeID int64 `json:"employee_id"`
}

type CreateAppointmentRequest struct {
	Client ClientRequest `json:"client"`

	BranchID   int64  `json:"branch_id"`
	EmployeeID int64  `json:"employee_id"`
	ServiceID  int64  `json:"service_id"`
	StartTime  string `json:"start_time"`

	Comment *string `json:"comment,omitempty"`
}

type ClientRequest struct {
	Email      *string `json:"email,omitempty"`
	Phone      string  `json:"phone"`
	Name       string  `json:"name"`
	Surname    string  `json:"surname"`
	TgUsername *string `json:"tg_username,omitempty"`
}
