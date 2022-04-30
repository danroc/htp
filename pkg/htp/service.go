package htp

type SyncTrace struct {
	Before func(i int) bool
	After  func(i int, round *SyncRound) bool
}

type SyncOptions struct {
	Count int
	Trace *SyncTrace
}

func Sync(client *SyncClient, model *SyncModel, trace *SyncTrace) error {
	for i := 0; ; i++ {
		if !trace.Before(i) {
			break
		}

		model.Sleep()

		round, err := client.Round()
		if err != nil {
			return err
		}

		if err := model.Update(round); err != nil {
			return err
		}

		if !trace.After(i, round) {
			break
		}
	}

	return nil
}

func SyncSystem(model *SyncModel) error {
	return syncSystem(model)
}
