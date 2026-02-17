package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var (
	indicatorTypes = []string{"ip", "domain", "url", "hash"}
	severities     = []string{"low", "medium", "high", "critical"}
	motivations    = []string{"financial", "espionage", "hacktivism", "destruction", "unknown"}
	countries      = []string{"CN", "RU", "IR", "KP", "US", "UA", "IL", "IN", "PK", "BR"}
	sources        = []string{"internal", "osint", "partner", "honeypot", "sandbox"}

	actorNames = []string{
		"APT-Dragon", "CyberBear", "SandWorm", "Lazarus", "Equation",
		"Turla", "FancyBear", "CozyBear", "DarkHotel", "Carbanak",
		"APT-Phoenix", "IceWolf", "ShadowNet", "GhostSpider", "CrimsonViper",
		"StealthHawk", "NightOwl", "SilentStorm", "BlackMamba", "RedScorpion",
		"CyberLynx", "DigitalPanther", "NeonSerpent", "IronMantis", "GoldenEagle",
		"SilverFox", "CobaltSpider", "OnyxPhoenix", "JadeRaven", "SapphireWolf",
		"RubyDragon", "EmeraldViper", "DiamondBear", "PlatinumHawk", "BronzeOwl",
		"CopperStorm", "TitaniumMamba", "CarbonScorpion", "SteelLynx", "ChromePanther",
		"ZincSerpent", "NickelMantis", "IronEagle", "LeadFox", "TinSpider",
		"AluminumPhoenix", "MagnesiumRaven", "CalciumWolf", "PotassiumDragon", "SodiumViper",
	}

	campaignNames = []string{
		"Operation ShadowNet", "Project Nightfall", "Campaign Tempest", "Mission Darkstar",
		"Operation CyberStorm", "Project IceBreaker", "Campaign Firewall", "Mission Blackout",
		"Operation RedSky", "Project GhostNet", "Campaign Stealthwind", "Mission Thunderbolt",
		"Operation NightHawk", "Project Venomstrike", "Campaign Frostbite", "Mission Ironclad",
		"Operation Specter", "Project Avalanche", "Campaign Blizzard", "Mission Typhoon",
	}

	domainTLDs     = []string{".com", ".net", ".org", ".io", ".co", ".xyz", ".info", ".biz", ".ru", ".cn"}
	domainPrefixes = []string{"malware", "phishing", "c2", "exfil", "payload", "dropper", "loader", "beacon", "rat", "backdoor"}
)

func main() {
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=threat_intel sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database. Starting seed...")

	actorIDs, err := seedThreatActors(db, 50)
	if err != nil {
		log.Fatalf("Failed to seed threat actors: %v", err)
	}
	log.Printf("Created %d threat actors", len(actorIDs))

	campaignIDs, err := seedCampaigns(db, 100, actorIDs)
	if err != nil {
		log.Fatalf("Failed to seed campaigns: %v", err)
	}
	log.Printf("Created %d campaigns", len(campaignIDs))

	indicatorIDs, err := seedIndicators(db, 10000)
	if err != nil {
		log.Fatalf("Failed to seed indicators: %v", err)
	}
	log.Printf("Created %d indicators", len(indicatorIDs))

	if err := seedIndicatorCampaigns(db, indicatorIDs, campaignIDs); err != nil {
		log.Fatalf("Failed to seed indicator-campaign relationships: %v", err)
	}
	log.Println("Created indicator-campaign relationships")

	if err := seedIndicatorActors(db, indicatorIDs, actorIDs); err != nil {
		log.Fatalf("Failed to seed indicator-actor relationships: %v", err)
	}
	log.Println("Created indicator-actor relationships")

	log.Println("Seed completed successfully!")
}

