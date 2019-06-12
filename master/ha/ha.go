package ha

// Campaigner is a generic leadership compaigner for HA election
type Campaigner interface {
	WaitElection() (<-chan bool, <-chan error, error)

	CurrentLeader() (string, error)
}
