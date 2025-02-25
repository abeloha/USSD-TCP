package jobs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/abeloha/USSDTCP/pkg/logger"
	"github.com/joho/godotenv"
)



func getLogger(channel string) (*logger.Logger, error) {
	// Load .env file
	 godotenv.Load()

	// Initialize logger
	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = "./logs" // default path
	}

	if (channel == "error") {
			return logger.New(logPath + "/monitoring/errors")
	}

			
	return logger.New(logPath + "/monitoring/logs")
	

	
}

type PostMetricData struct {
	URL      string
	Metric   string
	Value    interface{}
	Context1 interface{}
	Context2 interface{}
	Details  interface{}
}



// all interface is nullable string
func NewPostMetricData(metric string, value int, context1, context2, details interface{}) *PostMetricData {
	return &PostMetricData{
		URL:      "http://164.92.240.63:8000/api/update_metrics",
		Metric:   metric,
		Value:    value,
		Context1: context1,
		Context2: context2,
		Details:  details,
	}
}

func (p *PostMetricData) Handle() {



	errorLogger, err := getLogger("error")

	monitoringStatus := os.Getenv("MONITORING_STATUS")
	if monitoringStatus == "INACTIVE" {
		return
	}

	
	if errorLogger != nil {
		errorLogger.Error("Test2")
	}

	data := map[string]interface{}{
		"api_key":   os.Getenv("MONITORING_API_KEY"),
		"metric":    p.Metric,
		"value":     p.Value,
		"context_1": p.Context1,
		"context_2": p.Context2,
		"log":       p.Details,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		if errorLogger != nil {
		errorLogger.Error("Failed to marshal data: %v", err)
		}
		return
	}

	req, err := http.NewRequest("POST", p.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		if errorLogger != nil {
		errorLogger.Error("Failed to create request: %v", err)
		}
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if errorLogger != nil {
		errorLogger.Error("Failed to post metric data: %v", err)
		}
		return
	}
	defer resp.Body.Close()

	logMsg := ""
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logMsg = fmt.Sprint("Metric data posted successfully. Status: %v", resp.Status)
		errorLogger.Error(logMsg)
	} else {
		logMsg = fmt.Sprint("Failed to post metric data. Status: %v", resp.Status)
		errorLogger.Error(logMsg)
	}
}