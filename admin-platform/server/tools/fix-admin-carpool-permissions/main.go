package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v3"
)

type fixConfigFile struct {
	Mysql struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		DBName   string `yaml:"db-name"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"mysql"`
}

type fixRouteInfo struct {
	Method string
	Path   string
}

func main() {
	authorityFlag := flag.String("authority", "888", "comma-separated authority IDs to patch")
	apply := flag.Bool("apply", false, "write missing policies to casbin_rule")
	flag.Parse()

	authorityIDs := parseFixAuthorityIDs(*authorityFlag)
	if len(authorityIDs) == 0 {
		log.Fatal("authority cannot be empty")
	}

	cfg, err := readFixConfig(filepath.Join("configs", "config.yaml"))
	if err != nil {
		log.Fatal(err)
	}

	db, err := openFixDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	routes, err := collectFixCarpoolRoutes(filepath.Join("router", "carpool"))
	if err != nil {
		log.Fatal(err)
	}
	if len(routes) == 0 {
		log.Fatal("no carpool routes found")
	}

	mode := "DRY-RUN"
	if *apply {
		mode = "APPLY"
	}
	fmt.Printf("mode=%s authorities=%s carpool_routes=%d\n", mode, strings.Join(authorityIDs, ","), len(routes))

	totalInserted := 0
	for _, authorityID := range authorityIDs {
		missing, existing, err := diffFixPolicies(db, authorityID, routes)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("\nauthority %s: existing=%d missing=%d\n", authorityID, existing, len(missing))
		for _, route := range missing {
			fmt.Printf("  %s %s\n", route.Method, route.Path)
		}

		if !*apply || len(missing) == 0 {
			continue
		}

		inserted, err := insertFixPolicies(db, authorityID, missing)
		if err != nil {
			log.Fatal(err)
		}
		totalInserted += inserted
		fmt.Printf("  inserted=%d\n", inserted)
	}

	if !*apply {
		fmt.Println("\nNo database changes were made. Re-run with -apply to insert the missing policies.")
		return
	}
	fmt.Printf("\nInserted %d casbin policies. Restart the admin server or call /api/freshCasbin to reload Casbin.\n", totalInserted)
}

func parseFixAuthorityIDs(input string) []string {
	seen := map[string]bool{}
	var ids []string
	for _, item := range strings.Split(input, ",") {
		id := strings.TrimSpace(item)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids
}

func readFixConfig(path string) (fixConfigFile, error) {
	var cfg fixConfigFile
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func openFixDB(cfg fixConfigFile) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Mysql.Username,
		cfg.Mysql.Password,
		cfg.Mysql.Host,
		cfg.Mysql.Port,
		cfg.Mysql.DBName,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func diffFixPolicies(db *sql.DB, authorityID string, routes []fixRouteInfo) ([]fixRouteInfo, int, error) {
	rows, err := db.Query("select v1, v2 from casbin_rule where ptype = 'p' and v0 = ?", authorityID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	policies := map[string]bool{}
	for rows.Next() {
		var path, method string
		if err := rows.Scan(&path, &method); err != nil {
			return nil, 0, err
		}
		policies[method+" "+path] = true
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var missing []fixRouteInfo
	existing := 0
	for _, route := range routes {
		if policies[route.Method+" "+route.Path] {
			existing++
			continue
		}
		missing = append(missing, route)
	}
	return missing, existing, nil
}

func insertFixPolicies(db *sql.DB, authorityID string, routes []fixRouteInfo) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare("insert into casbin_rule (ptype, v0, v1, v2) values ('p', ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	inserted := 0
	for _, route := range routes {
		exists, err := fixPolicyExists(tx, authorityID, route)
		if err != nil {
			return 0, err
		}
		if exists {
			continue
		}
		if _, err := stmt.Exec(authorityID, route.Path, route.Method); err != nil {
			return 0, err
		}
		inserted++
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return inserted, nil
}

func fixPolicyExists(tx *sql.Tx, authorityID string, route fixRouteInfo) (bool, error) {
	var count int
	err := tx.QueryRow(
		"select count(*) from casbin_rule where ptype = 'p' and v0 = ? and v1 = ? and v2 = ?",
		authorityID,
		route.Path,
		route.Method,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func collectFixCarpoolRoutes(dir string) ([]fixRouteInfo, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	groupRE := regexp.MustCompile(`Router\.Group\("([^"]+)"\)`)
	routeRE := regexp.MustCompile(`\.(GET|POST|PUT|DELETE|PATCH)\("([^"]+)"`)
	seen := map[string]bool{}
	var routes []fixRouteInfo

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") || file.Name() == "enter.go" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}
		text := string(b)

		var base string
		if match := groupRE.FindStringSubmatch(text); len(match) == 2 {
			base = "/" + strings.Trim(match[1], "/")
		}
		if base == "" {
			continue
		}

		for _, match := range routeRE.FindAllStringSubmatch(text, -1) {
			path := base + "/" + strings.Trim(match[2], "/")
			key := match[1] + " " + path
			if seen[key] {
				continue
			}
			seen[key] = true
			routes = append(routes, fixRouteInfo{Method: match[1], Path: path})
		}
	}

	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})
	return routes, nil
}
