package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/html"
)

var db *sql.DB

type Analysis struct {
	ID                int      `json:"id"`
	URL               string   `json:"url"`
	HTMLVersion       string   `json:"html_version"`
	Title             string   `json:"title"`
	H1Count           int      `json:"h1_count"`
	H2Count           int      `json:"h2_count"`
	H3Count           int      `json:"h3_count"`
	H4Count           int      `json:"h4_count"`
	H5Count           int      `json:"h5_count"`
	H6Count           int      `json:"h6_count"`
	InternalLinks     int      `json:"internal_links"`
	ExternalLinks     int      `json:"external_links"`
	InaccessibleLinks int      `json:"inaccessible_links"`
	BrokenLinks       []string `json:"broken_links"`
	HasLoginForm      bool     `json:"has_login_form"`
	Status            string   `json:"status"`
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Database configuration with environment variables
	dbHost := getEnvWithDefault("DB_HOST", "localhost")
	dbPort := getEnvWithDefault("DB_PORT", "3306")
	dbUser := getEnvWithDefault("DB_USER", "user")
	dbPass := getEnvWithDefault("DB_PASSWORD", "password")
	dbName := getEnvWithDefault("DB_NAME", "webtraffic")

	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)
	
	// Retry database connection
	for i := 0; i < 30; i++ {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Printf("Failed to open database: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err = db.Ping(); err != nil {
			log.Printf("Failed to ping database (attempt %d/30): %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Println("Successfully connected to database")
		break
	}

	if err != nil {
		log.Fatal("Failed to connect to database after 30 attempts:", err)
	}

	createTable()

	go startWorker()

	r := gin.Default()

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	api := r.Group("/api")
	api.Use(authMiddleware())
	{
		api.POST("/analyze", analyzeHandler)
		api.POST("/analyze/rerun", rerunHandler)
		api.POST("/analyze/start", startAnalysisHandler)
		api.POST("/analyze/stop", stopAnalysisHandler)
		api.GET("/analyses", getAnalysesHandler)
		api.DELETE("/analyses/:id", deleteAnalysisHandler)
	}

	port := getEnvWithDefault("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	r.Run(":" + port)
}

func createTable() {
	// Try to read schema.sql, if it doesn't exist, use embedded schema
	var queries []string
	if schemaBytes, err := ioutil.ReadFile("schema.sql"); err == nil {
		// Split the schema into separate queries
		queryStr := string(schemaBytes)
		queries = strings.Split(queryStr, ";")
	} else {
		// Embedded schema as fallback - separate queries
		queries = []string{
			`CREATE TABLE IF NOT EXISTS analyses (
				id INT AUTO_INCREMENT PRIMARY KEY,
				url VARCHAR(255) NOT NULL,
				html_version VARCHAR(255),
				title VARCHAR(255),
				h1_count INT DEFAULT 0,
				h2_count INT DEFAULT 0,
				h3_count INT DEFAULT 0,
				h4_count INT DEFAULT 0,
				h5_count INT DEFAULT 0,
				h6_count INT DEFAULT 0,
				internal_links INT DEFAULT 0,
				external_links INT DEFAULT 0,
				inaccessible_links INT DEFAULT 0,
				has_login_form BOOLEAN,
				status VARCHAR(255) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
			`CREATE TABLE IF NOT EXISTS broken_links (
				id INT AUTO_INCREMENT PRIMARY KEY,
				analysis_id INT,
				link TEXT,
				FOREIGN KEY (analysis_id) REFERENCES analyses(id) ON DELETE CASCADE
			)`,
		}
	}

	// Execute each query separately
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		
		_, err := db.Exec(query)
		if err != nil {
			log.Fatal("Failed to create tables:", err)
		}
	}
	
	log.Println("Database tables created successfully")
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
			return
		}

		if parts[1] == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token not found"})
			return
		}

		c.Next()
	}
}

func analyzeHandler(c *gin.Context) {
	var body struct {
		URL string `json:"url"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := db.Exec("INSERT INTO analyses (url, status) VALUES (?, ?)", body.URL, "queued")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func rerunHandler(c *gin.Context) {
	var body struct {
		ID int `json:"id"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	_, err := db.Exec("UPDATE analyses SET status = ? WHERE id = ?", "queued", body.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func startAnalysisHandler(c *gin.Context) {
	var body struct {
		ID int `json:"id"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	_, err := db.Exec("UPDATE analyses SET status = ? WHERE id = ?", "queued", body.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func stopAnalysisHandler(c *gin.Context) {
	var body struct {
		ID int `json:"id"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	_, err := db.Exec("UPDATE analyses SET status = ? WHERE id = ?", "stopped", body.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func getAnalysesHandler(c *gin.Context) {
    rows, err := db.Query("SELECT id, url, html_version, title, h1_count, h2_count, h3_count, h4_count, h5_count, h6_count, internal_links, external_links, inaccessible_links, has_login_form, status FROM analyses ORDER BY created_at DESC")
    if err != nil {
        log.Printf("Error querying analyses: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query analyses"})
        return
    }
    defer rows.Close()

    var analyses []Analysis
    for rows.Next() {
        var analysis Analysis
        var htmlVersion, title sql.NullString
        var hasLoginForm sql.NullBool 

        err := rows.Scan(
            &analysis.ID, &analysis.URL, &htmlVersion, &title, 
            &analysis.H1Count, &analysis.H2Count, &analysis.H3Count, &analysis.H4Count, &analysis.H5Count, &analysis.H6Count, 
            &analysis.InternalLinks, &analysis.ExternalLinks, &analysis.InaccessibleLinks, 
            &hasLoginForm, 
            &analysis.Status,
        )
        
        if err != nil {
            log.Printf("Error scanning analysis row: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan analysis row"})
            return
        }

        analysis.HTMLVersion = htmlVersion.String
        analysis.Title = title.String
        analysis.HasLoginForm = hasLoginForm.Bool 

        brokenLinksRows, err := db.Query("SELECT link FROM broken_links WHERE analysis_id = ?", analysis.ID)
        if err != nil {
            log.Printf("Error querying broken links for analysis ID %d: %v", analysis.ID, err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query broken links"})
            return
        }

        var brokenLinks []string
        for brokenLinksRows.Next() {
            var link string
            if err := brokenLinksRows.Scan(&link); err != nil {
                log.Printf("Error scanning broken link: %v", err)
                brokenLinksRows.Close() // Ważne: zamknij przed return
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan broken link"})
                return
            }
            brokenLinks = append(brokenLinks, link)
        }
        brokenLinksRows.Close() 
        
        analysis.BrokenLinks = brokenLinks
        analyses = append(analyses, analysis)
    }

    if err = rows.Err(); err != nil {
        log.Printf("Error after iterating analysis rows: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating analysis results"})
        return
    }

    c.JSON(http.StatusOK, analyses)
}


func deleteAnalysisHandler(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec("DELETE FROM analyses WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func startWorker() {
	for {
		rows, err := db.Query("SELECT id, url FROM analyses WHERE status = ?", "queued")
		if err != nil {
			log.Println("Worker error:", err)
			time.Sleep(10 * time.Second)
			continue
		}

		for rows.Next() {
			var id int
			var url string
			err := rows.Scan(&id, &url)
			if err != nil {
				log.Println("Worker error:", err)
				continue
			}

			go processAnalysis(id, url)
		}

		rows.Close()
		time.Sleep(10 * time.Second)
	}
}

func processAnalysis(id int, url string) {
	_, err := db.Exec("UPDATE analyses SET status = ? WHERE id = ?", "running", id)
	if err != nil {
		log.Println("Worker error:", err)
		return
	}

	// Check if the analysis has been stopped
	var status string
	db.QueryRow("SELECT status FROM analyses WHERE id = ?", id).Scan(&status)
	if status == "stopped" {
		return
	}

	analysis, err := analyzeURL(url)
	if err != nil {
		_, dbErr := db.Exec("UPDATE analyses SET status = ? WHERE id = ?", "error", id)
		if dbErr != nil {
			log.Println("Worker error:", dbErr)
		}
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Println("Worker error:", err)
		return
	}

	_, err = tx.Exec("UPDATE analyses SET html_version = ?, title = ?, h1_count = ?, h2_count = ?, h3_count = ?, h4_count = ?, h5_count = ?, h6_count = ?, internal_links = ?, external_links = ?, inaccessible_links = ?, has_login_form = ?, status = ? WHERE id = ?",
		analysis.HTMLVersion, analysis.Title, analysis.H1Count, analysis.H2Count, analysis.H3Count, analysis.H4Count, analysis.H5Count, analysis.H6Count, analysis.InternalLinks, analysis.ExternalLinks, analysis.InaccessibleLinks, analysis.HasLoginForm, "done", id)
	if err != nil {
		tx.Rollback()
		log.Println("Worker error:", err)
		return
	}

	for _, link := range analysis.BrokenLinks {
		_, err := tx.Exec("INSERT INTO broken_links (analysis_id, link) VALUES (?, ?)", id, link)
		if err != nil {
			tx.Rollback()
			log.Println("Worker error:", err)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Worker error:", err)
	}
}

func analyzeURL(urlStr string) (*Analysis, error) {
	log.Printf("Analyzing URL: %s", urlStr)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	analysis := &Analysis{
		URL: urlStr,
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if n.FirstChild != nil {
					analysis.Title = n.FirstChild.Data
				}
			case "h1":
				analysis.H1Count++
			case "h2":
				analysis.H2Count++
			case "h3":
				analysis.H3Count++
			case "h4":
				analysis.H4Count++
			case "h5":
				analysis.H5Count++
			case "h6":
				analysis.H6Count++
			case "a":
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						if strings.HasPrefix(attr.Val, "http") {
							analysis.ExternalLinks++
						} else {
							analysis.InternalLinks++
						}
					}
				}
			case "form":
				for _, attr := range n.Attr {
					if attr.Key == "action" && (strings.Contains(attr.Val, "login") || strings.Contains(attr.Val, "signin")) {
						analysis.HasLoginForm = true
						break
					}
				}
				// Also check for password input fields
				if !analysis.HasLoginForm {
					var checkForPassword func(*html.Node)
					checkForPassword = func(child *html.Node) {
						if child.Type == html.ElementNode && child.Data == "input" {
							for _, attr := range child.Attr {
								if attr.Key == "type" && attr.Val == "password" {
									analysis.HasLoginForm = true
									return
								}
							}
						}
						for c := child.FirstChild; c != nil; c = c.NextSibling {
							checkForPassword(c)
						}
					}
					checkForPassword(n)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	analysis.HTMLVersion = getHTMLVersion(doc)

	analysis.BrokenLinks = checkInaccessibleLinks(doc, analysis.URL)
	analysis.InaccessibleLinks = len(analysis.BrokenLinks)

	return analysis, nil
}

func checkInaccessibleLinks(doc *html.Node, baseURL string) []string {
	var links []string
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					link, err := url.Parse(attr.Val)
					if err != nil {
						continue
					}

					base, err := url.Parse(baseURL)
					if err != nil {
						continue
					}

					resolvedLink := base.ResolveReference(link)

					resp, err := client.Get(resolvedLink.String())
					if err != nil || (resp.StatusCode >= 400 && resp.StatusCode <= 599) {
						links = append(links, resolvedLink.String())
					}
					if resp != nil {
						resp.Body.Close()
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return links
}

func getHTMLVersion(doc *html.Node) string {
	var version string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.DoctypeNode {
			version = n.Data
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if strings.Contains(strings.ToLower(version), "html 5") || version == "html" {
		return "HTML5"
	} else if strings.Contains(strings.ToLower(version), "xhtml 1.1") {
		return "XHTML 1.1"
	} else if strings.Contains(strings.ToLower(version), "xhtml 1.0") {
		return "XHTML 1.0"
	} else if strings.Contains(strings.ToLower(version), "html 4.01") {
		return "HTML 4.01"
	} else if strings.Contains(strings.ToLower(version), "html 4.0") {
		return "HTML 4.0"
	} else {
		return "HTML5" // Default to HTML5 if no doctype found
	}
}