package stateful

import (
	"encoding/json"
	"fmt"
	"os"
)

type Stateful interface {
	SaveState() error
	LoadState() error
	Close() error
}

type JsonStateful struct {
	stateFile string
}

func (r *JsonStateful) prepareState() (any, error) {
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
	err = os.WriteFile(r.stateFile, data, 0o6)
	if err != nil {
		return err
	}

	return nil
}

func (r *JsonStateful) LoadState() error {
	return fmt.Errorf("not implemented")
}
