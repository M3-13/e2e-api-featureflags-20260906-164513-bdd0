VERDICT: CHANGES_REQUESTED

## Sicherheitsbericht

### Prüfumfang
Der Quellcode des Go-Backends wurde manuell auf Secrets, Injection, Authentifizierung/Autorisierung, Abhängigkeiten sowie Konfigurations- und Transportaspekte geprüft. Für diesen Projekttyp war kein automatisierter Scanner konfiguriert; es liegen keine externen Abhängigkeiten über die Go-Standardbibliothek hinaus vor. Das Fehlen von Scannerausgaben wird nicht als Befund gewertet.

### Positive Beobachtungen
- Keine hartkodierten Secrets oder Zugangsdaten im Repository.
- `AUTH_TOKEN` wird aus der Umgebung gelesen; bei fehlendem Token verhält sich die Middleware fail-closed.
- Tokenvergleich erfolgt mittels `crypto/subtle.ConstantTimeCompare`.
- Request-Body wird über `http.MaxBytesReader` auf 1 MiB begrenzt.
- Panic-Recovery antwortet generisch ohne Stacktrace oder interne Details.
- Logging protokolliert ausschließlich Methode, Pfad (ohne Query-String), Statuscode und Dauer; `user`-Werte erscheinen nicht im Log.
- Der In-Memory-Store speichert keine `user`-Werte.
- JSON-Antworten setzen `Cache-Control: no-store` und `X-Content-Type-Options: nosniff`.

---

### Befund 1 — Unbegrenzter Speicherverbrauch im Rate-Limiter
- **Schweregrad:** mittel
- **Datei/Stelle:** `internal/httpapi/ratelimit.go`, Typ `limiter`, Methode `allow`
- **Beschreibung:** Die `clients`-Map im Rate-Limiter erhält für jede neue Client-IP einen Eintrag, wird aber nie bereinigt. Bei einem langlebigen, öffentlich erreichbaren Dienst wächst die Map mit wechselnden IPs unbegrenzt an. Das ist ein Speicherleck und kann langfristig zu Speicherknappheit und Denial of Service führen.
- **Konkrete Behebung:** Abgelaufene Einträge regelmäßig entfernen. Beispielsweise in `allow` nach dem `Lock` alle Einträge löschen, deren `start` älter als `window` ist, oder einen periodischen Cleanup über `time.Ticker` implementieren. Ein einfacher erster Schritt:
  ```go
  func (l *limiter) allow(ip string) bool {
      l.mu.Lock()
      defer l.mu.Unlock()

      now := time.Now()
      if len(l.clients) > 0 {
          for k, e := range l.clients {
              if now.Sub(e.start) >= l.window {
                  delete(l.clients, k)
              }
          }
      }

      e, ok := l.clients[ip]
      if !ok {
          l.clients[ip] = &windowEntry{start: now, count: 1}
          return true
      }
      if now.Sub(e.start) >= l.window {
          l.clients[ip] = &windowEntry{start: now, count: 1}
          return true
      }
      if e.count >= l.rps {
          return false
      }
      e.count++
      return true
  }
  ```

---

### Befund 2 — Authentifizierungs-Middleware öffnet `isPublic`-Zweig zu breit über Suffix-Matching
- **Schweregrad:** niedrig
- **Datei/Stelle:** `internal/httpapi/auth.go`, Funktion `isPublic`
- **Beschreibung:** Die Prüfung `strings.HasSuffix(r.URL.Path, "/evaluate")` macht jeden Pfad, der auf `/evaluate` endet, ohne Token erreichbar. Aktuell existiert zwar nur `GET /flags/{key}/evaluate`, aber sobald später andere Routen mit diesem Suffix hinzukommen, wären sie unbeabsichtigt öffentlich. Zudem wird die HTTP-Methode nicht geprüft.
- **Konkrete Behebung:** Exaktes Routing statt Suffix-Matching. Da `Auth` vor dem `ServeMux` läuft, sollte die Prüfung manuell erfolgen, z. B.:
  ```go
  func isPublic(r *http.Request) bool {
      if r.Method != http.MethodGet {
          return false
      }
      if r.URL.Path == "/healthz" {
          return true
      }
      parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
      return len(parts) == 3 && parts[0] == "flags" && parts[2] == "evaluate"
  }
  ```

---

