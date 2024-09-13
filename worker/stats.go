package worker

import (
	"github.com/c9s/goprocinfo/linux"
	"log"
)

type Stats struct {
	MemStats  *linux.MemInfo
	DiskStats *linux.Disk
	CpuStats  *linux.CPUStat
	LoadStats *linux.LoadAvg
	TaskCount int
}

func (s *Stats) MemTotalKb() uint64 {
	return s.MemStats.MemTotal
}

func (s *Stats) MemAvailableKb() uint64 {
	return s.MemStats.MemAvailable
}

func (s *Stats) MemUsedKb() uint64 {
	return s.MemStats.MemTotal - s.MemStats.MemAvailable
}

func (s *Stats) MemUsedPercent() uint64 {
	return s.MemStats.MemAvailable / s.MemStats.MemTotal
}

func (s *Stats) DiskTotal() uint64 {
	return s.DiskStats.All
}

func (s *Stats) DiskFree() uint64 {
	return s.DiskStats.Free
}

func (s *Stats) DiskUsed() uint64 {
	return s.DiskStats.Used
}

func (s *Stats) CupUsage() float64 {
	idle := s.CpuStats.Idle + s.CpuStats.IOWait
	nonIdle := s.CpuStats.User + s.CpuStats.Nice + s.CpuStats.System + s.CpuStats.IRQ + s.CpuStats.SoftIRQ + s.CpuStats.Steal
	total := idle + nonIdle

	if total == 0 {
		return 0.00
	}

	return (float64(total) - float64(idle)) / float64(total)
}

func GetStats() *Stats {
	return &Stats{
		MemStats:  GetMemoryInfo(),
		DiskStats: GetDiskInfo(),
		CpuStats:  GetCpuStats(),
		LoadStats: GetLoadAvg(),
	}
}

func GetMemoryInfo() *linux.MemInfo {
	mem, err := linux.ReadMemInfo("/proc/meminfo")
	if err != nil {
		log.Printf("Error reading memory info: %v\n", err)
		return &linux.MemInfo{}
	}

	return mem
}

func GetDiskInfo() *linux.Disk {
	disk, err := linux.ReadDisk("/")
	if err != nil {
		log.Printf("Error reading from /: %v\n", err)
		return &linux.Disk{}
	}

	return disk
}

func GetCpuStats() *linux.CPUStat {
	file := "/proc/stat"
	stats, err := linux.ReadStat(file)
	if err != nil {
		log.Printf("Error reading from %s: %v\n", file, err)
		return &linux.CPUStat{}
	}

	return &stats.CPUStatAll
}

func GetLoadAvg() *linux.LoadAvg {
	file := "/proc/loadavg"
	la, err := linux.ReadLoadAvg(file)
	if err != nil {
		log.Printf("Error reading from %s: %v\n", file, err)
		return &linux.LoadAvg{}
	}

	return la
}
