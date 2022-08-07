package utils

type NoCopy struct{} //nolint:unused

func (*NoCopy) Lock()   {} //nolint:unused
func (*NoCopy) Unlock() {} //nolint:unused
