package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"backend-api-skillforge/internal/gladia"
)

type TranscriptUtterance struct {
	Speaker string   `json:"speaker"`
	Text    string   `json:"text"`
	Start   float64  `json:"start"`
	End     float64  `json:"end"`
	Words   []string `json:"words,omitempty"`
}

type TranscriptResponse struct {
	Transcripts []TranscriptUtterance `json:"transcripts"`
	Data        any                   `json:"data"`
}

// MeetingTranscript gère le flux: upload -> start -> poll -> normalisation -> réponse
func MeetingTranscript(c *gin.Context) {
	// Optionnel: restreindre par header secret
	if token := os.Getenv("WORKER_TOKEN"); token != "" {
		if c.GetHeader("X-Worker-Token") != token {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
	}

	// Lire multipart
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file manquant ou invalide", "detail": err.Error()})
		return
	}
	defer file.Close()

	optionsJSON := c.PostForm("options")
	var options map[string]any
	if strings.TrimSpace(optionsJSON) != "" {
		if err := json.Unmarshal([]byte(optionsJSON), &options); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "options JSON invalide", "detail": err.Error()})
			return
		}
	} else {
		options = map[string]any{}
	}

	client, err := gladia.NewClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Minute)
	defer cancel()

	// 1) Upload
	audioURL, resp, body, err := client.UploadAudio(ctx, header.Filename, file)
	if err != nil {
		status := http.StatusBadGateway
		if resp != nil && resp.StatusCode >= 400 {
			status = resp.StatusCode
		}
		c.Data(status, "application/json", body)
		return
	}

	// 2) Start transcription
	payload := make(map[string]any, len(options)+1)
	for k, v := range options {
		payload[k] = v
	}
	payload["audio_url"] = audioURL

	resultURL, resp2, body2, err := client.StartTranscription(ctx, payload)
	if err != nil {
		status := http.StatusBadGateway
		if resp2 != nil && resp2.StatusCode >= 400 {
			status = resp2.StatusCode
		}
		c.Data(status, "application/json", body2)
		return
	}

	// 3) Polling
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var finalJSON map[string]any
	for {
		select {
		case <-ctx.Done():
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "timeout polling Gladia"})
			return
		case <-ticker.C:
			resp3, body3, err := client.PollResult(ctx, resultURL)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
				return
			}
			if resp3.StatusCode < 200 || resp3.StatusCode >= 300 {
				c.Data(resp3.StatusCode, "application/json", body3)
				return
			}
			if err := json.Unmarshal(body3, &finalJSON); err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": "réponse Gladia invalide", "detail": err.Error(), "raw": string(body3)})
				return
			}
			status, _ := finalJSON["status"].(string)
			switch status {
			case "queued", "processing":
				continue
			case "done":
				// ok on sort de la boucle
			case "error":
				c.JSON(http.StatusBadGateway, finalJSON)
				return
			default:
				// status inconnu, on continue un peu
				continue
			}
			break
		}
		// break externe lorsque done
		break
	}

	// 4) Normalisation
	transcripts := normalizeTranscripts(finalJSON)
	normalized := normalizeSentiment(finalJSON)
	if normalized != nil {
		finalJSON["sentiment_analysis_normalized"] = normalized
	}

	c.JSON(http.StatusOK, TranscriptResponse{
		Transcripts: transcripts,
		Data:        finalJSON,
	})
}

func normalizeTranscripts(data map[string]any) []TranscriptUtterance {
	result, ok := data["result"].(map[string]any)
	if !ok {
		return nil
	}
	// Gladia peut retourner plusieurs formats; on tente quelques chemins usuels
	var segments []any
	if s, ok := result["utterances"].([]any); ok {
		segments = s
	} else if s, ok := result["segments"].([]any); ok {
		segments = s
	}
	out := make([]TranscriptUtterance, 0, len(segments))
	for _, raw := range segments {
		seg, _ := raw.(map[string]any)
		if seg == nil {
			continue
		}
		speaker := str(seg["speaker"]) // parfois "spk_0"
		if speaker == "" {
			speaker = str(seg["speaker_id"]) // fallback
		}
		text := str(seg["text"])
		start := f64(seg["start"])
		end := f64(seg["end"])
		var words []string
		if w, ok := seg["words"].([]any); ok {
			words = make([]string, 0, len(w))
			for _, wr := range w {
				wm, _ := wr.(map[string]any)
				if wm == nil {
					continue
				}
				t := str(wm["text"])
				if t != "" {
					words = append(words, t)
				}
			}
		}
		out = append(out, TranscriptUtterance{
			Speaker: speaker,
			Text:    text,
			Start:   start,
			End:     end,
			Words:   words,
		})
	}
	return out
}

func normalizeSentiment(data map[string]any) map[string]any {
	result, ok := data["result"].(map[string]any)
	if !ok {
		return nil
	}
	raw := result["sentiment_analysis"]
	if raw == nil {
		return nil
	}
	normalized := map[string]any{
		"utterances": []map[string]any{},
	}
	list := []map[string]any{}

	switch v := raw.(type) {
	case []any:
		for _, it := range v {
			m, _ := it.(map[string]any)
			if m == nil {
				continue
			}
			list = append(list, map[string]any{
				"text":       str(m["text"]),
				"sentiment":  str(m["sentiment"]),
				"confidence": f64(m["confidence"]),
			})
		}
	case map[string]any:
		// parfois groupé par speaker
		for _, it := range v {
			arr, _ := it.([]any)
			for _, e := range arr {
				m, _ := e.(map[string]any)
				if m == nil {
					continue
				}
				list = append(list, map[string]any{
					"text":       str(m["text"]),
					"sentiment":  str(m["sentiment"]),
					"confidence": f64(m["confidence"]),
				})
			}
		}
	default:
		// inconnu, on laisse tomber
		return nil
	}
	normalized["utterances"] = list
	return normalized
}

func str(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func f64(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case json.Number:
		f, _ := t.Float64()
		return f
	default:
		return 0
	}
}