func seedThreatActors(db *sql.DB, count int) ([]string, error) {
	var ids []string

	for i := 0; i < count; i++ {
		id := uuid.New().String()
		name := actorNames[i%len(actorNames)]
		if i >= len(actorNames) {
			name = fmt.Sprintf("%s-%d", name, i/len(actorNames))
		}

		firstSeen := randomTime(365 * 3)
		lastSeen := randomTimeBetween(firstSeen, time.Now())

		_, err := db.Exec(`
			INSERT INTO threat_actors (id, name, description, country, motivation, first_seen, last_seen, confidence_level)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (name) DO NOTHING
		`,
			id,
			name,
			fmt.Sprintf("Threat actor group known as %s, active in cyber operations", name),
			randomChoice(countries),
			randomChoice(motivations),
			firstSeen,
			lastSeen,
			rand.Intn(100),
		)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func seedCampaigns(db *sql.DB, count int, actorIDs []string) ([]string, error) {
	var ids []string
	statuses := []string{"active", "inactive", "historical"}

	for i := 0; i < count; i++ {
		id := uuid.New().String()
		name := campaignNames[i%len(campaignNames)]
		if i >= len(campaignNames) {
			name = fmt.Sprintf("%s Phase %d", name, i/len(campaignNames)+1)
		}

		startDate := randomTime(365 * 2)
		var endDate *time.Time
		status := randomChoice(statuses)
		if status != "active" {
			end := randomTimeBetween(startDate, time.Now())
			endDate = &end
		}

		_, err := db.Exec(`
			INSERT INTO campaigns (id, name, description, status, start_date, end_date, threat_actor_id, severity)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`,
			id,
			name,
			fmt.Sprintf("Campaign %s targeting various sectors", name),
			status,
			startDate,
			endDate,
			randomChoice(actorIDs),
			randomChoice(severities),
		)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func seedIndicators(db *sql.DB, count int) ([]string, error) {
	var ids []string

	for i := 0; i < count; i++ {
		id := uuid.New().String()
		indType := randomChoice(indicatorTypes)
		value := generateIndicatorValue(indType)

		firstSeen := randomTime(365)
		lastSeen := randomTimeBetween(firstSeen, time.Now())

		_, err := db.Exec(`
			INSERT INTO indicators (id, type, value, description, severity, confidence, first_seen, last_seen, is_active, source)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`,
			id,
			indType,
			value,
			fmt.Sprintf("Malicious %s indicator observed in the wild", indType),
			randomChoice(severities),
			rand.Intn(100),
			firstSeen,
			lastSeen,
			rand.Float32() > 0.2,
			randomChoice(sources),
		)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)

		if (i+1)%1000 == 0 {
			log.Printf("Created %d indicators...", i+1)
		}
	}

	return ids, nil
}

func seedIndicatorCampaigns(db *sql.DB, indicatorIDs, campaignIDs []string) error {
	for _, indID := range indicatorIDs {
		numCampaigns := rand.Intn(3) + 1
		usedCampaigns := make(map[string]bool)

		for j := 0; j < numCampaigns; j++ {
			campID := randomChoice(campaignIDs)
			if usedCampaigns[campID] {
				continue
			}
			usedCampaigns[campID] = true

			_, err := db.Exec(`
				INSERT INTO indicator_campaigns (indicator_id, campaign_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, indID, campID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func seedIndicatorActors(db *sql.DB, indicatorIDs, actorIDs []string) error {
	for _, indID := range indicatorIDs {
		if rand.Float32() > 0.6 {
			continue
		}

		numActors := rand.Intn(2) + 1
		usedActors := make(map[string]bool)

		for j := 0; j < numActors; j++ {
			actorID := randomChoice(actorIDs)
			if usedActors[actorID] {
				continue
			}
			usedActors[actorID] = true

			_, err := db.Exec(`
				INSERT INTO indicator_actors (indicator_id, actor_id, attribution_confidence)
				VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, indID, actorID, rand.Intn(100))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func generateIndicatorValue(indType string) string {
	switch indType {
	case "ip":
		return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
	case "domain":
		prefix := randomChoice(domainPrefixes)
		tld := randomChoice(domainTLDs)
		return fmt.Sprintf("%s%d%s", prefix, rand.Intn(10000), tld)
	case "url":
		domain := generateIndicatorValue("domain")
		paths := []string{"/payload", "/download", "/update", "/api", "/data", "/file"}
		return fmt.Sprintf("https://%s%s/%s", domain, randomChoice(paths), uuid.New().String()[:8])
	case "hash":
		chars := "abcdef0123456789"
		hash := make([]byte, 64)
		for i := range hash {
			hash[i] = chars[rand.Intn(len(chars))]
		}
		return string(hash)
	}
	return ""
}

func randomChoice[T any](slice []T) T {
	return slice[rand.Intn(len(slice))]
}

func randomTime(maxDaysAgo int) time.Time {
	daysAgo := rand.Intn(maxDaysAgo)
	return time.Now().AddDate(0, 0, -daysAgo)
}

func randomTimeBetween(start, end time.Time) time.Time {
	delta := end.Sub(start)
	randomDelta := time.Duration(rand.Int63n(int64(delta)))
	return start.Add(randomDelta)
}
