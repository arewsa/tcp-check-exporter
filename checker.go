package main

import (
    "net"
    "time"
)

func checkHost(host, port string, timeout time.Duration) (up bool, latency float64) {
    start := time.Now()
    conn, err := net.DialTimeout("tcp", host+":"+port, timeout)
    latency = time.Since(start).Seconds()

    if err != nil {
        return false, latency
    }
    conn.Close()
    return true, latency
}