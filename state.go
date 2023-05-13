package gotransactions

import "os"

type StateSyncer interface {
	SaveState(string, []byte) (error)
	LoadState(string) ([]byte, error)
	LoadAllStates() (map[string][]byte, error)
}

type FileStateSyncer struct {
}

func (f *FileStateSyncer) SaveState(guid string, state []byte) error {
	os.Mkdir("states", 0755)
	file, err := os.Create("states/" + guid)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(state)
	if err != nil {
		return err
	}

	return nil
}

func (f *FileStateSyncer) LoadState(guid string) ([]byte, error) {
	file, err := os.Open("states/" + guid)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	state := make([]byte, 1024)
	_, err = file.Read(state)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (f *FileStateSyncer) LoadAllStates() (map[string][]byte, error) {
	states := make(map[string][]byte)
	files, err := os.ReadDir("states")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		state, err := f.LoadState(file.Name())
		if err != nil {
			return nil, err
		}
		states[file.Name()] = state
	}

	return states, nil
}

var _ StateSyncer = (*FileStateSyncer)(nil)
