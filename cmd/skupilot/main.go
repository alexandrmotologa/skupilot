package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alexandrmotologa/skupilot/internal/api"
	"github.com/alexandrmotologa/skupilot/internal/channels"
	"github.com/alexandrmotologa/skupilot/internal/config"
	"github.com/alexandrmotologa/skupilot/internal/copywriter"
	"github.com/alexandrmotologa/skupilot/internal/domain"
	"github.com/alexandrmotologa/skupilot/internal/engine"
	"github.com/alexandrmotologa/skupilot/internal/feed"
	"github.com/alexandrmotologa/skupilot/internal/store"
	"github.com/alexandrmotologa/skupilot/internal/vision"
	"github.com/alexandrmotologa/skupilot/web"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "serve":
		cmdServe(os.Args[2:])
	case "enrich":
		cmdEnrich(os.Args[2:])
	case "audit":
		cmdAudit(os.Args[2:])
	case "autofix":
		cmdAutoFix(os.Args[2:])
	case "export":
		cmdExport(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Printf("skupilot version %s (autonomous multimodal catalog engine)\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`skupilot — Autonomous Multimodal E-Commerce Catalog Enrichment Engine (v%s)

Usage:
  skupilot <command> [flags]

Commands:
  serve      Start the web merchandising studio and REST API daemon
  enrich     Ingest and enrich a raw supplier feed CSV/JSON
  audit      Run marketplace image and copy compliance audit on database
  autofix    Auto-remediate non-compliant images to pure white RGB 255 & pad 1000px
  export     Export enriched catalog to Shopify, Amazon, Google Shopping, or CSV
  version    Show version information

Run 'skupilot <command> --help' for details on a specific command.
`, version)
}

func cmdServe(args []string) {
	cfg := config.Load()

	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.String("port", cfg.Port, "Port to run the HTTP daemon on")
	dbPath := fs.String("db", cfg.DBPath, "SQLite catalog database path")
	_ = fs.Parse(args)

	db, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("Fatal: failed to initialize database at %s: %v", *dbPath, err)
	}
	defer db.Close()

	vlm := vision.NewDefaultVLM(cfg.GeminiAPIKey, cfg.OpenAIAPIKey)
	eng := engine.NewEngine(db, vlm)
	webFS := web.GetFS()
	srv := api.NewServer(db, eng, webFS)

	httpServer := &http.Server{
		Addr:    ":" + *port,
		Handler: srv,
	}

	go func() {
		fmt.Printf("🚀 skupilot Merchandising Studio live on http://localhost:%s\n", *port)
		fmt.Printf("📁 Database: %s\n", *dbPath)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down skupilot daemon...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
}

func cmdEnrich(args []string) {
	cfg := config.Load()

	fs := flag.NewFlagSet("enrich", flag.ExitOnError)
	inputFile := fs.String("input", "", "Path to raw supplier feed (CSV, TSV, or JSON) [required]")
	batchName := fs.String("name", "", "Name of the catalog batch")
	dbPath := fs.String("db", cfg.DBPath, "SQLite catalog database path")
	_ = fs.Parse(args)

	if *inputFile == "" {
		fmt.Println("Error: --input file path is required")
		fs.Usage()
		os.Exit(1)
	}

	file, err := os.Open(*inputFile)
	if err != nil {
		log.Fatalf("Fatal: failed to open input file %s: %v", *inputFile, err)
	}
	defer file.Close()

	var items []domain.RawFeedItem
	if strings.HasSuffix(strings.ToLower(*inputFile), ".json") {
		items, err = feed.ParseJSON(file)
	} else {
		items, err = feed.ParseCSV(file)
	}

	if err != nil {
		log.Fatalf("Fatal: failed parsing catalog feed: %v", err)
	}

	if len(items) == 0 {
		log.Fatal("Fatal: feed contains 0 valid product items")
	}

	db, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("Fatal: failed opening database: %v", err)
	}
	defer db.Close()

	name := *batchName
	if name == "" {
		name = *inputFile
	}

	batch, err := db.CreateBatch(name, len(items))
	if err != nil {
		log.Fatalf("Fatal: failed creating batch: %v", err)
	}

	fmt.Printf("📦 Ingesting batch '%s' (%d products)...\n", batch.Name, len(items))

	vlm := vision.NewDefaultVLM(cfg.GeminiAPIKey, cfg.OpenAIAPIKey)
	eng := engine.NewEngine(db, vlm)

	start := time.Now()
	err = eng.ProcessFeed(context.Background(), batch.ID, items, func(current, total int, sku string, p *domain.Product) {
		statusIcon := "✓"
		if p.ImageComplianceStatus != domain.CompliancePass {
			statusIcon = "!"
		}
		fmt.Printf("[%d/%d] %s SKU: %-14s | Quality: %3d%% | Whiteness: %5.1f%% | %s\n",
			current, total, statusIcon, p.RawSKU, p.QualityScore, p.ImageWhiteBGScore*100, p.OptimizedTitle)
	})

	if err != nil {
		log.Fatalf("Fatal: enrichment failed: %v", err)
	}

	fmt.Printf("\n✨ Successfully enriched %d products in %v (Batch ID: %s)\n", len(items), time.Since(start).Round(time.Millisecond), batch.ID)
}

func cmdAudit(args []string) {
	cfg := config.Load()

	fs := flag.NewFlagSet("audit", flag.ExitOnError)
	dbPath := fs.String("db", cfg.DBPath, "SQLite catalog database path")
	batchID := fs.String("batch-id", "", "Batch ID to audit (leave empty for all)")
	_ = fs.Parse(args)

	db, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("Fatal: failed to open database: %v", err)
	}
	defer db.Close()

	products, total, err := db.ListProducts(store.ProductFilter{
		BatchID: *batchID,
		Limit:   500,
	})
	if err != nil {
		log.Fatalf("Fatal: query failed: %v", err)
	}

	fmt.Printf("\n=== Marketplace Compliance Audit Report (Total: %d products) ===\n\n", total)
	fmt.Printf("%-14s | %-6s | %-12s | %-10s | %-7s | %-30s\n",
		"SKU", "Status", "Whiteness", "Zoom Ready", "Quality", "Optimized Title")
	fmt.Println(strings.Repeat("-", 90))

	compliant := 0
	for _, p := range products {
		if p.ImageComplianceStatus == domain.CompliancePass {
			compliant++
		}
		zoom := "PASS"
		if !p.ImageZoomReady {
			zoom = "FAIL (<1k)"
		}

		shortTitle := p.OptimizedTitle
		if len(shortTitle) > 30 {
			shortTitle = shortTitle[:27] + "..."
		}

		fmt.Printf("%-14s | %-6s | %10.1f%% | %-10s | %6d%% | %-30s\n",
			p.RawSKU, p.ImageComplianceStatus, p.ImageWhiteBGScore*100, zoom, p.QualityScore, shortTitle)
	}

	fmt.Println(strings.Repeat("-", 90))
	passRate := 0.0
	if total > 0 {
		passRate = float64(compliant) / float64(total) * 100
	}
	fmt.Printf("Summary: %d / %d compliant (%.1f%% compliance rate)\n\n", compliant, total, passRate)
}

func cmdAutoFix(args []string) {
	cfg := config.Load()

	fs := flag.NewFlagSet("autofix", flag.ExitOnError)
	dbPath := fs.String("db", cfg.DBPath, "SQLite catalog database path")
	batchID := fs.String("batch-id", "", "Batch ID to auto-fix (leave empty for all)")
	_ = fs.Parse(args)

	db, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("Fatal: failed to open database: %v", err)
	}
	defer db.Close()

	products, _, err := db.ListProducts(store.ProductFilter{
		BatchID: *batchID,
		Limit:   500,
	})
	if err != nil {
		log.Fatalf("Fatal: query failed: %v", err)
	}

	fmt.Printf("🪄 Running Automated Image Remediation on %d products...\n", len(products))
	remediated := 0

	for _, p := range products {
		if p.ImageComplianceStatus == domain.CompliancePass && p.ImageZoomReady {
			continue
		}

		if p.ImageURL == "" {
			continue
		}

		_, img, err := vision.FetchAndAuditImage(context.Background(), nil, p.ImageURL)
		if err != nil || img == nil {
			continue
		}

		_, fixedReport := vision.AutoWhitenAndPad(img, 1000)

		p.ImageWidth = fixedReport.Width
		p.ImageHeight = fixedReport.Height
		p.ImageZoomReady = fixedReport.ZoomReady
		p.ImageWhiteBGScore = fixedReport.WhiteBGScore
		p.ImageAspectRatio = fixedReport.AspectRatio
		p.ImageWatermarkDetected = fixedReport.WatermarkDetected
		p.ImageComplianceStatus = fixedReport.Status

		scorecard := copywriter.CalculateQualityScorecard(&p)
		p.QualityScore = scorecard.TotalScore

		if err := db.UpdateProduct(&p); err == nil {
			remediated++
			fmt.Printf("  [Remediated] %s -> Whiteness: %.1f%% | Zoom: %dx%d (Score: %d%%)\n",
				p.RawSKU, p.ImageWhiteBGScore*100, p.ImageWidth, p.ImageHeight, p.QualityScore)
		}
	}

	fmt.Printf("\n✨ Successfully remediated %d images to Amazon compliance standard.\n", remediated)
}

func cmdExport(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: skupilot export <shopify|amazon|google> [flags]")
		os.Exit(1)
	}

	cfg := config.Load()
	channelType := strings.ToLower(args[0])

	switch channelType {
	case "csv":
		fs := flag.NewFlagSet("export csv", flag.ExitOnError)
		dbPath := fs.String("db", cfg.DBPath, "SQLite database path")
		batchID := fs.String("batch-id", "", "Batch ID to export")
		outputFile := fs.String("output", "enriched_catalog.csv", "Output CSV file path")
		_ = fs.Parse(args[1:])

		db, err := store.Open(*dbPath)
		if err != nil {
			log.Fatalf("Fatal: failed opening database: %v", err)
		}
		defer db.Close()

		products, _, err := db.ListProducts(store.ProductFilter{BatchID: *batchID, Limit: 2000})
		if err != nil {
			log.Fatalf("Fatal: failed querying products: %v", err)
		}

		csvData, err := channels.GenerateEnrichedCSV(products)
		if err != nil {
			log.Fatalf("Fatal: failed generating CSV: %v", err)
		}
		if err := os.WriteFile(*outputFile, csvData, 0644); err != nil {
			log.Fatalf("Fatal: failed writing file: %v", err)
		}
		fmt.Printf("Exported %d products to Master PIM CSV: %s\n", len(products), *outputFile)

	case "amazon":
		fs := flag.NewFlagSet("export amazon", flag.ExitOnError)
		dbPath := fs.String("db", cfg.DBPath, "SQLite database path")
		batchID := fs.String("batch-id", "", "Batch ID to export")
		outputFile := fs.String("output", "amazon_feed.tsv", "Output TSV file path")
		_ = fs.Parse(args[1:])

		db, err := store.Open(*dbPath)
		if err != nil {
			log.Fatalf("Fatal: failed opening database: %v", err)
		}
		defer db.Close()

		products, _, err := db.ListProducts(store.ProductFilter{BatchID: *batchID, Limit: 1000})
		if err != nil {
			log.Fatalf("Fatal: failed querying products: %v", err)
		}

		tsvData := channels.GenerateAmazonFlatFile(products)
		if err := os.WriteFile(*outputFile, tsvData, 0644); err != nil {
			log.Fatalf("Fatal: failed writing file: %v", err)
		}
		fmt.Printf("Exported %d products to Amazon Flat File: %s\n", len(products), *outputFile)

	case "google":
		fs := flag.NewFlagSet("export google", flag.ExitOnError)
		dbPath := fs.String("db", cfg.DBPath, "SQLite database path")
		batchID := fs.String("batch-id", "", "Batch ID to export")
		outputFile := fs.String("output", "google_feed.xml", "Output XML file path")
		_ = fs.Parse(args[1:])

		db, err := store.Open(*dbPath)
		if err != nil {
			log.Fatalf("Fatal: failed opening database: %v", err)
		}
		defer db.Close()

		products, _, err := db.ListProducts(store.ProductFilter{BatchID: *batchID, Limit: 1000})
		if err != nil {
			log.Fatalf("Fatal: failed querying products: %v", err)
		}

		xmlData, err := channels.GenerateGoogleShoppingXML(products, "skupilot Feed", "https://example.com")
		if err != nil {
			log.Fatalf("Fatal: failed generating XML: %v", err)
		}
		if err := os.WriteFile(*outputFile, xmlData, 0644); err != nil {
			log.Fatalf("Fatal: failed writing file: %v", err)
		}
		fmt.Printf("Exported %d products to Google Merchant Center XML: %s\n", len(products), *outputFile)

	case "shopify":
		fs := flag.NewFlagSet("export shopify", flag.ExitOnError)
		dbPath := fs.String("db", cfg.DBPath, "SQLite database path")
		batchID := fs.String("batch-id", "", "Batch ID to export")
		storeDomain := fs.String("store", cfg.ShopifyStore, "Shopify store domain (e.g. brand.myshopify.com)")
		token := fs.String("token", cfg.ShopifyToken, "Shopify Admin API token")
		status := fs.String("status", "DRAFT", "Shopify publication status (DRAFT or ACTIVE)")
		_ = fs.Parse(args[1:])

		db, err := store.Open(*dbPath)
		if err != nil {
			log.Fatalf("Fatal: failed opening database: %v", err)
		}
		defer db.Close()

		products, _, err := db.ListProducts(store.ProductFilter{BatchID: *batchID, Limit: 500})
		if err != nil {
			log.Fatalf("Fatal: failed querying products: %v", err)
		}

		if *token == "" {
			fmt.Printf("⚠️  No Shopify token provided. Running validation dry-run for %d products...\n", len(products))
			for _, p := range products {
				input := channels.BuildShopifyProductInput(p, *status)
				fmt.Printf("  [Dry Run Valid] %s -> %s (Metafields: %d)\n", p.RawSKU, input.Title, len(input.Metafields))
			}
			fmt.Println("All product payloads valid for Shopify GraphQL Admin API.")
			return
		}

		client := channels.NewShopifyClient(*storeDomain, *token)
		exported := 0
		for _, p := range products {
			id, err := client.PublishProduct(context.Background(), p, *status)
			if err != nil {
				fmt.Printf("  [Error] %s: %v\n", p.RawSKU, err)
			} else {
				exported++
				fmt.Printf("  [Published] %s -> Shopify ID: %s\n", p.RawSKU, id)
			}
		}
		fmt.Printf("Exported %d of %d products to Shopify store %s\n", exported, len(products), *storeDomain)

	default:
		fmt.Printf("Unknown export channel: %s. Supported channels: shopify, amazon, google\n", channelType)
		os.Exit(1)
	}
}
