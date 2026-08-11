package main

import (
	"database/sql"
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

type configFile struct {
	Mysql struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		DBName   string `yaml:"db-name"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"mysql"`
}

type routeInfo struct {
	Method string
	Path   string
}

func main() {
	cfg, err := readConfig(filepath.Join("configs", "config.yaml"))
	if err != nil {
		log.Fatal(err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Mysql.Username,
		cfg.Mysql.Password,
		cfg.Mysql.Host,
		cfg.Mysql.Port,
		cfg.Mysql.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	routes, err := collectCarpoolRoutes(filepath.Join("router", "carpool"))
	if err != nil {
		log.Fatal(err)
	}

	authorities, err := queryAuthorities(db)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("== authority summary ==")
	for _, line := range authorities {
		fmt.Println(line)
	}

	fmt.Println("\n== carpool route coverage ==")
	for _, authorityID := range []string{"888", "8881", "9528"} {
		missing, existing, err := diffPolicies(db, authorityID, routes)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("authority %s: carpool routes=%d, existing=%d, missing=%d\n", authorityID, len(routes), existing, len(missing))
		for _, item := range missing {
			fmt.Printf("  MISSING %s %s\n", item.Method, item.Path)
		}
	}
}

func readConfig(path string) (configFile, error) {
	var cfg configFile
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func queryAuthorities(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`
select u.id, u.username, u.authority_id, a.authority_name, a.default_router
from sys_users u
left join sys_authorities a on a.authority_id = u.authority_id
order by u.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var id int
		var username string
		var authorityID sql.NullInt64
		var authorityName, defaultRouter sql.NullString
		if err := rows.Scan(&id, &username, &authorityID, &authorityName, &defaultRouter); err != nil {
			return nil, err
		}
		lines = append(lines, fmt.Sprintf("user=%s id=%d authority=%d role=%s defaultRouter=%s",
			username, id, authorityID.Int64, authorityName.String, defaultRouter.String))
	}
	return lines, rows.Err()
}

func diffPolicies(db *sql.DB, authorityID string, routes []routeInfo) ([]routeInfo, int, error) {
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

	var missing []routeInfo
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

func collectCarpoolRoutes(dir string) ([]routeInfo, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	groupRE := regexp.MustCompile(`Router\.Group\("([^"]+)"\)`)
	routeRE := regexp.MustCompile(`\.(GET|POST|PUT|DELETE|PATCH)\("([^"]+)"`)
	seen := map[string]bool{}
	var routes []routeInfo

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
			routes = append(routes, routeInfo{Method: match[1], Path: path})
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
