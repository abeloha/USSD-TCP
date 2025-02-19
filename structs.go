package main

import "encoding/xml"

// XML Message Structures

type LogonRequest struct {
	XMLName       xml.Name `xml:"AUTHRequest"`
	RequestID     string   `xml:"requestId"`
	Username      string   `xml:"userName"`
	Password      string   `xml:"passWord"`
	ApplicationID string   `xml:"applicationId"`
}

// type USSDRequest struct {
// 	XMLName      xml.Name `xml:"USSDRequest"`
// 	RequestID    string   `xml:"requestId"`
// 	MSISDN       string   `xml:"msisdn"`
// 	StarCode     string   `xml:"starCode"`
// 	ClientID     string   `xml:"clientId"`
// 	Phase        int      `xml:"phase"`
// 	DCS          int      `xml:"dcs"`
// 	MsgType      int      `xml:"msgtype"`
// 	UserData     string   `xml:"userdata"`
// 	EndOfSession int      `xml:"EndofSession"`
// }

type USSDRequest struct {
	XMLName      xml.Name `xml:"USSDRequest"`
	RequestID    string   `xml:"requestId"`
	MSISDN       string   `xml:"msisdn"`
	IMSI         string   `xml:"imsi,omitempty"` // Optional field
	StarCode     string   `xml:"starCode"`
	ClientID     string   `xml:"clientId"`
	Phase        int      `xml:"phase"`
	DCS          int      `xml:"dcs"`
	MsgType      int      `xml:"msgtype"`
	UserData     string   `xml:"userdata,omitempty"` // Optional field
	EndOfSession int      `xml:"EndofSession"`
	ErrorCode    string   `xml:"errorCode,omitempty"` // Optional field
}

type EnquireLink struct {
	XMLName xml.Name `xml:"ENQRequest"`
}

// USSDMenuRequest represents the API request payload
type USSDMenuRequest struct {
	Telco      string `json:"telco"`
	Shortcode  string `json:"shortcode"`
	ProductID  int    `json:"product_id"`
	Phone      string `json:"phone"`
	Input      string `json:"input"`
	SessionID  string `json:"session_id"`
}

// USSDMenuResponse represents the API response payload
type USSDMenuResponse struct {
	Message  string `json:"message"`
	Continue bool   `json:"continue"`
}


