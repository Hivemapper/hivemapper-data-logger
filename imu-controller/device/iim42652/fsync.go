package iim42652

import (
	"fmt"
)

type Fsync struct {
	timedelta int16
	fsync_int bool
}

func NewFsync(timedelta int16, fsync_int bool) *Fsync {
	fsync := &Fsync{
		timedelta: timedelta,
		fsync_int: fsync_int,
	}

	return fsync
}

func (f *Fsync) String() string {
	return fmt.Sprintf("Fsync{FSYNC interrupt: %t, timedelta: %d}", f.fsync_int, f.timedelta)
}

func (i *IIM42652) GetFsync() (*Fsync, error) {
	i.registerLock.Lock()
	defer i.registerLock.Unlock()

	err := i.setBank(RegisterTmstFsyncH.Bank)
	if err != nil {
		return nil, fmt.Errorf("setting bank %s: %w", RegisterTmstFsyncH.Bank.String(), err)
	}

	// read three bytes starting at Bank 0, Register 2B
	msg := make([]byte, 4)
	result := make([]byte, 4)
	msg[0] = ReadMask | byte(RegisterTmstFsyncH.Address)
	if err := i.connection.Tx(msg, result); err != nil {
		return nil, fmt.Errorf("reading to SPI port: %w", err)
	}

	timedelta := int16(result[1])<<8 | int16(result[2])
	// read only bit 7 which is the UI_FSYNC_INT
	var mask byte = 1 << 6
	var fsync_int bool = (result[3] & mask) != 0

	fsync := NewFsync(timedelta, fsync_int)
	return fsync, nil
}