package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/abeloha/USSDTCP/pkg/logger"
	"github.com/joho/godotenv"
)

var (
	ServerAddress string
	Username      string
	Password      string
	ClientID      string
	AppLogger     *logger.Logger
	ErrorLogger   *logger.Logger
	RequestLogger *logger.Logger
	MenuLogger    *logger.Logger
)

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Read environment variables
	host := os.Getenv("SERVER_HOST")
	port := os.Getenv("SERVER_PORT")
	ServerAddress = net.JoinHostPort(host, port)

	Username = os.Getenv("USERNAME")
	Password = os.Getenv("PASSWORD")
	ClientID = os.Getenv("CLIENT_ID")

	// Validate required environment variables
	requiredVars := []string{"SERVER_HOST", "SERVER_PORT", "USERNAME", "PASSWORD", "CLIENT_ID"}
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			log.Fatalf("Missing required environment variable: %s", v)
		}
	}

	// Initialize logger
	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = "./logs" // default path
	}
	var err error
	AppLogger, err = logger.New(logPath + "/log")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	ErrorLogger, err = logger.New(logPath + "/errors")
	if err != nil {
		log.Fatalf("Failed to initialize error logger: %v", err)
	}

	RequestLogger, err = logger.New(logPath + "/requests")
	if err != nil {
		log.Fatalf("Failed to initialize request logger: %v", err)
	}

	MenuLogger, err = logger.New(logPath + "/menu")
	if err != nil {
		log.Fatalf("Failed to initialize menu logger: %v", err)
	}
}

// Generates a unique Request ID (timestamp-based)
func generateRequestID() string {
	return fmt.Sprintf("%010d", time.Now().UnixNano()/int64(time.Millisecond))
}

// Creates a properly formatted 19-byte header
func createHeader(sessionID string, length int) []byte {
	header := make([]byte, 32)
	copy(header[:16], sessionID)             // Use the provided session ID
	lengthStr := fmt.Sprintf("%03d", length) // Ensure message length is 3-digit
	copy(header[16:], lengthStr)
	return header
}

// Utility function to send a message
func sendMessage(conn net.Conn, message []byte, sessionID string) error {
	fullXML := message
	header := createHeader(sessionID, len(fullXML)+32) // 16-byte session ID
	fullMessage := append(header, fullXML...)

	// Log the message
	AppLogger.Info("[SEND] Request:\n%s\n", string(fullXML))
	_, err := conn.Write(fullMessage)
	return err
}

// Reads a response and logs the raw data
func readResponse(conn net.Conn) ([]byte, []byte, error) {
	// Set a read timeout to prevent indefinite blocking
	err := conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to set read deadline: %v", err)
	}
	defer conn.SetReadDeadline(time.Time{}) // Clear deadline after reading

	header := make([]byte, 19)
	_, err = conn.Read(header)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, nil, fmt.Errorf("read timeout: no message received")
		}
		return nil, nil, fmt.Errorf("failed to read header: %v", err)
	}

	length, err := strconv.Atoi(string(header[16:]))
	if err != nil {
		return nil, nil, fmt.Errorf("invalid message length: %v", err)
	}

	body := make([]byte, length-16) // Subtract session ID length
	_, err = conn.Read(body)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, nil, fmt.Errorf("read timeout: incomplete message")
		}
		return nil, nil, fmt.Errorf("failed to read body: %v", err)
	}

	return header, body, nil
}

func main() {

	defer cleanup()

	AppLogger.Info("Starting USSD TCP Application")

	// Connect to server
	conn, err := net.Dial("tcp", ServerAddress)
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
		AppLogger.Error("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Generate a unique Request ID (timestamp-based)
	requestID := generateRequestID()

	// Send Logon Request
	logon := LogonRequest{
		RequestID:     requestID,
		Username:      Username,
		Password:      Password,
		ApplicationID: ClientID,
	}

	logonXML, _ := xml.Marshal(logon)
	fmt.Println("Sending Logon Request...")
	if err := sendMessage(conn, logonXML, requestID); err != nil {
		log.Fatalf("Failed to send logon: %v", err)
		AppLogger.Error("Failed to send logon: %v", err)
	}

	// Read Logon Response
	header, body, err := readResponse(conn)
	if err != nil {
		AppLogger.Error("Error reading response: %v", err)
		ErrorLogger.Error("Error reading response: %v", err)
		log.Fatalf("Error reading response: %v", err)
	}

	// Log response
	AppLogger.Info("[FINAL RESPONSE] Header: %s", string(header))
	AppLogger.Info("[FINAL RESPONSE] Body: %s", string(body))

	// Extract session ID from header (First 16 bytes)
	sessionID := string(header[:16])
	AppLogger.Info("Extracted Session ID: %s", sessionID)

	// Create a channel to signal when to stop listening
	stopChan := make(chan struct{})
	defer close(stopChan)

	// Goroutine for continuous message listening
	go func() {
		for {
			select {
			case <-stopChan:
				return
			default:
				header, body, err := readResponse(conn)
				if err != nil {
					// AppLogger.Error("Error reading server message: %v", err)
					// Add a small delay to prevent tight loop on continuous errors
					time.Sleep(1 * time.Second)
					continue
				}

				AppLogger.Info("[SERVER MESSAGE] Body: %s", string(body))

				// Process the response
				go processServerMessage(header, body, conn)
			}
		}
	}()

	// Periodic Enquire Link Request
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		enquireLink := EnquireLink{}
		enqXML, _ := xml.Marshal(enquireLink)
		fmt.Println("Sending Enquire Link Request...")
		if err := sendMessage(conn, enqXML, sessionID); err != nil {
			log.Fatalf("Failed to send Enquire Link: %v", err)
		}
	}
}


