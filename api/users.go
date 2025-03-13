package api

import "io"

// Connects telegram id with BX id so I do not have to request contact every time
type UsersIdStore interface {
	Set(tgId int64, bxId int64)   // Stores bxId for tgId
	Get(tgId int64) (int64, bool) // Similar to map field existance check: first - value, second - does the value exists
	Save() error                  // Temp function because I do not catch interupt signal yet...
	io.Closer
}
