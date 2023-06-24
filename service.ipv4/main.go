package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"service.ipv4/db"
	"service.ipv4/models"
)

type Repository struct {
	// Pointer to *gorm.DB
	DB *gorm.DB
}

func isValidIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() != nil
}

func (rep *Repository) GetReports(w http.ResponseWriter, req *http.Request) {
	type results struct {
		IPAddress     string    `json:"ip_address" gorm:"primaryKey;unique; not null"`
		AddedAt       time.Time `json:"added_at"`
		UpdatedAt     time.Time `json:"updated_at"`
		Malicious     bool      `json:"malicious"`
		Comment       string    `json:"comment"`
		ReportAddedAt string    `json:"report_added_at"`
	}

	type report struct {
		Comment       string `json:"comment"`
		Malicious     bool   `json:"malicious"`
		ReportAddedAt string `json:"report_added_at"`
	}

	type response struct {
		IPAddress string    `json:"ip_address"`
		AddedAt   time.Time `json:"added_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Reports   []report  `json:"reports"`
	}

	var resultsList []results
	err := rep.DB.Table("ipv4_addresses").
		Joins("left join reports on reports.fk_ip_address = ipv4_addresses.ip_address").
		Select("ipv4_addresses.ip_address, ipv4_addresses.added_at, ipv4_addresses.updated_at, reports.malicious, reports.comment, reports.report_added_at").
		Scan(&resultsList).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the final response structure
	finalResponse := make([]response, 0)
	for _, res := range resultsList {
		found := false
		for i, fr := range finalResponse {
			if fr.IPAddress == res.IPAddress {
				finalResponse[i].Reports = append(finalResponse[i].Reports, report{
					Comment:       res.Comment,
					Malicious:     res.Malicious,
					ReportAddedAt: res.ReportAddedAt,
				})
				found = true
				break
			}
		}
		if !found {
			finalResponse = append(finalResponse, response{
				IPAddress: res.IPAddress,
				AddedAt:   res.AddedAt,
				UpdatedAt: res.UpdatedAt,
				Reports: []report{
					{
						Comment:       res.Comment,
						Malicious:     res.Malicious,
						ReportAddedAt: res.ReportAddedAt,
					},
				},
			})
		}
	}

	// Convert response to JSON
	responseData, err := json.Marshal(finalResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseData)
}

func (rep *Repository) GetReportsByIP(w http.ResponseWriter, req *http.Request) {
	ip := req.URL.Query().Get("ip")
	reportModels := &[]models.Reports{{FKIPAddress: ip}}

	err := rep.DB.Where("fk_ip_address = ?", ip).Find(&reportModels).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(reportModels)
}

func (rep *Repository) AddReport(w http.ResponseWriter, req *http.Request) {
	var id uuid.UUID
	var malicious bool

	err := req.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	id, err = uuid.NewRandom()
	if err != nil {
		log.Println("Could not generate ID")
	}

	ip := req.Form.Get("ip_address")
	comment := req.Form.Get("comment")
	malicious, err = strconv.ParseBool(req.Form.Get("malicious"))
	if err != nil {
		fmt.Println("could not parse string to bool")
	}

	// Check if IP address already exists
	var existingIP models.IPv4Addresses
	err = rep.DB.Where("ip_address = ?", ip).First(&existingIP).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "Failed to query IP address entry", http.StatusInternalServerError)
		return
	}

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		// IP address entry doesn't exist, create a new one
		ipEntry := &models.IPv4Addresses{
			IPAddress: ip,
			AddedAt:   time.Now(),
			UpdatedAt: time.Now(),
			ISP:       "",
			Country:   "",
			Region:    "",
			City:      "",
		}

		// Create the IP address entry in the database
		err = rep.DB.Create(ipEntry).Error
		if err != nil {
			http.Error(w, "Failed to create IP address entry", http.StatusInternalServerError)
			return
		}
	}

	reportEntry := &models.Reports{
		ID:            id,
		Malicious:     malicious,
		Comment:       comment,
		ReportAddedAt: time.Now(),
		FKIPAddress:   ip,
	}

	// Create report entry in database
	err = rep.DB.Create(reportEntry).Error
	if err != nil {
		http.Error(w, "Failed to create report entry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "ok"})
}

func (rep *Repository) DeleteReportByID(w http.ResponseWriter, req *http.Request) {
	// Parse report ID from request parameters or body
	reportID := req.FormValue("report_id")

	// Check if report ID is provided
	if reportID == "" {
		http.Error(w, "Report ID is required", http.StatusBadRequest)
		return
	}

	// Perform deletion
	result := rep.DB.Where("id = ?", reportID).Delete(&models.Reports{})
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Check the number of affected rows
	if result.RowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"message": "report id does not exist"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "report deleted"})
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	config := &db.Config{
		Host:     "ipv4_db",
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("SSL_Mode"),
	}

	db, err := db.NewConnection(config)
	if err != nil {
		log.Fatal("Could not load the Database")
	}

	err = models.Migrate(db)
	if err != nil {
		log.Fatal("could not migrate db")
	}

	r := Repository{
		DB: db,
	}

	http.HandleFunc("/api/v1/ipv4/reports", r.GetReports)
	http.HandleFunc("/api/v1/ipv4/getreportsbyip", r.GetReportsByIP)
	http.HandleFunc("/api/v1/ipv4/addreport", r.AddReport)
	http.HandleFunc("/api/v1/ipv4/deletereportbyid", r.DeleteReportByID)
	http.HandleFunc("/api/v1/ipv4/health", healthcheckHandler)
	log.Fatal(http.ListenAndServe(":4000", nil))
}
