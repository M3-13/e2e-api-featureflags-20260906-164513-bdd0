VERDICT: CHANGES_REQUESTED

## 1. DSGVO / Datenschutz

### 1.1 Unbegrenzte Speicherung von IP-Adressen im Rate-Limiter
**Schweregrad:** hoch  
**Datei:** `internal/httpapi/ratelimit.go`

**Befund:**  
Der Rate-Limiter speichert pro Client-IP ein `windowEntry` in der Map `limiter.clients`. Abgelaufene Fenster werden nie entfernt. Damit bleiben IP-Adressen – personenbezogene Daten nach Art. 4 Nr. 1 DSGVO – dauerhaft im Speicher, bis der Prozess endet. Das verletzt die Grundsätze der Datenminimierung und Speicherbegrenzung aus Art. 5 Abs. 1 lit. c und e DSGVO und führt außerdem zu einem unbegrenzten Speicherwachstum.

**Konkrete Abhilfe:**  
In `limiter.allow()` beim Zugriff abgelaufene Einträge löschen, z. B.:
```go
if e, ok := l.clients[ip]; ok && now.Sub(e.start) >= l.window {
    delete(l.clients, ip)
}
```
Zusätzlich optional eine harte Obergrenze für `len(l.clients)` einführen und bei Erreichen veraltete Einträge bereinigen. Die eigentliche Ratenbegrenzung bleibt dabei erhalten; es wird lediglich der State nach Fensterablauf entfernt.

### 1.2 Öffentlicher Evaluate-Endpunkt verarbeitet Nutzerkennungen ohne Authentifizierung
**Schweregrad:** mittel  
**Datei:** `internal/httpapi/auth.go`, `main.go`

**Befund:**  
`GET /flags/{key}/evaluate?user={id}` ist über `isPublic` ausdrücklich ohne Bearer-Token erreichbar. Der Dienst verarbeitet dabei eine `user`-Kennung als Eingabe. Zwar wird die Kennung nicht gespeichert oder geloggt, aber je nach Einsatz kann der Endpunkt unbefugten Dritten ermöglichen, Rollout-Status einzelner Nutzer abzufragen. Das berührt die Vertraulichkeit der Verarbeitung und die Frage der Rechtsgrundlage.

**Konkrete Abhilfe:**  
Evaluate standardmäßig hinter dieselbe Authentifizierung legen oder ausdrücklich per Konfiguration steuerbar machen, z. B.:
```go
evaluatePublic := os.Getenv("EVALUATE_PUBLIC") == "true"
```
`isPublic` entsprechend anpassen. Sofern der Endpunkt bewusst öffentlich bleiben muss: Zweck, Empfängerkreis und Rechtsgrundlage in `README.md` / `COMPLIANCE.md` dokumentieren und die bestehende Ratenbegrenzung beibehalten. Wichtig ist, dass die produktiv benötigte Abfrage weiterhin funktioniert – die Absicherung darf nicht stillschweigend den legitimen Client-Fluss brechen, sondern muss konfigurierbar sein.

### 1.3 Positive Befunde
- Logs enthalten ausschließlich Methode, Pfad ohne Query-String, Statuscode und Dauer: `internal/httpapi/middleware.go`.
- `user`-Werte werden weder im Store noch in Antworten von `/evaluate` gespeichert oder zurückgegeben: `internal/store/store.go`, `internal/httpapi/evaluate.go`.
- Bearer-Token-Vergleich erfolgt konstantzeitgeschützt: `internal/httpapi/auth.go`.
- Keine personenbezogenen Daten in 500-Fehlerantworten; Panic-Details werden nicht an den Client gegeben: `internal/httpapi/middleware.go`.

## 2. EU Cyber Resilience Act (CRA)

### 2.1 Fehlende Transportverschlüsselung
**Schweregrad:** hoch  
**Datei:** `main.go`

