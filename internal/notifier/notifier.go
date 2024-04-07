package notifier

type Notification struct {
	ID     string
	UserID string
	Domain string
	Status string
	Hours  int
}

type Notifier interface {
	Notify(to string, notification Notification) error
}
