package ha

// NewDummyCampaigner is exported
func NewDummyCampaigner() Campaigner {
	return new(dummy)
}

type dummy struct{}

func (d *dummy) WaitElection() (<-chan bool, <-chan error, error) {
	var (
		electResCh = make(chan bool, 1)
		electErrCh = make(chan error, 1)
	)

	go func() {
		electResCh <- true
		select {}
	}()

	return electResCh, electErrCh, nil
}

func (d *dummy) CurrentLeader() (string, error) {
	return "a dummpy ha campaigner", nil
}
