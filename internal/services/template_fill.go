package services

import (
	"context"
	"fmt"
	"html"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

var placeholderAlias = map[string][]string{
	"firstname":              {"prenom", "personal_info.firstname"},
	"lastname":               {"nom", "personal_info.lastname"},
	"title":                  {"title", "poste"},
	"age":                    {"age", "personal_info.age"},
	"availability":           {"availability", "personal_info.availability"},
	"mobilityareas":          {"mobilite", "personal_info.mobilite", "mobilite.zones"},
	"activityareas":          {"secteurs_activites[]", "secteurs_activites", "domaines_expertise[]"},
	"expertiseareas":         {"domaines_expertise[]", "domaines_expertise", "technical_skills[]"},
	"diplomas":               {"diplome_principal", "diplome", "education[].degree"},
	"languages.language":     {"languages[].language", "languages[]"},
	"languages.level":        {"languages[].level"},
	"tools.tool":             {"technical_skills_full_list[]", "technical_skills[]", "logiciels_full_list[]", "logiciels[]"},
	"references.company":     {"experiences[].company", "experiences[].employer"},
	"references.title":       {"experiences[].title", "experiences[].poste"},
	"references.startdate":   {"experiences[].start_date", "experiences[].startDate", "experiences[].date_debut"},
	"references.enddate":     {"experiences[].end_date", "experiences[].endDate", "experiences[].date_fin"},
	"references.duration":    {"experiences[].duration"},
	"references.description": {"experiences[].description", "experiences[].summary"},
	"summary":                {"summary", "personal_info.summary"},
}

type TemplateFillService struct{}

func NewTemplateFillService() (*TemplateFillService, error) {
	// Keep compatibility with existing callers; no external dependencies required anymore.
	if strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")) == "" {
		log.Printf("[TemplateFill] ANTHROPIC_API_KEY not set, running in deterministic mode")
	}
	return &TemplateFillService{}, nil
}

func (s *TemplateFillService) FillTemplate(ctx context.Context, sanitizedHTML string, variableDelimiter string, candidateData map[string]any) (string, error) {
	_ = ctx
	prefix, suffix := parseDelimiter(variableDelimiter)
	log.Printf("[TemplateFill] Starting fill (deterministic): html_len=%d delimiter=%s prefix=%q suffix=%q candidate_keys=%v", len(sanitizedHTML), variableDelimiter, prefix, suffix, keysFromMap(candidateData))

	decodedHTML := html.UnescapeString(sanitizedHTML)
	log.Printf("[TemplateFill] HTML decoded? contains '&#36;': %v", strings.Contains(sanitizedHTML, "&#36;"))

	placeholderRe := buildPlaceholderRegex(prefix, suffix)
	rawPlaceholders := placeholderRe.FindAllString(decodedHTML, -1)
	if len(rawPlaceholders) == 0 {
		log.Printf("[TemplateFill] No placeholders detected in template; returning original HTML")
		return sanitizedHTML, nil
	}

	uniquePlaceholders := dedupeStrings(rawPlaceholders)
	flatData := flattenData(candidateData)

	replacements := make(map[string]string, len(uniquePlaceholders))
	for _, placeholder := range uniquePlaceholders {
		token := stripPlaceholder(placeholder, prefix, suffix)
		values := collectValuesForToken(token, flatData)
		if len(values) == 0 {
			log.Printf("[TemplateFill] No values found for token=%s", token)
		} else {
			log.Printf("[TemplateFill] Values found for token=%s -> %v", token, values)
		}
		formatted := formatValues(values)
		replacements[placeholder] = formatted
		log.Printf("[TemplateFill] Replacement mapping: %s -> %q", placeholder, formatted)
	}

	filled := decodedHTML
	for placeholder, value := range replacements {
		filled = strings.ReplaceAll(filled, placeholder, value)
	}

	return filled, nil
}

