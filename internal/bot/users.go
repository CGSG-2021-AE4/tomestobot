package bot

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/CGSG-2021-AE4/tomestobot/api"
)

// JSON Implementation for users id store

type jsonUsersIdStore struct {
	logger *slog.Logger

	filename string          // Storage filename
	ids      map[int64]int64 // Map where keys are tgIds, values are bxIds
}

func NewJsonUsersIdStore(logger *slog.Logger, filename string) api.UsersIdStore {
	// By default they are empty but we will fill them from file if no errors occurs
	ids := map[int64]int64{}

	// Read file
	if data, err := os.ReadFile(filename); err != nil {
		logger.Warn(fmt.Sprintf("Error while trying to read users id json file: %s\nWill create a new file", err.Error()))
	} else {
		// Parsing file
		// File contains of map[int64]int64 - ids
		if err := json.Unmarshal(data, &ids); err != nil {
			logger.Warn(fmt.Sprintf("Error while trying to parse users id json file: %s\nWill create a new file", err.Error()))
		}
	}
	return &jsonUsersIdStore{
		logger:   logger,
		filename: filename,
		ids:      ids,
	}
}

func (s *jsonUsersIdStore) Set(tgId int64, bxId int64) {
	s.ids[tgId] = bxId
}

func (s *jsonUsersIdStore) Get(tgId int64) (int64, bool) {
	id, ok := s.ids[tgId] // Allows only like this
	return id, ok
}

func (s *jsonUsersIdStore) Save() (outErr error) {
	defer func() {
		if outErr != nil {
			s.logger.Warn("Saving users json: " + outErr.Error())
		}
	}()

	// Convert to string
	data, err := json.Marshal(s.ids)
	if err != nil {
		return fmt.Errorf("marshal ids: %w", err)
	}

	// Open/Create file
	file, err := os.OpenFile(s.filename, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// Write to file
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write to file: %w", err)
	}
	return nil
}

func (s *jsonUsersIdStore) Close() (outErr error) {
	return s.Save()
}
