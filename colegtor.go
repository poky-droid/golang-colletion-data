package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

const csvFile = "dataset.csv"

func StartCollector(interval time.Duration) {
	// bikin file + header kalau belum ada
	initCSV()

	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			m := GetMatrixs()
			appendToCSV(m)
		}
	}()
}

func initCSV() {
	if _, err := os.Stat(csvFile); os.IsNotExist(err) {
		f, err := os.Create(csvFile)
		if err != nil {
			println("gagal bikin csv:", err.Error())
			return
		}
		defer f.Close()

		w := csv.NewWriter(f)
		defer w.Flush()

		w.Write([]string{
			"timestamp", "cpu", "ram", "disk",
			"net_in", "net_out", "request_count", "request_time",
		})
	}
}

func appendToCSV(m Matrixs) {
	f, err := os.OpenFile(csvFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		println("gagal buka csv:", err.Error())
		return
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	row := []string{
		time.Now().Format(time.RFC3339),
		fmt.Sprintf("%.2f", m.CpuInfo),
		fmt.Sprintf("%.2f", m.MemInfo),
		fmt.Sprintf("%.2f", m.DiskInfo),
		fmt.Sprintf("%.0f", m.Netin),
		fmt.Sprintf("%.0f", m.Netout),
		fmt.Sprintf("%d", m.RequestCount),
		fmt.Sprintf("%.2f", m.RequestTime),
	}

	w.Write(row)
}
