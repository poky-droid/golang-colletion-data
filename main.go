package main

import (
	"bufio"
	"os"
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
		AllowOriginFunc: func(origin string) bool {
			allowed := origin == "https://dasboard-nine-gamma.vercel.app"
			println("CORS check -> origin diterima: ["+origin+"] | panjang:", len(origin), "| diizinkan:", allowed)
			return allowed
		},
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
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		println("error opening /proc/net/dev:", err.Error())
		return 0, 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var netIn, netOut float64
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		// 2 baris pertama adalah header, skip
		if lineNum <= 2 {
			continue
		}

		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		iface := strings.TrimSpace(parts[0])
		// skip loopback interface
		if iface == "lo" {
			continue
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 9 {
			continue
		}

		// Format /proc/net/dev per interface:
		// bytes packets errs drop fifo frame compressed multicast | bytes packets errs drop fifo colls carrier compressed
		// index 0 = Ibytes (RX), index 8 = Obytes (TX)
		if ib, err := strconv.ParseFloat(fields[0], 64); err == nil {
			netIn += ib
		}
		if ob, err := strconv.ParseFloat(fields[8], 64); err == nil {
			netOut += ob
		}
	}

	if err := scanner.Err(); err != nil {
		println("error reading /proc/net/dev:", err.Error())
		return 0, 0
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
