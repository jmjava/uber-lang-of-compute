package store

import (
	"database/sql"
	"errors"
	"os"
)

func notFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, os.ErrNotExist) {
		return true
	}
	return err.Error() == "snapshot not found"
}

func refuseSealedMutation(getErr error, existingTime, existingData string, existingSealed bool, timeSlice, data string, sealed bool) error {
	if getErr != nil {
		if notFound(getErr) {
			return nil
		}
		return getErr
	}
	if !existingSealed {
		return nil
	}
	if sealed && existingTime == timeSlice && existingData == data {
		return nil
	}
	return ErrSealedOverwrite
}

func refuseMemoConflict(found bool, storedHash, storedOut, outputHash, output string) error {
	if !found {
		return nil
	}
	if storedHash == outputHash && storedOut == output {
		return nil
	}
	return ErrMemoConflict
}