func keysFromMap(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func dedupeStrings(items []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

func flattenData(data interface{}) map[string]string {
	result := make(map[string]string)
	flattenDataRecursive("", data, result)
	return result
}

func flattenDataRecursive(prefix string, value interface{}, out map[string]string) {
	switch v := value.(type) {
	case map[string]any:
		for key, child := range v {
			next := key
			if prefix != "" {
				next = prefix + "." + key
			}
			flattenDataRecursive(next, child, out)
		}
	case []any:
		for idx, child := range v {
			next := fmt.Sprintf("%s[%d]", prefix, idx)
			flattenDataRecursive(next, child, out)
		}
	default:
		if prefix != "" {
			out[prefix] = fmt.Sprintf("%v", v)
		}
	}
}

func collectValuesForToken(token string, flat map[string]string) []string {
	tokenLower := strings.ToLower(token)

	log.Printf("[TemplateFill] collectValuesForToken token=%s flat_size=%d", token, len(flat))

	if paths, ok := placeholderAlias[tokenLower]; ok {
		log.Printf("[TemplateFill] Using alias paths %v", paths)
		if values := collectValuesForPaths(paths, flat); len(values) > 0 {
			return values
		}
	}

	// Heuristic fallback: look for keys containing the token
	candidates := make([]string, 0)
	for key, value := range flat {
		kLower := strings.ToLower(key)
		if kLower == tokenLower ||
			strings.HasSuffix(kLower, "."+tokenLower) ||
			strings.Contains(kLower, tokenLower) {
			log.Printf("[TemplateFill] Heuristic match token=%s key=%s value=%q", token, key, value)
			candidates = append(candidates, value)
		}
	}
	if len(candidates) > 0 {
		return candidates
	}

	log.Printf("[TemplateFill] Fallback empty for token=%s", token)
	return []string{""}
}

func collectValuesForPaths(paths []string, flat map[string]string) []string {
	collected := make([]string, 0)
	for _, pattern := range paths {
		values := collectValuesForPattern(pattern, flat)
		if len(values) > 0 {
			collected = append(collected, values...)
		}
	}
	return collected
}

func collectValuesForPattern(pattern string, flat map[string]string) []string {
	if pattern == "" {
		return nil
	}
	lowerPattern := strings.ToLower(pattern)
	log.Printf("[TemplateFill] collectValuesForPattern pattern=%s flat_size=%d", pattern, len(flat))

	results := make([]string, 0)
	if strings.Contains(pattern, "[]") {
		parts := strings.Split(pattern, "[]")
		prefix := strings.ToLower(parts[0])
		suffix := ""
		if len(parts) > 1 {
			suffix = strings.ToLower(parts[1])
		}

		type pair struct {
			key   string
			value string
		}
		matches := make([]pair, 0)
		for key, value := range flat {
			kLower := strings.ToLower(key)
			if !strings.HasPrefix(kLower, prefix) {
				log.Printf("[TemplateFill] skip (prefix) key=%s", key)
				continue
			}
			if !strings.Contains(kLower, "[") || !strings.Contains(kLower, "]") {
				log.Printf("[TemplateFill] skip (no index) key=%s", key)
				continue
			}
			if suffix != "" && !strings.HasSuffix(kLower, suffix) {
				log.Printf("[TemplateFill] skip (suffix) key=%s", key)
				continue
			}
			matches = append(matches, pair{key, value})
		}
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].key < matches[j].key
		})
		for _, match := range matches {
			results = append(results, match.value)
		}
	} else {
		for key, value := range flat {
			kLower := strings.ToLower(key)
			if kLower == lowerPattern ||
				strings.HasSuffix(kLower, "."+lowerPattern) ||
				strings.Contains(kLower, lowerPattern) {
				log.Printf("[TemplateFill] scalar match pattern=%s key=%s value=%q", pattern, key, value)
				results = append(results, value)
			}
		}
	}
	if len(results) == 0 {
		log.Printf("[TemplateFill] no matches for pattern=%s", pattern)
	}
	return results
}

func formatValues(values []string) string {
	if len(values) == 0 {
		return ""
	}

	unique := make([]string, 0, len(values))
	seen := make(map[string]struct{})
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, exists := seen[v]; exists {
			continue
		}
		seen[v] = struct{}{}
		unique = append(unique, v)
	}

	if len(unique) == 0 {
		return ""
	}

	if len(unique) == 1 {
		return unique[0]
	}
	return strings.Join(unique, "<br/>")
}

func buildPlaceholderRegex(prefix, suffix string) *regexp.Regexp {
	if prefix == "" && suffix == "" {
		return regexp.MustCompile(`\$[A-Za-z0-9_.]+\$`)
	}
	pattern := fmt.Sprintf(`%s[A-Za-z0-9_.]+%s`, regexp.QuoteMeta(prefix), regexp.QuoteMeta(suffix))
	return regexp.MustCompile(pattern)
}

func stripPlaceholder(placeholder, prefix, suffix string) string {
	token := placeholder
	if prefix != "" && strings.HasPrefix(token, prefix) {
		token = strings.TrimPrefix(token, prefix)
	}
	if suffix != "" && strings.HasSuffix(token, suffix) {
		token = strings.TrimSuffix(token, suffix)
	}
	return token
}

func parseDelimiter(delimiter string) (string, string) {
	if delimiter == "" {
		return "$", "$"
	}
	lower := strings.ToLower(delimiter)
	idx := strings.Index(lower, "var")
	if idx >= 0 {
		prefix := delimiter[:idx]
		suffix := delimiter[idx+3:]
		if prefix == "" && suffix == "" {
			return "$", "$"
		}
		return prefix, suffix
	}
	// Treat as same prefix/suffix
	return delimiter, delimiter
}
