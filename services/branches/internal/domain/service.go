package domain

type Service struct {
	ID              int64
	BranchID        int64
	Name            string
	DurationMinutes int
	Price           float64
}
