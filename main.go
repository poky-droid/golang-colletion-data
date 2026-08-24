package main

import (
	"bufio"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

func main() {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Use(RequestMiddleware())
	StartCollector(10 * time.Second) // rekam tiap 10 detik

	router.GET("/Matrix", func(c *gin.Context) {
		matrixs := GetMatrixs()
		c.JSON(200, matrixs)
	})

	router.Run()
}

func GetMatrixs() Matrixs {
	cpuPercent, _ := cpu.Percent(time.Second, false)
	vmStat, _ := mem.VirtualMemory()
	diskStat, _ := disk.Usage("/")
	netIn, netOut := GetNetworkStats()

	return Matrixs{
		CpuInfo:      cpuPercent[0],
		MemInfo:      vmStat.UsedPercent,
		DiskInfo:     diskStat.UsedPercent,
		Netin:        netIn,
		Netout:       netOut,
		RequestCount: requestCount,
		RequestTime:  requestTime,
	}
}

// GetNetworkStats parsing manual dari `netstat -ib`, berdasarkan
// (bukan index tetap) supaya aman dari bug parsing gopsutil di macOS.
func GetNetworkStats() (float64, float64) {
	out, err := exec.Command("netstat", "-ib").Output()
	if err != nil {
		println("netstat error:", err.Error())
		return 0, 0
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	var header []string
	var netIn, netOut float64
	seen := map[string]bool{}

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}

		if header == nil {
			header = fields
			continue
		}

		if len(fields) < len(header) {
			continue
		}

		iface := fields[0]
		if iface == "lo0" || seen[iface] {
			continue
		}
		seen[iface] = true

		ibytesIdx := indexOf(header, "Ibytes")
		obytesIdx := indexOf(header, "Obytes")
		if ibytesIdx == -1 || obytesIdx == -1 {
			continue
		}

		if ib, err := strconv.ParseFloat(fields[ibytesIdx], 64); err == nil {
			netIn += ib
		}
		if ob, err := strconv.ParseFloat(fields[obytesIdx], 64); err == nil {
			netOut += ob
		}
	}

	return netIn, netOut
}

func indexOf(slice []string, target string) int {
	for i, s := range slice {
		if s == target {
			return i
		}
	}
	return -1
}

func RequestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestCount++

		c.Next()

		requestTime = time.Since(start).Seconds() * 1000 // simpan sebagai ms

		println("Request Count:", requestCount)
		println("Request Time:", int64(requestTime), "ms")
	}
}
