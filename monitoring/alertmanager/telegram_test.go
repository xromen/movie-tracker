package alertmanager

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"text/template"
	"unicode/utf16"
)

type alert struct {
	Status              string
	Labels, Annotations map[string]string
}

type alerts []alert

func (a alerts) Firing() alerts   { return a.withStatus("firing") }
func (a alerts) Resolved() alerts { return a.withStatus("resolved") }
func (a alerts) withStatus(status string) (result alerts) {
	for _, item := range a {
		if item.Status == status {
			result = append(result, item)
		}
	}
	return
}

func TestTelegramMessage(t *testing.T) {
	source, err := os.ReadFile("telegram.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err := template.New("telegram").Parse(string(source))
	if err != nil {
		t.Fatal(err)
	}
	render := func(items alerts) string {
		t.Helper()
		var output bytes.Buffer
		if err := tmpl.ExecuteTemplate(&output, "movie_tracker.telegram", struct{ Alerts alerts }{items}); err != nil {
			t.Fatal(err)
		}
		return output.String()
	}
	message := render(alerts{
		{Status: "firing", Labels: map[string]string{"job": "postgres", "instance": "postgres-exporter:9187", "__grafana_receiver__": "secret"}, Annotations: map[string]string{"summary": "Нет метрик PostgreSQL", "grafana_state_reason": "NoData", "__values__": "secret"}},
		{Status: "resolved", Labels: map[string]string{"alertname": "API восстановлен"}},
	})
	for _, want := range []string{"проблем — 1, восстановлено — 1", "ПРОБЛЕМА: Нет метрик PostgreSQL", "postgres-exporter:9187", "Причина: NoData", "ВОССТАНОВЛЕНО: API восстановлен"} {
		if !strings.Contains(message, want) {
			t.Fatalf("missing %q in %q", want, message)
		}
	}
	if strings.Contains(message, "secret") {
		t.Fatal("internal labels or annotations leaked")
	}
	var many alerts
	long := strings.Repeat("🎬", 5000)
	for i := 0; i < 100; i++ {
		many = append(many, alert{Status: "firing", Labels: map[string]string{"job": long, "instance": long}, Annotations: map[string]string{"summary": long, "grafana_state_reason": long}})
	}
	message = render(many)
	if length := len(utf16.Encode([]rune(message))); length > 4096 {
		t.Fatalf("message exceeds Telegram limit: %d", length)
	}
	if strings.Count(message, "ПРОБЛЕМА:") != 5 || !strings.Contains(message, "Показаны первые 5 из 100") {
		t.Fatal("group limit or truncation notice is missing")
	}
}
