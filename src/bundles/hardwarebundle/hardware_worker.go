package hardwarebundle

import (
	"time"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/sc-js/pour"
	"gorm.io/gorm"
)

func pollUsage(db *gorm.DB) {
	for {
		hw := hardwareUsage{}
		memory, err := memory.Get()
		if err == nil {
			time.Sleep(time.Second)
			hw.MemoryTotal = memory.Total
			hw.MemoryFree = memory.Free
			hw.MemoryUsed = memory.Used
		}

		before, err := cpu.Get()
		if err == nil {
			time.Sleep(time.Second)
			after, err := cpu.Get()
			if err == nil {
				total := float64(after.Total - before.Total)
				hw.CPUIdle = float64(after.Idle-before.Idle) / total * 100
				hw.CPUSystem = float64(after.System-before.System) / total * 100
				hw.CPUUser = float64(after.User-before.User) / total * 100
			}
		}

		db.Unscoped().Delete(&hardwareUsage{}, "created_at < ?", time.Now().Add(-time.Hour*24))
		pour.LogColor(true, pour.ColorYellow, "CPU:[", "User:", hw.CPUUser, "% Idle:", hw.CPUIdle, "% System", hw.CPUSystem, "% ] - MEM:[", "Total", hw.MemoryTotal, "Free", hw.MemoryFree, "Used", hw.MemoryUsed, "]")
		if err := db.Create(&hw).Error; err != nil {
			pour.LogErr(err)
		}
		time.Sleep(time.Minute * 15)
	}
}
