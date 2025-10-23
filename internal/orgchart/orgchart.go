package orgchart

import (
	"fmt"
	"sort"
	"strings"
)

// OrgNode représente un nœud de l'organigramme
type OrgNode struct {
	ID            string     `json:"id"`
	FirstName     string     `json:"firstName"`
	LastName      string     `json:"lastName"`
	Title         string     `json:"title"`
	TypeOf        int        `json:"typeOf"`
	MainManagerID string     `json:"mainManagerId,omitempty"`
	HRManagerID   string     `json:"hrManagerId,omitempty"`
	Subordinates  []*OrgNode `json:"subordinates,omitempty"`
}

// BuildOrgChartFromResources construit les arbres hiérarchiques et génère un graphe DOT
func BuildOrgChartFromResources(resources []map[string]any, includeHREdges bool) ([]*OrgNode, string, int) {
	// Index des nœuds par ID
	idToNode := make(map[string]*OrgNode)

	// 1) Hydrater tous les nœuds
	for _, r := range resources {
		id, _ := r["id"].(string)
		if strings.TrimSpace(id) == "" {
			continue
		}

		attrs, _ := r["attributes"].(map[string]any)
		rels, _ := r["relationships"].(map[string]any)

		firstName, _ := attrs["firstName"].(string)
		lastName, _ := attrs["lastName"].(string)
		title, _ := attrs["title"].(string)

		// typeOf est souvent décodé en float64 depuis JSON
		var typeOfInt int
		if v, ok := attrs["typeOf"]; ok {
			switch n := v.(type) {
			case float64:
				typeOfInt = int(n)
			case int:
				typeOfInt = n
			}
		}

		mainManagerID := extractRelationshipID(rels, "mainManager")
		hrManagerID := extractRelationshipID(rels, "hrManager")

		idToNode[id] = &OrgNode{
			ID:            id,
			FirstName:     firstName,
			LastName:      lastName,
			Title:         title,
			TypeOf:        typeOfInt,
			MainManagerID: mainManagerID,
			HRManagerID:   hrManagerID,
		}
	}

	// 2) Construire la hiérarchie (manager -> subordinates)
	var roots []*OrgNode
	orphanCount := 0
	for _, node := range idToNode {
		mgrID := strings.TrimSpace(node.MainManagerID)
		if mgrID == "" || idToNode[mgrID] == nil {
			// Pas de manager connu → racine
			roots = append(roots, node)
			continue
		}
		manager := idToNode[mgrID]
		manager.Subordinates = append(manager.Subordinates, node)
	}

	// Compter les orphelins: ressources ayant un manager inconnu (mgrID non vide mais absent de l'index)
	for _, node := range idToNode {
		if node.MainManagerID != "" {
			if _, ok := idToNode[node.MainManagerID]; !ok {
				orphanCount++
			}
		}
	}

	// Ordonner les enfants par nom pour stabilité visuelle
	var orderChildren func(n *OrgNode)
	orderChildren = func(n *OrgNode) {
		sort.SliceStable(n.Subordinates, func(i, j int) bool {
			li := strings.ToLower(n.Subordinates[i].LastName + " " + n.Subordinates[i].FirstName)
			lj := strings.ToLower(n.Subordinates[j].LastName + " " + n.Subordinates[j].FirstName)
			return li < lj
		})
		for _, c := range n.Subordinates {
			orderChildren(c)
		}
	}
	for _, r := range roots {
		orderChildren(r)
	}

	// 3) Générer le DOT
	dot := generateDOT(idToNode, roots, includeHREdges)
	return roots, dot, orphanCount
}

func extractRelationshipID(relationships map[string]any, relKey string) string {
	if relationships == nil {
		return ""
	}
	rel, _ := relationships[relKey].(map[string]any)
	if rel == nil {
		return ""
	}
	data := rel["data"]
	if data == nil {
		return ""
	}
	if m, ok := data.(map[string]any); ok {
		if id, ok := m["id"].(string); ok {
			return id
		}
	}
	return ""
}

func generateDOT(idToNode map[string]*OrgNode, roots []*OrgNode, includeHREdges bool) string {
	var b strings.Builder
	b.WriteString("digraph Org {\n")
	b.WriteString("\tgraph [rankdir=TB];\n")
	b.WriteString("\tnode [shape=box, style=rounded, fontsize=10, fontname=\"Helvetica\"];\n")

	// Déclarer tous les nœuds
	for _, node := range idToNode {
		label := fmt.Sprintf("%s %s\\n(id:%s)\\n%s\\ntype:%d", escape(node.FirstName), escape(node.LastName), escape(node.ID), escape(node.Title), node.TypeOf)
		b.WriteString(fmt.Sprintf("\t\"%s\" [label=\"%s\"];\n", escape(node.ID), label))
	}

	// Liens hiérarchiques (manager -> subordinate)
	for _, node := range idToNode {
		if node.MainManagerID != "" {
			if _, ok := idToNode[node.MainManagerID]; ok {
				b.WriteString(fmt.Sprintf("\t\"%s\" -> \"%s\";\n", escape(node.MainManagerID), escape(node.ID)))
			}
		}
	}

	// Liens RH en tirets
	if includeHREdges {
		for _, node := range idToNode {
			if node.HRManagerID != "" {
				if _, ok := idToNode[node.HRManagerID]; ok {
					b.WriteString(fmt.Sprintf("\t\"%s\" -> \"%s\" [style=dashed, color=gray50, label=\"HR\"];\n", escape(node.HRManagerID), escape(node.ID)))
				}
			}
		}
	}

	b.WriteString("}\n")
	return b.String()
}

func escape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
