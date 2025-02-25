package systemHealthController

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
)

type SystemHealthController struct {
	
}

func (c *SystemHealthController) Index(ctx *gin.Context) {
	cpuUsage := c.getCpuUsage()
	ramUsage := c.getRamUsage()
	diskUsage := c.getDiskUsage()
	dbActive := c.isDatabaseActive()
	dbConnections := c.getDatabaseConnections()
	redisHealth := c.getRedisHealth()

	ctx.JSON(200, gin.H{
		"cpu_usage":            cpuUsage,
		"ram_usage":            ramUsage,
		"disk_usage":           diskUsage,
		"db_active":            dbActive,
		"active_db_connections": dbConnections,
		"redis_active":         redisHealth,
	})


	

}

func (c *SystemHealthController) getCpuUsage() float64 {
	if runtime.GOOS == "windows" {
		return 0 // Implement CPU usage for Windows if required
	}
	out, err := exec.Command("sh", "-c", "cat /proc/loadavg | awk '{print $1}'").Output()
	if err != nil {
		return 0
	}
	load, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0
	}
	return load
}

func (c *SystemHealthController) getRamUsage() float64 {
	out, err := exec.Command("free", "-m").Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return 0
	}

	mem := strings.Fields(lines[1])
	if len(mem) < 3 {
		return 0
	}

	total, err1 := strconv.ParseFloat(mem[1], 64)
	used, err2 := strconv.ParseFloat(mem[2], 64)
	if err1 != nil || err2 != nil {
		return 0
	}

	return (used / total) * 100
}

func (c *SystemHealthController) getDiskUsage() map[string]interface{} {
	var stat syscall.Statfs_t
	syscall.Statfs("/", &stat)

	total := float64(stat.Blocks) * float64(stat.Bsize)
	free := float64(stat.Bfree) * float64(stat.Bsize)
	used := total - free
	percentage := (used / total) * 100

	return map[string]interface{}{
		"used":       fmt.Sprintf("%.2f GB", used/(1024*1024*1024)),
		"total":      fmt.Sprintf("%.2f GB", total/(1024*1024*1024)),
		"percentage": fmt.Sprintf("%.2f", percentage),
	}
}

func (c *SystemHealthController) isDatabaseActive() bool {
	return true
}

func (c *SystemHealthController) getDatabaseConnections() int {
	return 0
}

func (c *SystemHealthController) getRedisHealth() bool {
	return true
}