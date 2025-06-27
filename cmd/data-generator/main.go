package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync" // Import sync for identity simulation
	"time"
)

// --- Structures (matching tracker/types.go and src/track.ts payload) ---
type TrackingData struct {
	Type          string `json:"type"` // "page" or "event"
	Identity      string `json:"identity"`
	UserAgent     string `json:"ua"`
	Event         string `json:"event"`
	Category      string `json:"category"`
	Referrer      string `json:"referrer"`
	IsTouchDevice bool   `json:"isTouchDevice"`
	// OccuredAt is set by the server
}

type Tracking struct {
	SiteID string       `json:"site_id"`
	Action TrackingData `json:"tracking"`
}

// --- ---

var (
	logger *slog.Logger
)

// --- Data Generation Helpers ---

var sites = []string{"siteA", "siteB", "cool-blog", "news-corp"}

// Paths that tracker.page() would automatically track
var pagePaths = []string{
	"/", "/about", "/contact", "/products", "/products/1", "/products/2",
	"/blog", "/blog/post-1", "/blog/post-2", "/pricing", "/features", "/docs",
	// Simulate some hash changes
	"/#section1", "/about#team", "/pricing#enterprise",
}

// Custom events that would require manual tracker.track() calls
var customEvents = map[string][]string{ // category -> event names
	"click":    {"cta_button", "nav_link", "footer_link", "product_image", "add_to_cart"},
	"form":     {"contact_submit", "newsletter_signup", "login_attempt", "search_query"},
	"video":    {"play", "pause", "complete"},
	"download": {"whitepaper", "datasheet"},
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Safari/605.1.15",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Linux; Android 13; SM-G991U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0",
}

// External referrers only, plus direct visits
var referrers = []string{
	"https://www.google.com/", "https://www.bing.com/", "https://duckduckgo.com/",
	"https://t.co/", "https://www.facebook.com/", "https://news.ycombinator.com/",
	"https://www.reddit.com/", "https://github.com/",
	"", "", "", "", "", // Higher chance of direct visit
}

// --- Identity Simulation ---
// Simulate localStorage persistence for a subset of users
var simulatedUserStorage = make(map[int]string) // Temporary "user id" -> persistent tracker id
var storageMutex sync.Mutex

const numSimulatedUsers = 500     // How many distinct users to simulate identities for
const chanceOfReturningUser = 0.3 // 30% chance an event comes from a user with a stored ID

func getSimulatedIdentity() string {
	if rand.Float64() > chanceOfReturningUser {
		return "" // New user or user without stored ID
	}

	// Simulate a returning user
	tempUserID := rand.Intn(numSimulatedUsers)

	storageMutex.Lock()
	defer storageMutex.Unlock()

	persistentID, exists := simulatedUserStorage[tempUserID]
	if !exists {
		// First time we see this simulated user, give them a persistent ID
		persistentID = fmt.Sprintf("uid-%d-%d", tempUserID, rand.Intn(1000000))
		simulatedUserStorage[tempUserID] = persistentID
	}
	return persistentID
}

// --- ---

func randomElement(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	return slice[rand.Intn(len(slice))]
}

func generateEvent() Tracking {
	site := randomElement(sites)
	identity := getSimulatedIdentity() // Get potentially empty or persistent ID
	userAgent := randomElement(userAgents)
	isTouch := strings.Contains(userAgent, "iPhone") || strings.Contains(userAgent, "Android") || rand.Intn(20) == 0 // Simple UA check + random chance
	referrer := randomElement(referrers)

	var eventType, eventName, category string

	// Decide whether to generate a page view (automatic) or a custom event (manual)
	// Make page views much more common (e.g., 85% chance)
	if rand.Intn(100) < 85 {
		// --- Generate a Page View event (like tracker.page()) ---
		eventType = "page"
		category = "Page views"
		eventName = randomElement(pagePaths)
	} else {
		// --- Generate a Custom Event (like tracker.track()) ---
		eventType = "event"
		// Pick a random custom category
		customCategories := make([]string, 0, len(customEvents))
		for cat := range customEvents {
			customCategories = append(customCategories, cat)
		}
		if len(customCategories) == 0 { // Fallback if customEvents is empty
			category = "general"
			eventName = "fallback_event"
		} else {
			category = randomElement(customCategories)
			possibleEvents := customEvents[category]
			if len(possibleEvents) > 0 {
				eventName = randomElement(possibleEvents)
			} else {
				eventName = category + "_fallback" // Fallback event name
			}
		}
	}

	return Tracking{
		SiteID: site,
		Action: TrackingData{
			Type:          eventType,
			Identity:      identity,
			UserAgent:     userAgent,
			Event:         eventName,
			Category:      category,
			Referrer:      referrer,
			IsTouchDevice: isTouch,
		},
	}
}

// --- ---

func main() {
	// --- Logger Setup ---
	logLevel := new(slog.LevelVar) // Default to Info
	logLevel.Set(slog.LevelInfo)
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)
	// --- ---

	// --- Flags ---
	trackerURL := flag.String("url", "http://localhost:9876/track", "Tracker /track endpoint URL")
	numEvents := flag.Int("n", 500, "Number of events to send")                // Increased default
	delayMs := flag.Int("delay", 50, "Delay between requests in milliseconds") // Decreased default
	verbose := flag.Bool("v", false, "Enable debug logging")
	flag.Parse()

	if *verbose {
		logLevel.Set(slog.LevelDebug)
	}
	// --- ---

	rand.New(rand.NewSource(time.Now().UnixNano())) // Seed random number generator

	logger.Info("Starting data generator",
		slog.String("targetUrl", *trackerURL),
		slog.Int("count", *numEvents),
		slog.Int("delayMs", *delayMs))

	client := &http.Client{
		Timeout: 5 * time.Second, // Add a timeout to requests
	}
	delay := time.Duration(*delayMs) * time.Millisecond
	successCount := 0
	errorCount := 0

	for i := range *numEvents {
		event := generateEvent()

		jsonData, err := json.Marshal(event)
		if err != nil {
			logger.Error("Failed to marshal event to JSON", slog.Any("error", err), slog.Int("eventIndex", i))
			errorCount++
			continue // Skip this event
		}

		req, err := http.NewRequest("POST", *trackerURL, strings.NewReader(string(jsonData)))
		if err != nil {
			logger.Error("Failed to create HTTP request", slog.Any("error", err), slog.Int("eventIndex", i))
			errorCount++
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Failed to send request to tracker", slog.Any("error", err), slog.Int("eventIndex", i), slog.String("url", *trackerURL))
			errorCount++
		} else {
			logAttrs := []slog.Attr{
				slog.Int("eventIndex", i),
				slog.String("siteId", event.SiteID),
				slog.String("type", event.Action.Type),
				slog.String("category", event.Action.Category),
				slog.String("event", event.Action.Event),
				slog.String("identity", event.Action.Identity), // Log identity being sent
			}
			if resp.StatusCode == http.StatusOK {
				logger.LogAttrs(nil, slog.LevelDebug, "Event sent successfully", logAttrs...)
				successCount++
			} else {
				logAttrs = append(logAttrs, slog.Int("statusCode", resp.StatusCode))
				logger.LogAttrs(nil, slog.LevelWarn, "Tracker responded with non-OK status", logAttrs...)
				errorCount++
			}
			resp.Body.Close()
		}

		if i < *numEvents-1 && delay > 0 {
			time.Sleep(delay)
		}
	}

	logger.Info("Data generation complete",
		slog.Int("successCount", successCount),
		slog.Int("errorCount", errorCount))

	if errorCount > 0 {
		os.Exit(1)
	}
}
