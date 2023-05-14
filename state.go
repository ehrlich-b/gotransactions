package gotransactions

import (
	"encoding/gob"
	"os"
)

type StateSaver interface {
	SaveState(*Transaction) error
	LoadState(*Transaction) error
	GetAllStateGuids() ([]string, error)
}

type FileStateSaver struct {
}

func NewFileStateSaver() *FileStateSaver {
	return &FileStateSaver{}
}

func (fss *FileStateSaver) SaveState(t *Transaction) error {
	os.Mkdir("transactions", 0755)
	file := "transactions/" + t.guid
	t.SetState("", "currentStage", t.currentStage)
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	gobenc := gob.NewEncoder(f)
	err = gobenc.Encode(t.state)

	return err
}

func (fss *FileStateSaver) LoadState(t *Transaction) (error) {
	file := "transactions/" + t.guid
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	gobdec := gob.NewDecoder(f)
	err = gobdec.Decode(&t.state)
	t.currentStage = t.state["currentStage"].(int)

	return err
}

func (fss *FileStateSaver) GetAllStateGuids() ([]string, error) {
	files, err := os.ReadDir("transactions")
	if err != nil {
		return nil, err
	}
	guids := make([]string, 0)
	for _, file := range files {
		guids = append(guids, file.Name())
	}
	return guids, nil
}

var _ StateSaver = (*FileStateSaver)(nil)
