package hardwarebundle

import (
	"github.com/jaypipes/ghw/pkg/baseboard"
	"github.com/jaypipes/ghw/pkg/bios"
	"github.com/jaypipes/ghw/pkg/block"
	"github.com/jaypipes/ghw/pkg/chassis"
	"github.com/jaypipes/ghw/pkg/cpu"
	"github.com/jaypipes/ghw/pkg/gpu"
	"github.com/jaypipes/ghw/pkg/memory"
	"github.com/jaypipes/ghw/pkg/net"
	"github.com/jaypipes/ghw/pkg/pci"
	"github.com/jaypipes/ghw/pkg/product"
	"github.com/jaypipes/ghw/pkg/topology"
	"github.com/sc-js/core_backend/src/tools"
)

type hardwareUsage struct {
	tools.Model
	MemoryTotal uint64  `json:"memory_total"`
	MemoryUsed  uint64  `json:"memory_used"`
	MemoryFree  uint64  `json:"memory_free"`
	CPUUser     float64 `json:"cpu_user"`
	CPUSystem   float64 `json:"cpu_system"`
	CPUIdle     float64 `json:"cpu_idle"`
}

type hardwareInformation struct {
	CPU       cpu.Info       `json:"cpu"`
	Memory    memory.Info    `json:"memory"`
	GPU       gpu.Info       `json:"gpu"`
	Storage   block.Info     `json:"storage"`
	Network   net.Info       `json:"network"`
	Topology  topology.Info  `json:"topology"`
	Bios      bios.Info      `json:"bios"`
	PCI       pci.Info       `json:"pci"`
	Baseboard baseboard.Info `json:"baseboard"`
	Chassis   chassis.Info   `json:"chassis"`
	Product   product.Info   `json:"product"`
}
