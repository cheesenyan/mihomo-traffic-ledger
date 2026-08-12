package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

func observedSessionKey(conn connection, nowMS int64, existing string) string {
	if strings.TrimSpace(conn.Start) != "" {
		return conn.ID + "|" + conn.Start
	}
	if existing != "" {
		return existing
	}
	return fmt.Sprintf("%s|%d", conn.ID, nowMS)
}

func connectionStartedAt(conn connection, fallback int64) int64 {
	if parsed, err := time.Parse(time.RFC3339Nano, conn.Start); err == nil {
		return parsed.UnixMilli()
	}
	return fallback
}

func (s *service) persistConnectionSnapshot(now time.Time, payload *connectionsResponse) ([]trafficLog, error) {
	nowMS := now.UnixMilli()
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lastConnections == nil {
		s.lastConnections = make(map[string]connection)
	}
	if s.activeSessionKeys == nil {
		s.activeSessionKeys = make(map[string]string)
	}

	reset := payload.UploadTotal < s.lastUploadTotal || payload.DownloadTotal < s.lastDownloadTotal
	if reset {
		logCounterReset()
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if reset {
		if err := closeSessionKeys(tx, s.activeSessionKeys, nowMS); err != nil {
			return nil, err
		}
	}

	previous := s.lastConnections
	previousKeys := s.activeSessionKeys
	if reset {
		previous = map[string]connection{}
		previousKeys = map[string]string{}
	}

	current := make(map[string]connection, len(payload.Connections))
	currentKeys := make(map[string]string, len(payload.Connections))
	logs := make([]trafficLog, 0, len(payload.Connections))

	for _, conn := range payload.Connections {
		prev, hasPrev := previous[conn.ID]
		existingKey := previousKeys[conn.ID]
		key := observedSessionKey(conn, nowMS, existingKey)
		if hasPrev && existingKey != key {
			if _, err := tx.Exec(`UPDATE connection_sessions SET ended_at=? WHERE session_key=? AND ended_at IS NULL`, nowMS, existingKey); err != nil {
				return nil, err
			}
			hasPrev = false
		}

		uploadDelta, downloadDelta := conn.Upload, conn.Download
		if hasPrev {
			uploadDelta = conn.Upload - prev.Upload
			downloadDelta = conn.Download - prev.Download
		}
		if uploadDelta < 0 {
			uploadDelta = conn.Upload
		}
		if downloadDelta < 0 {
			downloadDelta = conn.Download
		}

		chains := sanitizeChains(conn.Chains)
		chainsJSON, err := json.Marshal(chains)
		if err != nil {
			return nil, err
		}
		process := defaultString(conn.Metadata.Process, "Unknown")
		host := defaultString(firstNonEmpty(conn.Metadata.Host, conn.Metadata.DestinationIP), "Unknown")
		outbound := outboundName(chains)
		route := routeType(chains)
		startedAt := connectionStartedAt(conn, nowMS)

		_, err = tx.Exec(`
			INSERT INTO connection_sessions
			(session_key, connection_id, mihomo_started_at, started_at, first_seen_at, last_seen_at, ended_at,
			 network, connection_type, source_ip, source_port, destination_ip, destination_port, host,
			 process, process_path, route_type, outbound, chains, rule, rule_payload, upload, download)
			VALUES (?, ?, ?, ?, ?, ?, NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(session_key) DO UPDATE SET
			 last_seen_at=excluded.last_seen_at, ended_at=NULL, network=excluded.network,
			 connection_type=excluded.connection_type, source_ip=excluded.source_ip, source_port=excluded.source_port,
			 destination_ip=excluded.destination_ip, destination_port=excluded.destination_port, host=excluded.host,
			 process=excluded.process, process_path=excluded.process_path, route_type=excluded.route_type,
			 outbound=excluded.outbound, chains=excluded.chains, rule=excluded.rule,
			 rule_payload=excluded.rule_payload, upload=excluded.upload, download=excluded.download
		`, key, conn.ID, conn.Start, startedAt, nowMS, nowMS,
			strings.TrimSpace(conn.Network), strings.TrimSpace(conn.ConnType), defaultString(conn.Metadata.SourceIP, "Inner"), strings.TrimSpace(conn.SourcePort),
			strings.TrimSpace(conn.Metadata.DestinationIP), strings.TrimSpace(conn.DestinationPort), host,
			process, strings.TrimSpace(conn.ProcessPath), route, outbound, string(chainsJSON), strings.TrimSpace(conn.Rule), strings.TrimSpace(conn.RulePayload),
			conn.Upload, conn.Download)
		if err != nil {
			return nil, err
		}

		if uploadDelta != 0 || downloadDelta != 0 {
			logs = append(logs, trafficLog{
				Timestamp: nowMS, SourceIP: defaultString(conn.Metadata.SourceIP, "Inner"), Host: host,
				DestinationIP: strings.TrimSpace(conn.Metadata.DestinationIP), Process: process,
				ProcessPath: strings.TrimSpace(conn.ProcessPath), RouteType: route, Outbound: outbound,
				Chains: chains, Rule: strings.TrimSpace(conn.Rule), RulePayload: strings.TrimSpace(conn.RulePayload),
				Upload: uploadDelta, Download: downloadDelta,
			})
		}
		current[conn.ID] = conn
		currentKeys[conn.ID] = key
	}

	for id, key := range previousKeys {
		if _, ok := currentKeys[id]; ok {
			continue
		}
		if _, err := tx.Exec(`UPDATE connection_sessions SET ended_at=? WHERE session_key=? AND ended_at IS NULL`, nowMS, key); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	s.lastConnections = current
	s.activeSessionKeys = currentKeys
	s.lastUploadTotal = payload.UploadTotal
	s.lastDownloadTotal = payload.DownloadTotal
	return logs, nil
}

func closeSessionKeys(tx *sql.Tx, keys map[string]string, endedAt int64) error {
	for _, key := range keys {
		if _, err := tx.Exec(`UPDATE connection_sessions SET ended_at=? WHERE session_key=? AND ended_at IS NULL`, endedAt, key); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) closeObservedSessions(now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := closeSessionKeys(tx, s.activeSessionKeys, now.UnixMilli()); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.lastConnections = make(map[string]connection)
	s.activeSessionKeys = make(map[string]string)
	return nil
}

var logCounterReset = func() {
	// Kept as a seam for tests and future structured status reporting.
	// The standard logger call stays here so credentials and payloads are never included.
	log.Printf("detected Mihomo counter reset; starting fresh baselines")
}