**Befund:**  
Der Server startet mit `srv.ListenAndServe()` und bietet ausschließlich HTTP. Bearer-Token, Flag-Daten und Evaluate-Entscheidungen würden bei einem Deployment außerhalb von `127.0.0.1` im Klartext übertragen. Der Default `127.0.0.1:8080` ist zwar sicher für lokale Entwicklung, aber das Produkt ist über `ADDR` veränderbar. Sicherheit by design/default nach CRA verlangt einen sicheren Transport oder eine klar dokumentierte, verpflichtende TLS-Terminierung.

**Konkrete Abhilfe:**  
TLS-Unterstützung ergänzen, z. B.:
```go
if os.Getenv("TLS_CERT_FILE") != "" && os.Getenv("TLS_KEY_FILE") != "" {
    srv.ListenAndServeTLS(os.Getenv("TLS_CERT_FILE"), os.Getenv("TLS_KEY_FILE"))
} else {
    srv.ListenAndServe()
}
```
Zusätzlich in `README.md` / `SECURITY.md` dokumentieren: Plain-HTTP nur für Loopback-Tests, produktiv zwingend TLS oder ein explizit als sicher dokumentierter TLS-terminierender Reverse Proxy. Die Funktionalität bleibt für lokale Entwicklung ohne Zertifikate erhalten.

### 2.2 SBOM- und Update-Dokumentation
**Schweregrad:** mittel  
**Datei:** `go.mod`, `README.md`, `SECURITY.md`, `COMPLIANCE.md`

**Befund:**  
Der CRA verlangt für Produkte mit digitalen Elementen Transparenz über Abhängigkeiten und Update-/Patch-Fähigkeit. `go.mod` verwendet offenbar nur die Standardbibliothek, was positiv ist. Eine explizit dokumentierte SBOM/Abhängigkeitsliste und ein dokumentierter Update-Prozess sind im sichtbaren Code jedoch nicht erkennbar.

**Konkrete Abhilfe:**  
In `README.md` oder `COMPLIANCE.md` die Go-Version, die vollständige Abhängigkeitsliste und den vorgesehenen Patch-/Update-Prozess dokumentieren. Falls nicht bereits vorhanden, in `SECURITY.md` eine Verantwortlichkeits- und Meldezeile für Sicherheitsupdates ergänzen.

### 2.3 Positive CRA-Befunde
- Security-by-design-Maßnahmen sind vorhanden: Body-Limit (`LimitBody`), Panic-Recovery (`Recover`), Authentifizierung (`Auth`), generische Fehlerantworten, Timeouts und Header-Härtung (`Cache-Control`, `X-Content-Type-Options`).
- Der In-Memory-Store ist durch `sync.RWMutex` gegen parallele Zugriffe geschützt.

## 3. EU AI Act

**Befund:**  
Keine KI-Funktion, kein Modell, kein automatisiertes Entscheidungssystem im Sinne des AI Act. Die Rollout-Entscheidung ist ein deterministischer FNV-1a-Hash über `key` und `user`. Der EU AI Act ist daher nicht einschlägig.

## 4. Pflichttexte & UI

**Befund:**  
Das Produkt ist ein reines HTTP-Backend ohne öffentliche Web-Oberfläche. Es gibt keine Cookie-Banner-, Legal-Notice- oder Widerrufsbelehrungspflichten im Code. Ein Impressum/datenschutzrechtlicher Hinweis ist Sache des Betreibers und gehört nicht in den Backend-Code. Kein Befund.

## 5. Barrierefreiheit

**Befund:**  
Keine öffentliche Web-UI, daher keine WCAG/BITV/EAA-Bewertung erforderlich. Die REST-API bietet JSON-Antworten und keine grafische Oberfläche. Kein Befund.

---

**Gesamteinschätzung:**  
Das Produkt setzt die wesentlichen Datenschutzanforderungen für `user`-Daten bereits vorbildlich um. Offen sind insbesondere die unbegrenzte IP-Speicherung im Rate-Limiter und die fehlende Transportverschlüsselung. Beides ist durch gezielte Änderungen behebbar, ohne die Produktfunktionen zu brechen. Daher: `CHANGES_REQUESTED`.