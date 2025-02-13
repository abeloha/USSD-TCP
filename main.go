package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net"
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
		logPath = "./logs"  // default path
	}
	var err error
	AppLogger, err = logger.New(logPath)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
}

// XML Message Structures

type LogonRequest struct {
	XMLName       xml.Name `xml:"AUTHRequest"`
	RequestID     string   `xml:"requestId"`
	Username      string   `xml:"userName"`
	Password      string   `xml:"passWord"`
	ApplicationID string   `xml:"applicationId"`
}

type USSDRequest struct {
	XMLName      xml.Name `xml:"USSDRequest"`
	RequestID    string   `xml:"requestId"`
	MSISDN       string   `xml:"msisdn"`
	StarCode     string   `xml:"starCode"`
	ClientID     string   `xml:"clientId"`
	Phase        int      `xml:"phase"`
	DCS          int      `xml:"dcs"`
	MsgType      int      `xml:"msgtype"`
	UserData     string   `xml:"userdata"`
	EndOfSession int      `xml:"EndofSession"`
}

type EnquireLink struct {
	XMLName xml.Name `xml:"ENQRequest"`
}

// Generates a unique Request ID (timestamp-based)
func generateRequestID() string {
	return fmt.Sprintf("%010d", time.Now().UnixNano()/int64(time.Millisecond))
}

// Creates a properly formatted 19-byte header
func createHeader(sessionID string, length int) []byte {
	header := make([]byte, 32)
	copy(header[:16], sessionID) // Use the provided session ID
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
	AppLogger.Info("[SEND] Full Message:\n%s\n", string(fullMessage))

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

	AppLogger.Info("[RECEIVE] Header: %s\n", header)
	AppLogger.Info("[RECEIVE] Raw Response:\n%s\n", string(body))

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
		RequestID:    requestID,
		Username:     Username,
		Password:     Password,
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
		log.Fatalf("Error reading response: %v", err)
		AppLogger.Error("Error reading response: %v", err)
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
					AppLogger.Error("Error reading server message: %v", err)
					// Optional: Add a small delay to prevent tight loop on continuous errors
					time.Sleep(1 * time.Second)
					continue
				}

				AppLogger.Info("[SERVER MESSAGE] Reading")
				AppLogger.Info("[SERVER MESSAGE] Header: %s", string(header))
				AppLogger.Info("[SERVER MESSAGE] Body: %s", string(body))
			}
		}
	}()

	// Periodic Enquire Link Request
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		enquireLink := EnquireLink{}
		enqXML, _ := xml.Marshal(enquireLink)
		fmt.Println("Sending Enquire Link Request...")
		if err := sendMessage(conn, enqXML, sessionID); err != nil {
			log.Fatalf("Failed to send Enquire Link: %v", err)
		}

		// Optional: Read and log Enquire Link Response
		header, body, err := readResponse(conn)
		if err != nil {
			AppLogger.Info("Error reading Enquire Link response: %v", err)
		} else {
			AppLogger.Info("[ENQUIRE LINK RESPONSE] Header: %s", string(header))
			AppLogger.Info("[ENQUIRE LINK RESPONSE] Body: %s", string(body))
		}
	}
}

func sendUssdMessage(conn net.Conn, msisdn, starCode, sessionID, clientID, message string) {
	// Send USSD Request
	ussdReq := USSDRequest{
		RequestID:    generateRequestID(),
		MSISDN:       msisdn,
		StarCode:     starCode,
		ClientID:     clientID,
		Phase:        2,
		DCS:          15,
		MsgType:      4, // Mobile-Originated USSD
		UserData:     message,
		EndOfSession: 0,
	}

	ussdXML, _ := xml.Marshal(ussdReq)
	fmt.Println("Sending USSD Request...")
	if err := sendMessage(conn, ussdXML, sessionID); err != nil {
		log.Fatalf("Failed to send USSD request: %v", err)
		AppLogger.Error("Failed to send USSD request: %v", err)
	}

	// Read USSD Response
	header, body, err := readResponse(conn)
	if err != nil {
		AppLogger.Info("Error reading response: %v", err)
	}

	// Log USSD Response
	AppLogger.Info("[USSD RESPONSE] Header: %s", string(header))
	AppLogger.Info("[USSD RESPONSE] Body: %s", string(body))
}


// function to perform general cleanup
func cleanup() {
	// Close the logger when the application exits
	if AppLogger != nil {
		AppLogger.Close()
	}
}