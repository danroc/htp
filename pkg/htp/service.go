package htp

type SyncOptions struct {
	Count int
	Trace func(round *SyncRound)
}

func Sync(client *SyncClient, model *SyncModel, options *SyncOptions) error {
	for i := 0; i < options.Count; i++ {
		model.Sleep()

		round, err := client.Round()
		if err != nil {
			return err
		}

		if err := model.Update(round); err != nil {
			return err
		}

		if options.Trace != nil {
			options.Trace(round)
		}
	}

	return nil
}
