package stateful

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Stateful interface {
	SaveState() error
	LoadState() error
	Close() error
}

type JsonStateful struct {
	stateFile string
}

func (r *JsonStateful) prepareState() (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *JsonStateful) SaveState() error {
	res, err := r.prepareState()
	if err != nil {
		return err
	}

	data, err := json.Marshal(res)
	if err != nil {
		return err
	}
	err = os.WriteFile(r.stateFile, data, 06)
	if err != nil {
		return err
	}

	return nil
}

func (r *JsonStateful) LoadState() error {
	return fmt.Errorf("not implemented")
}

func SaveState(ctx context.Context, state Stateful) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		time.Sleep(1 * time.Minute)
		state.SaveState()
	}
}