// processServerMessage checks if the message matches a USSDRequest, parses it, and logs it
func processServerMessage(header []byte, body []byte, conn net.Conn) {

	// Try to parse the XML body into USSDRequest
	var ussdRequest USSDRequest
	err := xml.Unmarshal(body, &ussdRequest)
	if err != nil || ussdRequest.XMLName.Local != "USSDRequest" {
		// not a valid USSDRequest
		return
	}

	// Log the parsed USSDRequest
	RequestLogger.Info("[INFO] Received USSD Request: %+v\n", ussdRequest)

	// Handle the USSD request
	handleUSSDRequest(ussdRequest, conn)
}

// handleUSSDRequest processes the parsed USSD request
func handleUSSDRequest(req USSDRequest, conn net.Conn) {
	
	if req.ErrorCode != "" {
		AppLogger.Info("Error code: %s for %s with code %s\n", req.ErrorCode, req.MSISDN, req.RequestID)
		return
	}

	if req.EndOfSession == 0 {
		handleMenuRequest(req, conn)
	} else {
		AppLogger.Info("USSD session ended for %s with code %s\n", req.MSISDN, req.RequestID)
	}
}


// getUSSDMenu calls the API and logs the request/response
func handleMenuRequest(req USSDRequest, conn net.Conn) {

	if (req.MsgType != 1) {
		AppLogger.Error("Invalid message type of %d for %s with code %s\n", req.MsgType, req.MSISDN, req.RequestID)
		return
	}

	if (req.UserData == "") {
		AppLogger.Error("Invalid input of %s for %s with code %s\n", req.UserData, req.MSISDN, req.RequestID)
		return
	}

	AppLogger.Info("[INFO] Continuing USSD session for %s with code %s\n", req.MSISDN, req.RequestID)

	//apiResponse, err := getUSSDMenu(req)
	apiResponse, err := getUSSDMenuMock(req)
	if err != nil {
		MenuLogger.Error("[ERROR] Failed to get USSD menu: %v\n", err)
		return
	}

	// Store response as variables
	ussdMessage := apiResponse.Message
	ussdContinue := apiResponse.Continue

	// Output stored response (for debugging)
	MenuLogger.Info("USSD Response Message:", ussdMessage)
	MenuLogger.Info("USSD Continue:", ussdContinue)

	// You can now use `ussdMessage` and `ussdContinue` for further processing.

	// send response back to client
	response := USSDResponse{
		RequestID:    req.RequestID,
		MSISDN:       req.MSISDN,
		StarCode:     req.StarCode,
		ClientID:     req.ClientID,
		Phase:        req.Phase,
		DCS:          req.DCS,
		MsgType:      0,
		UserData:     ussdMessage,
		EndOfSession: 0,
	}

	if !ussdContinue {
		response.EndOfSession = 1
	} 

	messageXML, _ := xml.Marshal(response)
	MenuLogger.Info("Sending ussd Request... for %s with code %s\n", req.MSISDN, req.RequestID)
	if err := sendMessage(conn, messageXML, response.RequestID); err != nil {
		MenuLogger.Error("Failed to ussd request message: %v", err)
	}

}


func getUSSDMenuMock(req USSDRequest) (*USSDMenuResponse, error) {
	var apiResponse USSDMenuResponse
	apiResponse.Continue = true
	apiResponse.Message = "This menu is coming soon"

	return &apiResponse, nil
}

func getUssdMenu(req USSDRequest) (*USSDMenuResponse, error){

		MenuLogger.Info("[INFO] Getting USSD menu for %s with code %s\n and request ID %s", req.MSISDN, req.StarCode, req.RequestID)


	// Prepare API request payload
	apiRequest := USSDMenuRequest{
		Telco:     "MTN", // Hardcoded for now; adjust as needed
		Shortcode: "*" + req.StarCode + "#",
		ProductID: 2,
		Phone:     req.MSISDN,
		Input:     req.UserData,
		SessionID: req.RequestID,
	}

	// Convert to JSON
	requestBody, err := json.Marshal(apiRequest)
	if err != nil {
		MenuLogger.Error("[ERROR] Failed to marshal request: %v\n", err)
		return nil, err
	}

	// API URL
	apiURL := os.Getenv("USSD_API_URL")
	if apiURL == "" {
		MenuLogger.Error("[ERROR] USSD menu url not set")
		return nil, errors.New("ussd menu url not set")
	}

	// Make HTTP request
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		MenuLogger.Error("[ERROR] Failed to call USSD menu API: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		MenuLogger.Error("[ERROR] Failed to read response: %v\n", err)
		return nil, err
	}

	// Log request and response
	MenuLogger.Info("[INFO] USSD Menu API Request: %s\n", string(requestBody))
	MenuLogger.Info("[INFO] USSD Menu API Response: %s\n", string(responseBody))

	// Parse JSON response
	var apiResponse USSDMenuResponse
	err = json.Unmarshal(responseBody, &apiResponse)
	if err != nil {
		log.Printf("[ERROR] Failed to parse response JSON: %v\n", err)
		return nil, err
	}

	return &apiResponse, nil
}

// function to perform general cleanup
func cleanup() {
	// Close the logger when the application exits
	if AppLogger != nil {
		AppLogger.Close()
	}
	if MenuLogger != nil {
		MenuLogger.Close()
	}
	if ErrorLogger != nil {
		ErrorLogger.Close()
	}
	if RequestLogger != nil {
		RequestLogger.Close()
	}
}