### Befund 3 — Rate-Limit läuft vor der Authentifizierung; unauthentifizierte Anfragen können das Kontingent aufbrauchen
- **Schweregrad:** niedrig
- **Datei/Stelle:** `main.go`, Middleware-Kette in `newHandler`
- **Beschreibung:** Die Kette ist `Logging(Recover(LimitBody(RateLimit(Auth(mux)))))`. Damit zählt der Rate-Limiter auch unauthentifizierte Anfragen, bevor `Auth` sie ablehnt. Ein unauthentifizierter Client kann so das Limit einer IP ausschöpfen und anschließend legitime, authentifizierte Nutzer hinter derselben IP blockieren.
- **Konkrete Behebung:** Getrennte Ratenbegrenzung für öffentliche und geschützte Routen einführen. Beispielsweise den öffentlichen Evaluate- und Health-Pfad über einen eigenen Limiter schützen und für Verwaltungsrouten die Reihenfolge so wählen, dass Auth vor dem Rate-Limit greift. Alternativ den Limiter erst nach erfolgreicher Authentifizierung für alle nicht-öffentlichen Routen anwenden.

---

### Befund 4 — `clientIP` ignoriert `X-Forwarded-For` und kann hinter Reverse-Proxies zu einer gemeinsamen IP führen
- **Schweregrad:** niedrig
- **Datei/Stelle:** `internal/httpapi/ratelimit.go`, Funktion `clientIP`
- **Beschreibung:** Die bewusste Entscheidung, `X-Forwarded-For` zu ignorieren, ist vertretbar, wenn der Dienst direkt erreichbar ist. Hinter einem Reverse-Proxy oder Load-Balancer wird `RemoteAddr` jedoch zur Proxy-IP; dann teilen sich alle Clients denselben Rate-Limit-Bucket. Ein einzelner Client kann dadurch alle Nutzer hinter dem Proxy ausbremsen.
- **Konkrete Behebung:** Optional einen vertrauenswürdigen Proxy konfigurierbar machen, z. B. über eine Umgebungsvariable wie `TRUSTED_PROXY_CIDR`, und nur bei Verbindungen aus diesem Netz `X-Forwarded-For` auswerten. Andernfalls sollte die Dokumentation klar darauf hinweisen, dass der Dienst nicht ungeschützt hinter einem Proxy mit geteilter `RemoteAddr` betrieben werden sollte.

---

### Hinweis — Öffentlicher Evaluate-Endpunkt ermöglicht Flag-Key-Enumeration
- **Schweregrad:** informativ
- **Datei/Stelle:** `internal/httpapi/evaluate.go`
- **Beschreibung:** Der Evaluate-Endpunkt ist öffentlich und liefert für bekannte Keys `200` bzw. `false`, für unbekannte Keys `404`. Dadurch ist grundsätzlich eine Enumeration vorhandener Flag-Keys möglich. Dies ist durch die Acceptance Criteria (`AC-07`) explizit so vorgesehen und darf daher nicht ohne Spezifikationsänderung geändert werden.
- **Konkrete Behebung:** Keine Änderung im aktuellen Produkt. Falls die Vertraulichkeit der Flag-Keys später höher priorisiert wird, müsste die Anforderung angepasst werden, z. B. durch zusätzliche Authentifizierung des Evaluate-Pfads oder durch eine einheitliche Antwort bei unbekannten Keys.

---

### Abhängigkeiten
Laut `go.mod` werden ausschließlich Module der Go-Standardbibliothek verwendet. Es sind keine bekannten kritischen oder hohen Schwachstellen in Abhängigkeiten erkennbar. Ein automatisierter Schwachstellen-Scan war für diesen Projekttyp nicht hinterlegt; die manuelle Prüfung deckt die sichtbaren Abhängigkeiten ab.

### Gesamtbewertung
Es wurden keine hoch- oder kritisch riskanten Schwachstellen festgestellt. Die Implementierung erfüllt die wesentlichen Sicherheitsanforderungen: Secrets werden nicht eingecheckt, Auth ist fail-closed, Inputs werden validiert, der Body wird begrenzt, interne Fehler bleiben generisch und personenbezogene `user`-Werte werden nicht geloggt oder gespeichert.

Empfohlen wird jedoch die Behebung des unbegrenzten Speicherwachstums im Rate-Limiter sowie die präzise Begrenzung des öffentlichen Evaluate-Pfads. Da es sich um Härtungsbedarf mit mindestens einem mittleren Befund handelt, lautet das Ergebnis `CHANGES_REQUESTED`.