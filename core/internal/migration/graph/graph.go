package graph

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/netpiedev/drift/core/internal/migration/parser"
)

type Node struct {
	ID    string   `json:"id"`
	Edges []string `json:"edges"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
}

func BuildMigrationGraph(migrations []parser.Migration) Graph {
	byVersion := map[string]struct{}{}
	versions := make([]string, 0)
	for _, m := range migrations {
		if m.Direction != parser.DirectionUp {
			continue
		}
		if _, ok := byVersion[m.Version]; ok {
			continue
		}
		byVersion[m.Version] = struct{}{}
		versions = append(versions, m.Version)
	}
	sort.Strings(versions)
	nodes := make([]Node, 0, len(versions))
	for i, v := range versions {
		node := Node{ID: v}
		if i > 0 {
			node.Edges = append(node.Edges, versions[i-1])
		}
		nodes = append(nodes, node)
	}
	return Graph{Nodes: nodes}
}

func BuildTableDependencyGraph(ctx context.Context, db *sql.DB) (Graph, error) {
	const q = `
SELECT tc.table_name, ccu.table_name AS foreign_table_name
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
  ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage AS ccu
  ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_schema='public';`
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return Graph{}, fmt.Errorf("query table dependencies: %w", err)
	}
	defer rows.Close()

	adj := map[string][]string{}
	for rows.Next() {
		var table string
		var dependsOn string
		if err := rows.Scan(&table, &dependsOn); err != nil {
			return Graph{}, fmt.Errorf("scan dependency row: %w", err)
		}
		adj[table] = append(adj[table], dependsOn)
		if _, ok := adj[dependsOn]; !ok {
			adj[dependsOn] = []string{}
		}
	}
	if err := rows.Err(); err != nil {
		return Graph{}, fmt.Errorf("iterate dependencies: %w", err)
	}

	nodes := make([]Node, 0, len(adj))
	for id, edges := range adj {
		nodes = append(nodes, Node{ID: id, Edges: edges})
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	return Graph{Nodes: nodes}, nil
}

func ToJSON(g Graph) (string, error) {
	b, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ToTerminal(g Graph) string {
	if len(g.Nodes) == 0 {
		return "(empty graph)"
	}
	out := ""
	for _, n := range g.Nodes {
		out += n.ID + "\n"
		for _, e := range n.Edges {
			out += "  -> " + e + "\n"
		}
	}
	return out
}
