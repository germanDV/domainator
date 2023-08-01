package plans

// Plan represents a subscription plan.
type Plan struct {
	ID           int
	Name         string
	Price        int
	DomainsLimit int
	CertsLimit   int
}
