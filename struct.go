package main

type Matrixs struct {
	CpuInfo      float64 `json:"cpu_info"`
	MemInfo      float64 `json:"mem_info"`
	DiskInfo     float64 `json:"disk_info"`
	Netin        float64 `json:"net_in"`
	Netout       float64 `json:"net_out"`
	RequestCount int     `json:"request_count"`
	RequestTime  float64 `json:"request_time"`
}

var requestCount int
var requestTime float64
