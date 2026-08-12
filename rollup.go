package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

var ledgerGranularities = []string{"hour", "day", "week", "month"}

func rollupBounds(timestamp int64, granularity string) (int64, int64, error) {
	t := time.UnixMilli(timestamp).In(time.Local)
	var start, end time.Time
	switch granularity {
	case "hour":
		start = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
		end = start.Add(time.Hour)
	case "day":
		start = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		end = start.AddDate(0, 0, 1)
	case "week":
		start = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		daysSinceMonday := (int(start.Weekday()) + 6) % 7
		start = start.AddDate(0, 0, -daysSinceMonday)
		end = start.AddDate(0, 0, 7)
	case "month":
		start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		end = start.AddDate(0, 1, 0)
	default:
		return 0, 0, fmt.Errorf("unsupported rollup granularity %q", granularity)
	}
	return start.UnixMilli(), end.UnixMilli(), nil
}

func upsertTrafficRollups(tx *sql.Tx, entry aggregatedEntry) error {
	for _, grain := range ledgerGranularities {
		start, end, err := rollupBounds(entry.BucketStart, grain)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`
			INSERT INTO traffic_rollups
			(granularity, bucket_start, bucket_end, host, destination_ip, process, process_path,
			 route_type, outbound, chains, rule, rule_payload, upload, download, count)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(granularity, bucket_start, process, process_path, host, destination_ip,
			 route_type, outbound, chains, rule, rule_payload)
			DO UPDATE SET upload=traffic_rollups.upload+excluded.upload,
			 download=traffic_rollups.download+excluded.download, count=traffic_rollups.count+excluded.count
		`, grain, start, end, entry.Host, entry.DestinationIP, entry.Process, entry.ProcessPath,
			entry.RouteType, entry.Outbound, entry.Chains, entry.Rule, entry.RulePayload,
			entry.Upload, entry.Download, entry.Count)
		if err != nil {
			return err
		}
	}
	return nil
}

func rollupDimensionColumn(dimension string) (string, error) {
	switch dimension {
	case "process":
		return "process", nil
	case "outbound":
		return "outbound", nil
	case "host":
		return "host", nil
	case "routeType":
		return "route_type", nil
	default:
		return "", fmt.Errorf("unsupported rollup dimension %q", dimension)
	}
}

func (s *service) queryTrafficRollups(granularity, dimension string, start, end int64) ([]aggregatedData, error) {
	validGranularity := false
	for _, candidate := range ledgerGranularities {
		if granularity == candidate {
			validGranularity = true
			break
		}
	}
	if !validGranularity {
		return nil, fmt.Errorf("unsupported rollup granularity %q", granularity)
	}
	column, err := rollupDimensionColumn(dimension)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`SELECT %s, SUM(upload), SUM(download), SUM(upload + download), SUM(count)
		FROM traffic_rollups WHERE granularity=? AND bucket_start < ? AND bucket_end > ?
		GROUP BY %s ORDER BY SUM(upload + download) DESC`, column, column)
	rows, err := s.db.Query(query, granularity, end, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []aggregatedData{}
	for rows.Next() {
		var item aggregatedData
		if err := rows.Scan(&item.Label, &item.Upload, &item.Download, &item.Total, &item.Count); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (s *service) handleTrafficRollups(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	start, end, err := parseTimeRange(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	data, err := s.queryTrafficRollups(r.URL.Query().Get("period"), r.URL.Query().Get("dimension"), start, end)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}
