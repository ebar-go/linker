package system

import "runtime"

func GetMem() uint64 {
	var memStat runtime.MemStats
	runtime.ReadMemStats(&memStat)
	return memStat.Sys
}
